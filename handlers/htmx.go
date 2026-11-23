package handlers

import (
	"bytes"
	"context"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/neouuuu/TrabajoPracticoEspecial/views"
)

func esHTMX(r *http.Request) bool { return r.Header.Get("HX-Request") == "true" }

func restaurarHistorialHTMX(r *http.Request) bool {
	return r.Header.Get("HX-History-Restore-Request") == "true"
}

func exigirHTMX(w http.ResponseWriter, r *http.Request) bool {
	if esHTMX(r) {
		return true
	}
	http.Error(w, "Este endpoint requiere una solicitud HTMX", http.StatusNotAcceptable)
	return false
}

// renderizar construye la respuesta completa antes de escribirla. Así, un error
// de templ nunca deja al navegador con una respuesta HTMX parcialmente escrita.
func renderizar(w http.ResponseWriter, r *http.Request, componentes ...templ.Component) error {
	var buffer bytes.Buffer
	for _, componente := range componentes {
		if componente == nil {
			continue
		}
		if err := componente.Render(r.Context(), &buffer); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buffer.WriteTo(w)
	return err
}

func renderizarError(_ http.ResponseWriter, _ *http.Request, err error) {
	log.Printf("error renderizando respuesta HTMX: %v", err)
}

// responderError mantiene los errores de negocio dentro del flujo HTMX.
func responderError(w http.ResponseWriter, mensaje string) {
	var buffer bytes.Buffer
	if err := views.Notificacion(mensaje, true).Render(context.Background(), &buffer); err != nil {
		log.Printf("error renderizando notificación: %v", err)
		http.Error(w, "No se pudo generar la respuesta.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buffer.WriteTo(w); err != nil {
		log.Printf("error enviando notificación HTMX: %v", err)
	}
}
