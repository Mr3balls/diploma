CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE platform_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE CHECK (code IN ('player', 'platform_admin'))
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT NOT NULL UNIQUE,
    nickname TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    avatar_url TEXT,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_nickname_lower ON users((lower(nickname)));
CREATE INDEX idx_users_email_lower ON users((lower(email)));

CREATE TABLE user_platform_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES platform_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, role_id)
);
CREATE INDEX idx_user_platform_roles_user_id ON user_platform_roles(user_id);

CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);

CREATE TABLE tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    discipline TEXT NOT NULL,
    description TEXT,
    rules TEXT,
    location TEXT,
    max_teams INT NOT NULL CHECK (max_teams >= 2),
    registration_deadline TIMESTAMPTZ,
    start_at TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('draft','registration_open','registration_closed','bracket_generated','in_progress','finished','cancelled')),
    visibility TEXT NOT NULL CHECK (visibility IN ('public','private')),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_tournaments_status ON tournaments(status);
CREATE INDEX idx_tournaments_visibility ON tournaments(visibility);
CREATE INDEX idx_tournaments_owner_user_id ON tournaments(owner_user_id);
CREATE INDEX idx_tournaments_deleted_at ON tournaments(deleted_at);

CREATE TABLE tournament_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner','manager')),
    assigned_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tournament_id, user_id, role)
);
CREATE INDEX idx_tournament_user_roles_tournament_user ON tournament_user_roles(tournament_id, user_id);

CREATE TABLE google_sheet_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE REFERENCES tournaments(id) ON DELETE CASCADE,
    sheet_url TEXT NOT NULL,
    spreadsheet_id TEXT NOT NULL,
    worksheet_name TEXT NOT NULL,
    status TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_google_sheet_links_tournament_id ON google_sheet_links(tournament_id);

CREATE TABLE import_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    sheet_link_id UUID NOT NULL REFERENCES google_sheet_links(id) ON DELETE CASCADE,
    started_by UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('pending','parsing','preview_ready','confirmed','failed')),
    summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_import_batches_tournament_id ON import_batches(tournament_id);

CREATE TABLE import_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL REFERENCES import_batches(id) ON DELETE CASCADE,
    row_number INT NOT NULL,
    raw_data_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    team_name TEXT,
    discipline TEXT,
    captain_nick TEXT,
    player_2_nick TEXT,
    player_3_nick TEXT,
    player_4_nick TEXT,
    player_5_nick TEXT,
    substitute_nick TEXT,
    status TEXT NOT NULL CHECK (status IN ('new','valid','needs_review','duplicate','rejected','confirmed')),
    validation_errors_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_import_rows_batch_id ON import_rows(batch_id);
CREATE INDEX idx_import_rows_status ON import_rows(status);

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','awaiting_confirmation','ready_for_review','approved','rejected')),
    approved_by_manager BOOLEAN,
    created_from_import_row_id UUID REFERENCES import_rows(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tournament_id, name)
);
CREATE INDEX idx_teams_tournament_id ON teams(tournament_id);
CREATE INDEX idx_teams_status ON teams(status);
CREATE INDEX idx_teams_deleted_at ON teams(deleted_at);

CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    nickname TEXT NOT NULL,
    member_role TEXT NOT NULL CHECK (member_role IN ('captain','player','substitute')),
    is_captain BOOLEAN NOT NULL DEFAULT FALSE,
    is_substitute BOOLEAN NOT NULL DEFAULT FALSE,
    confirmation_status TEXT NOT NULL CHECK (confirmation_status IN ('found','not_found','pending_confirmation','confirmed','declined','removed')),
    invited_at TIMESTAMPTZ,
    responded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_team_members_confirmation_status ON team_members(confirmation_status);
CREATE INDEX idx_team_members_deleted_at ON team_members(deleted_at);

CREATE TABLE brackets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE REFERENCES tournaments(id) ON DELETE CASCADE,
    format TEXT NOT NULL,
    seeding_method TEXT NOT NULL,
    status TEXT NOT NULL,
    generated_by UUID NOT NULL REFERENCES users(id),
    generated_at TIMESTAMPTZ NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX idx_brackets_tournament_id ON brackets(tournament_id);

CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    bracket_id UUID NOT NULL REFERENCES brackets(id) ON DELETE CASCADE,
    round_number INT NOT NULL CHECK (round_number >= 1),
    slot_index INT NOT NULL CHECK (slot_index >= 1),
    team1_id UUID REFERENCES teams(id),
    team2_id UUID REFERENCES teams(id),
    scheduled_at TIMESTAMPTZ,
    location_or_server TEXT,
    status TEXT NOT NULL CHECK (status IN ('scheduled','awaiting_confirmation','confirmed','reschedule_requested','issue_reported','in_progress','finished','cancelled')),
    team1_confirmation_status TEXT NOT NULL CHECK (team1_confirmation_status IN ('pending','ready_confirmed','reschedule_requested','issue_reported')),
    team2_confirmation_status TEXT NOT NULL CHECK (team2_confirmation_status IN ('pending','ready_confirmed','reschedule_requested','issue_reported')),
    winner_team_id UUID REFERENCES teams(id),
    score_text TEXT,
    manager_comment TEXT,
    next_match_id UUID REFERENCES matches(id),
    source_match1_id UUID REFERENCES matches(id),
    source_match2_id UUID REFERENCES matches(id),
    is_bye BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_matches_tournament_id ON matches(tournament_id);
CREATE INDEX idx_matches_bracket_id ON matches(bracket_id);
CREATE INDEX idx_matches_round_slot ON matches(round_number, slot_index);
CREATE INDEX idx_matches_next_match_id ON matches(next_match_id);
CREATE INDEX idx_matches_deleted_at ON matches(deleted_at);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('added_to_team','team_participation_confirmed','team_participation_declined','match_assigned','match_time_changed','match_rescheduled','match_cancelled','result_submitted','result_confirmed','tournament_finished')),
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    action_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    acted_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_deleted_at ON notifications(deleted_at);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id),
    tournament_id UUID REFERENCES tournaments(id) ON DELETE SET NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    action_type TEXT NOT NULL,
    description TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_tournament_id ON audit_logs(tournament_id);
CREATE INDEX idx_audit_logs_entity_type_entity_id ON audit_logs(entity_type, entity_id);
