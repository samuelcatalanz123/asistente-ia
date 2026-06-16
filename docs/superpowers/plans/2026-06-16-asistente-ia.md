# Asistente IA Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a mobile AI chat app (Flutter) backed by a Go proxy server that hides the AI key and calls the free Groq API, deployed free on Render.

**Architecture:** The Flutter app sends the conversation to a Go HTTP server (`POST /chat`). The Go server holds the secret `GROQ_API_KEY`, forwards the conversation to Groq's OpenAI-compatible API, and returns the reply. The AI call sits behind an `AIClient` interface so the HTTP handler can be tested with a fake (no real network in tests). The server is deployed to Render's free tier; the app points at that public URL.

**Tech Stack:** Go (standard library `net/http`, `encoding/json`), Flutter/Dart (`http` package), Groq API (model `llama-3.3-70b-versatile`), Render (hosting).

---

## File Structure

```
asistente-ia/
├── server/
│   ├── go.mod
│   ├── types.go          # ChatRequest, ChatResponse, Message
│   ├── aiclient.go       # AIClient interface
│   ├── groq.go           # GroqClient implements AIClient (real network)
│   ├── handler.go        # NewChatHandler, healthHandler
│   ├── handler_test.go   # handler tests using a fake AIClient
│   └── main.go           # wires GroqClient + handlers, starts server
├── app/                  # Flutter project (flutter create)
│   ├── lib/
│   │   ├── models/message.dart
│   │   ├── services/chat_service.dart
│   │   ├── screens/chat_screen.dart
│   │   └── main.dart
│   └── test/
│       ├── chat_service_test.dart
│       └── chat_screen_test.dart
├── render.yaml           # Render deploy config
└── README.md
```

---

## PHASE 1 — Go Backend

### Task 1: Initialize Go module and shared types

**Files:**
- Create: `server/go.mod`
- Create: `server/types.go`

- [ ] **Step 1: Create the Go module**

Run:
```bash
cd ~/Repos/asistente-ia/server && go mod init asistente-ia-server
```
Expected: creates `go.mod` with `module asistente-ia-server` and a Go version line.

- [ ] **Step 2: Create the shared types**

Create `server/types.go`:
```go
package main

// Message is one turn in the conversation.
type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// ChatRequest is what the Flutter app sends to POST /chat.
type ChatRequest struct {
	Messages []Message `json:"messages"`
}

// ChatResponse is what POST /chat returns.
type ChatResponse struct {
	Reply string `json:"reply"`
}

// ErrorResponse is returned on failures.
type ErrorResponse struct {
	Error string `json:"error"`
}
```

- [ ] **Step 3: Verify it compiles**

Run:
```bash
cd ~/Repos/asistente-ia/server && go build ./...
```
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

```bash
cd ~/Repos/asistente-ia && git add server/go.mod server/types.go && git commit -m "feat(server): init go module and shared types"
```

---

### Task 2: AIClient interface

**Files:**
- Create: `server/aiclient.go`

- [ ] **Step 1: Define the interface**

Create `server/aiclient.go`:
```go
package main

// AIClient talks to an AI provider. The HTTP handler depends on this
// interface so tests can substitute a fake without real network calls.
type AIClient interface {
	Complete(messages []Message) (string, error)
}
```

- [ ] **Step 2: Verify it compiles**

Run:
```bash
cd ~/Repos/asistente-ia/server && go build ./...
```
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

```bash
cd ~/Repos/asistente-ia && git add server/aiclient.go && git commit -m "feat(server): add AIClient interface"
```

---

### Task 3: Health endpoint (TDD)

**Files:**
- Create: `server/handler.go`
- Create: `server/handler_test.go`

- [ ] **Step 1: Write the failing test**

Create `server/handler_test.go`:
```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd ~/Repos/asistente-ia/server && go test ./...
```
Expected: FAIL — `undefined: healthHandler`.

- [ ] **Step 3: Write minimal implementation**

Create `server/handler.go`:
```go
package main

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
cd ~/Repos/asistente-ia/server && go test ./...
```
Expected: PASS (`ok asistente-ia-server`).

- [ ] **Step 5: Commit**

```bash
cd ~/Repos/asistente-ia && git add server/handler.go server/handler_test.go && git commit -m "feat(server): add /health endpoint"
```

---

### Task 4: Chat handler with a fake AIClient (TDD)

