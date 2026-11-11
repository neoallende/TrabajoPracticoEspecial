package tests

// Este archivo NO contiene tests en si mismos (no hay funciones Test*),
// sino funciones auxiliares que usan los demas archivos *_test.go para
// no repetir la logica de conexion a la base ni la creacion de datos
// de prueba (usuarios/eventos) que necesitan otros tests como entrada.

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	// esto importa el codigo Go autogenerado con sqlc
	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"

	_ "github.com/lib/pq" // Driver para Postgres
)

// conectarDB abre una conexion a la base de datos de testing (levantada
// por docker-compose) y devuelve las Queries generadas por sqlc listas
// para usar. La conexion se cierra sola al terminar el test gracias a
// t.Cleanup, asi que no hace falta hacer defer en cada test.
func conectarDB(t *testing.T) *db.Queries {
	t.Helper()

	connStr := "postgres://root:secretpassword@localhost:5432/db?sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("No se pudo conectar a la DB: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return db.New(conn)
}

// crearUsuarioDePrueba crea un usuario auxiliar para usar como dato de
// entrada en tests que lo necesiten (por ejemplo, para reservar un
// ticket). El "sufijo" se usa para que el email generado sea unico y no
// choque con el de otro test que corra contra la misma base.
func crearUsuarioDePrueba(t *testing.T, queries *db.Queries, sufijo string) db.Usuario {
	t.Helper()
	ctx := context.Background()

	email := fmt.Sprintf("carlos-%s-%d@tango.com", sufijo, time.Now().UnixNano())
	usuario, err := queries.CrearUsuario(ctx, db.CrearUsuarioParams{
		Nombre: "Carlos Gardel",
		Email:  email,
	})
	if err != nil {
		t.Fatalf("Error al crear usuario de prueba: %v", err)
	}

	return usuario
}

// crearEventoDePrueba crea un evento auxiliar para usar como dato de
// entrada en tests que lo necesiten (por ejemplo, para reservar un
// ticket).
func crearEventoDePrueba(t *testing.T, queries *db.Queries, nombre string) db.Evento {
	t.Helper()
	ctx := context.Background()

	evento, err := queries.CrearEvento(ctx, db.CrearEventoParams{
		Nombre:    nombre,
		Capacidad: 100,
		Fecha:     time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Error al crear evento de prueba: %v", err)
	}

	return evento
}
