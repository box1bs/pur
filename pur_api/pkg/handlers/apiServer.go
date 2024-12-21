package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/box1bs/pur/pur_api/pkg/config"
	"github.com/box1bs/pur/pur_api/pkg/crawler"
	"github.com/box1bs/pur/pur_api/pkg/model"
	_ "github.com/box1bs/pur/pur_api/pkg/summarize"
	"github.com/go-co-op/gocron"
)

type APIServer struct {
	ListenAddr 			string
	Store	   			config.Storage
	Accuracy   			int
	CorcurrencyCount 	int
	SummaryServAddr		string
}

func NewAPIServer(listenAddr string, store config.Storage, accuracy, corcurrencyCount int, summaryAddr string) *APIServer {
	return &APIServer{
		ListenAddr: listenAddr,
		Store: store,
		Accuracy: accuracy,
		CorcurrencyCount: corcurrencyCount,
		SummaryServAddr: summaryAddr,
	}
}

func (s *APIServer) Run() {
	router := http.NewServeMux()

	if err := s.Store.InitMigrate(); err != nil {
		log.Fatalf("Failed to initialize store migration: %v", err)
	}

	router.HandleFunc("/account/{id}", makeHTTPHandleFunc(s.HandleAccount))
	router.HandleFunc("/link/{id}", makeHTTPHandleFunc(s.HandleLinkWithId))
	router.HandleFunc("/link", makeHTTPHandleFunc(s.HandleLink))

	log.Println("PUR API server running on port: ", s.ListenAddr)

	sc := gocron.NewScheduler(time.UTC)

	sc.Every(1).Day().Do(func() {
		err := s.Store.DeleteObsoleteRecords()
		if err != nil {
			log.Println("error del obsolete records:", err)
		} else {
			log.Println("obsolete records successfuly deleted.")
		}
	})

	go sc.StartBlocking()

	http.ListenAndServe(":" + s.ListenAddr, router)
}

func (s *APIServer) HandleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.HandleGetAccount(w, r)
	case "POST":
		return s.HandleCreateAccount(w, r)
	case "DELETE":
		return s.HandleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed: %s", r.Method)
}

func (s *APIServer) HandleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accountId, err := varsHandleUUID(r, "id")
	if err != nil {
		return err
	}

	account, err := s.Store.GetAccountByID(accountId)
	if err != nil {
		return err
	}
	
	if err := WriteJSON(w, 200, account); err != nil {
		return err
	}

	return nil
}

func (s *APIServer) HandleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	accountId, err := varsHandleUUID(r, "id")
	if err != nil {
		return err
	}

	var account model.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		log.Printf("error decoding account request: %v with id: %s\n", err, accountId)
		return err
	}

	account.Id = accountId

	if err := s.Store.CreateAccount(account); err != nil {
		return nil
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

func (s *APIServer) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	accountId, err := varsHandleUUID(r, "id")
	if err != nil {
		return err
	}

	if err := s.Store.DeleteAccount(accountId); err != nil {
		return err
	}

	if err := s.Store.DeleteAllLinksById(accountId); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *APIServer) HandleLink(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "POST":
		return s.HandleSaveLink(w, r)
	case "PUT":
		return s.HandleUpdateLink(w, r)
	}

	return fmt.Errorf("method not allowed: %s", r.Method)
}

func (s *APIServer) HandleLinkWithId(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.HandleGetLinks(w, r)
	case "DELETE":
		return s.HandleDeleteLinkByUrl(w, r)
	}

	return fmt.Errorf("method not allowed: %s", r.Method)
}


func (s *APIServer) HandleSaveLink(w http.ResponseWriter, r *http.Request) error {
	var link model.Link
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		log.Printf("error decoding link: %v", err)
		return err
	}

	crawler := &crawler.Crawler{
		Client: &http.Client{
			Timeout: time.Millisecond * 200,
		},
		Types: make(map[string]int),
		Visited: make(map[string]string),
		MaxVisited: s.Accuracy,
		BaseUrl: link.Url,
		Mu: &sync.Mutex{},
		Wg: &sync.WaitGroup{},
		ConcurrencyControl: make(chan struct{}, s.CorcurrencyCount),
	}

	Type, err := crawler.Crawl()
	if err != nil  {
		log.Printf("error crawling link: %v", err)
	}

	//if CanSummarize(Type) {
	//	summary, err := summarize.NewSummarizeSender(link.Url, s.SummaryServAddr).Summarize()
	//	if err != nil {
	//		log.Printf("error summarize: %v, with type: %s", err, link.Type)
	//	}
	//
	//	link.Summary = summary
	//}

	link.Type = Type

	if err := s.Store.SaveLink(link); err != nil {
		log.Println(err)
		return err
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

func (s *APIServer) HandleUpdateLink(w http.ResponseWriter, r *http.Request) error {
	var link model.Link
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		log.Printf("error decoding link: %v", err)
		return err
	}

	if err := s.Store.UpdateLink(link); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (s *APIServer) HandleGetLinks(w http.ResponseWriter, r *http.Request) error {
	accountId, err := varsHandleUUID(r, "id")
	if err != nil {
		log.Printf("error getting uuid from path: %v", err)
		return err
	}

	links, err := s.Store.GetLinksByAccountID(accountId)
	if err != nil {
		log.Printf("error getting links by id: %v", err)
		return err
	}

	if err := WriteJSON(w, 200, links); err != nil {
		log.Printf("error sending json links: %v", err)
		return err
	}

	return nil
}

//Use link id for delete one link
func (s *APIServer) HandleDeleteLink(w http.ResponseWriter, r *http.Request) error {
	linkId, err := varsHandleUUID(r, "id")
	if err != nil {
		return err
	}

	if err := s.Store.DeleteLinkByID(linkId); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *APIServer) HandleDeleteLinkByUrl(w http.ResponseWriter, r *http.Request) error {
	uid, err := varsHandleUUID(r, "id")
	if err != nil {
		return err
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return err
	}

	if url, ok := resp["url"].(string); ok {
		if err := s.Store.DeleteRecordByUrl(uid, url); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid request body")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}