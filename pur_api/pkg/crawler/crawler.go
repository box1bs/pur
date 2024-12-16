package crawler

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type Crawler struct {
	Client				*http.Client
	Types           	map[string]int
	Visited				map[string]string
	MaxVisited 			int
	BaseUrl         	string
	Mu              	*sync.Mutex
	Wg              	*sync.WaitGroup
	ConcurrencyControl 	chan struct{}
}

func (c *Crawler) Crawl() (string, error) {
	t, err := getMimeType(c.BaseUrl, c.Client)
	if err != nil {
		return "", err
	}
	if t != "html" {
		return t, nil
	}

	links, err := c.extractLinks(c.BaseUrl)
	if err != nil {
		return "", err
	}

	for i, rawUrl := range links {
		if i >= c.MaxVisited {
			break
		}

		url, err := resolveURL(c.BaseUrl, rawUrl)
		if err != nil {
			log.Printf("Failed to resolve URL: %v\n", err)
			continue
		}

		if c.isVisited(url) {
			continue
		}

		c.Wg.Add(1)
		go func(url string) {
			defer c.Wg.Done()
			c.ConcurrencyControl <- struct{}{}
			defer func () {<- c.ConcurrencyControl}()

			mimeType, err := getMimeType(url, c.Client)
			if err != nil {
				log.Printf("error getting MIME-type for %s: %v\n", url, err)
				return
			}

			if mimeType == "html" {
				ogType := c.extractOGType(url)
				if ogType != "" {
					c.Mu.Lock()
					c.Types[ogType]++
					c.Visited[url] = ogType
					c.Mu.Unlock()
				}
			}
		}(url)
	}

	c.Wg.Wait()

	resultType := sortTypes(c.Types)
	if len(resultType) == 0 {
		return "unknown", nil
	}

	return resultType[0].t, nil
}

func (c *Crawler) isVisited(url string) bool {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if t, visited := c.Visited[url]; visited {
		c.Types[t]++
		return true
	}
	return false
} 

func (c *Crawler)extractOGType(url string) string {
	doc, err := getHTML(url, c.Client)
	if err != nil {
		return ""
	}

	var ogType string
	var findOGType func(*html.Node)
	findOGType = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, attr := range n.Attr {
				if attr.Key == "property" && attr.Val == "og:type" {
					for _, contentAttr := range n.Attr {
						if contentAttr.Key == "content" {
							ogType = contentAttr.Val
							return
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findOGType(c)
		}
	}

	findOGType(doc)

	return ogType
}

func (c *Crawler)extractLinks(url string) ([]string, error) {
	doc, err := getHTML(url, c.Client)
	if err != nil {
		return nil, err
	}

	var links []string
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if strings.HasPrefix(attr.Val, "#") || strings.HasPrefix(attr.Val, "mailto:") || strings.HasPrefix(attr.Val, "javascript:") {
						continue
					}
					links = append(links, attr.Val)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)

	return links, nil
}