-- 1. Get the email for the specific lead owner
select
  lead.id as lead_id,
  lead.obra_id,
  user.email,
  coalesce(lead.properties -> 'obra' ->> 'address', cop.address) as address,
  coalesce(lead.properties -> 'obra' ->> 'city', cop.city) as city,
  coalesce(lead.properties -> 'obra' ->> 'state', cop.state) as state,
  coalesce(lead.properties -> 'obra' ->> 'owner', cop.owner) as owner,
  coalesce(lead.properties -> 'obra' ->> 'professional', cop.professional) as professional,
  (lead.properties -> 'alert_at') / 1000 as alert_at,
  coalesce(lead.properties -> 'alerted', false) as alerted

from main.lead
join main.user on user.id = lead.owner_id
left join core.core_obras_plus cop on cop.id = lead.obra_id

where 1=1
  and (lead.properties -> 'alert_at') is not null
  and coalesce(lead.properties -> 'alerted', 'false') = 'false'
  and datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch') <= datetime()

UNION ALL

-- 2. Fallback: if no owner_id, get the team manager
select
  lead.id as lead_id,
  lead.obra_id,
  user.email,
  coalesce(lead.properties -> 'obra' ->> 'address', cop.address) as address,
  coalesce(lead.properties -> 'obra' ->> 'city', cop.city) as city,
  coalesce(lead.properties -> 'obra' ->> 'state', cop.state) as state,
  coalesce(lead.properties -> 'obra' ->> 'owner', cop.owner) as owner,
  coalesce(lead.properties -> 'obra' ->> 'professional', cop.professional) as professional,
  (lead.properties -> 'alert_at') / 1000 as alert_at,
  coalesce(lead.properties -> 'alerted', false) as alerted

from main.lead
join main.user on user.team_id = lead.team_id and user.manager = true
left join core.core_obras_plus cop on cop.id = lead.obra_id

where 1=1
  and (lead.owner_id is null or lead.owner_id = '')
  and (lead.properties -> 'alert_at') is not null
  and coalesce(lead.properties -> 'alerted', 'false') = 'false'
  and datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch') <= datetime()

UNION ALL

-- 3. Fallback: if no owner_id and no manager exists, get the first user
select
  lead.id as lead_id,
  lead.obra_id,
  user.email,
  coalesce(lead.properties -> 'obra' ->> 'address', cop.address) as address,
  coalesce(lead.properties -> 'obra' ->> 'city', cop.city) as city,
  coalesce(lead.properties -> 'obra' ->> 'state', cop.state) as state,
  coalesce(lead.properties -> 'obra' ->> 'owner', cop.owner) as owner,
  coalesce(lead.properties -> 'obra' ->> 'professional', cop.professional) as professional,
  (lead.properties -> 'alert_at') / 1000 as alert_at,
  coalesce(lead.properties -> 'alerted', false) as alerted

from main.lead
join main.user on user.team_id = lead.team_id
left join core.core_obras_plus cop on cop.id = lead.obra_id

where 1=1
  and (lead.owner_id is null or lead.owner_id = '')
  and (lead.properties -> 'alert_at') is not null
  and coalesce(lead.properties -> 'alerted', 'false') = 'false'
  and datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch') <= datetime()
  and lead.team_id not in (
    select distinct team_id from main.user where manager = true
  )

group by lead.id, lead.obra_id, lead.team_id
having user.rowid = min(user.rowid)
