-- Query para encontrar o token real
SELECT 
    id,
    numero_e164 as token_real,
    device_jid,
    conectado_em,
    provider
FROM conexoes_whatsapp 
WHERE device_jid IS NOT NULL 
ORDER BY conectado_em DESC
LIMIT 1;