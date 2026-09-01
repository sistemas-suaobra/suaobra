# Ciclo de busca e importação de novas obras

Este documento descreve **como obras entram no SuaObra**: da fonte pública (CAU/BR), passando pelos scrapers e pelo ETL, até o `core.db` que o app consulta no Obras+, nas campanhas e no CRM.

O app **não busca obras em tempo real**. Ele só lê um SQLite read-only (`core.db`) anexado ao PocketBase. Importar obras novas é um processo **offline**, feito à parte, e o resultado é um `core.db` atualizado no volume de produção.

O ciclo tem **dois passos obrigatórios**, nesta ordem:

1. **Puxar do CAU** (depois de achar RRTs que faltam na VPS): tipo, nome do proprietário, nome do profissional, tamanho, data de início, data de término, estado, cidade e endereço. O CAU **não traz telefone nem e-mail**.
2. **Triangular** com a planilha **`CONTATOS/contatos_suaobra/contatos_unificados.csv`**. A obra **só é elegível** se achar contato do **proprietário**, do **profissional**, ou **dos dois**. Sem contato de nenhum, não entra. **Se for elegível, importa na hora no `core.db` de produção.**

```
VPS (core.db de produção)
        │
        │  0. MAX(obra_number) + lacunas
        │     RRT maior = obra mais nova
        ▼
CAU/BR (página do RRT)
        │
        │  1. puxar: tipo, owner, professional,
        │     size, start_date, end_date,
        │     state, city, address
        ▼
  obra puxada (ainda sem contato)
        │
        │  2. triangular × contatos_unificados.csv
        │     (nome + cidade/estado)
        ▼
  contato do owner e/ou do professional
        │
        ├─ proprietário OU profissional OU ambos
        │     → ELEGÍVEL → INSERT/UPDATE no core.db de PRODUÇÃO
        └─ nenhum dos dois
              → NÃO elegível → não importa
```

---

## 1. Conceitos

### RRT

**RRT** = Registro de Responsabilidade Técnica, emitido pelo **CAU/BR** (Conselho de Arquitetura e Urbanismo do Brasil). Cada RRT tem um número sequencial (em geral **8 dígitos**).

Exemplo: **`15486006`**

```
https://acesso.caubr.gov.br/autenticidade/rrt?numero=15486006&retificador=
```

`15486007` é mais novo que `15486006`; `15485999` é mais antigo. Buscar “obras novas” = ir na **VPS de produção**, olhar quais números já estão no `core.db`, **preencher lacunas** no intervalo conhecido e, principalmente, avançar para **RRTs maiores que o máximo da base**.

### Dois bancos, papéis distintos

| Banco | Onde | Papel |
|---|---|---|
| `data.db` (PocketBase) | `data/main/data.db` | Auth, teams, CRM, campanhas, leads favoritados. **Não** contém o catálogo de obras. |
| `core.db` | volume `data/core.db` (Dokploy) ou `data/core/core.db` (layout legado) | Catálogo Obras+: obras + telefones + e-mails. Read-only no app. |

O backend faz `ATTACH DATABASE '.../core.db' AS 'core'` e consulta `core.core_obras_plus`.

### ID da obra

No scraper ZenRows o `id` é determinístico:

```
id = "obra_" + md5(número do RRT)
```

Isso é o `obra_id` usado em `lead`, `campanha_destinatarios` e `obra_note`. **Não mude a fórmula** ao reimportar: leads e campanhas já apontam para esse `id`.

### Datas de listagem

| Campo | Significado |
|---|---|
| `first_listing_date` | Data de registro no CAU (extraída do HTML) **ou**, se vier vazia da API, a data em que a obra entrou na nossa base. No **UPDATE**, este campo **não é sobrescrito**. |
| `last_listing_date` | Data da última vez que o scraper viu/atualizou o registro (`YYYY-MM-DD` de “hoje”). |
| `_sling_loaded_at` | Unix timestamp da última carga (legado do Sling; o scraper também preenche). |

O filtro “Mais recente / Mais antiga” do Obras+ usa `first_listing_date`.

---

## 2. Campos que precisamos puxar do CAU

Fonte: página pública, ex. RRT **`15486006`**:

```
https://acesso.caubr.gov.br/autenticidade/rrt?numero=15486006&retificador=
```

Parser: `go-concorrente/api_zenrows/internal/extract.go`.

**Lista do que puxamos do CAU** (é só isso para o ciclo de importação):

| O que puxar | Campo no `core_obras_plus` | Label no HTML do CAU |
|---|---|---|
| Tipo | `type` (e `activity`) | tabela Atividades |
| Nome do proprietário | `owner` | Nome/Razão Social |
| Nome do profissional | `professional` | Arquiteto(a) e Urbanista |
| Tamanho | `size` (+ `unidade`) | metragem na tabela Atividades |
| Data de início | `start_date` | Data de Início |
| Data de término | `end_date` | Previsão de Término |
| Estado | `state` | Cidade/UF |
| Cidade | `city` | Cidade/UF |
| Endereço | `address` | Tipo de Logradouro + Logradouro + Número/Ano + Complemento (+ bairro) |

O CAU **não traz telefone nem e-mail**. Depois desses campos, triangular com `contatos_unificados.csv` (§7). Só vira elegível — e **só então entra no banco de produção** — se houver contato do proprietário, do profissional, ou dos dois.

O `obra_number` (Número do RRT) também é obrigatório: é a chave do registro. `bairro` entra no endereço montado. `first_listing_date` (Data de Registro) é extraída para ordenação no Obras+, mas não entra na lista de cruzamento.

### 2.1 Pessoas (chaves do cruzamento)

São os nomes que vamos bater na planilha de contatos.

| Label no HTML do CAU | Campo no `core_obras_plus` | Quem é | Para que serve |
|---|---|---|---|
| **Nome/Razão Social:** | `owner` | Proprietário (pessoa física ou empresa) | Cruzar contato do **proprietário** |
| **Arquiteto(a) e Urbanista:** | `professional` | Profissional responsável pelo RRT | Cruzar contato do **profissional** |

