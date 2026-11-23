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

-- name: EliminarEvento :execrows
DELETE FROM eventos
WHERE id = $1;
-- name: ListarEventosConReservas :many
SELECT e.id, e.nombre, e.capacidad, e.fecha, COUNT(t.id)::BIGINT AS reservas
FROM eventos e
LEFT JOIN tickets t ON t.id_evento = e.id
GROUP BY e.id
ORDER BY e.fecha ASC;


-- name: ObtenerEventoConReservas :one
SELECT e.id, e.nombre, e.capacidad, e.fecha, COUNT(t.id)::BIGINT AS reservas
FROM eventos e
LEFT JOIN tickets t ON t.id_evento = e.id
WHERE e.id = $1
GROUP BY e.id;
