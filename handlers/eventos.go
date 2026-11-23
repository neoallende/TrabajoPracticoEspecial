package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

// GuardarEvento construye y devuelve el handler que procesa solicitudes POST "/eventos". El mismo handler permite crear un evento nuevo o modificar uno existente. La decisión se toma según venga vacío o no el campo "id" del formulario.
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

		// time.Local hace que la fecha se interprete usando la zona horaria local del servidor.
		fecha, err := time.ParseInLocation(
			"2006-01-02T15:04",
			r.FormValue("fecha"),
			time.Local,
		)

		if err != nil {
			http.Error(w, "Fecha inválida", http.StatusBadRequest)
			return
		}

		// Obtiene el campo "id" del formulario.
		// Si está vacío, se interpreta que se quiere crear un evento.
		// Si tiene un valor, se interpreta que se quiere editar ese evento.
		id := r.FormValue("id")

		// Rama de creación: se ejecuta cuando el formulario no contiene un ID.
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

			// Consulta cuántas reservas tiene actualmente el evento.
			reservadas, err := queries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))

			if err != nil {
				http.Error(w, "No se pudieron consultar las reservas del evento", http.StatusInternalServerError)
				return
			}

			// Comprueba que la capacidad nueva no sea inferior a las reservas existentes.
			if int64(capacidad) < reservadas {
				http.Error(
					w,
					fmt.Sprintf("La capacidad no puede ser menor que las reservas existentes (%d).", reservadas),
					http.StatusBadRequest,
				)
				return
			}

			// Actualiza el evento existente con los nuevos valores.
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

		// Aplica Post/Redirect/Get: luego del POST redirige al navegador a la página principal.
		// Conserva el usuario que estaba utilizando la aplicación mediante usuario_id.
		http.Redirect(w, r, "/?usuario_id="+usuarioID, http.StatusSeeOther)
	}
}

func EliminarEvento(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		usuarioID := r.FormValue("usuario_id")

		if !EsAdministrador(r.Context(), queries, usuarioID) {
			http.Error(w, "No tenés permisos para realizar esta acción", http.StatusForbidden)
			return
		}

		eventoID, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || eventoID < 1 {
			http.Error(w, "ID de evento inválido", http.StatusBadRequest)
			return
		}

		// Ejecutamos la consulta SQL para eliminar el evento.
		// Por ON DELETE CASCADE, también se eliminan sus tickets.
		filas, err := queries.EliminarEvento(r.Context(), int32(eventoID))
		if err != nil {
			http.Error(w, "No se pudo eliminar el evento", http.StatusInternalServerError)
			return
		}

		if filas == 0 {
			http.Error(w, "No se encontró el evento", http.StatusNotFound)
			return
		}

		// Redirigimos a Home mediante PRG.
		http.Redirect(
			w,
			r,
			"/?usuario_id="+url.QueryEscape(usuarioID),
			http.StatusSeeOther,
		)
	}
}
