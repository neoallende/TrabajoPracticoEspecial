package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

// ReservarTicket construye y devuelve el handler que procesa solicitudes POST "/tickets".La transacción permite comprobar la disponibilidad y crear la reserva como una operación coordinada.
func ReservarTicket(queries *db.Queries, conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Formulario inválido", http.StatusBadRequest)
			return
		}

		usuarioID, err := strconv.ParseInt(r.FormValue("usuario_id"), 10, 32)

		if err != nil || usuarioID < 1 {
			http.Error(w, "Primero ingresá un ID de usuario válido", http.StatusBadRequest)
			return
		}

		eventoID, err := strconv.ParseInt(r.FormValue("evento_id"), 10, 32)

		if err != nil || eventoID < 1 {
			http.Error(w, "Evento inválido", http.StatusBadRequest)
			return
		}

		// Inicia una transacción en la base de datos usando el contexto de la solicitud. Las consultas asociadas a tx forman parte de esta transacción.
		tx, err := conn.BeginTx(r.Context(), nil)

		if err != nil {
			http.Error(w, "No se pudo iniciar la reserva", http.StatusInternalServerError)
			return
		}

		// Programa un Rollback para cuando la función termine.
		defer tx.Rollback()

		// Crea una instancia de db.Queries que ejecutará sus consultas usando esta transacción, en lugar de ejecutarlas directamente sobre la conexión general.
		txQueries := queries.WithTx(tx)

		// Busca el evento y lo bloquea para evitar que otra transacción modifique simultáneamente su disponibilidad mientras se está realizando esta reserva.
		evento, err := txQueries.BloquearEvento(r.Context(), int32(eventoID))

		if err != nil {
			if err == sql.ErrNoRows {
				// Responde HTTP 404 porque el evento no existe.
				http.Error(w, "No se encontró el evento", http.StatusNotFound)
			} else {
				// Para cualquier otro error, responde HTTP 500.
				http.Error(w, "No se pudo consultar el evento", http.StatusInternalServerError)
			}
			return
		}

		// Consulta, dentro de la misma transacción, cuántos tickets están reservados para el evento cuyo registro se encuentra bloqueado.
		reservadas, err := txQueries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))

		if err != nil {
			http.Error(w, "Error al consultar la disponibilidad", http.StatusInternalServerError)
			return
		}

		// Compara la cantidad de reservas actuales con la capacidad total del evento.
		if reservadas >= int64(evento.Capacidad) {
			// Redirige a la página principal conservando el usuario y agregando un mensaje.
			http.Redirect(
				w,
				r,
				fmt.Sprintf("/?usuario_id=%d&mensaje=%s", usuarioID, url.QueryEscape("El evento no tiene más lugares disponibles.")),
				http.StatusSeeOther,
			)
			return
		}

		// Inserta el ticket asociando el usuario con el evento.
		_, err = txQueries.ReservarTicket(r.Context(), db.ReservarTicketParams{
			IDUsuario: int32(usuarioID),
			IDEvento:  int32(eventoID),
		})

		if err != nil {
			http.Error(w, "No se pudo realizar la reserva.", http.StatusBadRequest)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "No se pudo confirmar la reserva", http.StatusInternalServerError)
			return
		}

		// Aplica Post/Redirect/Get después de reservar correctamente.
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/?usuario_id=%d", usuarioID),
			http.StatusSeeOther,
		)
	}
}

// CancelarTicket construye y devuelve el handler que procesa solicitudes POST "/tickets/{id}/cancelar".

func CancelarTicket(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ticketID, err := strconv.ParseInt(r.PathValue("id"), 10, 32)

		if err != nil || ticketID < 1 {
			http.Error(w, "ID de ticket inválido", http.StatusBadRequest)
			return
		}

		// Obtiene el parámetro "usuario_id" de la query.
		usuarioID, err := strconv.ParseInt(r.URL.Query().Get("usuario_id"), 10, 32)

		if err != nil || usuarioID < 1 {
			http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
			return
		}

		// Ejecuta la consulta que elimina el ticket indicado
		filas, err := queries.EliminarTicket(r.Context(), db.EliminarTicketParams{
			ID:        int32(ticketID),
			IDUsuario: int32(usuarioID),
		})

		if err != nil {
			http.Error(w, "No se pudo cancelar el ticket", http.StatusInternalServerError)
			return
		}

		if filas == 0 {
			http.Error(w, "No se encontró el ticket", http.StatusNotFound)
			return
		}

		// Aplica Post/Redirect/Get después de cancelar el ticket.
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/?usuario_id=%d", usuarioID),
			http.StatusSeeOther,
		)
	}
}
