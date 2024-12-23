package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ReqResource struct {
	Addr 	string
	Client 	*http.Client
}

type link struct {
	Id 			string	`json:"user_id"`
	Url 		string	`json:"url"`
	Summary 	string	`json:"summary"`
	Description string	`json:"description"`
}

func (rr *ReqResource) SaveLink(id uuid.UUID, url, description string) error {
	l := link{
		Id: id.String(),
		Url: url,
		Description: description,
	}

	req, err := json.Marshal(l)
	if err != nil {
		return err
	}

	resp, err := rr.Client.Post(rr.Addr, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}

func (rr *ReqResource) DeleteLink(Url string) error {
	reqBody, err := json.Marshal(map[string]string{
		"url": Url,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, rr.Addr, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := rr.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}

func (rr *ReqResource) GetAllLinks() ([]link, error) {
	resp, err := rr.Client.Get(rr.Addr)
	if err != nil {
		log.Printf("error getting links: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var sendedData interface{}
	if err := json.NewDecoder(resp.Body).Decode(&sendedData); err != nil {
		log.Printf("error decode links: %v", err)
		return nil, err
	}

	var links []link

	switch data := sendedData.(type) {
	case []interface{}:
		for _, item := range data {
            if obj, ok := item.(map[string]interface{}); ok {
                links = append(links, mapToLink(obj))
            }
        }

	case map[string]interface{}:
		links = append(links, mapToLink(data))

	default:
		return nil, fmt.Errorf("unexpected JSON format")
	}

	return links, nil
}

func mapToLink(obj map[string]interface{}) link {
	var l link
	if id, ok := obj["id"].(string); ok {
		l.Id = id
	}
	if url, ok := obj["url"].(string); ok {
		l.Url = url
	}
    if summary, ok := obj["summary"].(string); ok {
		l.Summary = summary
    }
	if description, ok := obj["description"].(string); ok {
		l.Description = description
	}
	return l
}

func (l *link) PresentLink() string{
	if l.Summary != "" {
		return fmt.Sprintf("Your resource: %s,\ndescription: %s,\nsummary: %s\n", l.Url, l.Description, l.Summary)
	}
	return fmt.Sprintf("Your resource: %s,\ndescription: %s\n", l.Url, l.Description)
}