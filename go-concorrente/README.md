# Processador Concorrente de RRTs

Este programa processa RRTs (Registros de Responsabilidade Técnica) de forma concorrente, fazendo chamadas para a API do CAUBR hospedada no Google Cloud Run.

## 🚀 Funcionalidades

- **Processamento concorrente**: Múltiplos workers processando RRTs simultaneamente
- **Resiliência**: Retry automático com backoff exponencial
- **Validação robusta**: Verificação de dados e regras de negócio
- **Logging detalhado**: Acompanhamento completo do processamento
- **Rate limiting**: Controle de frequência para não sobrecarregar a API

## 📋 Pré-requisitos

- Go 1.19 ou superior
- SQLite3
- Conexão com internet para acessar a API

## ⚙️ Configuração

### Variáveis de Ambiente (Opcionais)

```bash
# URL da API (padrão: https://caubr-api-rem6jzjgfa-uc.a.run.app)
export CAUBR_API_URL="https://sua-api-customizada.com"

# Outras configurações podem ser ajustadas no código fonte
```

### Configurações Internas

- **Timeout da API**: 30 segundos
- **Máximo de retries**: 5 tentativas
- **Workers**: 1 (para evitar sobrecarga da API)
- **Delay entre requisições**: 2-5 segundos (adaptativo)

## 🏃‍♂️ Como Usar

1. **Compilar o programa**:
   ```bash
   go build -o main .
   ```

2. **Executar**:
   ```bash
   ./main
   ```

3. **Inserir os números da RRT**:
   - Digite o número inicial da RRT
   - Digite o número final da RRT

## 📊 Regras de Processamento

### ✅ RRTs Aceitos
- Tipo diferente de "RRT MÍNIMO"
- Data de término superior a 01/09/2023 (2 anos antes de 01/09/2025)
- Campos obrigatórios preenchidos (proprietário, endereço, cidade)

### ❌ RRTs Ignorados
- Tipo "RRT MÍNIMO"
- Data de término anterior a 01/09/2023
- Campos obrigatórios vazios
- Dados inválidos ou malformados

## 🔍 Logs e Monitoramento

O programa fornece logs detalhados com emojis para facilitar o acompanhamento:

- 🔄 Processando RRT
- ✅ Sucesso
- ❌ Erro
- ⚠️ Ignorado
- 📥 INSERT no banco
- 🔄 UPDATE no banco

## 🗄️ Banco de Dados

- **Arquivo**: `./go-concorrente/core.db`
- **Tabela**: `core_obras_plus`
- **Modo**: WAL (Write-Ahead Logging) para melhor performance concorrente

## 🛠️ Tratamento de Erros

### Níveis de Retry
1. **Nível HTTP**: Timeout, conexão falhou
2. **Nível API**: Status HTTP não-200
3. **Nível Aplicação**: Erro na API (não tenta novamente)
4. **Nível Banco**: Busy timeout e transações

### Estratégias de Resiliência
- **Backoff exponencial**: Delay aumenta com tentativas falhidas
- **Timeout configurável**: Evita travamentos
- **Validação de dados**: Garante integridade antes de salvar
- **Transações seguras**: Rollback automático em caso de erro

## 📈 Métricas

Ao final do processamento, o programa exibe:
- Total de sucessos
- Total de erros
- Total de ignorados
- Taxa de sucesso percentual

## 🔧 Personalização

Para modificar configurações, edite as constantes no início do arquivo:

```go
const (
    DEFAULT_API_URL = "https://caubr-api-rem6jzjgfa-uc.a.run.app"
    REQUEST_TIMEOUT = 30 * time.Second
    MAX_RETRIES     = 5
    RETRY_DELAY     = 2 * time.Second
    BACKOFF_FACTOR  = 2.0
)
```

## 🚨 Troubleshooting

### Problema: "API não responde"
**Solução**: Verifique se a URL da API está correta e se há conectividade

### Problema: "Database busy"
**Solução**: O programa já tem busy_timeout configurado. Se persistir, reduza o número de workers

### Problema: "Timeout na API"
**Solução**: Aumente o REQUEST_TIMEOUT ou verifique a performance da API

### Problema: "Muitos erros consecutivos"
**Solução**: Verifique se a API está funcionando corretamente ou se há rate limiting
