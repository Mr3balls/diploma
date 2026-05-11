ALTER TABLE matches
    ADD COLUMN bracket_section TEXT NOT NULL DEFAULT 'WB'
        CHECK (bracket_section IN ('WB','LB','GF')),
    ADD COLUMN loser_next_match_id UUID REFERENCES matches(id),
    ADD COLUMN loser_next_slot     INT  NOT NULL DEFAULT 0;

CREATE INDEX idx_matches_bracket_section ON matches(bracket_section);
CREATE INDEX idx_matches_loser_next_match_id ON matches(loser_next_match_id);
