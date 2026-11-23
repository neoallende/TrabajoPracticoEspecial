package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func GuardarEvento(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usuarioID := r.FormValue("usuario_id")
		if !EsAdministrador(r.Context(), queries, usuarioID) {
			http.Error(w, "No tenés permisos para realizar esta acción", http.StatusForbidden)
			return
		}
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
		fecha, err := time.ParseInLocation("2006-01-02T15:04", r.FormValue("fecha"), time.Local)
		if err != nil {
			http.Error(w, "Fecha inválida", http.StatusBadRequest)
			return
		}

		id := r.FormValue("id")
		if id == "" {
			_, err = queries.CrearEvento(r.Context(), db.CrearEventoParams{Nombre: nombre, Capacidad: int32(capacidad), Fecha: fecha})
		} else {
			eventoID, parseErr := strconv.ParseInt(id, 10, 32)
			if parseErr != nil || eventoID < 1 {
				http.Error(w, "ID de evento inválido", http.StatusBadRequest)
				return
			}
			reservadas, errReserva := queries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))
			if errReserva != nil {
				http.Error(w, "No se pudieron consultar las reservas del evento", http.StatusInternalServerError)
				return
			}
			if int64(capacidad) < reservadas {
				http.Error(w, fmt.Sprintf("La capacidad no puede ser menor que las reservas existentes (%d).", reservadas), http.StatusBadRequest)
				return
			}
			_, err = queries.EditarEvento(r.Context(), db.EditarEventoParams{ID: int32(eventoID), Nombre: nombre, Capacidad: int32(capacidad), Fecha: fecha})
		}
		if err != nil {
			http.Error(w, "No se pudo guardar el evento", http.StatusInternalServerError)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			renderHomeFragment(w, r, queries, usuarioID, "", "Evento guardado correctamente.")
			return
		}
		http.Redirect(w, r, "/?usuario_id="+usuarioID, http.StatusSeeOther)
	}
}

func EliminarEvento(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usuarioID := r.URL.Query().Get("usuario_id")
		if usuarioID == "" {
			usuarioID = r.FormValue("usuario_id")
		}
		if !EsAdministrador(r.Context(), queries, usuarioID) {
			http.Error(w, "No tenés permisos para realizar esta acción", http.StatusForbidden)
			return
		}
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || id < 1 {
			http.Error(w, "ID de evento inválido", http.StatusBadRequest)
			return
		}
		filas, err := queries.EliminarEvento(r.Context(), int32(id))
		if err != nil {
			http.Error(w, "No se pudo eliminar el evento", http.StatusInternalServerError)
			return
		}
		if filas == 0 {
			http.Error(w, "No se encontró el evento", http.StatusNotFound)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			renderHomeFragment(w, r, queries, usuarioID, "", "Evento eliminado correctamente.")
			return
		}
		http.Redirect(w, r, "/?usuario_id="+usuarioID, http.StatusSeeOther)
	}
}
