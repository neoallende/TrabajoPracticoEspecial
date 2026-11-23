package handlers

import (
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func CrearUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !exigirHTMX(w, r) {
			return
		}
		if err := r.ParseForm(); err != nil {
			responderError(w, "Formulario inválido.")
			return
		}
		nombre, email := r.FormValue("nombre"), r.FormValue("email")
		if nombre == "" || email == "" {
			responderError(w, "Completá el nombre y el email.")
			return
		}
		usuario, err := queries.CrearUsuario(r.Context(), db.CrearUsuarioParams{Nombre: nombre, Email: email})
		if err != nil {
			responderError(w, "No se pudo crear el usuario. Verificá que el email no esté registrado.")
			return
		}
		actualizarSesion(w, r, queries, strconv.Itoa(int(usuario.ID)))
	}
}
