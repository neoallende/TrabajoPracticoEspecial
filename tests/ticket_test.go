package tests

import (
	"context"
	"testing"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

// TestReservarTicket prueba unicamente la reserva de un ticket, usando
// un usuario y un evento de prueba creados a traves de los helpers.
func TestReservarTicket(t *testing.T) {
	queries := conectarDB(t)
	ctx := context.Background()

	usuario := crearUsuarioDePrueba(t, queries, "ticket")
	evento := crearEventoDePrueba(t, queries, "Recital de Rock")

	ticket, err := queries.ReservarTicket(ctx, db.ReservarTicketParams{
		IDUsuario: usuario.ID,
		IDEvento:  evento.ID,
	})
	if err != nil {
		t.Fatalf("Error al reservar ticket: %v", err)
	}

	if ticket.IDUsuario != usuario.ID {
		t.Errorf("El ticket debia pertenecer al usuario %d, pertenece a %d", usuario.ID, ticket.IDUsuario)
	}
	if ticket.IDEvento != evento.ID {
		t.Errorf("El ticket debia pertenecer al evento %d, pertenece a %d", evento.ID, ticket.IDEvento)
	}
}
