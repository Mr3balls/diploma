UPDATE tournaments
SET max_participants = max_teams
WHERE max_participants = 0 AND max_teams > 0 AND deleted_at IS NULL;
