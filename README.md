# Asistente IA 🤖

App de chat con IA hecha con **Flutter** (móvil) y un **backend en Go** que protege la clave
de la IA y llama a la API gratuita de **Groq** (modelos Llama). Backend desplegado en **Render**.

🌐 **Backend en producción:** https://asistente-ia-xh5v.onrender.com

## Arquitectura

```
App Flutter  →  Backend Go (/chat)  →  Groq API  →  respuesta
```

La app envía la conversación al backend; el backend añade la clave secreta (guardada como
variable de entorno, nunca en el código) y reenvía la petición a Groq.

## Backend (Go)

```bash
cd server
GROQ_API_KEY=tu_clave PORT=8090 go run .
# Comprobar:  curl localhost:8090/health
```

## App (Flutter)

```bash
cd app
# Usa por defecto el backend en Render. Para apuntar a un backend local:
flutter run --dart-define=BACKEND_URL=http://10.0.2.2:8090
```

## Tests

```bash
cd server && go test ./...     # 5 tests
cd app    && flutter test      # 3 tests
```

## Stack

- **Flutter / Dart** — app móvil multiplataforma
- **Go** (librería estándar) — backend proxy seguro
- **Groq API** (Llama 3.3) — IA, capa gratuita
- **Render** — hosting gratuito del backend
