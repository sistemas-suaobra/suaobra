-- legacy, stage_rank is currently removed.
-- Keeping in case we need it in the future.
update list_lead 
  set stage_rank = case
    when id = {:list_lead_id}
      then {:new_rank}
    when stage_rank >= {:new_rank}
      then stage_rank + 1
    else stage_rank
  end,
  stage_id = {:new_stage_id} -- set new stage_id fir list_lead_id
where stage_id = {:new_stage_id} or id = {:list_lead_id}