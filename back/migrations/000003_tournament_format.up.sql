ALTER TABLE tournaments
  ADD COLUMN format TEXT NOT NULL DEFAULT 'single_elimination',
  ADD COLUMN group_count INTEGER;

ALTER TABLE tournaments
  ADD CONSTRAINT tournaments_format_check
    CHECK (format IN ('single_elimination', 'double_elimination', 'group_stage', 'group_de')),
  ADD CONSTRAINT tournaments_group_count_check
    CHECK (group_count IS NULL OR group_count BETWEEN 2 AND 4);
