-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, created_at, updated_at, expires_at)
VALUES ($1, $2, now(), now(), $3)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now(),
    updated_at = now()
WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;