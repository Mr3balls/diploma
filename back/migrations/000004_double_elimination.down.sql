DROP INDEX IF EXISTS idx_matches_loser_next_match_id;
DROP INDEX IF EXISTS idx_matches_bracket_section;

ALTER TABLE matches
    DROP COLUMN IF EXISTS loser_next_slot,
    DROP COLUMN IF EXISTS loser_next_match_id,
    DROP COLUMN IF EXISTS bracket_section;
