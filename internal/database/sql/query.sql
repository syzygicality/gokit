-- name: ListAliasesByAlias :many
SELECT * FROM aliases ORDER BY alias DESC;

-- name: ListAliasesByRecency :many
SELECT * FROM aliases ORDER BY updated_at DESC;

-- name: ListAliasesByCreation :many
SELECT * FROM aliases ORDER BY created_at DESC;

-- name: GetAlias :one
SELECT * FROM aliases WHERE alias = ?;

-- name: UpsertAlias :exec
INSERT INTO aliases (alias, repo_url, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(alias)
DO UPDATE SET 
    repo_url = excluded.repo_url,
    updated_at = excluded.updated_at;

-- name: DeleteAlias :execrows
DELETE FROM aliases WHERE alias = ?;

