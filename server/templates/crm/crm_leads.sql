select
  lead.id as lead_id,
  list_lead.id as list_lead_id,
  list_stage.id as stage_id,
  lead.obra_id,
  user.email as owner_email,
  coalesce(nullif(nullif(user.properties, ''), 'null') ->> 'name', user.name) as owner_name,
  cop.address,
  cop.bairro,
  cop.city,
  cop.state,
  cop.owner,
  cop.professional,
  cop.size,
  unixepoch(cop.start_date) as start_date,
  unixepoch(cop.end_date) as end_date,
  unixepoch(lead.favorited_at) as favorited_at,
  lead.properties as lead_properties,
  list_lead.properties as list_lead_properties,
  unixepoch(lead.updated) as lead_updated,
  unixepoch(list_lead.updated) as list_lead_updated,
  row_number() over (
    order by
      case
        -- when alert is in past
        when datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch') < datetime()
          then (lead.properties -> 'alert_at')
      end asc,
      list_lead.updated asc
  ) as rank1
from main.list_lead
inner join main.list on list.id = list_lead.list_id
inner join main.lead on lead.id = list_lead.lead_id
inner join main.list_stage on list_stage.id = list_lead.stage_id
left join main.user on user.id = lead.owner_id
left join core.core_obras_plus cop on cop.id = lead.obra_id

where list.team_id = {:team_id}
  and {user_filter}
  and ( list_stage.id = {:stage_id} or list_lead.id = {:list_lead_id} )

order by rank1 desc