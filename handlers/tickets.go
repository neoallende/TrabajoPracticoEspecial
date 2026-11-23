package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func ReservarTicket(queries *db.Queries, conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !exigirHTMX(w, r) {
			return
		}
		if err := r.ParseForm(); err != nil {
			responderError(w, "Formulario inválido.")
			return
		}
		usuarioID, err := obtenerUsuarioID(r.FormValue("usuario_id"))
		if err != nil {
			responderError(w, "Primero seleccioná un usuario válido.")
			return
		}
		eventoID, err := strconv.ParseInt(r.FormValue("evento_id"), 10, 32)
		if err != nil || eventoID < 1 {
			responderError(w, "Evento inválido.")
			return
		}
		tx, err := conn.BeginTx(r.Context(), nil)
		if err != nil {
			responderError(w, "No se pudo iniciar la reserva.")
			return
		}
		defer tx.Rollback()
		txQueries := queries.WithTx(tx)
		evento, err := txQueries.BloquearEvento(r.Context(), int32(eventoID))
		if err != nil {
			responderError(w, "No se encontró el evento.")
			return
		}
		reservadas, err := txQueries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))
		if err != nil {
			responderError(w, "No se pudo consultar la disponibilidad.")
			return
		}
		if reservadas >= int64(evento.Capacidad) {
			responderError(w, "El evento no tiene más lugares disponibles.")
			return
		}
		ticket, err := txQueries.ReservarTicket(r.Context(), db.ReservarTicketParams{IDUsuario: usuarioID, IDEvento: int32(eventoID)})
		if err != nil {
			responderError(w, "No se pudo realizar la reserva.")
			return
		}
		if err := tx.Commit(); err != nil {
			responderError(w, "No se pudo confirmar la reserva.")
			return
		}
		vista, err := obtenerEventoVista(r.Context(), queries, int32(eventoID))
		if err != nil {
			responderError(w, "La reserva se realizó, pero no se pudo actualizar el evento.")
			return
		}
		tickets, err := queries.ListarTicketsPorUsuario(r.Context(), usuarioID)
		if err != nil {
			responderError(w, "La reserva se realizó, pero no se pudo actualizar la lista de tickets.")
			return
		}
		var ticketVista db.ListarTicketsPorUsuarioRow
		for _, t := range tickets {
			if t.TicketID == ticket.ID {
				ticketVista = t
				break
			}
		}
		if err := renderizar(w, r,
			views.Notificacion("", false),
			views.EventoItem(vista, strconv.Itoa(int(usuarioID)), false, "outerHTML"),
			views.TicketItem(ticketVista, strconv.Itoa(int(usuarioID)), "beforeend:#lista-tickets"),
			views.EstadoTickets(false, strconv.Itoa(int(usuarioID))),
			views.EstadoEventos(false),
		); err != nil {
			renderizarError(w, r, err)
		}
	}
}

func CancelarTicket(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !exigirHTMX(w, r) {
			return
		}
		ticketID, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || ticketID < 1 {
			responderError(w, "ID de ticket inválido.")
			return
		}
		usuarioID, err := obtenerUsuarioID(r.URL.Query().Get("usuario_id"))
		if err != nil {
			responderError(w, "ID de usuario inválido.")
			return
		}
		ticket, err := queries.ObtenerTicketPorID(r.Context(), int32(ticketID))
		if err != nil || ticket.IDUsuario != usuarioID {
			responderError(w, "No se encontró el ticket.")
			return
		}
		filas, err := queries.EliminarTicket(r.Context(), db.EliminarTicketParams{ID: int32(ticketID), IDUsuario: usuarioID})
		if err != nil {
			responderError(w, "No se pudo cancelar el ticket.")
			return
		}
		if filas == 0 {
			responderError(w, "No se encontró el ticket.")
			return
		}
		vista, err := obtenerEventoVista(r.Context(), queries, ticket.IDEvento)
		if err != nil {
			responderError(w, "El ticket se canceló, pero no se pudo actualizar el evento.")
			return
		}
		tickets, err := queries.ListarTicketsPorUsuario(r.Context(), usuarioID)
		if err != nil {
			responderError(w, "El ticket se canceló, pero no se pudo actualizar la lista.")
			return
		}
		if err := renderizar(w, r,
			views.Notificacion("", false),
			views.TicketEliminar(int32(ticketID)),
			views.EventoItem(vista, strconv.Itoa(int(usuarioID)), false, "outerHTML"),
			views.EstadoTickets(len(tickets) == 0, strconv.Itoa(int(usuarioID))),
		); err != nil {
			renderizarError(w, r, err)
		}
	}
}
