with recursive generate_dates(value) as (
  select julianday({:start_date})
  union all
  select value+1 from generate_dates
   where value+1<julianday({:end_date})
)

, visits as (
  select 
    julianday(date(coalesce(nullif(visited_at, ''), nullif(favorited_at, '')))) as date_value,
    count(1) as visit_cnt
  from main.lead
  where coalesce(nullif(visited_at, ''), nullif(favorited_at, '')) is not null
    and {where_clause_visits}
  group by 1
)

, leads as (
  select 
    julianday(date(favorited_at)) as date_value,
    count(1) as lead_cnt
  from main.lead
  where nullif(favorited_at, '') is not null
    and {where_clause_leads}
  group by 1
)

select
  date(dates.value) as date,
  coalesce(visit_cnt, 0) as visit_cnt,
  sum(coalesce(visit_cnt, 0)) over (rows unbounded preceding) as visit_cnt_cumulative,
  coalesce(lead_cnt, 0) as lead_cnt,
  sum(coalesce(lead_cnt, 0)) over (rows unbounded preceding) as lead_cnt_cumulative
from generate_dates dates
left join visits on dates.value = visits.date_value
left join leads on dates.value = leads.date_value
order by 1