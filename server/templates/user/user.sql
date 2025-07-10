select
  u.id,
  u.email,
  -- u.legacy_id,
  json_object(
    'id', t.id,
    'name', t.name,
    'active', case t.active when 1 then json('true') when 0 then json('false') end,
    'blocked', case t.blocked when 1 then json('true') when 0 then json('false') end,
    'cities', json(t.cities),
    'export', t.export,
    'entitlements', json(t.entitlements)
  ) as team
from main.user u
join main.team t on u.team_id = t.id
where u.id = {:id}