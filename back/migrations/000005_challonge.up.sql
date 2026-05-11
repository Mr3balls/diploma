-- 000005_challonge.up.sql
-- Challonge-style tournament extensions

-- Extend tournament status to include Challonge lifecycle values
ALTER TABLE tournaments DROP CONSTRAINT IF EXISTS tournaments_status_check;
ALTER TABLE tournaments ADD CONSTRAINT tournaments_status_check CHECK (
    status IN (
        'draft', 'registration_open', 'registration_closed',
        'bracket_generated', 'in_progress', 'finished', 'cancelled',
        'ready', 'completed'
    )
);

-- URL slug for Challonge tournaments (6-32 alphanumeric)
ALTER TABLE tournaments ADD COLUMN IF NOT EXISTS slug TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tournaments_slug ON tournaments(slug) WHERE slug IS NOT NULL;

-- MaxParticipants for Challonge mode (supplements existing max_teams)
ALTER TABLE tournaments ADD COLUMN IF NOT EXISTS max_participants INT NOT NULL DEFAULT 0;

-- Global match number (for display ordering)
ALTER TABLE matches ADD COLUMN IF NOT EXISTS global_number INT;

-- Participants (individual players, Challonge-style)
CREATE TABLE IF NOT EXISTS participants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    name          TEXT NOT NULL,
    seed          INT  NOT NULL DEFAULT 0,
    status        TEXT NOT NULL DEFAULT 'active'
                  CHECK (status IN ('active', 'eliminated', 'champion')),
    final_rank    INT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, name)
);
CREATE INDEX IF NOT EXISTS idx_participants_tournament ON participants(tournament_id);
CREATE INDEX IF NOT EXISTS idx_participants_user       ON participants(user_id) WHERE user_id IS NOT NULL;

-- Participant columns on matches (nullable; NULL for team-based tournaments)
ALTER TABLE matches ADD COLUMN IF NOT EXISTS participant1_id      UUID REFERENCES participants(id) ON DELETE SET NULL;
ALTER TABLE matches ADD COLUMN IF NOT EXISTS participant2_id      UUID REFERENCES participants(id) ON DELETE SET NULL;
ALTER TABLE matches ADD COLUMN IF NOT EXISTS winner_participant_id UUID REFERENCES participants(id) ON DELETE SET NULL;

-- Tournament members (organizer/co_organizer/participant/viewer roles)
-- Separate from tournament_user_roles (owner/manager) to avoid collisions.
CREATE TABLE IF NOT EXISTS tournament_members (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role          TEXT NOT NULL CHECK (role IN ('organizer', 'co_organizer', 'participant', 'viewer')),
    joined_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tournament_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_tournament_members_tournament ON tournament_members(tournament_id);
CREATE INDEX IF NOT EXISTS idx_tournament_members_user       ON tournament_members(user_id);

-- Match action log (replaces/supplements audit_logs for per-match events)
CREATE TABLE IF NOT EXISTS match_log_entries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    match_id      UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    action        TEXT NOT NULL,
    actor_id      UUID NOT NULL REFERENCES users(id),
    detail        JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_match_log_tournament ON match_log_entries(tournament_id);
CREATE INDEX IF NOT EXISTS idx_match_log_match      ON match_log_entries(match_id);

-- Participant-submitted result reports (await organizer approval)
CREATE TABLE IF NOT EXISTS result_reports (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id       UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    reported_by_id UUID NOT NULL REFERENCES users(id),
    winner_id      UUID NOT NULL REFERENCES participants(id),
    score1         INT  NOT NULL DEFAULT 0,
    score2         INT  NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_result_reports_match ON result_reports(match_id);

-- Co-organizer invitations (7-day expiry, token-based accept)
CREATE TABLE IF NOT EXISTS co_organizer_invites (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    invitee_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invited_by_id UUID NOT NULL REFERENCES users(id),
    token         TEXT NOT NULL UNIQUE,
    status        TEXT NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'accepted', 'rejected', 'expired')),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_co_organizer_invites_token   ON co_organizer_invites(token);
CREATE INDEX IF NOT EXISTS idx_co_organizer_invites_invitee ON co_organizer_invites(invitee_id);
