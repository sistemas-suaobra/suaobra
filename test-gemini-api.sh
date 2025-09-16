#!/bin/bash

# Script para testar a integração com a API do Gemini
# Execute: chmod +x test-gemini-api.sh && ./test-gemini-api.sh

echo "🔍 Testando integração com a API do Gemini..."

# Verificar se as variáveis estão definidas
if [ -z "$GEMINI_API_KEY" ]; then
    echo "❌ GEMINI_API_KEY não está definida"
    echo "   Configure a variável de ambiente: export GEMINI_API_KEY=sua_chave_aqui"
    echo "   Obtenha sua chave em: https://aistudio.google.com/app/apikey"
    exit 1
fi

echo "✅ Variável de ambiente encontrada:"
echo "   GEMINI_API_KEY: ${GEMINI_API_KEY:0:20}..."

# Testar conectividade com a API do Gemini
echo ""
echo "🌐 Testando conectividade com a API do Gemini..."

# Payload de teste
test_payload='{
  "contents": [
    {
      "parts": [
        {
          "text": "Teste de conectividade. Responda apenas: OK"
        }
      ]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "topK": 40,
    "topP": 0.95,
    "maxOutputTokens": 10
  }
}'

# Fazer requisição de teste
response=$(curl -s -w "%{http_code}" \
    -H "Content-Type: application/json" \
    -d "$test_payload" \
    "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent?key=$GEMINI_API_KEY")

# Extrair código HTTP
http_code="${response: -3}"
response_body="${response%???}"

echo "HTTP Status: $http_code"

if [ "$http_code" = "000" ]; then
    echo "❌ Não foi possível conectar à API do Gemini"
    echo "   Verifique sua conexão com a internet"
elif [ "$http_code" = "400" ]; then
    echo "❌ Erro na requisição (HTTP 400)"
    echo "   Verifique o formato da requisição"
elif [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
    echo "❌ Erro de autenticação (HTTP $http_code)"
    echo "   Verifique se a GEMINI_API_KEY está correta"
    echo "   Obtenha uma nova chave em: https://aistudio.google.com/app/apikey"
elif [ "$http_code" = "429" ]; then
    echo "⚠️  Rate limit excedido (HTTP 429)"
    echo "   Tente novamente em alguns minutos"
elif [ "$http_code" = "200" ]; then
    echo "✅ API do Gemini está funcionando!"
    echo ""
    echo "📄 Resposta da API:"
    echo "$response_body" | python3 -m json.tool 2>/dev/null || echo "$response_body"
else
    echo "⚠️  Resposta inesperada (HTTP $http_code)"
    echo "   Resposta: $response_body"
fi

echo ""
echo "📋 Próximos passos:"
echo "   1. Configure a variável GEMINI_API_KEY no seu .env"
echo "   2. Teste a funcionalidade através da interface do suaobra-app"
echo "   3. Monitore os logs para verificar se está funcionando corretamente"