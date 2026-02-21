-- name: CreateTxn :one
INSERT INTO transactions (id, type)
VALUES ($1, $2)
RETURNING *;

-- name: GetTransactionById :one
SELECT *
FROM transactions
WHERE id = $1;

-- name: GetTransactionByType :many
SELECT *
FROM transactions
WHERE type = $1
ORDER BY
  CASE WHEN $4 = 'asc' THEN created_at END ASC,
  CASE WHEN $4 = 'desc' THEN created_at END DESC,
  created_at DESC
LIMIT COALESCE($2, 5)
OFFSET COALESCE($3, 0);