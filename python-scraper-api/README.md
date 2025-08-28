# RRT Scraper API

API para extração de dados de RRT (Registro de Responsabilidade Técnica) do site do CAU BR.

## 🚀 Quick Start

### Usando Docker (Recomendado)

```bash
# Build e deploy automático
./deploy.sh

# Ou manualmente:
docker-compose up -d
```

### Desenvolvimento Local

```bash
# Instalar dependências
pip install -r requirements.txt

# Instalar playwright
playwright install chromium

# Executar
python app.py
```

## 📋 API Endpoints

### Health Check
```
GET /health
```
Retorna o status da API.

### Obter dados de RRT
```
GET /rrt/{numero}
```
Extrai dados do RRT especificado.

**Exemplo:**
```bash
curl http://localhost:5000/rrt/9999967
```

**Resposta:**
```json
{
  "obra_number": "9999967",
  "professional": "PATRICIA DE SOUZA ANTUNES",
  "owner": "",
  "address": "RUA CENTO E VINTE E SETE, LOTE 38, QUADRA 533, JARDIM ATLÂNTICO LESTE (ITAIPUAÇU) - MARICÁ, RJ",
  "bairro": "JARDIM ATLÂNTICO LESTE (ITAIPUAÇU)",
  "city": "MARICÁ",
  "state": "RJ",
  "start_date": "2020-09-27",
  "end_date": "2022-09-27",
  "activity": "2.1.1 - EXECUÇÃO DE OBRA",
  "type": "RRT SIMPLES",
  "size": 142.2,
  "unidade": "m² - metro quadrado",
  "first_listing_date": "2020-09-23"
}
```

## 🐳 Docker

### Build
```bash
docker build -t rrt-scraper-api .
```

### Run
```bash
docker run -p 5000:5000 rrt-scraper-api
```

### Docker Compose
```bash
docker-compose up -d
```

## 🔧 Configuração

### Variáveis de Ambiente

- `FLASK_ENV`: Ambiente (development/production)
- `PYTHONUNBUFFERED`: Para logs em tempo real

### Timeouts

- Timeout de navegação: 60 segundos
- Timeout de elemento: 60 segundos
- Workers Gunicorn: 2
- Timeout Gunicorn: 120 segundos

## 📊 Monitoramento

A API inclui health check em `/health` que pode ser usado para monitoramento com ferramentas como:
- Docker health check
- Kubernetes liveness/readiness probes
- Load balancers

## 🛠️ Desenvolvimento

### Estrutura de Arquivos
```
python-scraper-api/
├── app.py              # Aplicação principal
├── requirements.txt    # Dependências Python
├── Dockerfile         # Imagem Docker
├── docker-compose.yml # Compose para desenvolvimento
├── deploy.sh         # Script de deploy
├── .dockerignore     # Arquivos ignorados no build
└── README.md         # Documentação
```

### Dependências Principais
- Flask: Framework web
- Playwright: Automação de navegador
- Gunicorn: Servidor WSGI para produção

## 🚨 Troubleshooting

### Container não inicia
```bash
# Verificar logs
docker-compose logs

# Verificar status
docker-compose ps
```

### API não responde
```bash
# Verificar health check
curl http://localhost:5000/health

# Verificar se a porta está aberta
netstat -tulpn | grep 5000
```

### Problemas com Playwright
```bash
# Reinstalar navegadores
playwright install chromium
```