Os dois nomes precisam ser gravados **como o CAU escreveu** (depois normalizamos acento/caixa para o match, ver §7.3). Elegível se achar contato de **proprietário, profissional ou ambos**; nesse caso **já importa no `core.db` de produção**.

### 2.2 Localização (desempate no cruzamento)

Um mesmo nome pode aparecer várias vezes na base de contatos (homônimos, várias cidades). Cidade e UF da obra servem para **preferir o contato da mesma praça**.

| Label no HTML do CAU | Campo | Uso |
|---|---|---|
| **Cidade/UF:** (dois `<b>`) | `city`, `state` | Filtro do Obras+ **e** ranking do contato (mesmo município da obra primeiro) |
| **Bairro:** | `bairro` | Filtro de bairro no app; entra também no endereço montado |
| **Tipo de Logradouro:** + **Logradouro:** + **Número/Ano:** + **Complemento:** | `address` | Endereço único concatenado |

Endereço montado:

```
{tipo}, {logradouro}, {número}, {complemento}, {bairro}, {cidade} - {UF}
```

### 2.3 Identificação e datas da obra

| Label no HTML do CAU | Campo | Uso |
|---|---|---|
| **Número do RRT:** | `obra_number` | Chave de upsert. Sem este campo a extração falha. `id` = `obra_` + md5(número) |
| **Data de Registro:** | `first_listing_date` | Quando o RRT entrou no CAU; ordenação “mais recente” no Obras+. Formato gravado: `YYYY-MM-DD` |
| **Data de Início:** | `start_date` | Início do prazo da etapa (§2.7) |
| **Previsão de Término:** | `end_date` | Fim do prazo (CAU); regra de qualidade (obra antiga); etapa |

Datas no HTML vêm `DD/MM/YYYY` e o parser inverte para `YYYY-MM-DD`.

### 2.4 Atividade e tamanho

Vêm da **primeira linha** da tabela “Atividades” no HTML.

| Origem | Campo | Uso |
|---|---|---|
| Texto da 1ª célula | `activity` | Tipo de serviço; entra no **tipo ART** da etapa (§2.7) |
| Mesmo texto (trecho após ` - `) | `type` | Classificação; filtro; regra “RRT MÍNIMO”; tipo ART |
| 2ª célula (`{n} / m²`) | `size`, `unidade` | Filtro de metragem **e** faixa de prazo da etapa |

### 2.5 O que o CAU **não** entrega

Não puxar (não existe na página do RRT):

- telefone do proprietário
- telefone do profissional
- e-mail do proprietário
- e-mail do profissional
- CPF/CNPJ, CREA/CAU do profissional, área do terreno, valor da obra

Esses contatos saem da **nossa planilha** (`core_obras_plus_phone` / `core_obras_plus_email`) no cruzamento (§7).

### 2.6 Checklist mínimo para aceitar um RRT

No scrape, precisa ter no mínimo o que o CAU entrega para o ciclo:

- tipo (`type` / `activity`)
- nome do proprietário (`owner`)
- nome do profissional (`professional`)
- tamanho (`size`)
- data de início e data de término (`start_date`, `end_date`)
- estado e cidade (`state`, `city`)
- endereço (`address`)
- número do RRT (`obra_number`)

Isso **ainda não torna a obra elegível**. Elegível só depois de triangular com `contatos_unificados.csv`: contato de proprietário **ou** profissional **ou** ambos. Se for elegível, **já importa no banco de produção** (§7).

### 2.7 Etapa da obra — INICIO, ESTRUTURA, ACABAMENTO, FINALIZADA

A etapa **não vem do CAU**. **A metragem (`size`) define o prazo; o prazo define a etapa.**

Cadeia:

```
metragem (m²) + tipo ART  →  prazo em meses  →  terço do tempo decorrido  →  ETAPA
```

Sem `size` não dá para escolher a coluna da tabela. Obra pequena e obra grande **do mesmo tipo**, começando no mesmo dia, ficam em etapas diferentes porque o prazo muda.

Os únicos valores válidos são estes quatro:

| Etapa | Quando |
|---|---|
| **INICIO** | Ainda não começou, ou está no **primeiro terço** do prazo |
| **ESTRUTURA** | **Segundo terço** do prazo (mais de 33% e até 66%) |
| **ACABAMENTO** | **Último terço** do prazo (mais de 66%), ainda dentro da data de fim |
| **FINALIZADA** | Hoje **depois** da data de término |

Não usar outros rótulos (ex. “em andamento”, “obra nova”).

#### Faixas de metragem (o que escolhe o prazo)

Campo do CAU: `size` em m² (§2.4). Três colunas:

| Faixa | Quando usar | Exemplo |
|---|---|---|
| **até 250 m²** | `size` ≤ 250 | casa 180 m² |
| **250 a 500 m²** | 250 < `size` ≤ 500 | 400 m² |
| **acima de 1.000 m²** | `size` > 1.000 | 1.500 m² |

A planilha não traz faixa **500–1.000 m²**. Na operação, tratar como a faixa vizinha combinada (em geral a de **acima de 1.000 m²** para não subestimar o prazo).

#### Prazo total — tabela por tipo ART × metragem

O **tipo ART** escolhe a linha; a **metragem** escolhe a coluna (quando o tipo varia por tamanho). Tipos com prazo fixo ignoram a metragem.

