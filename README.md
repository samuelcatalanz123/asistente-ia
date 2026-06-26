# Asistente IA 🤖

Asistente de chat con **Inteligencia Artificial**, multiplataforma (web, móvil y escritorio),
construido con **Flutter** y un **backend en Go**. Es **multimodal**: además de chatear,
**genera imágenes** 🎨 y **entiende las fotos que le subes** 👁️. Responde en *streaming*
(palabra a palabra), formatea el código, habla por voz y guarda tus conversaciones.
La versión web es una **PWA instalable** 📲: se añade a la pantalla de inicio, abre a
pantalla completa y funciona sin conexión.

[![CI](https://github.com/samuelcatalanz123/asistente-ia/actions/workflows/ci.yml/badge.svg)](https://github.com/samuelcatalanz123/asistente-ia/actions/workflows/ci.yml)

🌐 **Pruébalo en vivo:** https://asistente-ia-xh5v.onrender.com

---

## 📸 Capturas

| Pantalla de inicio | Genera y entiende imágenes 🎨👁️ |
|--------------------|----------------------------------|
| ![Inicio](docs/captura-inicio.png) | ![Chat](docs/captura-chat.png) |

---

## ✨ Funciones

- 💬 **Chat con IA** en *streaming* (la respuesta aparece palabra a palabra, como ChatGPT)
- 📲 **App instalable (PWA)**: instálala en el móvil o el escritorio; abre a pantalla completa y **funciona sin conexión** (*service worker*)
- 🎨 **Genera imágenes**: pídele *"imagen de un gato"* o *"dibuja un dragón"* y la crea con IA — con **estilos** (realista, anime, acuarela…), **descargar** 💾 y **compartir** 📤
- 👁️ **Entiende imágenes (visión)**: súbele una foto (o **tómala con la cámara** 📸) y te la describe o responde sobre ella
- ⚡ **Dos cerebros**: modelo *Rápido* (responde al instante) o *Inteligente* (mejor respuesta)
- 🎙️ **Manos libres**: modo de voz continua para hablar sin tocar la pantalla
- 📝 **Resumir** y 🌐 **traducir** con un toque, sobre tu texto o la última respuesta
- 💻 **Formato de código**: bloques con resaltado y botón de *copiar* — experto en programación
- 🎤 **Voz**: dicta tus mensajes por micrófono y escucha las respuestas en voz alta
- 🌍 **Bilingüe**: responde en el mismo idioma en el que le escribes (español / inglés)
- 🗂️ **Varias conversaciones**: panel lateral con **buscador** 🔍 y **renombrar** ✏️ (estilo ChatGPT)
- ⭐ **Favoritos**, 🖼️ **galería** de imágenes y 📑 **exportar la conversación a PDF**
- 🌙 **Modo oscuro** y diseño moderno y responsive
- 🛡️ **Seguridad**: la clave de la IA vive solo en el servidor; *rate limiting* por IP
- 🛠️ **Backend de producción**: *graceful shutdown* (no corta peticiones al reiniciar), *middleware* de **logging** y de **recuperación de pánico**, y `/health` con versión y *uptime* para monitoreo
- ⚙️ **CI + tests**: **35 tests** de Go (handlers, streaming, visión, rate limit, middlewares) + tests de Flutter, ejecutados automáticamente en cada cambio

## 🏗️ Arquitectura

```
App Flutter / Web  →  Backend Go  →  [recovery → logging → CORS → rate limit]  →  Groq API
                      (/chat, /chat/stream)                                        ├─ Llama 3.x  (texto)
                                                                                   └─ Llama 4 Scout  (visión 👁️)
App (web)           →  Pollinations  →  imágenes generadas 🎨
```

Cada petición pasa por una cadena de *middlewares* (recuperación de pánico, logging,
CORS y *rate limiting*) antes de llegar al handler. La app envía la conversación al
backend en Go, que añade la clave secreta (guardada como variable de entorno,
**nunca en el código**) y reenvía la consulta a Groq, devolviendo la respuesta en
*streaming* mediante **Server-Sent Events (SSE)**. Cuando subes una foto, el backend usa el **modelo de visión**
de Groq para describirla; y las imágenes que pides (*"dibuja…"*) se generan con
**Pollinations** (gratis, sin clave).

## 🧰 Stack

| Capa | Tecnología |
|------|------------|
| Frontend | **Flutter / Dart** (móvil, escritorio) + cliente web en HTML/JS |
| Web app | **PWA** instalable (manifest + *service worker*, funciona offline) |
| Backend | **Go** (librería estándar, sin frameworks) |
| IA | **Groq API** — Llama 3.x (texto) y Llama 4 Scout (visión 👁️); **Pollinations** para generar imágenes 🎨 |
| Hosting | **Render** (backend) con **CI/CD**: el deploy se dispara solo cuando los tests pasan |
| Contenedores | **Docker** (build multi-stage, imagen mínima) |
| Calidad | **GitHub Actions** (CI), **35 tests** en Go + tests en Flutter |
| Producción | *graceful shutdown*, *middlewares* (recovery, logging), `/health` con monitoreo |

## ▶️ Cómo ejecutarlo

**Backend (Go):**
```bash
cd server
GROQ_API_KEY=tu_clave PORT=8090 go run .
# Comprobar:  curl localhost:8090/health
```

**App (Flutter):**
```bash
cd app
flutter pub get
flutter run                                  # usa por defecto el backend en Render
# Backend local:  flutter run --dart-define=BACKEND_URL=http://10.0.2.2:8090
```

**Backend con Docker 🐳:**
```bash
cd server
docker build -t asistente-ia .
docker run -p 8080:8080 -e GROQ_API_KEY=tu_clave asistente-ia
```

## ✅ Tests

```bash
cd server && go test ./...     # tests del backend (handlers, streaming, rate limit)
cd app    && flutter test      # tests de la app
```

## 📂 Estructura

```
asistente-ia/
├── server/   Backend en Go (proxy seguro + streaming SSE + rate limiting)
├── app/      App Flutter (móvil, escritorio) + cliente web
└── .github/  Integración continua (CI)
```

---

Hecho por **Samuel Catalán** — proyecto de aprendizaje full-stack (Go + Flutter + IA).
