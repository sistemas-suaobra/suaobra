select
  strftime('%Y-%m', date({:month})) as month,
  count(distinct case
    when l.visited_at >= {:month}
      and l.visited_at < date({:month}, '+1 month')
    then l.id
    when coalesce(nullif(l.visited_at, ''), '') = ''
      and l.favorited_at >= {:month}
      and l.favorited_at < date({:month}, '+1 month')
    then l.id
  end) as visit_cnt,
  count(distinct case
    when ll.id is not null
      and l.favorited_at >= {:month}
      and l.favorited_at < date({:month}, '+1 month')
    then l.id
  end) as lead_cnt,
  count(distinct case
    when ll.id is not null
      and (
        cast(ls."order" as integer) >= 3
        or upper(ls.name) like '%PROPOST%'
        or upper(ls.name) like '%OPORTUN%'
        or upper(ls.name) like '%NEGOCI%'
        or upper(ls.name) like '%VEND%'
        or upper(ls.name) like '%FECHAD%'
      )
      and (
        (l.favorited_at >= {:month} and l.favorited_at < date({:month}, '+1 month'))
        or (l.visited_at >= {:month} and l.visited_at < date({:month}, '+1 month'))
        or exists (
          select 1
          from main.lead_activity activity
          where activity.lead_id = l.id
            and activity.type = 'history'
            and activity.created >= {:month}
            and activity.created < date({:month}, '+1 month')
            and activity.properties ->> 'stage_id' = ls.id
        )
      )
    then l.id
  end) as opportunity_cnt,
  count(distinct case
    when ll.id is not null
      and (
        upper(ls.name) like '%VEND%'
        or upper(ls.name) like '%FECHAD%'
        or cast(ls."order" as integer) = (
          select max(cast(stage."order" as integer))
          from main.list_stage stage
          where stage.list_id = ls.list_id
        )
      )
      and (
        (l.favorited_at >= {:month} and l.favorited_at < date({:month}, '+1 month'))
        or (l.visited_at >= {:month} and l.visited_at < date({:month}, '+1 month'))
        or exists (
          select 1
          from main.lead_activity activity
          where activity.lead_id = l.id
            and activity.type = 'history'
            and activity.created >= {:month}
            and activity.created < date({:month}, '+1 month')
            and activity.properties ->> 'stage_id' = ls.id
        )
      )
    then l.id
  end) as sold_cnt
from main.lead l
left join main.list_lead ll on l.id = ll.lead_id
left join main.list_stage ls on ls.id = ll.stage_id
where {where_clause}
  and (
    (l.favorited_at >= {:month} and l.favorited_at < date({:month}, '+1 month'))
    or (l.visited_at >= {:month} and l.visited_at < date({:month}, '+1 month'))
    or exists (
      select 1
      from main.lead_activity activity
      where activity.lead_id = l.id
        and activity.type = 'history'
        and activity.created >= {:month}
        and activity.created < date({:month}, '+1 month')
    )
  )