| Tipo ART | Prazo fixo (qualquer m²) | até 250 m² | 250 a 500 m² | acima de 1.000 m² |
|---|---|---|---|---|
| Execução de obra | — | **12 meses** | **18 meses** | **28 meses** |
| Avaliação | **3 meses** | | | |
| Execução de obra interiores | **10 meses** | | | |
| Acompanhamento de obra ou serviço técnico | **igual ao tempo da obra** (prazo da execução de obra **naquela metragem**) | | | |
| Execução de reforma de edificação | **10 meses** | | | |
| Execução de instalações de ventilação, exaustão e climatização | **1 mês** | | | |
| Projeto de estrutura metálica | **2 meses** | | | |
| Execução de estrutura metálica | **1 ano** | | | |
| Projeto arquitetônico | — | **6 meses** | **10 meses** | **18 meses** |

Tipo ART casa com `activity` / `type` puxados do CAU.

```
data_inicio  = start_date do CAU  (Data de Início)
prazo        = tabela[tipo ART][faixa de size]     ← a metragem entra aqui
data_fim     = data_inicio + prazo
```

#### Do prazo para a etapa (terços)

Com `data_inicio`, `data_fim` e a data de hoje:

```
se hoje > data_fim                         → FINALIZADA
se hoje < data_inicio                      → INICIO
senão:
  percentual = (hoje - data_inicio) / (data_fim - data_inicio)
  percentual ≤ 33%                         → INICIO
  percentual ≤ 66%                         → ESTRUTURA
  senão                                    → ACABAMENTO
```

No produto o corte em terços já está em `calculateObraStage` (`frontend/src/components/obras-plus/ObraPlusPage.tsx`) e no export (`scripts/export_obras_60d.py`). A tabela tipo ART × **metragem** é a regra de negócio do prazo.

#### Exemplo — mesma data, metragem diferente, etapa diferente

Tipo **execução de obra**, início `2025-01-01`, hoje `2025-10-01` (9 meses decorridos):

| `size` | Faixa | Prazo | % decorrido | Etapa |
|---|---|---|---|---|
| **200 m²** | até 250 | 12 meses | 75% | **ACABAMENTO** |
| **400 m²** | 250 a 500 | 18 meses | 50% | **ESTRUTURA** |
| **1.500 m²** | acima de 1.000 | 28 meses | 32% | **INICIO** |

É por isso que a metragem define a etapa: ela muda o denominador (prazo), não só um rótulo.

---

## 3. Evolução das ferramentas (o que existe no repo)

O pipeline foi construído em camadas. Várias estão gitignored ou incompletas; a que vale para operação hoje está marcada.

| Ferramenta | Pasta / arquivo | Status |
|---|---|---|
| dbt → CSV.gz → Sling → SQLite | `store/sling/build.sqlite.core.yaml`, `store/coredb.go`, `scripts/build.asset.sh` | Pipeline **histórico de rebuild completo**. `cmd/etl/` está no `.gitignore`. |
| Playwright (Python) | `automacao-caubr/automacao_caubr.py` (histórico git) | Primeira automação: abre a página do CAU e imprime o texto. |
| Puppeteer + browser ZenRows | `node-scraper-rrt/` (gitignored) | Scrape via WebSocket ZenRows. |
| API CAUBR no Cloud Run | `go-concorrente/main.go` | Go concorrente chama `https://caubr-api-*.run.app/caubr?numero=`. Aplica regras de negócio. |
| ZenRows HTTP + goquery | `go-concorrente/api_zenrows/` | **Pipeline atual de captura.** Fetch HTML → parse → INSERT/UPDATE no `core.db`. |
| `go-concorrente/main_zenrows.go` | mesmo módulo, arquivo paralelo | **Incompleto** (TODO de parse HTML; retorna `nil`). Não usar. |

Também gitignored (legado, não versionado): `caubr-api/`, `caubr-api-v2/`, `python_scraper_api/`, `cno/`, `process_cno_data.py`, `contatos/`.

---

## 4. Pipeline A — Captura incremental de RRTs (operação do dia a dia)

Este é o ciclo para **buscar obras novas** e gravá-las num `core.db` existente.

### 4.1 Descobrir quais RRTs faltam (sempre na VPS)

A fonte da verdade é o `core.db` **de produção**, não o banco local. Sem SSH na VPS você não sabe o que já entrou na plataforma.

Dois trabalhos distintos:

| Trabalho | O que é | Prioridade |
|---|---|---|
| **Fronteira (obras novas)** | RRTs **maiores** que o `MAX(obra_number)` da base | **Sempre.** Número maior = obra mais recente no CAU |
| **Lacunas** | Números **no meio** do intervalo que nunca gravamos (falha, captcha, lote pulado) | Entrar nesses buracos para não deixar obra antiga de fora |

Não dá para “adivinhar” o último RRT emitido pelo CAU: o site não publica um ranking. O jeito operacional é **consultar o que já temos na VPS** e daí:

1. anotar o máximo;
2. listar buracos (lacunas);
3. scrapar `máximo + 1` em diante até a página passar a responder `no_data` com frequência (ainda não emitido, ou sem página).

#### Acessar o `core.db` na VPS

SSH no servidor Dokploy. O arquivo costuma estar no volume bind:

```
/etc/dokploy/compose/<projeto>/files/suaobra/data/core.db
```

(layout legado: `.../files/suaobra/data/core/core.db`.) Se não achar:

```bash
docker ps --format '{{.Names}}'
docker inspect <container-backend> --format '{{range .Mounts}}{{.Source}} -> {{.Destination}}{{println}}{{end}}'
find /etc/dokploy -name core.db -type f 2>/dev/null
```

Consultar **somente leitura**. Não rode `VACUUM` nem writes pesados no arquivo vivo. Para queries grandes, faça snapshot:

```bash
sqlite3 /caminho/files/suaobra/data/core.db ".backup '/tmp/core-snapshot.db'"
sqlite3 /tmp/core-snapshot.db
```

No arquivo vivo (consultas rápidas):

```bash
sqlite3 /caminho/files/suaobra/data/core.db
```

#### Fronteira — o RRT mais recente que ainda não está na plataforma

