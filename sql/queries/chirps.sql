-- name: CreateChirps :one

INSERT INTO chirps (id, created_at, updated_at, body, user_id) 
VALUES (
    gen_random_uuid(), 
    now(), 
    now(), 
    $1, 
    $2)
    
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: ListChirps :many
SELECT * FROM chirps
WHERE sqlc.narg('author_id')::uuid IS NULL
    OR user_id = sqlc.narg('author_id')
ORDER BY
    CASE WHEN @sort_desc::bool THEN created_at END DESC,
    created_at ASC;

-- name: GetSingleChirp :one
SELECT * FROM chirps WHERE id = $1;

-- name: GetChirpsByAuthor :many
SELECT * FROM chirps WHERE user_id = $1 ORDER BY created_at ASC;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1 AND user_id = $2;