-- Script para adicionar campos faltantes na tabela rudderstack
-- Execute este script manualmente no banco de dados

-- Adicionar campo receivedAt
ALTER TABLE rudderstack ADD COLUMN receivedAt TEXT DEFAULT "";

-- Adicionar campo request_ip
ALTER TABLE rudderstack ADD COLUMN request_ip TEXT DEFAULT "";

-- Adicionar campo rudderId
ALTER TABLE rudderstack ADD COLUMN rudderId TEXT DEFAULT "";

-- Adicionar campo timestamp
ALTER TABLE rudderstack ADD COLUMN timestamp TEXT DEFAULT "";

-- Verificar se os campos foram adicionados
.schema rudderstack
