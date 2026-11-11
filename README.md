# IAccess — rama `tp5-revision-controlada`

Trabajo Práctico Especial: aplicación web desarrollada con Go, PostgreSQL y [templ](https://templ.guide/). Go genera el HTML en el servidor y los formularios se envían con `POST` + redirect, sin JavaScript en el frontend.

## Stack

- **Go** (`net/http` con `http.ServeMux`) como servidor.
- **[templ](https://templ.guide/)**: vistas tipadas que se compilan a Go.
- **PostgreSQL 15** + **[sqlc](https://sqlc.dev/)** para consultas tipadas, con el driver [lib/pq](https://github.com/lib/pq).
- **Docker / Docker Compose** para la base y la aplicación.

---

## Estructura

```
.
├── main.go                 # Conexión a la BD, rutas y servidor
├── handlers/
│   └── ssr.go              # Handlers: leen formularios, llaman a sqlc y redirigen / renderizan
├── views/
│   ├── layout.templ        # Estructura HTML común
│   ├── home.templ          # Página principal
│   └── *_templ.go          # Código generado por templ
├── db/
│   ├── schema.sql          # Esquema de la base de datos
│   ├── queries/            # Consultas SQL (entrada de sqlc)
│   └── sqlc/               # Código Go generado por sqlc (no se versiona)
├── static/styles.css       # Estilos
├── tests/                  # Tests de la capa de datos contra Postgres real
├── sqlc.yaml
├── Dockerfile              # Build multi-stage
├── docker-compose.yml      # Servicios: db (Postgres) y app (servidor Go)
├── run.sh                  # Todo en contenedores
├── dev.sh                  # BD en contenedor + Go en tu máquina
└── Makefile                # `make test`
```

---

## Rutas

| Método | Ruta | Qué hace | Respuesta |
|---|---|---|---|
| `GET` | `/` | Página principal. Acepta `?usuario_id=N` y `?editar=N` | HTML |
| `POST` | `/usuarios` | Crea un usuario | `303` → `/?usuario_id=<nuevo>` |
| `POST` | `/eventos` | Crea o edita un evento | `303` → `/` |
| `POST` | `/tickets` | Reserva un ticket si el evento tiene cupo | `303` → `/?usuario_id=N` |
| `POST` | `/tickets/{id}/cancelar` | Cancela un ticket del usuario indicado en la query | `303` → `/?usuario_id=N` |
| `GET` | `/styles.css` | Hoja de estilos | CSS |

---

## Cómo ejecutarlo

Clonar la rama:

```bash
git clone --branch tp5-revision-controlada https://github.com/neoallende/TrabajoPracticoEspecial.git
cd TrabajoPracticoEspecial
```

La aplicación queda disponible en <http://localhost:8080>. Los dos modos usan los puertos `8080` y `5432`: no ejecutes ambos al mismo tiempo.

### Todo en contenedores (`run.sh`)

Requiere **Docker** y el plugin **Docker Compose**. Go, sqlc y templ se instalan y ejecutan dentro de la construcción de la imagen.

```bash
./run.sh
```

Construye la imagen y levanta PostgreSQL y la aplicación. Al presionar `Ctrl+C`, se detienen y eliminan los contenedores, pero los datos de PostgreSQL se conservan en un volumen.

Para detener el entorno y borrar también los datos persistidos:

```bash
./run.sh --down
```

### Go nativo + PostgreSQL en contenedor (`dev.sh`)

Requiere **Docker**, **Docker Compose**, **Go**, **sqlc** y **templ** instalados en tu máquina.

Instalación de las herramientas de generación (una sola vez):

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/a-h/templ/cmd/templ@latest
```

Iniciar el entorno:

```bash
./dev.sh
```

Cada inicio genera el código de sqlc y las vistas de templ, levanta PostgreSQL y ejecuta el servidor Go localmente. Al presionar `Ctrl+C`, se detiene el entorno y se conservan los datos.

Para detener PostgreSQL y borrar también los datos persistidos:

```bash
./dev.sh --down
```

### Persistencia de PostgreSQL

La base utiliza el volumen Docker `db_data`. Por eso, detener el entorno con `Ctrl+C` no borra los datos. La opción `--down` de cualquiera de los dos scripts ejecuta la eliminación de contenedores y volúmenes del proyecto; al iniciar nuevamente, PostgreSQL crea una base vacía y vuelve a aplicar `db/schema.sql`.

---

## Tests

Los tests de [`tests/`](tests/) prueban la capa de datos contra un PostgreSQL real. Requieren Go, sqlc, Docker y Make:

```bash
make test
```