```sql
SELECT
  MIN(CAST(obra_number AS INTEGER)) AS rrt_mais_antigo,
  MAX(CAST(obra_number AS INTEGER)) AS rrt_mais_recente,
  COUNT(*) AS obras_na_base
FROM core_obras_plus
WHERE obra_number GLOB '[0-9]*';
```

- `rrt_mais_recente` = maior número que **já** temos.
- O lote de obras novas começa em **`rrt_mais_recente + 1`**.
- Exemplo: se o `MAX` na VPS for `15486006`, a fronteira começa em **`15486007`**. Um bloco típico: inicial `15486007`, final `15487000`.
- O teto é empírico: suba em blocos (alguns milhares) até o scraper devolver `no_data` em sequência — o CAU ainda não emitiu aquele número. Dias depois o teto sobe; por isso a rotina é **reconsultar o MAX na VPS** a cada importação, não reusar o intervalo da vez passada.

Quanto maior o RRT, mais nova a obra. O critério para “o que scrapar de novo” é o **número que não está em `obra_number`**, não uma data no app.

#### Lacunas — RRTs no meio que nunca entraram

`COUNT(*)` muito menor que `(MAX - MIN + 1)` é normal: muitos números não existem, são RRT MÍNIMO ou falharam no scrape. Ainda assim dá para achar **buracos consecutivos** (lotes que pulamos):

```sql
-- Maiores lacunas (números em falta entre dois RRTs que já temos).
-- next_n - n > 1  →  há pelo menos um número não gravado no meio.
SELECT
  n + 1           AS lacuna_inicio,
  next_n - 1      AS lacuna_fim,
  next_n - n - 1  AS qtd_faltando
FROM (
  SELECT
    CAST(obra_number AS INTEGER) AS n,
    LEAD(CAST(obra_number AS INTEGER)) OVER (
      ORDER BY CAST(obra_number AS INTEGER)
    ) AS next_n
  FROM core_obras_plus
  WHERE obra_number GLOB '[0-9]*'
)
WHERE next_n - n > 1
ORDER BY n DESC          -- lacunas mais perto da fronteira primeiro
LIMIT 50;
```

Como ler cada linha:

- `lacuna_inicio` … `lacuna_fim` = intervalo para mandar no scraper (modo 1 ou lista).
- `qtd_faltando` enorme (dezenas de milhares) muitas vezes **não** é “esquecemos um lote”: é faixa em que o CAU não tem RRT válido. Vale **amostrar** (ex. 100 números) antes de varrer a lacuna inteira.
- Lacunas **perto do MAX** (`ORDER BY n DESC`) são as mais úteis: obras relativamente novas que falharam.

Conferir se um número específico já está na plataforma:

```sql
SELECT id, obra_number, city, state, owner, professional, first_listing_date
FROM core_obras_plus
WHERE obra_number = '15486006';
```

Vazio = o RRT `15486006` não está na base → entra na lista de scrape.

#### O que mandar para o scraper

Depois das queries na VPS, a fila fica assim:

1. **Obrigatório:** `MAX+1` até um teto um pouco acima (RRT mais novo que ainda não existe na plataforma).
2. **Lacunas escolhidas** (sobretudo as próximas da fronteira) e o `failed_rrts_zenrows.log` de rodadas anteriores.

O scraper pede isso na CLI:

```
1 - Intervalo (número inicial e final)     ← fronteira ou uma lacuna
2 - Lista personalizada (vírgula ou .txt)  ← falhas + números pontuais
```

Exemplo modo 1 (fronteira a partir de `15486007`):

```
Digite o número inicial da RRT: 15486007
Digite o número final da RRT:   15487000
```

Exemplo modo 2 (lacunas pontuais):

```
15486006,15486014,15486082
```

RRTs que falharem depois de todos os retries vão para `failed_rrts_zenrows.log` (um número por linha) e voltam no modo 2 — são lacunas conhecidas.

### 4.2 Rodar o scraper ZenRows

Diretório: `go-concorrente/api_zenrows/`.

```bash
cd go-concorrente/api_zenrows
go run .
# escolhe modo 1, informa início e fim
```

O binário espera um `core.db` **no diretório de trabalho** (`DB_PATH = "core.db"`). Copie o snapshot de produção para lá **antes** de rodar, senão você cria/atualiza um banco local desconectado do catálogo real.

Fluxo interno de **cada** RRT:

```
RRT N
  │
  ├─ GET ZenRows
  │    url alvo = acesso.caubr.gov.br/autenticidade/rrt?numero=N
  │    js_render=true
  │    wait_for=.title-acction
  │    premium_proxy=true
  │    proxy_country=br
  │    timeout HTTP = 200s
  │
  ├─ Classificar HTML
  │    título da home do CAU/BR     → captcha_block  (retry com backoff)
  │    "Nenhum dado encontrado"     → no_data        (ignora, sem retry)
  │    "Não foi possível localizar" → no_data
  │    HTTP ≠ 200                   → erro
  │
  ├─ Limpar HTML (remove <svg> e <img>)
  ├─ Extrair campos com goquery (labels <span> + <b>)
  ├─ id = obra_ + md5(número do RRT)
  └─ INSERT ou UPDATE em core_obras_plus (chave = obra_number)
```

Concorrência:

- **10 workers**
- **3 s** de pausa após sucesso
- backoff em erro / timeout / captcha (até 4 tentativas)
- circuit breaker: 3 timeouts consecutivos globais abortam o lote
- `PRAGMA journal_mode=WAL` e `busy_timeout=30000` no SQLite

A chave da ZenRows está **hardcoded** no código (`ZENROWS_API_KEY`). Não commitar chaves novas; o ideal é passar por variável de ambiente.

### 4.3 Extração (implementação)

Lista canônica dos campos: **§2**. O parser lê labels `<span>` + `<b>` no HTML.

Logs úteis:

- `html_rrt.log` — HTML (sucesso e falha)
- `extract_result.log` — JSON extraído por RRT
- `failed_rrts_zenrows.log` — números para reprocessar

