-- name: CreateEmail :one
INSERT INTO email (
    user_id,
    thread_id,
    gmail_message_id,
    raw_mime_content,
    snippet,
    is_read,
    is_starred
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateEmail :one
UPDATE email
SET
    thread_id = ?,
    raw_mime_content = ?,
    snippet = ?,
    is_read = ?,
    is_starred = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteEmail :exec
DELETE FROM email WHERE id = ?;

-- name: ListEmails :many
SELECT * FROM email
WHERE user_id = COALESCE(?, user_id)
  AND is_read = COALESCE(?, is_read)
  AND is_starred = COALESCE(?, is_starred)
ORDER BY created_at DESC
LIMIT COALESCE(?, 100)
OFFSET COALESCE(?, 0);

-- name: GetEmailByID :one
SELECT * FROM email WHERE id = ?;