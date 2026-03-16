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
  date(julianday(cop.end_date) + 400) as end_date,
  cop.type,
  cop.activity,
  cop.size,
  cop.unidade,
  cop.has_owner_phone,
  cop.has_owner_email,
  cop.has_professional_phone,
  cop.has_professional_email,
  l.visited_at,
  l.favorited_at,
  l.excluded_at,
  l.owner_contact_pending_at,
  l.professional_contact_pending_at,
  l.owner_contacted_at,
  l.professional_contacted_at,
  (select 1 from main.campanha_destinatarios where team_id = '{teamId}' and obra_id = cop.id and contato_tipo = 'OWNER' and status = 'ENVIADO' limit 1) as owner_enviado_em,
  (select 1 from main.campanha_destinatarios where team_id = '{teamId}' and obra_id = cop.id and contato_tipo = 'PROFISSIONAL' and status = 'ENVIADO' limit 1) as professional_enviado_em,
  n.id is not null as has_note

from core.core_obras_plus cop

left join main.lead l
  on cop.id = l.obra_id
  and l.team_id = '{teamId}'

left join main.obra_note n
  on cop.id = n.obra_id

where 1=1
  and cop.city = '{city}'
  and cop.size >= {sizeMin}
  and cop.size <= {sizeMax}
  and {statusCond}
  and {neighborhoodCond}
  and {filterCond}
  and {dateFilterCond}

order by {order}
limit {itemPerPage}
offset {offset}