**Files:**
- Modify: `server/handler.go`
- Modify: `server/handler_test.go`

- [ ] **Step 1: Write the failing test**

Add to `server/handler_test.go` (keep existing imports; add `"bytes"` and `"strings"`):
```go
// fakeAI is a test double for AIClient.
type fakeAI struct {
	reply string
	err   error
	got   []Message
}

func (f *fakeAI) Complete(messages []Message) (string, error) {
	f.got = messages
	return f.reply, f.err
}

func TestChatHandlerReturnsReply(t *testing.T) {
	fake := &fakeAI{reply: "¡Hola! ¿En qué te ayudo?"}
	handler := NewChatHandler(fake)

	body := `{"messages":[{"role":"user","content":"hola"}]}`
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var resp ChatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.Reply != "¡Hola! ¿En qué te ayudo?" {
		t.Fatalf("unexpected reply: %q", resp.Reply)
	}
	if len(fake.got) != 1 || fake.got[0].Content != "hola" {
		t.Fatalf("handler did not pass messages through: %+v", fake.got)
	}
}

func TestChatHandlerRejectsEmptyMessages(t *testing.T) {
	handler := NewChatHandler(&fakeAI{reply: "x"})
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"messages":[]}`))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty messages, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd ~/Repos/asistente-ia/server && go test ./...
```
Expected: FAIL — `undefined: NewChatHandler`.

- [ ] **Step 3: Write minimal implementation**

Add to `server/handler.go` (add `"errors"` to imports if needed — it is not, only the function below):
```go
// NewChatHandler returns an http handler that forwards the conversation
// to the given AIClient and returns the reply as JSON.
func NewChatHandler(ai AIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "usa POST"})
			return
		}
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "json inválido"})
			return
		}
		if len(req.Messages) == 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "faltan mensajes"})
			return
		}
		reply, err := ai.Complete(req.Messages)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: "la IA no respondió, inténtalo de nuevo"})
			return
		}
		writeJSON(w, http.StatusOK, ChatResponse{Reply: reply})
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
cd ~/Repos/asistente-ia/server && go test ./...
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/Repos/asistente-ia && git add server/handler.go server/handler_test.go && git commit -m "feat(server): add /chat handler with AIClient"
```

---

### Task 5: Groq client (real implementation)

**Files:**
- Create: `server/groq.go`

- [ ] **Step 1: Implement GroqClient**

Create `server/groq.go`:
```go
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
```

- [ ] **Step 2: Verify it compiles and the interface is satisfied**

Run:
```bash
cd ~/Repos/asistente-ia/server && go build ./... && go vet ./...
```
Expected: no output, exit code 0. (If `GroqClient` did not satisfy `AIClient`, Task 6's `main.go` would fail to compile; we confirm there.)

- [ ] **Step 3: Commit**

```bash
cd ~/Repos/asistente-ia && git add server/groq.go && git commit -m "feat(server): add Groq client"
```

---

### Task 6: Wire up main with env var and CORS

**Files:**
- Create: `server/main.go`

- [ ] **Step 1: Write main.go**

Create `server/main.go`:
```go
package main

import (
	"log"
	"net/http"
	"os"
)

