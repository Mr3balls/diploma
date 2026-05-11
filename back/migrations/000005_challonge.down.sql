-- 000005_challonge.down.sql

DROP INDEX  IF EXISTS idx_co_organizer_invites_invitee;
DROP INDEX  IF EXISTS idx_co_organizer_invites_token;
DROP TABLE  IF EXISTS co_organizer_invites;

DROP INDEX  IF EXISTS idx_result_reports_match;
DROP TABLE  IF EXISTS result_reports;

DROP INDEX  IF EXISTS idx_match_log_match;
DROP INDEX  IF EXISTS idx_match_log_tournament;
DROP TABLE  IF EXISTS match_log_entries;

DROP INDEX  IF EXISTS idx_tournament_members_user;
DROP INDEX  IF EXISTS idx_tournament_members_tournament;
DROP TABLE  IF EXISTS tournament_members;

ALTER TABLE matches DROP COLUMN IF EXISTS winner_participant_id;
ALTER TABLE matches DROP COLUMN IF EXISTS participant2_id;
ALTER TABLE matches DROP COLUMN IF EXISTS participant1_id;

DROP INDEX  IF EXISTS idx_participants_user;
DROP INDEX  IF EXISTS idx_participants_tournament;
DROP TABLE  IF EXISTS participants;

ALTER TABLE matches      DROP COLUMN IF EXISTS global_number;
ALTER TABLE tournaments  DROP COLUMN IF EXISTS max_participants;
ALTER TABLE tournaments  DROP COLUMN IF EXISTS slug;
DROP INDEX  IF EXISTS idx_tournaments_slug;

ALTER TABLE tournaments DROP CONSTRAINT IF EXISTS tournaments_status_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_status_check CHECK (
    status IN (
        'draft', 'registration_open', 'registration_closed',
        'bracket_generated', 'in_progress', 'finished', 'cancelled'
    )
);
