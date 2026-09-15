-- Recorte: first_listing_date entre 2026-07-17 e 2026-09-15 (60 dias a partir de 15/09/2026).
-- Somente leitura no core.db (TEMP tables + SELECT). Não altera produção.

PRAGMA temp_store = FILE;
PRAGMA cache_size = -100000;

CREATE TEMP TABLE obras AS
SELECT
  id AS obra_id,
  obra_number,
  owner,
  professional,
  address,
  bairro,
  city,
  state,
  start_date,
  end_date,
  type,
  activity,
  size,
  unidade,
  first_listing_date
FROM core_obras_plus
WHERE first_listing_date >= '2026-07-17'
  AND first_listing_date <= '2026-09-15';

CREATE INDEX idx_tmp_obras_owner ON obras(owner);
CREATE INDEX idx_tmp_obras_prof ON obras(professional);
CREATE INDEX idx_tmp_obras_city ON obras(state, city);

CREATE TEMP TABLE nomes AS
SELECT owner AS nome FROM obras WHERE owner IS NOT NULL AND trim(owner) != ''
UNION
SELECT professional FROM obras WHERE professional IS NOT NULL AND trim(professional) != '';

CREATE INDEX idx_tmp_nomes ON nomes(nome);

CREATE TEMP TABLE phones_prep AS
SELECT
  p.uf,
  p.cidade,
  p.nome,
  CASE WHEN EXISTS (SELECT 1 FROM obras o WHERE o.professional = p.nome) THEN p.telefone END AS professional_telefone,
  CASE WHEN EXISTS (SELECT 1 FROM obras o WHERE o.owner = p.nome) THEN p.telefone END AS owner_telefone,
  p.poder_aquisitivo,
  CASE WHEN EXISTS (SELECT 1 FROM obras o WHERE o.state = p.uf AND o.city = p.cidade) THEN 1 ELSE 2 END AS rank
FROM core_obras_plus_phone p
INNER JOIN nomes n ON n.nome = p.nome;

CREATE TEMP TABLE phones AS
SELECT
  uf,
  cidade,
  nome,
  professional_telefone,
  owner_telefone,
  rank,
  row_number() OVER (PARTITION BY nome ORDER BY rank, professional_telefone DESC, owner_telefone DESC, poder_aquisitivo DESC, uf, cidade) AS row_num
FROM phones_prep;

CREATE TEMP TABLE phones_agg AS
SELECT
  nome,
  max(CASE WHEN row_num = 1 THEN owner_telefone END) AS owner_first_telefone,
  max(CASE WHEN row_num = 1 THEN professional_telefone END) AS professional_first_telefone,
  group_concat(DISTINCT CASE WHEN row_num > 1 THEN owner_telefone END) AS owner_telefones,
  group_concat(DISTINCT CASE WHEN row_num > 1 THEN professional_telefone END) AS professional_telefones
FROM phones
WHERE row_num <= 10
GROUP BY nome;

CREATE TEMP TABLE emails_prep AS
SELECT
  e.uf,
  e.cidade,
  e.nome,
  CASE WHEN EXISTS (SELECT 1 FROM obras o WHERE o.professional = e.nome) THEN e.email END AS professional_email,
  CASE WHEN EXISTS (SELECT 1 FROM obras o WHERE o.owner = e.nome) THEN e.email END AS owner_email,
  e.poder_aquisitivo,
  CASE WHEN EXISTS (SELECT 1 FROM obras o WHERE o.state = e.uf AND o.city = e.cidade) THEN 1 ELSE 2 END AS rank
FROM core_obras_plus_email e
INNER JOIN nomes n ON n.nome = e.nome;

CREATE TEMP TABLE emails AS
SELECT
  uf,
  cidade,
  nome,
  professional_email,
  owner_email,
  rank,
  row_number() OVER (PARTITION BY nome ORDER BY rank, poder_aquisitivo DESC, uf, cidade) AS row_num
FROM emails_prep;

CREATE TEMP TABLE emails_agg AS
SELECT
  nome,
  max(CASE WHEN row_num = 1 THEN owner_email END) AS owner_first_email,
  max(CASE WHEN row_num = 1 THEN professional_email END) AS professional_first_email,
  group_concat(DISTINCT CASE WHEN row_num > 1 THEN owner_email END) AS owner_emails,
  group_concat(DISTINCT CASE WHEN row_num > 1 THEN professional_email END) AS professional_emails
FROM emails
WHERE row_num <= 5
GROUP BY nome;

SELECT
  o.obra_number,
  o.address,
  o.bairro,
  o.city,
  o.state,
  o.size,
  o.unidade,
  o.type,
  o.activity,
  date(o.start_date) AS start_date,
  date(o.end_date) AS end_date,
  date(o.first_listing_date) AS first_listing_date,
  o.owner,
  o.professional,
  pao.owner_first_telefone,
  pap.professional_first_telefone,
  eao.owner_first_email,
  eap.professional_first_email,
  pao.owner_telefones,
  pap.professional_telefones,
  eao.owner_emails,
  eap.professional_emails
FROM obras o
LEFT JOIN phones_agg pao ON o.owner = pao.nome
LEFT JOIN phones_agg pap ON o.professional = pap.nome
LEFT JOIN emails_agg eao ON o.owner = eao.nome
LEFT JOIN emails_agg eap ON o.professional = eap.nome
ORDER BY o.first_listing_date DESC, o.start_date DESC, o.city, o.state;
