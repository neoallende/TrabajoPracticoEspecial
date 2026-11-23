package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

// CrearUsuario construye y devuelve el handler que procesa solicitudes POST "/usuarios"

func CrearUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Formulario inválido", http.StatusBadRequest)
			return
		}

		nombre := r.FormValue("nombre")

		email := r.FormValue("email")

		if nombre == "" || email == "" {
			http.Error(w, "Completá todos los campos", http.StatusBadRequest)
			return
		}

		// Ejecuta la consulta de creación y le pasa los datos requeridos. El resultado contiene el usuario creado, incluido su ID generado por la base de datos.
		usuario, err := queries.CrearUsuario(r.Context(), db.CrearUsuarioParams{
			Nombre: nombre,
			Email:  email,
		})

		if err != nil {
			http.Error(w, "Error al crear el usuario.", http.StatusInternalServerError)
			return
		}

		// Aplica el patrón Post/Redirect/Get:
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/?usuario_id=%d", usuario.ID),
			http.StatusSeeOther,
		)
	}
}

func EliminarUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		usuarioID, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
		if err != nil || usuarioID < 1 {
			http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
			return
		}

		// Ejecutamos la consulta SQL para eliminar al usuario.
		// PostgreSQL también eliminará sus tickets por ON DELETE CASCADE.
		filas, err := queries.EliminarUsuario(r.Context(), int32(usuarioID))
		if err != nil {
			http.Error(w, "No se pudo eliminar el usuario", http.StatusInternalServerError)
			return
		}

		if filas == 0 {
			http.Error(w, "No se encontró el usuario", http.StatusNotFound)
			return
		}

		// Redirigimos a Home mediante PRG y no pasamos usuario_id porque acabamos de eliminarlo.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
