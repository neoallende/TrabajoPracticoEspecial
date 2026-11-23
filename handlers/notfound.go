package handlers

import (
	"bytes"
	"log"
	"net/http"

	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	if err := views.NotFound().Render(r.Context(), &buffer); err != nil {
		log.Printf("error renderizando página 404: %v", err)
		http.Error(w, "No se pudo generar la página solicitada.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if _, err := buffer.WriteTo(w); err != nil {
		log.Printf("error enviando página 404: %v", err)
	}
}
