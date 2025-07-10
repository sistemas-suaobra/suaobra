with prep as (
  select
    cop.id as obra_id,
    cop.obra_number,
    cop.owner,
    cop.professional,
    cop.address,
    cop.bairro,
    cop.city,
    cop.state,
    cop.start_date,
    cop.end_date,
    cop.type,
    cop.activity,
    cop.size,
    cop.unidade,
    cop.has_owner_phone,
    cop.has_owner_email,
    cop.has_professional_phone,
    cop.has_professional_email,
    l.favorited_at,
    0 as score

  from core.core_obras_plus cop

  left join main.lead l
    on cop.id = l.obra_id
    and l.team_id = {:team_id}
    and {user_filter}

  where 1=1
    and cop.state = {:state}
    and cop.city = {:city}
    and {filters}
)

select *
from prep

order by score desc

limit 10