package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const groqURL = "https://api.groq.com/openai/v1/chat/completions"
const groqModel = "llama-3.3-70b-versatile"
const groqVisionModel = "meta-llama/llama-4-scout-17b-16e-instruct" // entiende imágenes

// modeloDeGroq traduce la elección del usuario ("rapido"/"inteligente") al
// nombre real del modelo en Groq.
func modeloDeGroq(modelo string) string {
	if modelo == "rapido" {
		return "llama-3.1-8b-instant" // más pequeño y veloz
	}
	return groqModel // "inteligente" (por defecto): el grande
}

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
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

// groqMessage es un mensaje en el formato que entiende Groq. El Content puede
// ser texto simple (string) o una lista de partes (texto + imagen) para visión.
type groqMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type contentPart struct {
	Type     string    `json:"type"` // "text" o "image_url"
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

// aGroqMessages convierte nuestros mensajes al formato de Groq. Si algún mensaje
// trae una imagen, arma el formato multimodal y devuelve hayImagen=true.
func aGroqMessages(messages []Message) (out []groqMessage, hayImagen bool) {
	out = make([]groqMessage, len(messages))
	for i, m := range messages {
		if m.Imagen != "" {
			hayImagen = true
			texto := m.Content
			if texto == "" {
				texto = "¿Qué hay en esta imagen? Descríbela en español."
			}
			out[i] = groqMessage{Role: m.Role, Content: []contentPart{
				{Type: "text", Text: texto},
				{Type: "image_url", ImageURL: &imageURL{URL: m.Imagen}},
			}}
		} else {
			out[i] = groqMessage{Role: m.Role, Content: m.Content}
		}
	}
	return out, hayImagen
}

type groqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// groqStreamChunk es cada trocito que llega cuando pedimos streaming.
type groqStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (c *GroqClient) Complete(ctx context.Context, messages []Message, modelo string) (string, error) {
	ms, hayImagen := aGroqMessages(messages)
	model := modeloDeGroq(modelo)
	if hayImagen {
		model = groqVisionModel // si hay foto, usa el modelo que "ve"
	}
	payload, err := json.Marshal(groqRequest{Model: model, Messages: ms})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(payload))
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
		return "", fmt.Errorf("groq %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
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

// StreamComplete pide la respuesta en modo streaming y llama a onChunk con
// cada trocito de texto según va llegando desde Groq.
func (c *GroqClient) StreamComplete(ctx context.Context, messages []Message, modelo string, onChunk func(string)) error {
	ms, hayImagen := aGroqMessages(messages)
	model := modeloDeGroq(modelo)
	if hayImagen {
		model = groqVisionModel // si hay foto, usa el modelo que "ve"
	}
	payload, err := json.Marshal(groqRequest{Model: model, Messages: ms, Stream: true})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("groq %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	// Groq envía líneas tipo "data: {json}" y termina con "data: [DONE]".
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk groqStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			onChunk(chunk.Choices[0].Delta.Content)
		}
	}
	return scanner.Err()
}
