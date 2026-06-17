package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	"vez de inventar. Tu objetivo es que la persona se sienta bien atendida y " +
	"ayudada. Tienes conocimientos especiales de PROGRAMACIÓN (lenguajes como Go, " +
	"Python, Dart/Flutter, JavaScript, bases de datos, despliegue en la nube, etc.): " +
	"cuando te pregunten de programar, explica con claridad, da ejemplos sencillos y, " +
	"si ayuda, muestra trozos de código. Aun así, sigues siendo un asistente amable " +
	"que también ayuda con cualquier otro tema."

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
	Stream   bool      `json:"stream,omitempty"`
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

// StreamComplete pide la respuesta en modo streaming y llama a onChunk con
// cada trocito de texto según va llegando desde Groq.
func (c *GroqClient) StreamComplete(messages []Message, onChunk func(string)) error {
	withPersonality := append([]Message{{Role: "system", Content: systemPrompt}}, messages...)
	payload, err := json.Marshal(groqRequest{Model: groqModel, Messages: withPersonality, Stream: true})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(payload))
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
		return fmt.Errorf("groq error %d", resp.StatusCode)
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
