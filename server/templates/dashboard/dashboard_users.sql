select
  u.id,
  u.email,
  coalesce(nullif(nullif(u.properties, ''), 'null') ->> 'name', u.name) as name,
  u.created
from user u
where u.team_id = {:team_id}
order by u.name 