# ETL Script - Contatos CSV para SQLite

## Descrição

Script Python para importar contatos de arquivos CSV para o banco de dados SQLite (`core.db`).

## Funcionalidades

- ✓ Lê arquivos CSV com dados de contatos
- ✓ Normaliza caracteres (uppercase, sem acentos)
- ✓ Gera `contact_id` seguindo a lógica TypeScript do projeto
- ✓ Gera `person_id` usando MD5 do nome
- ✓ Popula tabelas `core_obras_plus_phone` e `core_obras_plus_email`
- ✓ Processamento em lotes (batch) para melhor performance
- ✓ Tratamento de erros robusto

## Uso

### Uso básico (caminhos padrão):

```bash
python3 scripts/etl_contacts_csv.py
```

Caminhos padrão:
- CSV: `contatos/500.000 contatos_Brasilia_DF - -- contatoszap.com --.csv`
- Database: `data/backend/core.db`

### Uso com caminhos customizados:

```bash
python3 scripts/etl_contacts_csv.py <caminho_csv> <caminho_db>
```

Exemplo:
```bash
python3 scripts/etl_contacts_csv.py contatos/meu_arquivo.csv data/backend/core.db
```

## Formato do CSV

O CSV deve conter as seguintes colunas:

- `Nome` - Nome do contato (obrigatório)
- `Telefone` - Telefone do contato (opcional)
- `Email` - Email do contato (opcional)
- `Município` ou `Cidade` - Cidade do contato (opcional)
- `UF` - Estado do contato (opcional)
- `Bairro` - Bairro do contato (opcional)

## Regras de Normalização

### Texto (Nome, Cidade, UF, Bairro):
- Convertido para MAIÚSCULAS
- Acentos removidos
- Espaços extras removidos

### Telefone:
- Apenas dígitos (remove caracteres especiais)
- Formato: `5561992928641`

### Email:
- Convertido para minúsculas
- Espaços extras removidos

## IDs Gerados

### contact_id:
```python
# Lógica baseada no TypeScript:
hash = uuid4().replace('-', '')
type = 'phone' if telefone else 'email' if email else ''
contact_id = f"cont_{type}_{hash}"
```

Exemplos:
- `cont_phone_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`
- `cont_email_z9y8x7w6v5u4t3s2r1q0p9o8n7m6l5k4`

### person_id:
```python
person_id = md5(nome)
```

Exemplo:
- Nome: `JOAO SILVA`
- person_id: `5d41402abc4b2a76b9719d911017c592`

## Tabelas Populadas

### core_obras_plus_phone
Quando o contato tem telefone.

Campos:
- `contact_id` - ID único do contato
- `person_id` - MD5 do nome
- `company_id` - NULL (não usado atualmente)
- `nome` - Nome normalizado
- `telefone` - Telefone normalizado
- `cidade` - Cidade normalizada
- `uf` - UF normalizado
- `poder_aquisitivo` - NULL (não usado atualmente)
- `source` - Fonte dos dados
- `_sling_loaded_at` - Timestamp da importação

### core_obras_plus_email
Quando o contato tem email.

Campos: (mesma estrutura, com `email` ao invés de `telefone`)

## Performance

- Processamento em lotes de 1000 registros
- Commit a cada lote para evitar perda de dados
- Estimativa: ~100.000 registros/minuto

## Logs

O script exibe:
- Progresso a cada 10.000 linhas
- Inserções em lote
- Resumo final com estatísticas
- Erros encontrados

## Exemplo de Saída

```
Starting ETL process...
CSV file: contatos/500.000 contatos_Brasilia_DF - -- contatoszap.com --.csv
Database: data/backend/core.db
Processed 10000 rows...
Inserted 10000 phone records...
Processed 20000 rows...
...
============================================================
ETL Process Completed
============================================================
Total rows processed: 500000
Phone records inserted: 498523
Email records inserted: 0
Errors: 1477
============================================================

✓ ETL process completed successfully!
```

## Requisitos

- Python 3.6+
- Bibliotecas padrão (não requer instalação adicional):
  - csv
  - sqlite3
  - hashlib
  - uuid
  - unicodedata

## Notas

- O script não sobrescreve dados existentes, apenas insere novos registros
- Telefones e emails vazios são ignorados
- Registros sem nome são pulados
- Cada erro é logado individualmente sem interromper o processo
