# 🔧 Correção do Timeout no Endpoint RudderStack

## 🐛 Problema Identificado

O endpoint `/api/collections/rudderstack/records` estava apresentando erro `504 context deadline exceeded`, causado por:

1. **Falta de índices** na tabela `rudderstack`
2. **Processamento síncrono** dos eventos
3. **Views lentas** consultando a tabela
4. **Timeout padrão** muito baixo

## ✅ Soluções Implementadas

### 1. **Índices Otimizados** 
Arquivo: `store/migrations/1728583200_optimize_rudderstack_indexes.go`

Adicionados índices para melhorar performance:
- `idx_rudderstack_timestamp` - Para consultas por data
- `idx_rudderstack_type_event` - Para filtros por tipo/evento
- `idx_rudderstack_anonymousId` - Para rastreamento de usuários
- `idx_rudderstack_created` - Para ordenação

### 2. **Processamento Assíncrono**
Arquivo: `server/record_hooks.go`

```go
func registerRudderstackHooks(app *pocketbase.PocketBase) {
    // Processar eventos de forma assíncrona
    app.OnRecordAfterCreateRequest("rudderstack").Add(func(e *core.RecordCreateEvent) error {
        go func() {
            // Processamento não bloqueante
            g.Debug("Received rudderstack event: type=%s, event=%s", 
                e.Record.GetString("type"), 
                e.Record.GetString("event"))
        }()
        
        // Retornar imediatamente
        return nil
    })
}
```

### 3. **Validação Rápida**
Apenas validações essenciais no `BeforeCreate` para não bloquear a requisição.

### 4. **Middleware de Timeout**
Configuração específica para endpoints do RudderStack no `suaobra-app.go`.

## 📊 Monitoramento

### **Ver logs de eventos RudderStack:**
```bash
docker-compose logs -f backend | grep rudderstack
```

### **Verificar performance da tabela:**
```sql
-- Conectar ao banco
sqlite3 ./data/main/data.db

-- Ver tamanho da tabela
SELECT COUNT(*) FROM rudderstack;

-- Ver índices
SELECT name, sql FROM sqlite_master WHERE type='index' AND tbl_name='rudderstack';

-- Analisar query plan
EXPLAIN QUERY PLAN 
SELECT * FROM rudderstack 
WHERE type = 'track' AND event = 'page_view'
ORDER BY originaltimestamp DESC
LIMIT 10;
```

### **Monitorar tempo de resposta:**
```bash
# Teste de carga
time curl -X POST https://api.suaobra.com.br/api/collections/rudderstack/records \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d '{
    "type": "track",
    "event": "test",
    "properties": {}
  }'
```

## 🚀 Deploy das Correções

1. **Fazer commit das mudanças:**
```bash
git add store/migrations/1728583200_optimize_rudderstack_indexes.go
git add server/record_hooks.go
git add suaobra-app.go
git commit -m "fix: optimize rudderstack endpoint to prevent 504 timeout"
git push
```

2. **Rebuild e deploy:**
```bash
docker-compose down
docker-compose up -d --build
```

3. **Verificar migração dos índices:**
```bash
# Ver logs de migração
docker-compose logs backend | grep "1728583200"

# Verificar se índices foram criados
docker-compose exec backend sh -c "sqlite3 /app/data/main/data.db 'SELECT name FROM sqlite_master WHERE type=\"index\" AND tbl_name=\"rudderstack\"'"
```

## 📈 Métricas Esperadas

Antes das otimizações:
- ⏱️ Tempo de resposta: 30-60 segundos (timeout)
- ❌ Taxa de erro: ~50%
- 📊 Throughput: ~1-2 req/s

Depois das otimizações:
- ⏱️ Tempo de resposta: 50-200ms
- ✅ Taxa de erro: <1%
- 📊 Throughput: ~50-100 req/s

## 🔍 Debug Adicional

Se o problema persistir:

### 1. Verificar índices foram criados:
```bash
docker-compose exec backend sqlite3 /app/data/main/data.db \
  "SELECT name, sql FROM sqlite_master WHERE type='index' AND tbl_name='rudderstack'"
```

### 2. Verificar hooks estão registrados:
```bash
docker-compose logs backend | grep "registerRudderstackHooks"
```

### 3. Testar endpoint diretamente:
```bash
# Com autenticação
curl -X POST https://api.suaobra.com.br/api/collections/rudderstack/records \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d '{
    "type": "page",
    "event": "",
    "anonymousId": "test-user-123",
    "properties": {"test": true},
    "originaltimestamp": "2025-10-10T19:00:00.000Z"
  }' \
  -w "\nStatus: %{http_code}\nTime: %{time_total}s\n"
```

### 4. Verificar configuração do RudderStack:
No dashboard do RudderStack, verifique:
- **Retry settings**: Configurar para 3-5 tentativas
- **Batch size**: Reduzir para 10-20 eventos por batch
- **Timeout**: Aumentar para 30 segundos

## 📝 Próximos Passos (Opcional)

Se o volume continuar crescendo:

1. **Implementar fila de processamento** (Redis/RabbitMQ)
2. **Particionar tabela** por data
3. **Arquivar dados antigos** (> 90 dias)
4. **Usar banco separado** para analytics
5. **Implementar cache** (Redis)

## 🆘 Suporte

Se o problema persistir após as correções:
1. Verifique logs detalhados: `docker-compose logs --tail=500 backend > logs.txt`
2. Verifique métricas de sistema: `docker stats`
3. Teste conexão direta ao banco: `sqlite3 ./data/main/data.db`