Depois do scrape, a obra **ainda não é elegível**. Falta triangular com `contatos_unificados.csv`. Só o que passar nessa regra entra no `core.db` de produção (§7).

### 4.4 INSERT vs UPDATE

Chave de existência: `obra_number`.

- **Não existe** → INSERT com todos os campos, inclusive `first_listing_date` vinda do CAU e `last_listing_date = hoje`.
- **Já existe** → UPDATE de owner, professional, endereço, datas de obra, activity/type/size, `last_listing_date` e `_sling_loaded_at`. **`first_listing_date` permanece a original.**

Isso permite re-varrer um intervalo para corrigir dados sem “rejuvenescer” a obra na ordenação do Obras+.

### 4.5 Caminho alternativo: API CAUBR (Cloud Run)

`go-concorrente/main.go` chama:

```
GET {CAUBR_API_URL}/health
GET {CAUBR_API_URL}/caubr?numero={RRT}
```

Padrão: `https://caubr-api-rem6jzjgfa-uc.a.run.app` (override: `CAUBR_API_URL`).

Este caminho **aplica regras de negócio** que o ZenRows atual **não aplica** (ver §5). Use-o se a API Cloud Run estiver no ar; caso contrário, o caminho operacional é o ZenRows.

---

## 5. Regras de negócio (filtro de qualidade)

Implementadas em `go-concorrente/main.go` → `isValidForProcessing`. **Não estão** no `api_zenrows` atual — se rodar só o ZenRows, RRTs mínimos e obras antigas entram no banco. Filtrar depois no SQL, ou portar as regras para o `api_zenrows`.

Um RRT **entra** se:

1. `type` **não** é `"RRT MÍNIMO"`
2. `end_date` ≥ **2023-09-01** (dois anos antes de 2025-09-01; constante hardcoded)
3. `owner`, `address` e `city` preenchidos
4. Datas no formato `YYYY-MM-DD` (validação extra em `validateObraData`)

Um RRT é **ignorado** (não é erro) se falhar essas regras. Erro = falha HTTP/API/banco; vai para o log de falhas.

Limpeza posterior de obras velhas: `scripts/clean_databases.py` remove do `core.db` obras com `end_date < 2023-01-01` e, se confirmado, leads relacionados no `data.db`.

---

## 6. Pipeline B — Rebuild completo via CSV + Sling (legado / staging)

Usado quando se reconstrói o `core.db` do zero a partir de CSVs (dbt ou dump).

### 6.1 Origem dos CSVs

Três arquivos (gzip):

- `core_obras_plus.csv.gz`
- `core_obras_plus_phone.csv.gz`
- `core_obras_plus_email.csv.gz`

Em staging antigo, vinham do R2 (`R2/sua-obra/staging/*.csv.gz`) e/ou de `DBT_TARGET_FOLDER`.

### 6.2 Sling

Receita: `store/sling/build.sqlite.core.yaml`.

- Source: `file://./data/core/core_obras_plus.csv` (e phone/email)
- Target: `sqlite://./data/core/core.db`
- Mode: **full-refresh** (apaga e recarrega a tabela)
- Tipos forçados: `obra_number` string, `size` integer, datas timestamp/date, flags `has_*` bool
- Depois da carga, cria índices (id, city+state, bairro, owner, professional, nome+telefone, uf+cidade)

### 6.3 `scripts/build.asset.sh`

```
1. go run cmd/etl/main.go          # pasta cmd/ está no .gitignore
2. baixa sling CLI (darwin/arm64)
3. sling run -r store/sling/build.sqlite.core.yaml
4. mv ./data/core/core.db .
5. sqlite3 ./core.db "vacuum;"
6. gzip -f ./core.db               # gera core.db.gz
```

### 6.4 Staging → produção (scripts antigos)

`scripts/deploy.staging.sh`:

1. Baixa `*.csv.gz` do R2
2. Roda `build.asset.sh`
3. Sobe `suaobra-app` e `core.db.gz` para `R2/sua-obra/staging/`
4. No servidor, descompacta e faz swap do `core.db`

`scripts/deploy.production.sh`:

1. Copia o binário staging → production no R2
2. No servidor de prod, baixa `core.db.gz` **do bucket de staging**, descompacta e substitui o `core.db` do compose

Hoje o Dokploy usa volume bind (`../files/suaobra/data:/app/data`). O arquivo vivo de produção fica em:

```
.../files/suaobra/data/core.db
```

(layout legado: `.../files/suaobra/data/core/core.db`). Trocar o arquivo **com o app ligado** exige snapshot SQLite (`sqlite3 ... ".backup"`), não `cp` do `.db` solto (há WAL). Depois, **reiniciar o backend** para recarregar `ObrasCities`.

---

## 7. Triangular com `contatos_unificados.csv` e importar no banco de produção

Depois de puxar do CAU tipo, nomes, tamanho, datas, UF, cidade e endereço, **triangular** com a planilha unificada:

```
CONTATOS/contatos_suaobra/contatos_unificados.csv
```

Colunas: `nome`, `cidade`, `estado`, `telefone`, `email` (~14 milhões de linhas; união das planilhas da pasta `CONTATOS/contatos_suaobra`).

Regra de elegibilidade — a obra **só entra no banco de produção** se:

| Cruzamento | Elegível? | O que fazer |
|---|---|---|
| Contato do **proprietário e do profissional** | Sim | **Importar já** no `core.db` de produção |
| Contato **só do proprietário** | Sim | **Importar já** no `core.db` de produção |
| Contato **só do profissional** | Sim | **Importar já** no `core.db` de produção |
| **Nenhum** dos dois | **Não** | Não importa. Obra não vira elegível. |

“Contato” = telefone ou e-mail na planilha unificada. Match pelo **nome** (`owner` / `professional`); cidade e estado da obra desempatam homônimos.

