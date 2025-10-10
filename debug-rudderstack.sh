#!/bin/bash

echo "🔍 Testando endpoint RudderStack com debug completo"
echo "=================================================="
echo ""

URL="https://api.suaobra.test/api/collections/rudderstack/records"

echo "📡 URL: $URL"
echo "🔑 TOKEN: sua-obra-rudderstack"
echo ""

# Payload mínimo
PAYLOAD='{
  "type": "track",
  "event": "test_event",
  "anonymousId": "debug-test-'$(date +%s)'",
  "channel": "web",
  "messageId": "debug-msg-'$(date +%s)'",
  "userId": "debug-user",
  "originalTimestamp": "2025-10-10T20:00:00.000Z",
  "sentAt": "2025-10-10T20:00:00.000Z"
}'

echo "📦 Payload:"
echo "$PAYLOAD" | jq '.'
echo ""

echo "🚀 Enviando requisição..."
RESPONSE=$(curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d "$PAYLOAD" \
  -k -s -w "\n%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "📨 Status HTTP: $HTTP_CODE"
echo "📄 Response Body:"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ SUCESSO!"
else
    echo "❌ ERRO $HTTP_CODE"
    echo ""
    echo "🔍 Verificando logs do container..."
    docker-compose logs backend --tail=5 | grep -i "error\|failed" || echo "Nenhum erro encontrado nos logs"
fi
