-- Lembretes vencidos para aviso na plataforma (todas as telas).
-- alert_at é timestamp absoluto em ms. Inclui já e-mailados: o aviso
-- consome o lembrete (remove alert_at) e ele deixa de existir.
select
  lead.id as lead_id,
  coalesce(lead.properties -> 'obra' ->> 'owner', cop.owner, '') as owner,
  coalesce(lead.properties -> 'obra' ->> 'city', cop.city, '') as city,
  coalesce(lead.properties -> 'obra' ->> 'state', cop.state, '') as state,
  (lead.properties -> 'alert_at') as alert_at
from main.lead
left join core.core_obras_plus cop on cop.id = lead.obra_id
where lead.team_id = {:team_id}
  and {user_filter}
  and (lead.properties -> 'alert_at') is not null
  and cast((lead.properties -> 'alert_at') as real) > 0
  and datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch') <= datetime()
order by (lead.properties -> 'alert_at') asc
limit 20
