-- legacy, stage_rank is currently removed.
-- Keeping in case we need it in the future.
update list_lead 
  set stage_rank = case
    when id = {:list_lead_id}
      then 0
    when stage_rank >= {:old_rank}
      then stage_rank - 1
    else stage_rank
  end,
  stage_id = case
    when id = {:list_lead_id}
      then '' -- set empty to remove from stage
    else stage_id
  end
where stage_id = {:old_stage_id}