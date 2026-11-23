package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func GuardarEvento(queries *db.Queries, conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !exigirHTMX(w, r) {
			return
		}
		if err := r.ParseForm(); err != nil {
			responderError(w, "Formulario inválido.")
			return
		}
		usuarioID := r.FormValue("usuario_id")
		if !EsAdministrador(r.Context(), queries, usuarioID) {
			responderError(w, "No tenés permisos para realizar esta acción.")
			return
		}
		nombre := r.FormValue("nombre")
		capacidad, err := strconv.Atoi(r.FormValue("capacidad"))
		if err != nil || capacidad < 1 || nombre == "" {
			responderError(w, "Completá correctamente los datos del evento.")
			return
		}
		fecha, err := time.ParseInLocation("2006-01-02T15:04", r.FormValue("fecha"), time.Local)
		if err != nil {
			responderError(w, "La fecha del evento es inválida.")
			return
		}
		idText := r.FormValue("id")
		var evento db.Evento
		esNuevo := idText == ""
		if esNuevo {
			evento, err = queries.CrearEvento(r.Context(), db.CrearEventoParams{Nombre: nombre, Capacidad: int32(capacidad), Fecha: fecha})
		} else {
			id, parseErr := strconv.ParseInt(idText, 10, 32)
			if parseErr != nil || id < 1 {
				responderError(w, "ID de evento inválido.")
				return
			}

			// La edición bloquea la misma fila que bloquean las reservas.
			// De esta manera, capacidad y cantidad de tickets se verifican
			// dentro de una única sección crítica.
			tx, txErr := conn.BeginTx(r.Context(), nil)
			if txErr != nil {
				responderError(w, "No se pudo iniciar la edición del evento.")
				return
			}
			defer tx.Rollback()
			txQueries := queries.WithTx(tx)
			if _, err = txQueries.BloquearEvento(r.Context(), int32(id)); err != nil {
				responderError(w, "No se encontró el evento.")
				return
			}
			reservadas, countErr := txQueries.ObtenerCapacidadReservada(r.Context(), int32(id))
			if countErr != nil {
				responderError(w, "No se pudieron consultar las reservas.")
				return
			}
			if int64(capacidad) < reservadas {
				responderError(w, fmt.Sprintf("La capacidad no puede ser menor que las reservas existentes (%d).", reservadas))
				return
			}
			evento, err = txQueries.EditarEvento(r.Context(), db.EditarEventoParams{ID: int32(id), Nombre: nombre, Capacidad: int32(capacidad), Fecha: fecha})
			if err != nil {
				responderError(w, "No se pudo guardar el evento.")
				return
			}
			if err = tx.Commit(); err != nil {
				responderError(w, "No se pudo confirmar la edición del evento.")
				return
			}
		}
		if err != nil {
			responderError(w, "No se pudo guardar el evento.")
			return
		}
		vista, err := obtenerEventoVista(r.Context(), queries, evento.ID)
		if err != nil {
			responderError(w, "El evento se guardó, pero no se pudo actualizar la vista.")
			return
		}
		w.Header().Set("HX-Push-Url", "/?usuario_id="+usuarioID)
		swap := "outerHTML"
		if esNuevo {
			swap = "beforeend:#lista-eventos"
		}
		if err := renderizar(w, r,
			views.Notificacion("", false),
			views.EventoItem(vista, usuarioID, true, swap),
			views.EstadoEventos(false),
			views.PanelFormEvento(usuarioID, true, nil),
		); err != nil {
			renderizarError(w, r, err)
		}
	}
}

func EliminarEvento(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !exigirHTMX(w, r) {
			return
		}
		usuarioID := r.URL.Query().Get("usuario_id")
		if !EsAdministrador(r.Context(), queries, usuarioID) {
			responderError(w, "No tenés permisos para realizar esta acción.")
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || id < 1 {
			responderError(w, "ID de evento inválido.")
			return
		}
		uid, err := obtenerUsuarioID(usuarioID)
		if err != nil {
			responderError(w, "ID de usuario inválido.")
			return
		}
		tickets, err := queries.ListarTicketsPorUsuario(r.Context(), uid)
		if err != nil {
			responderError(w, "No se pudieron consultar los tickets asociados al usuario.")
			return
		}
		filas, err := queries.EliminarEvento(r.Context(), int32(id))
		if err != nil {
			responderError(w, "No se pudo eliminar el evento.")
			return
		}
		if filas == 0 {
			responderError(w, "No se encontró el evento.")
			return
		}
		eventos, err := queries.ListarEventosConReservas(r.Context())
		if err != nil {
			responderError(w, "El evento se eliminó, pero no se pudo actualizar la lista.")
			return
		}
		actuales, err := queries.ListarTicketsPorUsuario(r.Context(), uid)
		if err != nil {
			responderError(w, "El evento se eliminó, pero no se pudo actualizar la lista de tickets.")
			return
		}
		componentes := []templ.Component{
			views.Notificacion("", false),
			views.EventoEliminar(int32(id)),
			views.EstadoEventos(len(eventos) == 0),
		}
		for _, ticket := range tickets {
			if ticket.EventoID == int32(id) {
				componentes = append(componentes, views.TicketEliminar(ticket.TicketID))
			}
		}
		componentes = append(componentes, views.EstadoTickets(len(actuales) == 0, usuarioID))
		if err := renderizar(w, r, componentes...); err != nil {
			renderizarError(w, r, err)
		}
	}
}