Não gravar no `core.db` de prod a obra cru (só CAU) e enriquecer depois: **triangular primeiro; se elegível, importa**. O destino é o mesmo arquivo da VPS (`.../files/suaobra/data/core.db`), tabelas `core_obras_plus` + phone/email.

```
CAU: owner + professional
        │
        ▼
contatos_unificados.csv  (nome → telefone / e-mail)
        │
        ├─ proprietário OU profissional OU ambos
        │     → ELEGÍVEL → INSERT/UPDATE produção
        └─ nenhum
              → NÃO ELEGÍVEL → descarta (não vai pro banco de prod)
```

```
                    core_obras_plus (veio do CAU)
                    ┌─────────────────────────────────┐
                    │ owner        = "JOAO DA SILVA"  │
                    │ professional = "MARIA ARQUITETA"│
                    │ city / state = CAMPINAS / SP    │
                    └──────────────┬──────────────────┘
                                   │
              match exato de nome  │  (preferir mesmo city+UF)
                    ┌──────────────┴──────────────┐
                    ▼                             ▼
     core_obras_plus_phone            core_obras_plus_email
     (nossa planilha)                 (nossa planilha)
                    │                             │
                    └──────────────┬──────────────┘
                                   ▼
                    flags:
                      has_owner_phone / has_owner_email
                      has_professional_phone / has_professional_email
                    elegível se:
                      (has_owner_phone OU has_owner_email)
                      OU
                      (has_professional_phone OU has_professional_email)
```

| Pessoa no CAU | Campo na obra | Contato na planilha | Flag se achar |
|---|---|---|---|
| Proprietário | `owner` | telefone onde `nome = owner` | `has_owner_phone` |
| Proprietário | `owner` | e-mail onde `nome = owner` | `has_owner_email` |
| Profissional | `professional` | telefone onde `nome = professional` | `has_professional_phone` |
| Profissional | `professional` | e-mail onde `nome = professional` | `has_professional_email` |

Não há `obra_id` na planilha. O vínculo é **só o nome**.

### 7.1 De onde vem a planilha unificada

Não é o CAU. É a união das planilhas em `CONTATOS/contatos_suaobra/` (telefones_parte_*, emails_parte_*, PR, etc.) no arquivo:

**`CONTATOS/contatos_suaobra/contatos_unificados.csv`**

Esse CSV é a fonte para triangular. No `core.db` de produção os matches viram linhas em `core_obras_plus_phone` / `core_obras_plus_email` (e as flags `has_*` na obra).

Carga legada (ainda válida se for alimentar o SQLite a partir do CSV): `scripts/etl_contacts_csv.py`.

Colunas do `contatos_unificados.csv`:

| Coluna | Papel |
|---|---|
| `nome` | match com `owner` / `professional` |
| `telefone` / `email` | o que o CAU não traz |
| `cidade`, `estado` | desempate (mesma praça da obra primeiro) |

Nomes, cidade e estado no CSV já estão em MAIÚSCULAS sem acento; telefone só dígitos; e-mail minúsculo.

### 7.2 Como o match é feito (regra de ranking)

Usada no app (`core_obras_plus_phone.sql` / `_email.sql`), no export enriquecido e nas campanhas.

1. Filtrar contatos com `nome` **igual** ao `owner` (ou ao `professional`).
2. **Triangular com a praça da obra:** contatos com o mesmo `uf` + `cidade` da obra (`state` + `city`) ficam em primeiro.
3. Homônimos em outras cidades entram depois (ainda servem se não houver contato local).
4. Deduplicar o mesmo telefone/e-mail (`row_number` por valor).
5. Limites ao entregar para o produto: **até 3 telefones** e **até 6 e-mails** por pessoa.

É isso que “triangular” significa aqui: **nome do CAU × nome do contato × cidade/UF da obra**.

Campanhas (`campanha_service.go`) repetem a mesma regra, teto de 3 telefones por destinatário (`OWNER` ou `PROFISSIONAL`).

### 7.3 Normalizar antes de cruzar

Se o CAU gravou `São Paulo` e a base de contatos tem `SAO PAULO`, o match **quebra**.

Depois do scrape, rodar `scripts/update_bairro_city.py` nas obras (remove acentos de `bairro` e `city`). O nome do owner/profissional também precisa estar no mesmo padrão da base de contatos (maiúsculo, sem acento). Se o scraper gravar com acento, o cruzamento não acha o telefone mesmo o nome “sendo o mesmo”.

Cidade no Obras+ é o nome em maiúsculas **sem acento** (ex.: `CAMPOS DO JORDAO`), validado contra `store.ObrasCities`. O frontend mapeia `SP-CAMPOS_DO_JORDAO` → `CAMPOS DO JORDAO` em `frontend/src/store/cities.ts`. Grafia diferente → **404 “cidade inválida”**.

### 7.4 Recalcular as flags na obra

Depois de garantir que phone/email estão carregados, `scripts/update_obras_contact_phone_email.py`:

1. Carrega o conjunto de `nome` em `core_obras_plus_phone` e em `core_obras_plus_email`.
2. Para cada obra com algum `has_*` zerado, testa:
   - `owner` está no set de telefones? → `has_owner_phone`
   - `owner` está no set de e-mails? → `has_owner_email`
   - idem para `professional`
3. Grava as quatro flags na linha da obra.

**Elegibilidade:** proprietário **ou** profissional **ou** ambos com telefone/e-mail. Se as quatro flags forem 0, **não vira elegível e não importa no banco de produção**.

Obra elegível → **INSERT/UPDATE já no `core.db` de produção** (`core_obras_plus` + contatos casados). Snapshot/WAL: não `cp` o arquivo vivo; gravar no SQLite anexado ou fazer swap com `.backup` e reiniciar o backend.

Esse script **não copia** o telefone para `core_obras_plus`. Só marca se existe match. O número em si continua na planilha e é resolvido na hora da API (`/query/obras-plus-contacts`).

