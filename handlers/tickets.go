package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func ReservarTicket(queries *db.Queries, conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Formulario inválido", http.StatusBadRequest)
			return
		}
		usuarioID, err := strconv.ParseInt(r.FormValue("usuario_id"), 10, 32)
		if err != nil || usuarioID < 1 {
			http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
			return
		}
		usuario, err := queries.ObtenerUsuarioPorID(r.Context(), int32(usuarioID))
		if err != nil || usuario.Rol != "user" {
			http.Error(w, "La cuenta no puede reservar tickets", http.StatusForbidden)
			return
		}
		eventoID, err := strconv.ParseInt(r.FormValue("evento_id"), 10, 32)
		if err != nil || eventoID < 1 {
			http.Error(w, "Evento inválido", http.StatusBadRequest)
			return
		}

		tx, err := conn.BeginTx(r.Context(), nil)
		if err != nil {
			http.Error(w, "No se pudo iniciar la reserva", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()
		txQueries := queries.WithTx(tx)

		evento, err := txQueries.BloquearEvento(r.Context(), int32(eventoID))
		if err != nil {
			http.Error(w, "No se encontró el evento", http.StatusNotFound)
			return
		}
		reservadas, err := txQueries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))
		if err != nil {
			http.Error(w, "Error al consultar la disponibilidad", http.StatusInternalServerError)
			return
		}
		if reservadas >= int64(evento.Capacidad) {
			if r.Header.Get("HX-Request") == "true" {
				renderHomeFragment(w, r, queries, fmt.Sprint(usuarioID), "", "El evento no tiene más lugares disponibles.")
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/?usuario_id=%d&mensaje=%s", usuarioID, url.QueryEscape("El evento no tiene más lugares disponibles.")), http.StatusSeeOther)
			return
		}
		_, err = txQueries.ReservarTicket(r.Context(), db.ReservarTicketParams{IDUsuario: int32(usuarioID), IDEvento: int32(eventoID)})
		if err != nil {
			http.Error(w, "No se pudo realizar la reserva", http.StatusBadRequest)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, "No se pudo confirmar la reserva", http.StatusInternalServerError)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			renderHomeFragment(w, r, queries, fmt.Sprint(usuarioID), "", "Ticket reservado correctamente.")
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/?usuario_id=%d", usuarioID), http.StatusSeeOther)
	}
}

func CancelarTicket(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketID, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || ticketID < 1 {
			http.Error(w, "ID de ticket inválido", http.StatusBadRequest)
			return
		}
		usuarioID, err := strconv.ParseInt(r.URL.Query().Get("usuario_id"), 10, 32)
		if err != nil || usuarioID < 1 {
			usuarioID, err = strconv.ParseInt(r.FormValue("usuario_id"), 10, 32)
		}
		if err != nil || usuarioID < 1 {
			http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
			return
		}
		filas, err := queries.EliminarTicket(r.Context(), db.EliminarTicketParams{ID: int32(ticketID), IDUsuario: int32(usuarioID)})
		if err != nil {
			http.Error(w, "No se pudo cancelar el ticket", http.StatusInternalServerError)
			return
		}
		if filas == 0 {
			http.Error(w, "No se encontró el ticket", http.StatusNotFound)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			renderHomeFragment(w, r, queries, fmt.Sprint(usuarioID), "", "Reserva cancelada correctamente.")
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/?usuario_id=%d", usuarioID), http.StatusSeeOther)
	}
}
