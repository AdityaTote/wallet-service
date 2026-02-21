-- name: CreateLedger :one
INSERT INTO ledgers (amount, transaction_id, wallet_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLedgerById :one
SELECT *
FROM ledgers
WHERE id = $1;

-- name: GetLedgerByWalletAndTnx :one
SELECT *
FROM ledgers
WHERE wallet_id = $1 AND transaction_id = $2;

-- name: GetLedgersByWalletId :many
SELECT *
FROM ledgers
WHERE wallet_id = $1
ORDER BY
  CASE WHEN $4 = 'asc'  THEN created_at END ASC,
  CASE WHEN $4 = 'desc' THEN created_at END DESC,
  created_at DESC
LIMIT COALESCE($2, 5)
OFFSET COALESCE($3, 0);

-- name: GetBalance :one
SELECT COALESCE(SUM(amount), 0)
FROM ledgers
WHERE wallet_id = $1;