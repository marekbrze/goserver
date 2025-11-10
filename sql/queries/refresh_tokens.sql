-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTokenInfo :one
SELECT t1.*, t2.id, t2.email
FROM refresh_tokens AS t1
INNER JOIN users AS t2 ON t1.user_id = t2.id
WHERE t1.token = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = $1, updated_at = $2
WHERE token = $3;
