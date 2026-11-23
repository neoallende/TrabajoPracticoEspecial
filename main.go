package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/handlers"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://root:secretpassword@localhost:5432/db?sslmode=disable"
	}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al abrir la BD: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("No se pudo conectar a la BD: %v", err)
	}

	queries := db.New(conn)
	mux := http.NewServeMux()

	// Página principal
	mux.HandleFunc("GET /", handlers.Home(queries))

	// Usuarios
	mux.HandleFunc("POST /usuarios", handlers.CrearUsuario(queries))
	mux.HandleFunc("POST /usuarios/{id}/eliminar", handlers.EliminarUsuario(queries))

	// Eventos (POST /eventos con un ID vacío crea un evento, con un ID no vacío edita el evento)
	mux.HandleFunc("POST /eventos", handlers.GuardarEvento(queries))
	mux.HandleFunc("POST /eventos/{id}/eliminar", handlers.EliminarEvento(queries))

	// Tickets
	mux.HandleFunc("POST /tickets", handlers.ReservarTicket(queries, conn))
	mux.HandleFunc("POST /tickets/{id}/cancelar", handlers.CancelarTicket(queries))

	// Rutas no encontradas
	mux.HandleFunc("/", handlers.NotFound)

	// CSS
	mux.Handle("GET /styles.css", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Servidor corriendo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
