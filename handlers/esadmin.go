package handlers

import (
	"context"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func EsAdministrador(ctx context.Context, queries *db.Queries, usuarioID string) bool {
	id, err := strconv.ParseInt(usuarioID, 10, 32)
	if err != nil || id <= 0 {
		return false
	}

	usuario, err := queries.ObtenerUsuarioPorID(ctx, int32(id))
	if err != nil {
		return false
	}

	return usuario.Rol == "admin"
}
