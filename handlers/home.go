package handlers

import (
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func Home(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "HX-Request")
		w.Header().Add("Vary", "HX-History-Restore-Request")
		usuarioID := r.URL.Query().Get("usuario_id")
		var esAdmin bool
		var tickets []db.ListarTicketsPorUsuarioRow
		if usuarioID != "" {
			id, err := obtenerUsuarioID(usuarioID)
			if err != nil {
				responderHomeError(w, r, err.Error(), http.StatusBadRequest)
				return
			}
			usuario, err := queries.ObtenerUsuarioPorID(r.Context(), id)
			if err != nil {
				responderHomeError(w, r, "No se encontró el usuario", http.StatusNotFound)
				return
			}
			esAdmin = usuario.Rol == "admin"
			tickets, err = queries.ListarTicketsPorUsuario(r.Context(), id)
			if err != nil {
				responderHomeError(w, r, "No se pudieron obtener los tickets", http.StatusInternalServerError)
				return
			}
		}
		var eventoEditar *db.Evento
		if editar := r.URL.Query().Get("editar"); editar != "" {
			if !esAdmin {
				responderHomeError(w, r, "No tenés permisos para editar eventos", http.StatusForbidden)
				return
			}
			id, err := strconv.ParseInt(editar, 10, 32)
			if err != nil || id < 1 {
				responderHomeError(w, r, "ID de evento inválido", http.StatusBadRequest)
				return
			}
			evento, err := queries.ObtenerEventoPorID(r.Context(), int32(id))
			if err != nil {
				responderHomeError(w, r, "No se encontró el evento", http.StatusNotFound)
				return
			}
			eventoEditar = &evento
			if esHTMX(r) && !restaurarHistorialHTMX(r) {
				if err := renderizar(w, r, views.PanelFormEventoPage(usuarioID, esAdmin, eventoEditar)); err != nil {
					renderizarError(w, r, err)
				}
				return
			}
		}
		eventos, err := queries.ListarEventosConReservas(r.Context())
		if err != nil {
			responderHomeError(w, r, "No se pudieron obtener los eventos", http.StatusInternalServerError)
			return
		}
		if esHTMX(r) && !restaurarHistorialHTMX(r) {
			if err := renderizar(w, r, views.PanelFormEventoPage(usuarioID, esAdmin, nil)); err != nil {
				renderizarError(w, r, err)
			}
			return
		}
		if err := renderizar(w, r, views.LayoutHTMX("IAccess", views.AppContent(eventos, usuarioID, tickets, eventoEditar, esAdmin))); err != nil {
			renderizarError(w, r, err)
		}
	}
}

func ActualizarSesion(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !exigirHTMX(w, r) {
			return
		}
		actualizarSesion(w, r, queries, r.URL.Query().Get("usuario_id"))
	}
}

func responderHomeError(w http.ResponseWriter, r *http.Request, mensaje string, status int) {
	if esHTMX(r) {
		responderError(w, mensaje)
		return
	}
	http.Error(w, mensaje, status)
}
