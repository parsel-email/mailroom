-- Migration Down

-- Revert the changes made in the corresponding .up.sql file

-- Drop indexes first (optional, but good practice to reverse index creation)
DROP INDEX IF EXISTS idx_history_user_id;
DROP INDEX IF EXISTS idx_email_label_user_id;
DROP INDEX IF EXISTS idx_email_label_label_id;
DROP INDEX IF EXISTS idx_email_label_email_id;
DROP INDEX IF EXISTS idx_thread_user_id;
DROP INDEX IF EXISTS idx_label_user_id;
DROP INDEX IF EXISTS idx_email_thread_id;
DROP INDEX IF EXISTS idx_email_user_id;

-- Drop the history table
DROP TABLE IF EXISTS history;

-- Drop the email_label junction table
DROP TABLE IF EXISTS email_label;

-- Drop the trigger for the email table
DROP TRIGGER IF EXISTS email_auto_updated_at;

-- Drop the email table (this was modified, so we drop the new version)
DROP TABLE IF EXISTS email;

-- Drop the trigger for the thread table
DROP TRIGGER IF EXISTS thread_auto_updated_at;

-- Drop the thread table
DROP TABLE IF EXISTS thread;

-- Drop the trigger for the label table
DROP TRIGGER IF EXISTS label_auto_updated_at;

-- Drop the label table
DROP TABLE IF EXISTS label;

-- Note: The user table is assumed to be managed by a different migration
-- and is not dropped here.
