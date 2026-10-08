select
  lead.id as id,
  coalesce(nullif(trim(cop.address), ''), nullif(trim(cop.owner), ''), 'Obra') as title,
  cop.city as city,
  cop.state as state,
  coalesce(nullif(lead.favorited_at, ''), nullif(lead.visited_at, ''), lead.created) as at
from main.lead lead
left join core.core_obras_plus cop on cop.id = lead.obra_id
where lead.team_id = {:team_id}
  and ifnull(lead.owner_id, '') = ''
  and ifnull(lead.excluded_at, '') = ''
order by at desc
limit 8
