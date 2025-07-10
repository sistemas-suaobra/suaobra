with obras as (
  {coreObrasPlusSQL}
)

, phones_prep as (
  select
    uf,
    cidade,
    nome,
    case
      when nome in ( select professional from obras)
        then telefone
    end as professional_telefone,
    case
      when nome in ( select owner from obras)
        then telefone
    end as owner_telefone,
    poder_aquisitivo,
    case
      when (uf || cidade) in ( select state || city from obras )
        then 1
      else 2
    end as rank
  from core.core_obras_plus_phone
  where nome in ( select distinct owner from obras)
    or nome in ( select distinct professional from obras) 
)

, phones as (
  select
    uf,
    cidade,
    nome,
    professional_telefone,
    owner_telefone,
    rank,
    row_number() over (partition by nome order by rank) as row_num
  from phones_prep
  where nome in ( select distinct owner from obras)
    or nome in ( select distinct professional from obras) 

  order by rank asc, professional_telefone desc, owner_telefone desc, poder_aquisitivo desc, uf, cidade
)
, phones_agg as (
  select
    nome,
    max(case when row_num = 1 then owner_telefone end) owner_first_telefone,
    max(case when row_num = 1 then professional_telefone end) professional_first_telefone,
    group_concat(distinct case when row_num > 1 then owner_telefone end) as owner_telefones,
    group_concat(distinct case when row_num > 1 then professional_telefone end) as professional_telefones
  from phones
  where row_num <= 10
  group by 1
)
, emails_prep as (
  select
    uf,
    cidade,
    nome,
    case
      when nome in ( select professional from obras)
        then email
    end as professional_email,
    case
      when nome in ( select owner from obras)
        then email
    end as owner_email,
    poder_aquisitivo,
    case
      when (uf || cidade) in ( select state || city from obras )
        then 1
      else 2
    end as rank
  from core.core_obras_plus_email
  where nome in ( select distinct owner from obras)
    or nome in ( select distinct professional from obras) 
)
, emails as (
  select
    uf,
    cidade,
    nome,
    professional_email,
    owner_email,
    rank,
    row_number() over (partition by nome order by rank) as row_num
  from emails_prep
  where nome in ( select distinct owner from obras)
    or nome in ( select distinct professional from obras) 

  order by rank asc, poder_aquisitivo desc, uf, cidade
)
, emails_agg as (
  select
    nome,
    max(case when row_num = 1 then owner_email end) owner_first_email,
    max(case when row_num = 1 then professional_email end) professional_first_email,
    group_concat(distinct case when row_num > 1 then owner_email end) as owner_emails,
    group_concat(distinct case when row_num > 1 then professional_email end) as professional_emails
  from emails
  where row_num <= 5
  group by 1
)

select
  obras.obra_id,
  obras.address,
  obras.bairro,
  obras.city,
  obras.state,
  obras.size,
  obras.owner,
  obras.professional,
  date(obras.start_date) as start_date,
  date(obras.end_date) as end_date,
  pao.owner_first_telefone,
  pap.professional_first_telefone,
  pao.owner_telefones,
  pap.professional_telefones,
  eao.owner_first_email,
  eap.professional_first_email,
  eao.owner_emails,
  eap.professional_emails
from obras
left join phones_agg pao on obras.owner = pao.nome
left join phones_agg pap on obras.professional = pap.nome
left join emails_agg eao on obras.owner = eao.nome
left join emails_agg eap on obras.professional = eap.nome
limit {itemPerPage}