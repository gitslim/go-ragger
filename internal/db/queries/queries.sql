-- name: UserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = @email;

