update lead
set {contacted_at_col} = datetime('now')
where team_id = '{team_id}' and lead_id = '{lead_id}'
