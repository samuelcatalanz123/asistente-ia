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

// systemPrompt define la personalidad del asistente. Se envía como primer
// mensaje (rol "system") en cada conversación para guiar cómo responde.
const systemPrompt = "Eres el asistente personal de Samuel, un asistente con " +
	"inteligencia artificial. Eres simpático, cercano y amable, como un buen " +
	"amigo que ayuda. IMPORTANTE: responde SIEMPRE en el mismo idioma en el que " +
	"te escribe la persona en su último mensaje. Si te escriben en inglés, " +
	"respondes en inglés; si te escriben en español, respondes en español. " +
	"Hablas de forma natural y relajada. Usas algún emoji de vez en cuando para " +
	"dar calidez, pero sin pasarte. Das respuestas claras y fáciles de entender, " +
	"sin tecnicismos innecesarios. Si no sabes algo, lo dices con sinceridad en " +
	"vez de inventar. Tu objetivo es que la persona se sienta bien atendida y ayudada."

// GroqClient calls Groq's OpenAI-compatible chat completions API.
type GroqClient struct {
	APIKey string
	URL    string
	HTTP   *http.Client
}

func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{
		APIKey: apiKey,
		URL:    groqURL,
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
	// Anteponemos la personalidad (mensaje "system") a la conversación.
	withPersonality := append([]Message{{Role: "system", Content: systemPrompt}}, messages...)
	payload, err := json.Marshal(groqRequest{Model: groqModel, Messages: withPersonality})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(payload))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading groq response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq error %d", resp.StatusCode)
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
