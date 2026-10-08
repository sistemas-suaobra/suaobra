select
  activity.id as id,
  coalesce(nullif(trim(cop.owner), ''), nullif(trim(cop.address), ''), 'Lead') as title,
  ifnull(stage.name, '') as stage,
  coalesce(
    nullif(nullif(actor.properties, ''), 'null') ->> 'name',
    nullif(actor.name, ''),
    nullif(activity.actor_email, '')
  ) as by_name
from main.lead_activity activity
inner join main.lead lead on lead.id = activity.lead_id
left join core.core_obras_plus cop on cop.id = lead.obra_id
left join main.user actor
  on actor.email = activity.actor_email
  and actor.team_id = activity.team_id
left join main.list_stage stage on stage.id = activity.properties ->> 'stage_id'
where activity.team_id = {:team_id}
  and activity.type = 'history'
  and {user_filter}
  and ifnull(stage.name, '') != ''
order by activity.created desc
limit 8
