-- name: CreateWallet :one
INSERT INTO wallets (owner_type, owner_id, asset_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWalletById :one
SELECT *
FROM wallets
WHERE id = $1;

-- name: GetWalletByOwner :one
SELECT *
FROM wallets
WHERE owner_id = $1;

-- name: GetSystemWallet :one
SELECT id
FROM wallets
WHERE owner_type = 'SYSTEM';

-- name: LockWallet :one
SELECT id
FROM wallets
WHERE owner_id = $1
FOR UPDATE;