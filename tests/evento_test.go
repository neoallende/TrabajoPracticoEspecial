package tests

import (
	"context"
	"testing"
	"time"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

// TestCrearEvento prueba unicamente la creacion de un evento.
func TestCrearEvento(t *testing.T) {
	queries := conectarDB(t)
	ctx := context.Background()

	evento, err := queries.CrearEvento(ctx, db.CrearEventoParams{
		Nombre:    "Recital de Rock",
		Capacidad: 100,
		Fecha:     time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Error al crear evento: %v", err)
	}

	if evento.ID == 0 {
		t.Errorf("Se esperaba un ID valido de evento, se obtuvo 0")
	}
	if evento.Capacidad != 100 {
		t.Errorf("Se esperaba capacidad 100, se obtuvo %d", evento.Capacidad)
	}
}
