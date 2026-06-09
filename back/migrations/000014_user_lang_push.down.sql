DROP TABLE IF EXISTS push_subscriptions;
ALTER TABLE users DROP COLUMN IF EXISTS notification_preferences;
ALTER TABLE users DROP COLUMN IF EXISTS lang;
