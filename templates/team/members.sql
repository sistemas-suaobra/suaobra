SELECT 
  u.id,
  u.email,
  u.created,
  u.updated,
  u.properties
FROM "user" u
WHERE u.team_id = :team_id
ORDER BY u.created DESC 