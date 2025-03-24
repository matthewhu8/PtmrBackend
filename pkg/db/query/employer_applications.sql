-- name: CreateEmployerApplication :one
INSERT INTO employer_applications (
    employer_id,
    candidate_id,
    message,
    application_status,
    created_at
) VALUES (
             $1, $2, $3, $4, $5
         ) RETURNING *;

-- name: UpdateEmployerApplication :one
UPDATE employer_applications
SET application_status = COALESCE(sqlc.narg(application_status), application_status)
WHERE employer_id = $1 AND candidate_id = $2
RETURNING *;

-- name: UpdateEmployerApplicationStatus :exec
UPDATE employer_applications
SET application_status = $3
WHERE candidate_id = $1 AND employer_id = $2;

-- name: DeleteEmployerApplication :exec
DELETE FROM employer_applications
WHERE employer_id = $1 AND candidate_id = $2;

-- name: GetEmployerApplicationsByCandidate :many
SELECT * FROM employer_applications
WHERE candidate_id = $1
ORDER BY created_at DESC;

-- name: GetEmployerApplicationsByStatusPending :many
SELECT * FROM employer_applications
WHERE application_status = 'pending' AND candidate_id = $1
ORDER BY created_at DESC;

-- name: GetEmployerApplicationsByStatusSubmitted :many
SELECT * FROM employer_applications
WHERE application_status = 'submitted' AND candidate_id = $1
ORDER BY created_at DESC;

-- name: GetEmployerApplicationsByStatusAccepted :many
SELECT * FROM employer_applications
WHERE application_status = 'accepted' AND candidate_id = $1
ORDER BY created_at DESC;

-- name: GetEmployerApplicationsByStatusRejected :many
SELECT * FROM employer_applications
WHERE application_status = 'rejected' AND candidate_id = $1
ORDER BY created_at DESC;
