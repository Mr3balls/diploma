CREATE TABLE IF NOT EXISTS tournament_messages (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content       TEXT NOT NULL CHECK (length(content) >= 1 AND length(content) <= 1000),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tournament_messages_tournament_created
    ON tournament_messages (tournament_id, created_at DESC);
