# Étape 1 — Build statique du binaire Go, sans métadonnées
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copie du code source
COPY binary/main.go .

# Fixe la date de build (pour un hash stable)
ENV SOURCE_DATE_EPOCH=0

# Compilation Go sans chemins locaux ni métadonnées
RUN go mod init gkeapi && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-buildid=" -o main .

# Étape 2 — Image distroless stable
FROM gcr.io/distroless/base-debian12@sha256:9e9b50d2048db3741f86a48d939b4e4cc775f5889b3496439343301ff54cdba8

# Copier uniquement le binaire compilé
COPY --from=builder /app/main /main

# Variables fixes pour stabilité
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

# Entrypoint reproductible
ENTRYPOINT ["/main"]

