# Estágio 1: Imagem base com o dado pesado
# Esta camada só será reconstruída se o arquivo core.db mudar.
# FROM alpine:latest AS data-base
# WORKDIR /app
# COPY data/core.db .


# Estágio 2: Builder para compilar a aplicação Go
# Esta camada será reconstruída a cada mudança no código-fonte.
FROM golang:1.20-alpine AS builder

# Instala as ferramentas de build
RUN apk add --no-cache build-base

WORKDIR /app

# Copia os arquivos de módulo e baixa as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código-fonte da aplicação
COPY suaobra-app.go .
COPY server/ ./server
COPY store/ ./store
COPY templates/ ./templates

# Define as flags para a compilação
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"

# Compila a aplicação
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=1 go build -ldflags="-s -w" -a -o suaobra-app .


# Estágio 3: Imagem final, combinando a base de dados e o binário
# Esta é a imagem que será usada em produção.
FROM data-base AS final

# Copia o binário compilado do estágio 'builder'.
# Esta é uma operação rápida, pois o binário é pequeno.
COPY --from=builder /app/suaobra-app .

EXPOSE 8090

# Executa a aplicação
CMD ["./suaobra-app", "serve", "--http", "0.0.0.0:8090"]
