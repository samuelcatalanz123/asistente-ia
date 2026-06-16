# Diseño: Asistente IA (App de Chat con IA)

**Fecha:** 2026-06-16
**Autor:** Samuel
**Objetivo:** Construir una app de chat con IA, profesional, para impresionar a un jefe y reforzar el CV (Flutter + Go).

---

## 1. Resumen

Una app móvil de **chat con inteligencia artificial** (estilo mini-ChatGPT) hecha en **Flutter**,
respaldada por un **backend en Go** que guarda la clave secreta de la IA y hace de intermediario
con una **API de IA gratuita (Groq)**. El backend se despliega gratis en la nube (Render).

La meta inicial es una **demo profesional** mostrada en un móvil real (Camino B, coste 0€),
dejando la app **lista para publicar** en Google Play y Apple App Store más adelante.

## 2. Decisiones tomadas (brainstorming)

- **Tipo de app:** chat con IA (factor "wow" + alineado con lo que Samuel ya estudia).
- **Coste de la IA:** 100% gratis → proveedor **Groq** (rápido, modelos Llama gratis, API compatible con OpenAI).
- **Arquitectura:** app Flutter **+ backend en Go** (más profesional y muestra dos tecnologías).
- **Publicación:** Camino B — demo gratis primero; pagar cuotas de tienda solo si se quiere publicar de verdad.

## 3. Arquitectura

```
  App Flutter  ── POST /chat ──▶  Backend Go  ── HTTPS ──▶  Groq API (IA)
  (UI de chat)                    (guarda la       ◀── respuesta ──
       ▲                           clave secreta)
       └────────── respuesta de la IA ──────────────────────────┘
```

**Flujo de datos:**
1. El usuario escribe un mensaje en la app.
2. La app envía `POST /chat` al backend Go con el mensaje y el historial de la conversación.
3. El backend añade la clave secreta (variable de entorno) y reenvía la petición a Groq.
4. Groq responde con el texto generado.
5. El backend devuelve ese texto a la app, que lo muestra en una burbuja de chat.

## 4. Componentes

### 4.1 App Flutter (`/app`)
- **Responsabilidad:** interfaz de chat y comunicación con el backend.
- **Pantallas:** una sola pantalla de chat.
- **Elementos UI:** lista de mensajes (burbujas usuario/IA), campo de texto, botón enviar,
  indicador "escribiendo…", manejo de errores (mensaje amable si falla la red).
- **Estado:** lista de mensajes en memoria + historial enviado al backend para dar contexto.
- **Depende de:** el endpoint `/chat` del backend (URL configurable).

### 4.2 Backend Go (`/server`)
- **Responsabilidad:** proxy seguro entre la app y la IA; oculta la clave.
- **Endpoint:** `POST /chat`
  - Entrada (JSON): `{ "messages": [ { "role": "user", "content": "..." } ] }`
  - Salida (JSON): `{ "reply": "texto de la IA" }`
- **Endpoint extra:** `GET /health` → `{ "status": "ok" }` (para comprobar que vive).
- **Secreto:** la clave de Groq se lee de la variable de entorno `GROQ_API_KEY` (nunca en el código).
- **CORS / config:** permitir peticiones desde la app; puerto desde variable `PORT`.
- **Depende de:** la API de Groq.

### 4.3 Despliegue (Render, gratis)
- El backend Go se sube a **Render** (capa gratuita).
- La variable `GROQ_API_KEY` se configura en el panel de Render (no se sube a git).
- La app Flutter apunta a la URL pública del backend.
- Nota: la capa gratis "duerme" tras inactividad; la primera petición tarda unos segundos. Aceptable para una demo.

## 5. Manejo de errores
- Sin internet o backend caído → la app muestra "No me pude conectar, inténtalo de nuevo".
- Error de la IA (límite gratuito, etc.) → el backend devuelve un mensaje claro, la app lo muestra.
- Entradas vacías → el botón enviar se desactiva.

## 6. Pruebas
- **Backend Go:** test del endpoint `/chat` con un cliente HTTP simulado (sin llamar a Groq real) y test de `/health`.
- **App Flutter:** widget test de la pantalla de chat (enviar un mensaje muestra la burbuja del usuario).
- **Prueba manual:** conversación real de extremo a extremo en un móvil/emulador.

## 7. Estructura de carpetas

```
asistente-ia/
├── app/        # proyecto Flutter
├── server/     # backend en Go
├── docs/       # este diseño y futuros planes
└── README.md   # cómo arrancar todo (para enseñar al jefe)
```

## 8. Coste
- IA (Groq): **0€** (capa gratuita).
- Backend (Render): **0€** (capa gratuita).
- Demo en móvil: **0€**.
- (Opcional, solo si se publica de verdad: Google Play 25€ una vez, Apple App Store 99€/año.)

## 9. Fuera de alcance (YAGNI por ahora)
- Cuentas de usuario / login.
- Guardar conversaciones en una base de datos.
- Voz, imágenes o pagos.
- Streaming de la respuesta palabra por palabra (se puede añadir después).
