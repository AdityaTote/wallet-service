-- name: GetAssetById :one
SELECT *
FROM assets
WHERE id = $1;

-- name: GetAssetByCode :one
SELECT *
FROM assets
WHERE code = $1;