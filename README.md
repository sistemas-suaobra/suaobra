# Sua Obra App

Aplicação web para gerenciamento de obras, leads e contatos.

## Stack Tecnológico

- **Frontend**: Astro, React, Nano Stores.
- **Backend**: Go, PocketBase (SQLite).
- **Infraestrutura**: Docker, Caddy.

## Pré-requisitos

Para rodar o projeto localmente, você precisará de:

- **Go** 1.24 ou superior
- **Node.js** 18 ou superior
- **Docker** (opcional, para rodar via container)

## Instalação e Execução Local

### 1. Backend (Go + PocketBase)

O backend é construído sobre o PocketBase. Ele gerencia autenticação, banco de dados e API.

1. Navegue para a raiz do projeto.
2. Configure as variáveis de ambiente:
   ```bash
   cp .env.example .env
   ```
   Edite o arquivo `.env` conforme necessário (chaves de API, etc). O padrão já funciona para dev básico.

3. Instale as dependências:
   ```bash
   go mod download
   ```
4. Execute o servidor:
   ```bash
   go run suaobra-app.go serve
   ```
   
   O servidor iniciará em `http://127.0.0.1:8090` (padrão) ou `http://0.0.0.0:8080` (configurações do ambiente podem variar).
   
   > **Nota sobre Banco de Dados**: O PocketBase utiliza SQLite. Ao rodar o backend pela primeira vez, ele criará automaticamente o banco de dados `data.db` e aplicará as migrações necessárias no diretório `data/main/`. Não é necessária configuração manual de banco de dados SQL.

### 2. Frontend (Astro)

1. Navegue para a pasta `frontend`:
   ```bash
   cd frontend
   ```
2. Instale as dependências:
   ```bash
   npm install
   ```
3. Inicie o servidor de desenvolvimento:
   ```bash
   npm run dev
   ```
   O frontend estará disponível em `http://localhost:3000` (ou a porta indicada no terminal).

4. (Opcional, recomendado para dev) Defina explicitamente a URL da API no frontend:
   crie `frontend/.env` com:
   ```bash
   PUBLIC_API_BASE_URL=http://127.0.0.1:8090
   ```
   Isso evita mismatch de porta quando o backend estiver rodando localmente.

## Teste local rápido (campanha + IA)

1. Suba backend e frontend localmente.
2. Crie uma campanha com **"Manter IA após resposta"** desativado.
3. Inicie a campanha e envie uma resposta pelo WhatsApp do lead.
4. Resultado esperado: a conversa deve ficar `PAUSADA` e a IA **não** deve responder automaticamente.

Para rodar os testes de backend relacionados:
```bash
scripts/test-smoke.ps1
```

## Executando com Docker

Se preferir rodar toda a aplicação em containers (simulando produção):

1. Certifique-se de que o Docker está rodando.
2. Adicione os seguintes domínios ao seu arquivo hosts (`/etc/hosts` no Linux/Mac ou `C:\Windows\System32\drivers\etc\hosts` no Windows) apontando para `127.0.0.1`:
   ```
   127.0.0.1 app.suaobra.test
   127.0.0.1 api.suaobra.test
   ```
3. Na raiz do projeto, execute:
   ```bash
   docker-compose up -d --build
   ```
4. Acesse a aplicação em `http://app.suaobra.test` e a API em `http://api.suaobra.test`.

## Estrutura do Projeto

- `/frontend`: Código fonte do cliente (Astro/React).
- `/server`, `/store`, `/cmd`: Código fonte do backend Go e extensões do PocketBase.
- `/data`: Diretório onde o banco de dados SQLite é persistido.
