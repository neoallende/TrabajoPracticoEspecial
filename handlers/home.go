package handlers

import (
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

// Home responde a las solicitudes GET "/".
func Home(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Consulta todos los eventos.
		eventos, err := queries.ListarEventos(r.Context())
		if err != nil {
			http.Error(w, "Error al obtener los eventos", http.StatusInternalServerError)
			return
		}

		if eventos == nil {
			eventos = []db.Evento{}
		}

		// Obtiene la cantidad de reservas de cada evento.
		reservas := make(map[int32]int64)

		for _, evento := range eventos {
			total, err := queries.ObtenerCapacidadReservada(r.Context(), evento.ID)
			if err != nil {
				http.Error(w, "Error al consultar las reservas", http.StatusInternalServerError)
				return
			}

			reservas[evento.ID] = total
		}

		// Obtiene el usuario seleccionado y su rol.
		usuarioID := r.URL.Query().Get("usuario_id")
		mensaje := r.URL.Query().Get("mensaje")
		esAdmin := false

		var tickets []db.ListarTicketsPorUsuarioRow

		if usuarioID != "" {
			id, err := strconv.ParseInt(usuarioID, 10, 32)
			if err != nil || id < 1 {
				http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
				return
			}

			usuario, err := queries.ObtenerUsuarioPorID(r.Context(), int32(id))
			if err != nil {
				http.Error(w, "No se encontró el usuario", http.StatusNotFound)
				return
			}

			esAdmin = usuario.Rol == "admin"

			// Consulta los tickets del usuario seleccionado.
			tickets, err = queries.ListarTicketsPorUsuario(r.Context(), int32(id))
			if err != nil {
				http.Error(w, "Error al obtener los tickets", http.StatusInternalServerError)
				return
			}
		}

		if tickets == nil {
			tickets = []db.ListarTicketsPorUsuarioRow{}
		}

		// Prepara el evento que se quiere editar.
		var eventoEditar *db.Evento

		if editar := r.URL.Query().Get("editar"); editar != "" {

			// Solamente un administrador puede acceder a la edición.
			if !esAdmin {
				http.Error(w, "No tenés permisos para editar eventos", http.StatusForbidden)
				return
			}

			id, err := strconv.ParseInt(editar, 10, 32)
			if err != nil || id < 1 {
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

		// Renderiza la página principal.
		err = views.Home(
			eventos,
			reservas,
			usuarioID,
			tickets,
			eventoEditar,
			mensaje,
			esAdmin,
		).Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error al renderizar la página", http.StatusInternalServerError)
		}
	}
}
