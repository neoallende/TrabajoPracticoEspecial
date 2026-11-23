package handlers

import (
	"net/http"

	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	views.NotFound().Render(r.Context(), w)
}
