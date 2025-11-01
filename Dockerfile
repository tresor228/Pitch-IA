# Dockerfile pour déploiement alternatif si nécessaire
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Compiler l'application
RUN go build -o pitch

# Image finale
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copier le binaire compilé
COPY --from=builder /app/pitch .
COPY --from=builder /app/views ./views

# Exposer le port
EXPOSE 8080

# Commande de démarrage
CMD ["./pitch"]

