-- name: GetUserByID :one
SELECT * FROM user WHERE id = ?;