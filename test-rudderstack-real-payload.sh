#!/bin/bash

# Script para testar o endpoint do RudderStack com payload REAL

# URL do endpoint
URL="${1:-http://localhost:8090/api/collections/rudderstack/records}"

echo "🧪 Testando com payload REAL do RudderStack..."
echo "URL: $URL"
echo ""

curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "TOKEN: sua-obra-rudderstack" \
  -d '{
    "anonymousId": "4a73f8bc-ec12-44a7-aba6-34dc3b419f4b",
    "channel": "web",
    "context": {
        "app": {
            "installType": "npm",
            "name": "RudderLabs JavaScript SDK",
            "namespace": "com.rudderlabs.javascript",
            "version": "3.23.0"
        },
        "campaign": {},
        "ip": "177.80.126.6",
        "library": {
            "name": "RudderLabs JavaScript SDK",
            "version": "3.23.0"
        },
        "locale": "pt-BR",
        "os": {
            "name": "",
            "version": ""
        },
        "page": {
            "initial_referrer": "https://suaobra.com.br/",
            "initial_referring_domain": "suaobra.com.br",
            "path": "/obras-plus/index.html",
            "referrer": "$direct",
            "referring_domain": "",
            "search": "",
            "tab_url": "https://app.suaobra.com.br/obras-plus/index.html",
            "title": "Obras+",
            "url": "https://app.suaobra.com.br/obras-plus/index.html"
        },
        "screen": {
            "density": 1,
            "height": 900,
            "innerHeight": 739,
            "innerWidth": 1440,
            "width": 1440
        },
        "sessionId": 1760126668316,
        "timezone": "GMT-0300",
        "traits": {
            "user": {
                "email": "sollartecengenharia@gmail.com",
                "id": "user_TrvM8fl3rMybnbv",
                "team_id": "team_QSaSl9FWEYtoeSN"
            }
        },
        "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
    },
    "event": "obras-plus-get-records",
    "integrations": {
        "All": true
    },
    "messageId": "eeda298b-3cc5-480f-a4e2-8f7a80ba45df",
    "originalTimestamp": "2025-10-10T20:20:14.987Z",
    "properties": {
        "city": {
            "city": "GUARULHOS",
            "id": "SP-GUARULHOS",
            "state": "SP"
        },
        "endDateFrom": "",
        "endDateTo": "",
        "filter": "",
        "itemsPerPage": 10,
        "neighborhood": [],
        "order": "first_listing_date-desc,start_date-desc",
        "page_number": 2,
        "records": 10,
        "rows_per_page": 10,
        "size": "0-9999999",
        "startDateFrom": "",
        "startDateTo": "",
        "state": "SP",
        "statuses": [
            "todos"
        ],
        "total": 8322,
        "user": {
            "email": "sollartecengenharia@gmail.com",
            "id": "user_TrvM8fl3rMybnbv",
            "legacy_id": "",
            "team_id": "team_QSaSl9FWEYtoeSN"
        }
    },
    "receivedAt": "2025-10-10T20:20:15.662Z",
    "request_ip": "177.80.126.6",
    "rudderId": "5089a451-e691-4d98-b8dd-6f44a0fe6fce",
    "sentAt": "2025-10-10T20:20:14.987Z",
    "timestamp": "2025-10-10T20:20:15.662Z",
    "type": "track",
    "userId": "user_TrvM8fl3rMybnbv"
  }' \
  -w "\n\n📊 Status: %{http_code}\n⏱️  Time: %{time_total}s\n\n" \
  -s | jq '.' 2>/dev/null || cat

echo "✅ Teste concluído!"
