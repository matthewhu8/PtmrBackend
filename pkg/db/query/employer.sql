-- name: CreateEmployer :one
INSERT INTO employers (
    username,
    business_name,
    business_email,
    business_phone,
    location,
    industry,
    profile_photo,
    business_description
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8
         ) RETURNING *;

-- name: GetEmployer :one
SELECT * FROM employers
WHERE id = $1 LIMIT 1;

-- name: ListEmployers :many
SELECT * FROM employers
WHERE username = $1
ORDER BY id
LIMIT $2
    OFFSET $3;

-- name: UpdateEmployer :one
UPDATE employers
SET business_name =  COALESCE(sqlc.narg(business_name), business_name),
    business_email = COALESCE(sqlc.narg(business_email), business_email),
    business_phone = COALESCE(sqlc.narg(business_phone), business_phone),
    location = COALESCE(sqlc.narg(location), location),
    industry = COALESCE(sqlc.narg(industry), industry),
    profile_photo = COALESCE(sqlc.narg(profile_photo), profile_photo),
    business_description = COALESCE(sqlc.narg(business_description), business_description)
WHERE
    id = $1
RETURNING *;

-- name: DeleteEmployer :exec
DELETE FROM employers
WHERE id = $1;

-- name: GetEmployerIdByUsername :one
SELECT id FROM employers
WHERE username = $1 LIMIT 1;

-- name: AddJobListing :exec
UPDATE employers
SET job_listings = array_append(job_listings, sqlc.narg(job_id))
WHERE id = $1;
