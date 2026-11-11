-- CTE limita o histórico antes dos joins; sem core.* para não travar se core.db cair.
with recent as (
  select
    activity.id as id,
    activity.lead_id as lead_id,
    activity.actor_email as actor_email,
    activity.team_id as team_id,
    activity.properties ->> 'stage_id' as stage_id,
    activity.created as created
  from main.lead_activity activity
  inner join main.lead lead on lead.id = activity.lead_id
  where activity.team_id = {:team_id}
    and activity.type = 'history'
    and {user_filter}
  order by activity.created desc
  limit 40
)
select
  recent.id as id,
  coalesce(
    nullif(trim(nullif(lead.properties -> 'obra' ->> 'owner', 'null')), ''),
    nullif(trim(nullif(lead.properties -> 'obra' ->> 'professional', 'null')), ''),
    nullif(trim(nullif(json_extract(lead.properties, '$.contacts[0].name'), 'null')), ''),
    nullif(trim(nullif(lead.properties -> 'obra' ->> 'address', 'null')), ''),
    'Lead'
  ) as title,
  ifnull(stage.name, '') as stage,
  coalesce(
    nullif(nullif(actor.properties, ''), 'null') ->> 'name',
    nullif(actor.name, ''),
    nullif(recent.actor_email, '')
  ) as by_name
from recent
inner join main.lead lead on lead.id = recent.lead_id
left join main.user actor
  on actor.email = recent.actor_email
  and actor.team_id = recent.team_id
left join main.list_stage stage on stage.id = recent.stage_id
where ifnull(stage.name, '') != ''
order by recent.created desc
limit 8
