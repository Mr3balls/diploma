-- 000006_registration_mode.up.sql
ALTER TABLE tournaments
    ADD COLUMN IF NOT EXISTS registration_mode TEXT NOT NULL DEFAULT 'team'
    CHECK (registration_mode IN ('team', 'individual'));

-- Relax max_teams constraint so individual tournaments can use 0
ALTER TABLE tournaments DROP CONSTRAINT IF EXISTS tournaments_max_teams_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_max_teams_check CHECK (max_teams >= 0);
