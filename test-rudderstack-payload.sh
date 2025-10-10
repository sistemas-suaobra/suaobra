#!/bin/bash

# Script para testar o endpoint do RudderStack

# URL do endpoint
URL="${1:-http://localhost:8090/api/collections/rudderstack/records}"

# Payloads de teste baseados nos formatos do RudderStack

echo "🧪 Testando payload tipo 'page'..."
curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d '{
    "messageId": "test-msg-page-001",
    "type": "page",
    "anonymousId": "test-anonymous-user-123",
    "originalTimestamp": "2025-10-10T20:00:00.000Z",
    "context": {},
    "properties": {}
  }' \
  -w "\nStatus: %{http_code}\nTime: %{time_total}s\n\n"

echo "🧪 Testando payload tipo 'track'..."
curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d '{
    "messageId": "test-msg-track-001",
    "type": "track",
    "event": "button_click",
    "anonymousId": "test-anonymous-user-123",
    "originalTimestamp": "2025-10-10T20:00:00.000Z",
    "context": {"page": {"url": "https://suaobra.com.br/test"}},
    "properties": {"button": "signup"}
  }' \
  -w "\nStatus: %{http_code}\nTime: %{time_total}s\n\n"

echo "🧪 Testando payload tipo 'identify'..."
curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d '{
    "messageId": "test-msg-identify-001",
    "type": "identify",
    "userId": "user-123",
    "anonymousId": "test-anonymous-user-123",
    "originalTimestamp": "2025-10-10T20:00:00.000Z",
    "context": {},
    "properties": {"email": "test@example.com"}
  }' \
  -w "\nStatus: %{http_code}\nTime: %{time_total}s\n\n"

echo "🧪 Testando payload MÍNIMO (sem campos opcionais)..."
curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d '{
    "type": "page"
  }' \
  -w "\nStatus: %{http_code}\nTime: %{time_total}s\n\n"

echo "✅ Testes concluídos!"
