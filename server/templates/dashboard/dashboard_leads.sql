select
  l.obra_id,
  cop.address,
  cop.bairro,
  cop.city,
  cop.state,
  cop.owner,
  cop.professional,
  cop.size,
  l.favorited_at
from main.lead l
inner join core.core_obras_plus cop on l.obra_id = cop.id
where l.favorited_at > ''
  and {where_clause}
order by l.favorited_at desc
limit 25
