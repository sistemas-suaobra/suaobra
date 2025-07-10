with prep as (
  select
    *,
    row_number() over(partition by nome order by uf, coalesce(poder_aquisitivo, 1) desc) as row_num
  from core.core_obras_plus_phone

  where nome in ({names})
)

select distinct
  contact_id,
  person_id,
  company_id,
  nome,
  telefone as telephone,
  cidade as city,
  uf as state
from prep
where row_num = 1