// withCORS allows the Flutter app (any origin) to call the API.
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("falta la variable GROQ_API_KEY")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var ai AIClient = NewGroqClient(apiKey)

	http.HandleFunc("/health", withCORS(healthHandler))
	http.HandleFunc("/chat", withCORS(NewChatHandler(ai)))

	log.Printf("servidor escuchando en :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

- [ ] **Step 2: Verify everything compiles and all tests pass**

Run:
```bash
cd ~/Repos/asistente-ia/server && go build ./... && go test ./...
```
Expected: build succeeds; tests PASS.

- [ ] **Step 3: Manual smoke test with a real key**

Run (replace with the real key from Task 10):
```bash
cd ~/Repos/asistente-ia/server && GROQ_API_KEY=sk-xxxxx go run . &
sleep 2
curl -s localhost:8080/health
curl -s -X POST localhost:8080/chat -H "Content-Type: application/json" -d '{"messages":[{"role":"user","content":"di hola en una palabra"}]}'
```
Expected: `{"status":"ok"}` then a JSON `{"reply":"..."}` with a short greeting. Stop the server with `kill %1` afterward.

- [ ] **Step 4: Commit**

```bash
cd ~/Repos/asistente-ia && git add server/main.go && git commit -m "feat(server): wire up main with CORS and env config"
```

---

## PHASE 2 — Flutter App

### Task 7: Create the Flutter project

**Files:**
- Create: `app/` (generated)

- [ ] **Step 1: Generate the project**

Run:
```bash
cd ~/Repos/asistente-ia && flutter create --org com.samuel --project-name asistente_ia app
```
Expected: Flutter scaffolds the `app/` directory.

- [ ] **Step 2: Add the http dependency**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter pub add http
```
Expected: `http` added to `pubspec.yaml`.

- [ ] **Step 3: Verify it builds**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter analyze
```
Expected: "No issues found!" (warnings about the default counter app are fine).

- [ ] **Step 4: Commit**

```bash
cd ~/Repos/asistente-ia && git add app && git commit -m "feat(app): scaffold flutter project with http"
```

---

### Task 8: Message model

**Files:**
- Create: `app/lib/models/message.dart`

- [ ] **Step 1: Write the model**

Create `app/lib/models/message.dart`:
```dart
class Message {
  final String role; // "user" or "assistant"
  final String content;

  const Message({required this.role, required this.content});

  bool get isUser => role == 'user';

  Map<String, String> toJson() => {'role': role, 'content': content};
}
```

- [ ] **Step 2: Verify it analyzes clean**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter analyze lib/models/message.dart
```
Expected: no issues.

- [ ] **Step 3: Commit**

```bash
cd ~/Repos/asistente-ia && git add app/lib/models/message.dart && git commit -m "feat(app): add Message model"
```

---

### Task 9: Chat service (TDD)

**Files:**
- Create: `app/lib/services/chat_service.dart`
- Create: `app/test/chat_service_test.dart`

- [ ] **Step 1: Write the failing test**

Create `app/test/chat_service_test.dart`:
```dart
import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:asistente_ia/models/message.dart';
import 'package:asistente_ia/services/chat_service.dart';

void main() {
  test('sendMessages returns the reply from the backend', () async {
    final mockClient = MockClient((request) async {
      expect(request.url.path, '/chat');
      final body = jsonDecode(request.body);
      expect(body['messages'][0]['content'], 'hola');
      return http.Response(jsonEncode({'reply': 'buenas'}), 200);
    });

    final service = ChatService(
      baseUrl: 'http://test.local',
      client: mockClient,
    );

    final reply = await service.sendMessages([
      const Message(role: 'user', content: 'hola'),
    ]);

    expect(reply, 'buenas');
  });

  test('sendMessages throws a friendly error on failure', () async {
    final mockClient = MockClient((request) async {
      return http.Response(jsonEncode({'error': 'boom'}), 502);
    });
    final service = ChatService(baseUrl: 'http://test.local', client: mockClient);

    expect(
      () => service.sendMessages([const Message(role: 'user', content: 'x')]),
      throwsA(isA<Exception>()),
    );
  });
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter test test/chat_service_test.dart
```
Expected: FAIL — `chat_service.dart` / `ChatService` not found.

- [ ] **Step 3: Write minimal implementation**

Create `app/lib/services/chat_service.dart`:
```dart
import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/message.dart';

class ChatService {
  final String baseUrl;
  final http.Client client;

  ChatService({required this.baseUrl, http.Client? client})
      : client = client ?? http.Client();

  Future<String> sendMessages(List<Message> messages) async {
    final response = await client.post(
      Uri.parse('$baseUrl/chat'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'messages': messages.map((m) => m.toJson()).toList()}),
    );

    if (response.statusCode != 200) {
      throw Exception('No me pude conectar, inténtalo de nuevo');
    }
    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data['reply'] as String;
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter test test/chat_service_test.dart
```
Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
cd ~/Repos/asistente-ia && git add app/lib/services/chat_service.dart app/test/chat_service_test.dart && git commit -m "feat(app): add ChatService"
```

---

### Task 10: Chat screen UI

**Files:**
- Create: `app/lib/screens/chat_screen.dart`
- Create: `app/test/chat_screen_test.dart`

- [ ] **Step 1: Write the failing widget test**

Create `app/test/chat_screen_test.dart`:
```dart
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/testing.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:asistente_ia/services/chat_service.dart';
import 'package:asistente_ia/screens/chat_screen.dart';

void main() {
  testWidgets('typing and sending shows the user bubble', (tester) async {
    final mockClient = MockClient((request) async {
      return http.Response(jsonEncode({'reply': 'respuesta IA'}), 200);
    });
    final service = ChatService(baseUrl: 'http://test.local', client: mockClient);

    await tester.pumpWidget(MaterialApp(home: ChatScreen(service: service)));

    await tester.enterText(find.byType(TextField), 'hola mundo');
    await tester.tap(find.byIcon(Icons.send));
    await tester.pump(); // user bubble appears immediately

    expect(find.text('hola mundo'), findsOneWidget);

    await tester.pumpAndSettle(); // wait for async reply
    expect(find.text('respuesta IA'), findsOneWidget);
  });
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter test test/chat_screen_test.dart
```
Expected: FAIL — `chat_screen.dart` / `ChatScreen` not found.

- [ ] **Step 3: Write the implementation**

Create `app/lib/screens/chat_screen.dart`:
```dart
import 'package:flutter/material.dart';
import '../models/message.dart';
import '../services/chat_service.dart';

class ChatScreen extends StatefulWidget {
  final ChatService service;
  const ChatScreen({super.key, required this.service});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final _controller = TextEditingController();
  final _messages = <Message>[];
  bool _loading = false;

  Future<void> _send() async {
    final text = _controller.text.trim();
    if (text.isEmpty || _loading) return;

    setState(() {
      _messages.add(Message(role: 'user', content: text));
      _loading = true;
      _controller.clear();
    });

    try {
      final reply = await widget.service.sendMessages(_messages);
      setState(() => _messages.add(Message(role: 'assistant', content: reply)));
    } catch (e) {
      setState(() => _messages
          .add(Message(role: 'assistant', content: 'No me pude conectar, inténtalo de nuevo')));
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Asistente IA')),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(12),
              itemCount: _messages.length,
              itemBuilder: (context, i) {
                final m = _messages[i];
                return Align(
                  alignment: m.isUser ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    margin: const EdgeInsets.symmetric(vertical: 4),
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: m.isUser ? Colors.blue.shade100 : Colors.grey.shade200,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(m.content),
                  ),
                );
              },
            ),
          ),
          if (_loading) const Padding(
            padding: EdgeInsets.all(8),
            child: Text('escribiendo…'),
          ),
          Padding(
            padding: const EdgeInsets.all(8),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _controller,
                    decoration: const InputDecoration(
                      hintText: 'Escribe un mensaje…',
                      border: OutlineInputBorder(),
                    ),
                    onSubmitted: (_) => _send(),
                  ),
                ),
                IconButton(icon: const Icon(Icons.send), onPressed: _send),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter test test/chat_screen_test.dart
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd ~/Repos/asistente-ia && git add app/lib/screens/chat_screen.dart app/test/chat_screen_test.dart && git commit -m "feat(app): add chat screen UI"
```

---

### Task 11: Wire up main.dart with backend URL

**Files:**
- Modify: `app/lib/main.dart` (replace generated content)

- [ ] **Step 1: Replace main.dart**

Overwrite `app/lib/main.dart`:
```dart
import 'package:flutter/material.dart';
import 'services/chat_service.dart';
import 'screens/chat_screen.dart';

// Para desarrollo local con emulador Android usa http://10.0.2.2:8080
// En producción, reemplaza por la URL pública de Render (Tarea 13).
const backendUrl = String.fromEnvironment(
  'BACKEND_URL',
  defaultValue: 'http://10.0.2.2:8080',
);

void main() {
  runApp(const AsistenteApp());
}

class AsistenteApp extends StatelessWidget {
  const AsistenteApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Asistente IA',
      theme: ThemeData(colorSchemeSeed: Colors.blue, useMaterial3: true),
      home: ChatScreen(service: ChatService(baseUrl: backendUrl)),
    );
  }
}
```

- [ ] **Step 2: Verify the whole app analyzes and tests pass**

Run:
```bash
cd ~/Repos/asistente-ia/app && flutter analyze && flutter test
```
Expected: "No issues found!" and all tests PASS.

- [ ] **Step 3: Commit**

```bash
cd ~/Repos/asistente-ia && git add app/lib/main.dart && git commit -m "feat(app): wire up main with configurable backend url"
```

---

## PHASE 3 — Deploy (free) and end-to-end

### Task 12: Get a free Groq API key

**Files:** none (manual)

- [ ] **Step 1: Create the key**

Manual steps:
1. Go to https://console.groq.com in a browser.
2. Sign up / log in (free, no card required).
3. Open "API Keys" → "Create API Key". Copy it (starts with `gsk_`).
4. Save it somewhere safe; it will go into Render as `GROQ_API_KEY`. **Never commit it to git.**

- [ ] **Step 2: Verify it works locally**

Run the smoke test from Task 6 Step 3 with the real key. Expected: a real reply from the AI.

---

### Task 13: Deploy the Go backend to Render

**Files:**
- Create: `render.yaml`

- [ ] **Step 1: Add a Render blueprint**

Create `render.yaml`:
```yaml
services:
  - type: web
    name: asistente-ia-server
    runtime: go
    rootDir: server
    buildCommand: go build -o app .
    startCommand: ./app
    envVars:
      - key: GROQ_API_KEY
        sync: false
```

- [ ] **Step 2: Push the repo to GitHub**

Run (create the repo first with `gh repo create` or via the website):
```bash
cd ~/Repos/asistente-ia && gh repo create asistente-ia --public --source=. --push
```
Expected: repo created and pushed.

- [ ] **Step 3: Create the service on Render**

Manual steps:
1. Go to https://render.com, sign up (free), connect GitHub.
2. New → "Blueprint" → pick the `asistente-ia` repo → Render reads `render.yaml`.
3. When prompted, paste the `GROQ_API_KEY` value (the `gsk_...` key).
4. Deploy. Wait for "Live". Copy the public URL (e.g. `https://asistente-ia-server.onrender.com`).

- [ ] **Step 4: Verify the deployed backend**

Run (replace with your URL):
```bash
curl -s https://asistente-ia-server.onrender.com/health
curl -s -X POST https://asistente-ia-server.onrender.com/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"hola"}]}'
```
Expected: `{"status":"ok"}` then a JSON reply. (First call may take ~30s while the free instance wakes up.)

- [ ] **Step 5: Commit**

```bash
cd ~/Repos/asistente-ia && git add render.yaml && git commit -m "chore: add Render deploy blueprint" && git push
```

---

### Task 14: Run the app against the live backend on a real device

**Files:** none (manual)

- [ ] **Step 1: Run pointing at the live URL**

Run (replace with your Render URL):
```bash
cd ~/Repos/asistente-ia/app && flutter run --dart-define=BACKEND_URL=https://asistente-ia-server.onrender.com
```
Expected: app launches on the connected device/emulator.

- [ ] **Step 2: Manual end-to-end test**

1. Type "Hola, ¿quién eres?" and send.
2. Confirm the user bubble appears, "escribiendo…" shows, then the AI reply appears.
3. Send a follow-up to confirm conversation context works.

Expected: a natural multi-turn conversation. This is the demo you show your boss.

---

## PHASE 4 — Polish

### Task 15: README for the demo

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write the README**

Create `README.md`:
```markdown
# Asistente IA

App de chat con IA hecha con **Flutter** (móvil) y un **backend en Go** que protege la clave
de la IA y llama a la API gratuita de **Groq** (modelos Llama). Backend desplegado en Render.

## Arquitectura
App Flutter  →  Backend Go (`/chat`)  →  Groq API  →  respuesta

## Backend (Go)
```bash
cd server
GROQ_API_KEY=tu_clave go run .
# Health:  curl localhost:8080/health
```

## App (Flutter)
```bash
cd app
flutter run --dart-define=BACKEND_URL=https://TU-URL.onrender.com
```

## Tests
```bash
cd server && go test ./...
cd app && flutter test
```
```

- [ ] **Step 2: Commit and push**

```bash
cd ~/Repos/asistente-ia && git add README.md && git commit -m "docs: add README" && git push
```

---

## Self-Review Notes

- **Spec coverage:** chat UI (Task 10), Go proxy + secret via env (Tasks 4–6), Groq free provider (Task 5/12), Render deploy (Task 13), `/health` (Task 3), error handling app+server (Tasks 4, 9, 10), tests both sides (Tasks 3, 4, 9, 10), README (Task 15), demo on device = Camino B (Task 14). All spec sections covered.
- **Out of scope kept out:** no auth, no database, no streaming — matches spec §9.
- **Type consistency:** `Message{role,content}`, `ChatRequest.messages`, `ChatResponse.reply`, `AIClient.Complete`, `ChatService.sendMessages`, `ChatScreen(service:)` used consistently across Go and Dart.
```
