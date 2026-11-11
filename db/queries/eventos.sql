-- name: CrearEvento :one
INSERT INTO eventos (nombre, capacidad, fecha)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ObtenerEventoPorID :one
SELECT * FROM eventos
WHERE id = $1 LIMIT 1;

-- name: ListarEventos :many
SELECT * FROM eventos
ORDER BY fecha ASC;

-- name: EditarEvento :one
UPDATE eventos
SET nombre = $2,
    capacidad = $3,
    fecha = $4
WHERE id = $1
RETURNING *;

-- name: BloquearEvento :one
-- Bloquea la fila del evento durante la transacción para coordinar reservas concurrentes.
SELECT * FROM eventos
WHERE id = $1
FOR UPDATE;
