-- name: CreateUser :one
INSERT INTO users (
  id,
  created_at,
  updated_at,
  email,
  hashed_password
) VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING users.id, users.created_at, users.updated_at, users.email;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
where users.email = $1;
