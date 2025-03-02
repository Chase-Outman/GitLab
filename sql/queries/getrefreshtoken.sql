-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE $1 = token;