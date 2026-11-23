-- name: CrearUsuario :one
INSERT INTO usuarios (nombre, email)
VALUES ($1, $2)
RETURNING *;

-- name: EliminarUsuario :execrows
DELETE FROM usuarios
WHERE id = $1;

-- name: ObtenerUsuarioPorID :one
SELECT *
FROM usuarios
WHERE id = $1;