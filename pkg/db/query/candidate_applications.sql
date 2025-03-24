-- name: CreateCandidateApplication :one
INSERT INTO candidate_applications (
    candidate_id,
    employer_id,
    elasticsearch_doc_id,
    job_doc_id,
    application_status
) VALUES (
             $1, $2, $3, $4, $5
         ) RETURNING *;

-- name: UpdateCandidateApplication :one
UPDATE candidate_applications
SET application_status = COALESCE(sqlc.narg(application_status), application_status),
    elasticsearch_doc_id = COALESCE(sqlc.narg(elasticsearch_doc_id), elasticsearch_doc_id)
WHERE candidate_id = $1 AND employer_id = $2
RETURNING *;

-- name: UpdateCandidateApplicationStatus :exec
UPDATE candidate_applications
SET application_status = $3
WHERE candidate_id = $1 AND employer_id = $2;

-- name: DeleteCandidateApplication :exec
DELETE FROM candidate_applications
WHERE candidate_id = $1 AND employer_id = $2;

-- name: GetCandidateApplicationsByEmployer :many
SELECT * FROM candidate_applications
WHERE employer_id = $1
ORDER BY created_at DESC;

-- name: GetCandidateApplicationsByStatusPending :many
SELECT * FROM candidate_applications
WHERE application_status = 'pending' AND candidate_id = $1
ORDER BY created_at DESC;

-- name: GetCandidateApplicationsByStatusSubmitted :many
SELECT * FROM candidate_applications
WHERE application_status = 'submitted' AND candidate_id = $1
ORDER BY created_at DESC;


-- name: GetCandidateApplicationsByStatusAccepted :many
SELECT * FROM candidate_applications
WHERE application_status = 'accepted' AND candidate_id = $1
ORDER BY created_at DESC;


-- name: GetCandidateApplicationsByStatusRejected :many
SELECT * FROM candidate_applications
WHERE application_status = 'rejected' AND candidate_id = $1
ORDER BY created_at DESC;
