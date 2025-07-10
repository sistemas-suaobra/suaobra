-- legacy, stage_rank is currently removed.
-- Keeping in case we need it in the future.

-- adjust ranking when the stage is the same
update list_lead 
  set stage_rank = case
    when id = {:list_lead_id}
      then {:new_rank}
    when {:new_rank} < {:old_rank} and stage_rank >= {:new_rank} and stage_rank < {:old_rank}
      then stage_rank + 1
    when {:new_rank} > {:old_rank} and stage_rank <= {:new_rank} and stage_rank > {:old_rank}
      then stage_rank - 1
    else stage_rank
  end
where stage_id = {:stage_id}