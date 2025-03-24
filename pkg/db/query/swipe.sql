-- name: CreateCandidateSwipe :exec
INSERT INTO candidate_swipes (
    candidate_id,
    job_id,
    swipe
) VALUES (
             $1, $2, $3
         );

-- name: DeleteCandidateSwipe :exec
DELETE FROM candidate_swipes
WHERE candidate_id = $1 AND job_id = $2;

-- name: GetCandidateSwipe :many
SELECT * FROM candidate_swipes
WHERE job_id = $1 and candidate_id = $2
ORDER BY created_at DESC;

-- name: GetJobIDsByCandidate :many
SELECT job_id
FROM candidate_swipes
WHERE candidate_id = $1
ORDER BY created_at DESC;

-- name: GetRejectedJobIdsByCandidate :many
SELECT job_id
FROM candidate_swipes
WHERE candidate_id = $1 AND swipe = 'reject';

-- name: CreateEmployerSwipes :exec
INSERT INTO employer_swipes (
    employer_id,
    candidate_id,
    swipe
) VALUES (
             $1, $2, $3
         );

-- name: DeleteEmployerSwipe :exec
DELETE FROM employer_swipes
WHERE employer_id = $1 AND candidate_id = $2;

-- name: GetEmployerSwipe :one
SELECT * FROM employer_swipes
WHERE candidate_id = $1 and employer_id = $2
ORDER BY created_at DESC;

-- name: GetCandidateIDsByEmployer :many
SELECT candidate_id
FROM employer_swipes
WHERE employer_id = $1
ORDER BY created_at DESC;

-- name: GetRejectedCandidateIdsByEmployer :many
SELECT candidate_id
FROM employer_swipes
WHERE employer_id = $1 AND swipe = 'reject';