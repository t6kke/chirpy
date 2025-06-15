-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, revoked_at, user_id)
VALUES ($1, NOW(), NOW(), $2, NULL, $3)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT user_id from refresh_tokens
WHERE revoked_at is NULL and expires_at > NOW() and token = $1;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1
RETURNING *;
