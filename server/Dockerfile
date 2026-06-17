# ---- Etapa 1: construir ----
# Usamos una imagen con Go para compilar el servidor.
FROM golang:1.26 AS build
WORKDIR /app

# Copiamos primero los archivos de dependencias (mejor uso de la caché).
COPY go.mod ./
RUN go mod download

# Copiamos el resto del código y compilamos un binario estático.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o servidor .

# ---- Etapa 2: imagen final (mínima) ----
# Solo llevamos el binario ya compilado: imagen pequeña y segura.
FROM alpine:3.20
WORKDIR /app

# Certificados para poder llamar por HTTPS a la API de Groq.
RUN apk add --no-cache ca-certificates

COPY --from=build /app/servidor .

# El servidor escucha en el puerto que le diga la variable PORT (por defecto 8080).
EXPOSE 8080
CMD ["./servidor"]
