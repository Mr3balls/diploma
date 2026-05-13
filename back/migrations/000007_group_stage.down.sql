DROP INDEX IF EXISTS idx_matches_group_id;
DROP INDEX IF EXISTS idx_bracket_group_members_group_id;
DROP INDEX IF EXISTS idx_bracket_groups_bracket_id;
ALTER TABLE matches DROP COLUMN IF EXISTS group_id;
DROP TABLE IF EXISTS bracket_group_members;
DROP TABLE IF EXISTS bracket_groups;