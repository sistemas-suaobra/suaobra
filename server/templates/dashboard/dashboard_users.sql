select
  u.id,
  u.email,
  coalesce(nullif(nullif(u.properties, ''), 'null') ->> 'name', u.name) as name,
  u.created
from user u
where u.team_id = {:team_id}
  and coalesce(u.properties ->> 'active', '1') not in ('0', 'false')
order by u.name 