-- Create the label table
CREATE TABLE IF NOT EXISTS label (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    gmail_label_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT, -- 'system' or 'user'
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES "user"(provider_id) ON DELETE CASCADE,
    UNIQUE (user_id, gmail_label_id)
);

CREATE TRIGGER IF NOT EXISTS label_auto_updated_at
BEFORE UPDATE ON label
FOR EACH ROW
BEGIN
    UPDATE label SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create the thread table
CREATE TABLE IF NOT EXISTS thread (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    gmail_thread_id TEXT NOT NULL,
    snippet TEXT,
    history_id_ref TEXT, -- Stores the historyId from Gmail API for the thread
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES "user"(provider_id) ON DELETE CASCADE,
    UNIQUE (user_id, gmail_thread_id)
);

CREATE TRIGGER IF NOT EXISTS thread_auto_updated_at
BEFORE UPDATE ON thread
FOR EACH ROW
BEGIN
    UPDATE thread SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create the email table to store MIME email data from the Gmail API (Modified)
-- This definition REPLACES the original email table definition.
CREATE TABLE IF NOT EXISTS email (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    thread_id INTEGER,                     -- Foreign key to the thread table
    gmail_message_id TEXT NOT NULL,        -- The unique ID of the message from the Gmail API
    raw_mime_content TEXT NOT NULL,
    snippet TEXT,                          -- Snippet of the message
    is_read BOOLEAN DEFAULT FALSE,         -- Read status of the email
    is_starred BOOLEAN DEFAULT FALSE,      -- Starred status of the email
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    server_timestamp DATETIME NOT NULL,             -- New column to store the server timestamp of the email
    FOREIGN KEY (user_id) REFERENCES "user"(provider_id) ON DELETE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES thread(id) ON DELETE SET NULL,
    UNIQUE (user_id, gmail_message_id)     -- gmail_message_id should be unique per user
);

CREATE TRIGGER IF NOT EXISTS email_auto_updated_at
BEFORE UPDATE ON email
FOR EACH ROW
BEGIN
    UPDATE email SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create the email_label junction table (many-to-many relationship between emails and labels)
CREATE TABLE IF NOT EXISTS email_label (
    email_id INTEGER NOT NULL,
    label_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL, -- Ensures data consistency and can simplify user-specific queries
    PRIMARY KEY (email_id, label_id),
    FOREIGN KEY (email_id) REFERENCES email(id) ON DELETE CASCADE,
    FOREIGN KEY (label_id) REFERENCES label(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES "user"(provider_id) ON DELETE CASCADE
);

-- Create the history table (to track Gmail history records for a user)
CREATE TABLE IF NOT EXISTS history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    gmail_history_id TEXT NOT NULL, -- The history ID from Gmail API
    processed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- When this history record was processed
    FOREIGN KEY (user_id) REFERENCES "user"(provider_id) ON DELETE CASCADE,
    UNIQUE (user_id, gmail_history_id)
);

-- Optional: Add indexes for foreign keys and frequently queried columns
-- Note: UNIQUE constraints often create indexes automatically.
CREATE INDEX IF NOT EXISTS idx_email_user_id ON email(user_id);
CREATE INDEX IF NOT EXISTS idx_email_thread_id ON email(thread_id);

CREATE INDEX IF NOT EXISTS idx_label_user_id ON label(user_id);

CREATE INDEX IF NOT EXISTS idx_thread_user_id ON thread(user_id);

CREATE INDEX IF NOT EXISTS idx_email_label_email_id ON email_label(email_id);
CREATE INDEX IF NOT EXISTS idx_email_label_label_id ON email_label(label_id);
CREATE INDEX IF NOT EXISTS idx_email_label_user_id ON email_label(user_id);

CREATE INDEX IF NOT EXISTS idx_history_user_id ON history(user_id);

