-- name: GetUserFiles :many
SELECT * FROM files 
WHERE ownerid = $1;

-- name: GetAllFiles :many
SELECT * FROM files;

-- name: PutFile :exec
INSERT INTO files 
(ownerid, objkey, filename, id, size, expiresat)
VALUES
($1, $2, $3, $4, $5, $6);

-- name: DeleteFile :exec
DELETE FROM files 
WHERE objkey = $1;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUserSpace :one
SELECT space FROM users 
WHERE username = $1;

-- name: PutUser :exec
INSERT INTO users (username, space)
VALUES ($1, 0)
    ON CONFLICT (username) DO NOTHING;

-- name: RecalculateUserSpace :exec
UPDATE users SET space = coalesce((
    SELECT sum(size)
    FROM files
    WHERE files.ownerid = users.username
),0)
WHERE username = $1;