### 7.5 Ordem obrigatória neste passo

```
1. Obras puxadas do CAU
2. Normalizar city/bairro/nomes
3. Triangular com contatos_unificados.csv
4. Elegível = proprietário OU profissional OU ambos
5. Se elegível → importar já no core.db de PRODUÇÃO
```

Inverter 3 e 4 faz a obra ficar marcada “sem telefone” até rodar o recálculo de novo.

---

## 8. Como o app consome o `core.db`

Boot (`suaobra-app.go` → `store.SetPocketBaseDB`):

1. Abre `data.db`
2. `AttachCoreDb()` procura o arquivo (mínimo 4 KB) em:
   - `CORE_DB_PATH` (se definido)
   - `{dataDir}/../core.db`
   - `{dataDir}/../core/core.db`
   - `/app/data/core.db`
   - `/app/data/core/core.db`
   - `./data/core/core.db`
3. `ATTACH` tenta caminho absoluto e depois relativo (`../core.db`)
4. Cria índices se faltarem (`city`, `city+size`, `city+first_listing_date+start_date`, nome em phone/email)
5. `loadCities()` preenche `store.ObrasCities`

Antes de cada query Obras+, `EnsureCoreReady()` reanexa se o SQLite tiver soltado o attach.

Rotas que leem o catálogo:

| Rota | Uso |
|---|---|
| `GET /query/obras-plus` | Listagem paginada (Obras+) |
| `GET /query/leads-plus` | Mesmo SQL (modal de campanha) |
| `GET /query/obras-plus-neighborhood` | Bairros da cidade |
| `GET /query/obras-plus-contacts` | Telefones/e-mails por nome |
| `GET /query/obras-plus-export` | CSV enriquecido |
| `GET /query/cities` | Lista `ObrasCities` |

Filtros relevantes da listagem: cidade (obrigatória), bairro, tamanho, status (em andamento, com telefone/e-mail, visitada, favorita…), datas de início/fim da obra.

Dev (`ENV=development`): `store/coredb.go` `importDevData()` copia os três `csv.gz` de `DBT_TARGET_FOLDER` para `./data/core/` e roda Sling se os arquivos forem mais novos que `store/dev_data_timestamp`.

---

## 9. Schema mínimo de `core.db`

### `core_obras_plus`

```
id, obra_number, owner, professional,
address, bairro, city, state,
start_date, end_date, activity, type, size, unidade,
has_owner_phone, has_owner_email, has_professional_phone, has_professional_email,
first_listing_date, last_listing_date, _sling_loaded_at
```

O scraper **não preenche** os `has_*` nem grava telefone/e-mail na obra. Isso é a triangulação do §7. Obras recém-inseridas ficam com esses campos nulos/0 até o recálculo.

### `core_obras_plus_phone`

```
contact_id, person_id, company_id, nome, telefone,
cidade, uf, poder_aquisitivo, source, _sling_loaded_at
```

### `core_obras_plus_email`

Igual, com `email` no lugar de `telefone`.

Índices que o app garante no attach (além dos do Sling):

- `idx_core_obras_plus_city (city)`
- `idx_core_obras_plus_city_size (city, size)`
- `idx_core_obras_plus_city_listing (city, first_listing_date, start_date)`
- `idx_core_obras_plus_phone_nome (nome)`
- `idx_core_obras_plus_email_nome (nome)`

Sem índice de `city`, o Obras+ faz full scan (o `core.db` de produção tem da ordem de **vários GB**).

---

## 10. Receita operacional — “tem obra nova, atualiza a base”

Passo a passo recomendado para incrementar o catálogo de produção:

0. **SSH na VPS** e descobrir o que falta (§4.1): `MAX(obra_number)` (fronteira — obras mais novas) e lacunas no meio. Sem isso o scrape não sabe de onde começar.

1. **Snapshot consistente** do `core.db` no volume Dokploy (app ligado, WAL ativo):

   ```bash
   sqlite3 /caminho/files/suaobra/data/core.db ".backup '/tmp/core-snapshot.db'"
   ```

   Não use `cp core.db` com o processo escrito.

2. Baixar o snapshot para a máquina que vai scrapar (o scrape **grava nesse arquivo**, não no da VPS direto).

3. Copiar o snapshot para `go-concorrente/api_zenrows/core.db`.

4. Rodar o scraper: primeiro a **fronteira** (`MAX+1` … teto), depois lacunas / `failed_rrts_zenrows.log` no modo lista.

5. (Recomendado) Filtrar lixo que o ZenRows não descarta:

   ```sql
   DELETE FROM core_obras_plus WHERE type = 'RRT MÍNIMO';
   DELETE FROM core_obras_plus WHERE end_date IS NOT NULL AND end_date < '2023-09-01';
   ```

6. **Normalizar** cidade/bairro (e conferir nomes sem acento) — pré-requisito do cruzamento:

   ```bash
   python3 scripts/update_bairro_city.py
   ```

   Conferir se cidades novas estão em MAIÚSCULO sem acento e existem em `frontend/src/store/cities.ts`. Se a cidade não estiver no `cityMap`, o time não consegue selecioná-la no frontend mesmo com dados no core.

7. **Triangular com `contatos_unificados.csv`** e **importar no banco de produção**:

   Arquivo: `CONTATOS/contatos_suaobra/contatos_unificados.csv`.

   - Casar `owner` / `professional` da obra com `nome` (preferir mesma cidade/estado).
   - **Elegível** se tiver contato do proprietário **ou** do profissional **ou** dos dois.
   - **Não elegível** se não achar nenhum → não grava em produção.
   - **Se elegível → já importa** no `core.db` da VPS (`.../files/suaobra/data/core.db`): obra + telefone/e-mail casados + flags `has_*`.
   - Reiniciar o backend depois do import para recarregar `ObrasCities`.

