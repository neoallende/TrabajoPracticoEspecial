package handlers

import (
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

type homeData struct {
	eventos      []db.Evento
	reservas     map[int32]int64
	usuarioID    string
	tickets      []db.ListarTicketsPorUsuarioRow
	eventoEditar *db.Evento
	mensaje      string
	esAdmin      bool
}

func cargarHome(r *http.Request, queries *db.Queries, usuarioID, editar, mensaje string) (homeData, int, error) {
	data := homeData{
		usuarioID: usuarioID,
		mensaje:   mensaje,
		reservas:  make(map[int32]int64),
		tickets:   []db.ListarTicketsPorUsuarioRow{},
	}

	eventos, err := queries.ListarEventos(r.Context())
	if err != nil {
		return data, http.StatusInternalServerError, err
	}
	if eventos == nil {
		eventos = []db.Evento{}
	}
	data.eventos = eventos

	for _, evento := range eventos {
		total, err := queries.ObtenerCapacidadReservada(r.Context(), evento.ID)
		if err != nil {
			return data, http.StatusInternalServerError, err
		}
		data.reservas[evento.ID] = total
	}

	if usuarioID != "" {
		id, err := strconv.ParseInt(usuarioID, 10, 32)
		if err != nil || id < 1 {
			return data, http.StatusBadRequest, http.ErrNoCookie
		}
		usuario, err := queries.ObtenerUsuarioPorID(r.Context(), int32(id))
		if err != nil {
			return data, http.StatusNotFound, err
		}
		data.esAdmin = usuario.Rol == "admin"
		if !data.esAdmin {
			data.tickets, err = queries.ListarTicketsPorUsuario(r.Context(), int32(id))
			if err != nil {
				return data, http.StatusInternalServerError, err
			}
			if data.tickets == nil {
				data.tickets = []db.ListarTicketsPorUsuarioRow{}
			}
		}
	}

	if editar != "" {
		if !data.esAdmin {
			return data, http.StatusForbidden, http.ErrNoCookie
		}
		id, err := strconv.ParseInt(editar, 10, 32)
		if err != nil || id < 1 {
			return data, http.StatusBadRequest, http.ErrNoCookie
		}
		evento, err := queries.ObtenerEventoPorID(r.Context(), int32(id))
		if err != nil {
			return data, http.StatusNotFound, err
		}
		data.eventoEditar = &evento
	}

	return data, http.StatusOK, nil
}

func renderHomeFragment(w http.ResponseWriter, r *http.Request, queries *db.Queries, usuarioID, editar, mensaje string) {
	data, status, err := cargarHome(r, queries, usuarioID, editar, mensaje)
	if err != nil {
		http.Error(w, "No se pudo actualizar la página: "+err.Error(), status)
		return
	}
	if err := views.HomeContent(data.eventos, data.reservas, data.usuarioID, data.tickets, data.eventoEditar, data.mensaje, data.esAdmin).Render(r.Context(), w); err != nil {
		http.Error(w, "Error al renderizar la página", http.StatusInternalServerError)
	}
}

func Home(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usuarioID := r.URL.Query().Get("usuario_id")
		data, status, err := cargarHome(r, queries, usuarioID, r.URL.Query().Get("editar"), r.URL.Query().Get("mensaje"))
		if err != nil {
			http.Error(w, "No se pudo cargar la página: "+err.Error(), status)
			return
		}

		if r.Header.Get("HX-Request") == "true" {
			if err := views.HomeContent(data.eventos, data.reservas, data.usuarioID, data.tickets, data.eventoEditar, data.mensaje, data.esAdmin).Render(r.Context(), w); err != nil {
				http.Error(w, "Error al renderizar la página", http.StatusInternalServerError)
			}
			return
		}

		if err := views.Home(data.eventos, data.reservas, data.usuarioID, data.tickets, data.eventoEditar, data.mensaje, data.esAdmin).Render(r.Context(), w); err != nil {
			http.Error(w, "Error al renderizar la página", http.StatusInternalServerError)
		}
	}
}
