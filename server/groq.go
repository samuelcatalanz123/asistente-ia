package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const groqURL = "https://api.groq.com/openai/v1/chat/completions"
const groqModel = "llama-3.3-70b-versatile"

// GroqClient calls Groq's OpenAI-compatible chat completions API.
type GroqClient struct {
	APIKey string
	HTTP   *http.Client
}

func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{
		APIKey: apiKey,
		HTTP:   &http.Client{Timeout: 30 * time.Second},
	}
}

type groqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type groqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func (c *GroqClient) Complete(messages []Message) (string, error) {
	payload, err := json.Marshal(groqRequest{Model: groqModel, Messages: messages})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, groqURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq error %d: %s", resp.StatusCode, string(body))
	}
	var parsed groqResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}
