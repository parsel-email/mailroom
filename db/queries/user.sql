-- name: GetUserByID :one
SELECT * FROM user WHERE id = ?;

-- name: GetUserByProviderID :one
SELECT * FROM user WHERE provider = ? AND provider_id = ?;

-- name: CreateUser :one
INSERT INTO user (
    id,
    email,
    provider,
    provider_id
) VALUES (
    ?,
    ?,
    ?,
    ?
)
RETURNING *;