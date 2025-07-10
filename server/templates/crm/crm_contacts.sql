with owner_phone_records as (

  select distinct
    nome as name,
    contact_id,
    person_id,
    company_id,
    telefone as telephone,
    null as email,
    cidade as city,
    uf as state,
    row_number() over(partition by telefone order by coalesce(poder_aquisitivo, 1) desc) as rank_num
  from core.core_obras_plus_phone
  where nome in (
    select owner
    from core.core_obras_plus
    where id = {:obra_id}
  )

  order by
    -- poder_aquisitivo desc,
    telefone desc,
    uf,
    cidade
)

, professional_phone_records as (

  select distinct
    nome as name,
    contact_id,
    person_id,
    company_id,
    telefone as telephone,
    null as email,
    cidade as city,
    uf as state,
    row_number() over(partition by telefone order by coalesce(poder_aquisitivo, 1) desc) as rank_num
  from core.core_obras_plus_phone
  where nome in (
    select professional
    from core.core_obras_plus
    where id = {:obra_id}
  )

  order by
    -- poder_aquisitivo desc,
    telefone desc,
    uf,
    cidade
)

, owner_email_records as (

  select distinct
    nome as name,
    contact_id,
    person_id,
    company_id,
    null as telephone,
    email,
    cidade as city,
    uf as state,
    row_number() over(partition by email order by coalesce(poder_aquisitivo, 1) desc) as rank_num
  from core.core_obras_plus_email
  where nome in (
    select owner
    from core.core_obras_plus
    where id = {:obra_id}
  )

  order by
    poder_aquisitivo desc,
    email desc,
    uf,
    cidade
)

, professional_email_records as (

  select distinct
    nome as name,
    contact_id,
    person_id,
    company_id,
    null as telephone,
    email,
    cidade as city,
    uf as state,
    row_number() over(partition by email order by coalesce(poder_aquisitivo, 1) desc) as rank_num
  from core.core_obras_plus_email
  where nome in (
    select professional
    from core.core_obras_plus
    where id = {:obra_id}
  )

  order by
    poder_aquisitivo desc,
    email desc,
    uf,
    cidade
)

, owner_email_records_limited as ( select * from owner_email_records where rank_num = 1 limit 4 )

, professional_email_records_limited as ( select * from professional_email_records where rank_num = 1 limit 4 )

, owner_phone_records_limited as ( select * from owner_phone_records where rank_num = 1 limit 4 )

, professional_phone_records_limited as ( select * from professional_phone_records where rank_num = 1 limit 4 )

select * from owner_email_records_limited
union all
select * from professional_email_records_limited
union all
select * from owner_phone_records_limited
union all
select * from professional_phone_records_limited