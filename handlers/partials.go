package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func obtenerUsuarioID(valor string) (int32, error) {
	id, err := strconv.ParseInt(valor, 10, 32)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("ID de usuario inválido")
	}
	return int32(id), nil
}

func obtenerEventoVista(ctx context.Context, queries *db.Queries, id int32) (db.ListarEventosConReservasRow, error) {
	evento, err := queries.ObtenerEventoConReservas(ctx, id)
	if err != nil {
		return db.ListarEventosConReservasRow{}, err
	}
	return db.ListarEventosConReservasRow{ID: evento.ID, Nombre: evento.Nombre, Capacidad: evento.Capacidad, Fecha: evento.Fecha, Reservas: evento.Reservas}, nil
}

func actualizarSesion(w http.ResponseWriter, r *http.Request, queries *db.Queries, usuarioID string) {
	eventos, err := queries.ListarEventosConReservas(r.Context())
	if err != nil {
		responderError(w, "No se pudieron obtener los eventos.")
		return
	}
	esAdmin := false
	rol := "No hay una sesión seleccionada."
	var tickets []db.ListarTicketsPorUsuarioRow
	if usuarioID != "" {
		id, err := obtenerUsuarioID(usuarioID)
		if err != nil {
			responderError(w, err.Error())
			return
		}
		usuario, err := queries.ObtenerUsuarioPorID(r.Context(), id)
		if err != nil {
			responderError(w, "No se encontró el usuario.")
			return
		}
		esAdmin = usuario.Rol == "admin"
		if esAdmin {
			rol = "Sesión seleccionada: Administrador"
		} else {
			rol = "Sesión seleccionada: Usuario"
		}
		tickets, err = queries.ListarTicketsPorUsuario(r.Context(), id)
		if err != nil {
			responderError(w, "No se pudieron obtener los tickets.")
			return
		}
	}
	destino := "/"
	if usuarioID != "" {
		destino += "?usuario_id=" + usuarioID
	}
	w.Header().Set("HX-Push-Url", destino)
	componentes := []templ.Component{
		views.Notificacion("", false),
		views.SelectorUsuario(usuarioID, rol),
		views.PanelFormEvento(usuarioID, esAdmin, nil),
	}
	for _, evento := range eventos {
		componentes = append(componentes, views.EventoItem(evento, usuarioID, esAdmin, "outerHTML"))
	}
	componentes = append(componentes, views.EstadoEventos(len(eventos) == 0))
	componentes = append(componentes, views.PanelTickets(usuarioID, tickets))
	if err := renderizar(w, r, componentes...); err != nil {
		renderizarError(w, r, err)
	}
}
