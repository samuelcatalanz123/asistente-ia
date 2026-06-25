package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// Sin imágenes: cada mensaje debe quedar como texto plano y hayImagen=false.
func TestAGroqMessagesSoloTexto(t *testing.T) {
	in := []Message{
		{Role: "user", Content: "hola"},
		{Role: "assistant", Content: "¿qué tal?"},
	}

	out, hayImagen := aGroqMessages(in)

	if hayImagen {
		t.Fatal("no había imágenes, pero hayImagen=true")
	}
	if len(out) != 2 {
		t.Fatalf("esperaba 2 mensajes, obtuve %d", len(out))
	}
	if s, ok := out[0].Content.(string); !ok || s != "hola" {
		t.Fatalf("el contenido de texto no se conservó: %v", out[0].Content)
	}
}

// Con una foto: hayImagen=true y el contenido es texto + image_url con la URL correcta.
func TestAGroqMessagesConImagen(t *testing.T) {
	const dataURL = "data:image/jpeg;base64,AAAA"
	in := []Message{
		{Role: "user", Content: "¿qué es esto?", Imagen: dataURL},
	}

	out, hayImagen := aGroqMessages(in)

	if !hayImagen {
		t.Fatal("había una imagen, pero hayImagen=false")
	}
	partes, ok := out[0].Content.([]contentPart)
	if !ok {
		t.Fatalf("el contenido con imagen debería ser []contentPart, fue %T", out[0].Content)
	}
	if len(partes) != 2 {
		t.Fatalf("esperaba 2 partes (texto + imagen), obtuve %d", len(partes))
	}
	if partes[0].Type != "text" || partes[0].Text != "¿qué es esto?" {
		t.Fatalf("la parte de texto está mal: %+v", partes[0])
	}
	if partes[1].Type != "image_url" || partes[1].ImageURL == nil || partes[1].ImageURL.URL != dataURL {
		t.Fatalf("la parte de imagen está mal: %+v", partes[1])
	}
}

// Si se sube una foto sin escribir nada, debe ponerse una pregunta por defecto.
func TestAGroqMessagesImagenSinTexto(t *testing.T) {
	in := []Message{
		{Role: "user", Content: "", Imagen: "data:image/png;base64,BBBB"},
	}

	out, _ := aGroqMessages(in)

	partes := out[0].Content.([]contentPart)
	if partes[0].Text == "" {
		t.Fatal("con foto y sin texto debería haber una pregunta por defecto")
	}
}

// El resultado debe serializarse al JSON que espera Groq (text + image_url anidado).
func TestAGroqMessagesJSONValido(t *testing.T) {
	out, _ := aGroqMessages([]Message{
		{Role: "user", Content: "mira", Imagen: "data:image/jpeg;base64,CCCC"},
	})

	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("no se pudo serializar: %v", err)
	}
	j := string(b)
	for _, esperado := range []string{`"type":"text"`, `"type":"image_url"`, `"image_url":{"url":`} {
		if !strings.Contains(j, esperado) {
			t.Fatalf("el JSON no contiene %q. JSON: %s", esperado, j)
		}
	}
}
