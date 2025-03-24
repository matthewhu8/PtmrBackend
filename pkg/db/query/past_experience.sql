-- name: CreatePastExperience :one
INSERT INTO past_experiences (
    candidate_id,
    industry,
    employer,
    job_title,
    start_date,
    end_date,
    present,
    description
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8
         ) RETURNING *;

-- name: GetPastExperience :one
SELECT * FROM past_experiences
WHERE id = $1 LIMIT 1;

-- name: ListPastExperiences :many
SELECT * FROM past_experiences
WHERE candidate_id = $1
ORDER BY id
LIMIT $2
    OFFSET $3;

-- name: UpdatePastExperience :one
UPDATE past_experiences
SET industry = COALESCE(sqlc.narg(industry), industry),
    employer = COALESCE(sqlc.narg(employer), employer),
    job_title = COALESCE(sqlc.narg(job_title), job_title),
    start_date = COALESCE(sqlc.narg(start_date), start_date),
    end_date = COALESCE(sqlc.narg(end_date), end_date),
    present = COALESCE(sqlc.narg(present), present),
    description = COALESCE(sqlc.narg(description), description)
WHERE
    id = $1
RETURNING *;

-- name: DeletePastExperience :exec
DELETE FROM past_experiences
WHERE id = $1 and candidate_id = $2;