8. Validar:

    ```
    GET /query/cities
    GET /query/leads-plus?city=NOME DA CIDADE&itemsPerPage=20
    ```

    Esperado: HTTP 200, obras novas na ordenação por `first_listing_date`. Logs do backend: `core.db anexado com sucesso` e `Loaded N cities from core.db`.

Times só enxergam cidades listadas em `team.cities` (JSON no PocketBase). Incluir a chave tipo `SP-CIDADE_NOVA` no time depois de garantir o nome no `cityMap` e no core.

---

## 11. Dev local e mocks

- Colocar um `core.db` em `./data/core/core.db` (ou `CORE_DB_PATH`) e `go run suaobra-app.go serve`.
- `ENV=development` + `DBT_TARGET_FOLDER` dispara Sling se os CSV.gz forem mais novos que o timestamp.
- `scripts/generate_mock_core.py` gera um `core.db` fake (20 obras) a partir do schema de um banco de origem — útil sem o dump de produção.
- Não commitar `core.db`, `*.csv.gz`, `data/`.

---

## 12. Troubleshooting

| Sintoma | Causa típica |
|---|---|
| Não sei qual RRT scrapar | Ir na VPS, rodar o `MAX` + query de lacunas (§4.1). Banco local não é a plataforma. |
| Scrape “não acha obra nova” | Intervalo abaixo do MAX (já estava na base) ou teto curto demais. Sempre começar em `MAX+1`. |
| Obras+ vazio / restart loop | `core.db` ausente, &lt;4 KB, ou ATTACH com path absoluto inválido. Ver log `tentando attach core.db`. |
| Obra nova sem telefone no card / não elegível | Cruzamento não achou contato de **nenhum** dos dois (owner e professional); ou nome do CAU ≠ `nome` na planilha (acento, espaço, razão social). |
| Flag `has_owner_phone=1` mas a API não devolve número | Nome bateu no set, mas o SQL de resolução filtra/deduplica diferente — conferir `nome` + `uf`/`cidade`. |
| Captcha em massa no scraper | ZenRows bloqueado; o código já faz retry em `captcha_block`. Reduzir workers ou pausar. |
| `Database busy` | Muitos writers no mesmo SQLite; o scraper já usa WAL + busy_timeout. Rodar com 1 processo. |
| Cidade nova no SQL mas não no dropdown | Falta no `cityMap` e/ou em `team.cities`; backend não reiniciado. |
| `go-concorrente/main_zenrows.go` “não grava nada” | Arquivo incompleto (parse HTML TODO). Use `api_zenrows/`. |
| Rebuild Sling apagou obras incrementais | Mode `full-refresh`: o YAML **substitui** as tabelas pelos CSV. Não misturar rebuild full com captura incremental sem unir os CSVs antes. |

---

## 13. Mapa de arquivos

| Arquivo | Papel |
|---|---|
| `go-concorrente/api_zenrows/main.go` | Orquestra workers ZenRows |
| `go-concorrente/api_zenrows/internal/fetch.go` | HTTP ZenRows + detecção captcha/sem dados |
| `go-concorrente/api_zenrows/internal/extract.go` | Parse HTML → campos do §2 |
| `go-concorrente/api_zenrows/internal/db.go` | INSERT/UPDATE `core_obras_plus` (ainda sem contato) |
| `scripts/etl_contacts_csv.py` | Carrega a base de contatos (phone/email) |
| `scripts/update_obras_contact_phone_email.py` | Triangula nomes CAU × base e grava `has_*` |
| `scripts/update_bairro_city.py` | Remove acentos de cidade/bairro (pré-match) |
| `server/templates/core/core_obras_plus_phone.sql` | Resolução de telefone no app (nome + cidade/UF) |
| `server/templates/core/core_obras_plus_email.sql` | Resolução de e-mail no app |
| `go-concorrente/main.go` | Caminho API Cloud Run + regras de negócio |
| `go-concorrente/README.md` | README do processador de RRTs (API) |
| `store/db.go` | ATTACH, índices, `loadCities` |
| `store/coredb.go` | Import Sling em development |
| `store/sling/build.sqlite.core.yaml` | Receita Sling CSV → SQLite |
| `store/load.core.data.sql` | Cópia core → main + índices (legado) |
| `scripts/build.asset.sh` | ETL + Sling + gzip do `core.db` |
| `scripts/deploy.staging.sh` / `deploy.production.sh` | R2 + swap do `core.db` no servidor |
| `scripts/clean_databases.py` | Purge de obras antigas |
| `scripts/generate_mock_core.py` | `core.db` fake para dev |
| `server/routes_obas_plus.go` | API Obras+ / validateCity |
| `frontend/src/components/obras-plus/ObraPlusPage.tsx` | `calculateObraStage` — INICIO / ESTRUTURA / ACABAMENTO / FINALIZADA |
| `scripts/export_obras_60d.py` | Mesma regra de etapa no Excel |
| `frontend/src/store/cities.ts` | `UF-CIDADE` → nome usado na query |
| `CONTATOS/contatos_suaobra/contatos_unificados.csv` | Planilha unificada para triangular (nome, cidade, estado, telefone, email) |

---

## 14. Resumo

1. **Na VPS**, ver o que já está no `core.db`: preencher **lacunas** e buscar RRTs **maiores que o MAX** (número maior = obra mais nova).
2. **Puxar do CAU:** tipo, nome do proprietário, nome do profissional, tamanho, data de início, data de término, estado, cidade e endereço. O CAU **não traz** telefone nem e-mail.
3. **Calcular a etapa** (§2.7): a **metragem** escolhe o prazo; sai **INICIO**, **ESTRUTURA**, **ACABAMENTO** ou **FINALIZADA**.
4. **Triangular** com `CONTATOS/contatos_suaobra/contatos_unificados.csv`. Elegível se tiver contato do **proprietário ou do profissional ou de ambos**. Sem nenhum, não entra.
5. **Se elegível, importar já no `core.db` de produção** e reiniciar o backend.
