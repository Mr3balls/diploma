CREATE TABLE bracket_groups (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bracket_id    UUID NOT NULL REFERENCES brackets(id) ON DELETE CASCADE,
    tournament_id UUID NOT NULL,
    name          TEXT NOT NULL,
    position      INT  NOT NULL
);

CREATE TABLE bracket_group_members (
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES bracket_groups(id) ON DELETE CASCADE,
    team_id  UUID NOT NULL,
    wins     INT  NOT NULL DEFAULT 0,
    losses   INT  NOT NULL DEFAULT 0,
    draws    INT  NOT NULL DEFAULT 0,
    points   INT  NOT NULL DEFAULT 0
);

ALTER TABLE matches ADD COLUMN group_id UUID REFERENCES bracket_groups(id);

CREATE INDEX idx_bracket_groups_bracket_id ON bracket_groups(bracket_id);
CREATE INDEX idx_bracket_group_members_group_id ON bracket_group_members(group_id);
CREATE INDEX idx_matches_group_id ON matches(group_id);
