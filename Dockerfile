FROM golang:1.24.0-alpine AS builder

WORKDIR /app

# Instalar dependencias necesarias
RUN apk add --no-cache gcc musl-dev

# Copiar archivos de módulos y descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN go build -o app .
RUN chmod +x app

# Imagen final más pequeña
FROM alpine:latest

WORKDIR /app

# Instalar dependencias necesarias incluyendo postgresql-client
RUN apk add --no-cache tzdata ca-certificates postgresql-client

# Copiar binario compilado desde el builder
COPY --from=builder /app/app .

# Copiar plantillas y archivos estáticos
COPY templates/ /app/templates/
COPY static/ /app/static/
COPY .env /app/.env
COPY wait-for-postgres.sh /app/wait-for-postgres.sh
RUN chmod +x /app/wait-for-postgres.sh

# Exponer puerto
EXPOSE 4332

# Comando para iniciar la aplicación
CMD ["./app"]