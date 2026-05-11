ALTER TABLE tournaments DROP COLUMN IF EXISTS registration_mode;
ALTER TABLE tournaments DROP CONSTRAINT IF EXISTS tournaments_max_teams_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_max_teams_check CHECK (max_teams >= 2);
