with prep as (
  select
    *,
    row_number() over(partition by email order by coalesce(poder_aquisitivo, 1) desc) as row_num
  from core.core_obras_plus_email

  where nome = '{nome}' and uf = '{uf}' and cidade = '{cidade}'
)

select distinct
  contact_id,
  person_id,
  company_id,
  email,
  cidade as city,
  uf as state
from prep

where row_num = 1

order by
  case
    when uf = '{uf}'
      and cidade = '{cidade}'
      then 1
    else 2
  end asc,
  poder_aquisitivo desc,
  uf,
  cidade

limit 6