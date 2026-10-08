select
  lead.id as id,
  coalesce(nullif(trim(cop.owner), ''), nullif(trim(cop.address), ''), 'Lead') as title,
  cop.city as city,
  cop.state as state,
  coalesce(
    nullif(nullif(user.properties, ''), 'null') ->> 'name',
    nullif(user.name, ''),
    nullif(user.email, '')
  ) as owner,
  case
    when datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch') < datetime('now')
      then 'Atrasado'
    else 'Pendente'
  end as status
from main.lead lead
left join core.core_obras_plus cop on cop.id = lead.obra_id
left join main.user on user.id = lead.owner_id
where lead.team_id = {:team_id}
  and {user_filter}
  and (lead.properties -> 'alert_at') is not null
  and cast((lead.properties -> 'alert_at') as real) > 0
order by
  case
    when datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch') < datetime('now') then 0
    else 1
  end,
  (lead.properties -> 'alert_at') asc
limit 8
