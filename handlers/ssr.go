package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func Home(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventos, err := queries.ListarEventos(r.Context())
		if err != nil {
			http.Error(w, "Error al obtener los eventos", http.StatusInternalServerError)
			return
		}

		if eventos == nil {
			eventos = []db.Evento{}
		}

		reservas := make(map[int32]int64)
		for _, evento := range eventos {
			total, err := queries.ObtenerCapacidadReservada(r.Context(), evento.ID)
			if err != nil {
				http.Error(w, "Error al consultar las reservas", http.StatusInternalServerError)
				return
			}
			reservas[evento.ID] = total
		}

		usuarioID := r.URL.Query().Get("usuario_id")
		mensaje := r.URL.Query().Get("mensaje")
		var tickets []db.ListarTicketsPorUsuarioRow

		if usuarioID != "" {
			id, err := strconv.ParseInt(usuarioID, 10, 32)
			if err != nil {
				http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
				return
			}

			tickets, err = queries.ListarTicketsPorUsuario(r.Context(), int32(id))
			if err != nil {
				http.Error(w, "Error al obtener los tickets", http.StatusInternalServerError)
				return
			}
		}

		if tickets == nil {
			tickets = []db.ListarTicketsPorUsuarioRow{}
		}

		var eventoEditar *db.Evento
		if editar := r.URL.Query().Get("editar"); editar != "" {
			id, err := strconv.ParseInt(editar, 10, 32)
			if err != nil {
				http.Error(w, "ID de evento inválido", http.StatusBadRequest)
				return
			}

			evento, err := queries.ObtenerEventoPorID(r.Context(), int32(id))
			if err != nil {
				http.Error(w, "No se pudo encontrar el evento", http.StatusNotFound)
				return
			}
			eventoEditar = &evento
		}

		err = views.Home(
			eventos,
			reservas,
			usuarioID,
			tickets,
			eventoEditar,
			mensaje,
		).Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error al renderizar la página", http.StatusInternalServerError)
		}
	}
}

func CrearUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Formulario inválido", http.StatusBadRequest)
			return
		}

		nombre := r.FormValue("nombre")
		email := r.FormValue("email")

		if nombre == "" || email == "" {
			http.Error(w, "Completá todos los campos", http.StatusBadRequest)
			return
		}

		usuario, err := queries.CrearUsuario(r.Context(), db.CrearUsuarioParams{
			Nombre: nombre,
			Email:  email,
		})
		if err != nil {
			http.Error(w, "Error al crear el usuario. Revisá si el email ya existe.", http.StatusInternalServerError)
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/?usuario_id=%d", usuario.ID),
			http.StatusSeeOther,
		)
	}
}

func GuardarEvento(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Formulario inválido", http.StatusBadRequest)
			return
		}

		nombre := r.FormValue("nombre")
		capacidad, err := strconv.Atoi(r.FormValue("capacidad"))
		if err != nil || capacidad <= 0 || nombre == "" {
			http.Error(w, "Datos del evento inválidos", http.StatusBadRequest)
			return
		}

		fecha, err := time.ParseInLocation(
			"2006-01-02T15:04",
			r.FormValue("fecha"),
			time.Local,
		)
		if err != nil {
			http.Error(w, "Fecha inválida", http.StatusBadRequest)
			return
		}

		id := r.FormValue("id")

		if id == "" {
			_, err = queries.CrearEvento(r.Context(), db.CrearEventoParams{
				Nombre:    nombre,
				Capacidad: int32(capacidad),
				Fecha:     fecha,
			})
		} else {
			eventoID, parseErr := strconv.Atoi(id)
			if parseErr != nil {
				http.Error(w, "ID de evento inválido", http.StatusBadRequest)
				return
			}

			reservadas, err := queries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))
			if err != nil {
				http.Error(w, "No se pudieron consultar las reservas del evento", http.StatusInternalServerError)
				return
			}
			if int64(capacidad) < reservadas {
				http.Error(
					w,
					fmt.Sprintf("La capacidad no puede ser menor que las reservas existentes (%d).", reservadas),
					http.StatusBadRequest,
				)
				return
			}

			_, err = queries.EditarEvento(r.Context(), db.EditarEventoParams{
				ID:        int32(eventoID),
				Nombre:    nombre,
				Capacidad: int32(capacidad),
				Fecha:     fecha,
			})
		}

		if err != nil {
			http.Error(w, "No se pudo guardar el evento", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/?usuario_id="+r.FormValue("usuario_id"), http.StatusSeeOther)
	}
}

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

		tx, err := conn.BeginTx(r.Context(), nil)
		if err != nil {
			http.Error(w, "No se pudo iniciar la reserva", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		txQueries := queries.WithTx(tx)

		evento, err := txQueries.BloquearEvento(r.Context(), int32(eventoID))
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No se encontró el evento", http.StatusNotFound)
			} else {
				http.Error(w, "No se pudo consultar el evento", http.StatusInternalServerError)
			}
			return
		}

		reservadas, err := txQueries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))
		if err != nil {
			http.Error(w, "Error al consultar la disponibilidad", http.StatusInternalServerError)
			return
		}

		if reservadas >= int64(evento.Capacidad) {
			http.Redirect(
				w,
				r,
				fmt.Sprintf("/?usuario_id=%d&mensaje=%s", usuarioID, url.QueryEscape("El evento no tiene más lugares disponibles.")),
				http.StatusSeeOther,
			)
			return
		}

		_, err = txQueries.ReservarTicket(r.Context(), db.ReservarTicketParams{
			IDUsuario: int32(usuarioID),
			IDEvento:  int32(eventoID),
		})
		if err != nil {
			http.Error(w, "No se pudo realizar la reserva. Revisá el ID de usuario.", http.StatusBadRequest)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "No se pudo confirmar la reserva", http.StatusInternalServerError)
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/?usuario_id=%d", usuarioID),
			http.StatusSeeOther,
		)
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
			http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
			return
		}

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

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/?usuario_id=%d", usuarioID),
			http.StatusSeeOther,
		)
	}
}
