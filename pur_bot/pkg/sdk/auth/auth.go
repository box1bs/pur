package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type AccountData struct {
	Id   	uuid.UUID
	Name 	string
	Client 	*http.Client
}

func (a *AccountData) Authorizate() error {
	req, err := json.Marshal(map[string]string{
		"name": a.Name,
	})
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("http://localhost:8080/account/%s", a.Id.String())

	resp, err := a.Client.Post(addr, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}

func (a *AccountData) DeleteAccount() error {
	uri, err := url.Parse(fmt.Sprintf("http://localhost:8080/account/%s", a.Id.String()))
	if err != nil {
		return err
	}

	resp, err := a.Client.Do(&http.Request{
		Method: http.MethodDelete,
		URL: uri,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("invalid status code")
	}

	return nil
}