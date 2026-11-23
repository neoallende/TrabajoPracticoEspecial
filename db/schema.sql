DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS eventos;
DROP TABLE IF EXISTS usuarios;

CREATE TABLE usuarios (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    rol VARCHAR(20) NOT NULL DEFAULT 'user'
        CHECK (rol IN ('user', 'admin'))
);

CREATE TABLE eventos (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    capacidad INT NOT NULL,
    fecha TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE tickets (
    id SERIAL PRIMARY KEY,
    id_usuario INT NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    id_evento INT NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
    fecha TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Usuario administrador creado previamente.
INSERT INTO usuarios (nombre, email, rol)
VALUES ('Administrador', 'admin@eventos.com', 'admin');

-- Usuarios comunes de prueba
INSERT INTO usuarios (nombre, email)
VALUES
    ('Juan Pérez', 'juan@gmail.com'),
    ('María López', 'maria@gmail.com'),
    ('Pedro Gómez', 'pedro@gmail.com');

-- Eventos de prueba
INSERT INTO eventos (nombre, capacidad, fecha)
VALUES
    ('Concierto de Rock', 100, '2026-12-15 20:00:00-03'),
    ('Festival de Música', 250, '2026-12-20 18:00:00-03'),
    ('Conferencia de Tecnología', 80, '2026-12-22 09:00:00-03'),
    ('Torneo de Videojuegos', 50, '2026-12-28 16:00:00-03'),
    ('Obra de Teatro', 120, '2027-01-05 21:00:00-03');

-- Tickets de prueba
INSERT INTO tickets (id_usuario, id_evento)
VALUES
    (2, 1), 
    (2, 2), 
    (3, 1), 
    (3, 3), 
    (4, 4);