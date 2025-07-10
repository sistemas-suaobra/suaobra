select
  strftime('%Y-%m', date({:month})) as month,
  count(coalesce(nullif(l.visited_at, ''), l.id)) as visit_cnt,
  count(ll.id) as lead_cnt,
  count(case when ls."order" in (3, 4) then 1 end) as opportunity_cnt,
  count(case when ls."order" = 4 then 1 end) as sold_cnt
from main.lead l
left join list_lead ll on l.id = ll.lead_id
left join list_stage ls on ls.id = ll.stage_id
where {where_clause}
  and (
    strftime('%Y-%m', date(l.favorited_at)) = strftime('%Y-%m', date({:month}))
    or strftime('%Y-%m', date(l.visited_at)) = strftime('%Y-%m', date({:month}))
  )