-- name: CreateCandidate :one
INSERT INTO candidates (
    username,
    full_name,
    phone_number,
    education,
    location,
    skill_set,
    certificates,
    industry_of_interest,
    job_preference,
    time_availability,
    account_verified,
    resume_file,
    profile_photo,
    description
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
         ) RETURNING *;

-- name: GetCandidate :one
SELECT * FROM candidates
WHERE id = $1 LIMIT 1;

-- name: ListCandidates :many
SELECT * FROM candidates
WHERE username = $1
ORDER BY id
LIMIT $2
    OFFSET $3;

-- name: UpdateCandidate :one
UPDATE candidates
SET full_name = COALESCE(sqlc.narg(full_name), full_name),
    phone_number = COALESCE(sqlc.narg(phone_number), phone_number),
    education = COALESCE(sqlc.narg(education), education),
    location = COALESCE(sqlc.narg(location), location),
    skill_set = COALESCE(sqlc.narg(skill_set), skill_set),
    certificates = COALESCE(sqlc.narg(certificates), certificates),
    industry_of_interest = COALESCE(sqlc.narg(industry_of_interest), industry_of_interest),
    job_preference = COALESCE(sqlc.narg(job_preference), job_preference),
    time_availability = COALESCE(sqlc.narg(time_availability), time_availability),
    account_verified = COALESCE(sqlc.narg(account_verified), account_verified),
    resume_file = COALESCE(sqlc.narg(resume_file), resume_file),
    profile_photo = COALESCE(sqlc.narg(profile_photo), profile_photo),
    description = COALESCE(sqlc.narg(description), description)
WHERE
    id = $1
RETURNING *;

-- name: DeleteCandidate :exec
DELETE FROM candidates
WHERE id = $1;

-- name: GetCandidateIdByUsername :one
SELECT id FROM candidates
WHERE username = $1 LIMIT 1;
