-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id)
VALUES (gen_random_uuid(), $1, $2)
RETURNING *;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id = $1;

-- name: GetAllChirps :many
SELECT * FROM chirps;

-- name: GetUserChirps :many
SELECT * FROM chirps WHERE user_id = $1;

-- name: DeleteChirp :execrows
DELETE FROM chirps WHERE id = $1;
