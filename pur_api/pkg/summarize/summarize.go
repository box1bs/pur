package summarize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SummarizeSender struct {
	ArticleUrl string
	ServerAddr string
}

func NewSummarizeSender(articleUrl, serverAddr string) *SummarizeSender {
	return &SummarizeSender{
		ArticleUrl: articleUrl,
		ServerAddr: serverAddr,
	}
}

func (s *SummarizeSender) Summarize() (string, error) {
	req, err := json.Marshal(map[string]string{
		"url": s.ArticleUrl,
	})
	if err != nil {
		return "", fmt.Errorf("error creating JSON: %v", err)
	}

	resp, err := http.Post(s.ServerAddr, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	if errMsg, exists := response["error"].(string); exists {
		return "", fmt.Errorf("server responded with an error: %s", errMsg)
	}

	summary, exists := response["summary"].(string)
	if !exists {
		return "", fmt.Errorf("summary not found in response")
	}

	return summary, nil
}
