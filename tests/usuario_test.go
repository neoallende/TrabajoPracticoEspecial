package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

// TestCrearUsuario prueba unicamente la creacion de un usuario.
func TestCrearUsuario(t *testing.T) {
	queries := conectarDB(t)
	ctx := context.Background()

	email := fmt.Sprintf("carlos-usuario-%d@tango.com", time.Now().UnixNano())
	usuario, err := queries.CrearUsuario(ctx, db.CrearUsuarioParams{
		Nombre: "Carlos Gardel",
		Email:  email,
	})
	if err != nil {
		t.Fatalf("Error al crear usuario: %v", err)
	}

	if usuario.ID == 0 {
		t.Errorf("Se esperaba un ID valido de usuario, se obtuvo 0")
	}
	if usuario.Email != email {
		t.Errorf("Se esperaba el email %q, se obtuvo %q", email, usuario.Email)
	}
}
