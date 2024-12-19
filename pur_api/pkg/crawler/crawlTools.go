package crawler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

type types struct {
	t 		string
	count 	int
}

func sortTypes(t map[string]int) []types {
	typesSlice := []types{}
	for t, count := range t {
		typesSlice = append(typesSlice, types{t: t, count: count})
	}
	sort.Slice(typesSlice, func(i, j int) bool {
		return typesSlice[i].count > typesSlice[j].count
	})
	return typesSlice
}

func getMimeType(url string, c *http.Client) (string, error) {
	req, err := c.Head(url)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	log.Printf("req body: %v", req.StatusCode)

	if req.StatusCode > 399 && req.Header == nil {
		return "", fmt.Errorf("invalid status code with empty headers")
	}

	contentTypes := req.Header["Content-Type"]
	if len(contentTypes) == 0 {
		return "", fmt.Errorf("empty headers on: %s", url)
	}

	return analyzeContentType(contentTypes), nil
}

func analyzeContentType(mimeTypes []string) string {
	for _, mimeType := range mimeTypes {
		mimeType = strings.ToLower(mimeType)

		switch {
		case strings.HasPrefix(mimeType, "text/html"):
			return "html"
		case strings.HasPrefix(mimeType, "video/"):
			return "video"
		case strings.HasPrefix(mimeType, "image/"):
			return "image"
		case strings.HasPrefix(mimeType, "application/pdf"):
			return "pdf"
		case strings.HasPrefix(mimeType, "application/json"):
			return "json"
		case strings.HasPrefix(mimeType, "application/xml"):
			return "xml"
		default:
			continue
		}
	}

	return "unknown"
}

func getHTML(url string, c *http.Client) (*html.Node, error) {
	req, err := c.Get(url)
	if err != nil {
		log.Printf("error getting response from: %s\n", url)
		return nil, err
	}
	defer req.Body.Close()

	if req.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code")
	}

	doc, err := html.Parse(req.Body)
	if err != nil {
		log.Printf("error parsing page %s: %v\n", url, err)
		return nil, err
	}

	return doc, err
}

func resolveURL(base, ref string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid base URL %s: %v", base, err)
	}

	refURL, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("invalid reference URL %s: %v", ref, err)
	}

	return baseURL.ResolveReference(refURL).String(), nil
}