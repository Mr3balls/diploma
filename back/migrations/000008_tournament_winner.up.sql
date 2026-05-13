ALTER TABLE tournaments ADD COLUMN winner_team_id UUID REFERENCES teams(id) ON DELETE SET NULL;
ALTER TABLE tournaments ADD COLUMN winner_participant_id UUID REFERENCES participants(id) ON DELETE SET NULL;
