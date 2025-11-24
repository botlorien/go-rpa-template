# Estágio de Build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compila ambos os binários
RUN go build -o /dist/api cmd/api/main.go
RUN go build -o /dist/worker cmd/cli/main.go

# Estágio Final (Imagem Leve)
FROM alpine:latest

# Instala Chromium e dependências para o Rod funcionar
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    tzdata

# Configura variável para o Rod achar o binário
ENV ROD_BIN=/usr/bin/chromium-browser

WORKDIR /root/

COPY --from=builder /dist/api .
COPY --from=builder /dist/worker .
COPY --from=builder /app/config/config.yaml . 

# Define variáveis padrão
ENV TZ=America/Sao_Paulo

# O padrão é rodar a API, mas pode ser sobrescrito
CMD ["./api"]