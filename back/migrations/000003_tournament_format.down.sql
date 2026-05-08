ALTER TABLE tournaments
  DROP CONSTRAINT IF EXISTS tournaments_format_check,
  DROP CONSTRAINT IF EXISTS tournaments_group_count_check,
  DROP COLUMN IF EXISTS format,
  DROP COLUMN IF EXISTS group_count;
