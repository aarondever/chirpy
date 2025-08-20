-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps 
WHERE @userID::TEXT = '' OR user_id::TEXT = @userID
ORDER BY CASE 
            WHEN @sort::TEXT = 'desc' THEN created_at
            END DESC,
        CASE
            WHEN @sort IS NULL OR @sort::TEXT = 'asc' THEN created_at
            END ASC;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1 AND user_id = $2;

-- name: GetChirpById :one
SELECT * FROM chirps WHERE id = $1;