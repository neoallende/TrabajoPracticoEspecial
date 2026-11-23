package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func CrearUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Formulario inválido", http.StatusBadRequest)
			return
		}
		nombre, email := r.FormValue("nombre"), r.FormValue("email")
		if nombre == "" || email == "" {
			http.Error(w, "Completá todos los campos", http.StatusBadRequest)
			return
		}
		usuario, err := queries.CrearUsuario(r.Context(), db.CrearUsuarioParams{Nombre: nombre, Email: email})
		if err != nil {
			http.Error(w, "Error al crear el usuario", http.StatusInternalServerError)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Push-Url", fmt.Sprintf("/?usuario_id=%d", usuario.ID))
			renderHomeFragment(w, r, queries, fmt.Sprint(usuario.ID), "", "Usuario creado correctamente.")
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/?usuario_id=%d", usuario.ID), http.StatusSeeOther)
	}
}

func EliminarUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || id < 1 {
			http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
			return
		}
		filas, err := queries.EliminarUsuario(r.Context(), int32(id))
		if err != nil {
			http.Error(w, "No se pudo eliminar el usuario", http.StatusInternalServerError)
			return
		}
		if filas == 0 {
			http.Error(w, "No se encontró el usuario", http.StatusNotFound)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Push-Url", "/")
			renderHomeFragment(w, r, queries, "", "", "El usuario fue eliminado.")
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
