-- name: CrearUsuario :one
INSERT INTO usuarios (nombre, email)
VALUES ($1, $2)
RETURNING *;
