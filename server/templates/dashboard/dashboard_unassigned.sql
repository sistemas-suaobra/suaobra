select
  lead.id as id,
  coalesce(
    nullif(trim(nullif(lead.properties -> 'obra' ->> 'address', 'null')), ''),
    nullif(trim(nullif(lead.properties -> 'obra' ->> 'owner', 'null')), ''),
    nullif(trim(nullif(lead.properties -> 'obra' ->> 'professional', 'null')), ''),
    'Obra'
  ) as title,
  nullif(lead.properties -> 'obra' ->> 'city', 'null') as city,
  nullif(lead.properties -> 'obra' ->> 'state', 'null') as state,
  coalesce(nullif(lead.favorited_at, ''), nullif(lead.visited_at, ''), lead.created) as at
from main.lead lead
where lead.team_id = {:team_id}
  and ifnull(lead.owner_id, '') = ''
  and ifnull(lead.excluded_at, '') = ''
order by at desc
limit 8
