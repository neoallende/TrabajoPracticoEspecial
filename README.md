# IAccess — rama `tp5-revision-controlada`

Trabajo Práctico Especial: aplicación web para la **gestión de eventos y reservas de tickets**, desarrollada con Go, PostgreSQL y [templ](https://templ.guide/).

## Dominio de la aplicación

IAccess permite administrar eventos y gestionar las reservas de tickets de los usuarios.

La aplicación contempla dos tipos de cuentas:

- **Administrador (`admin`):** puede consultar, crear, editar y eliminar eventos. No puede reservar tickets, consultar las reservas de los usuarios, crear usuarios ni eliminar su propia cuenta.
- **Usuario (`user`):** puede consultar los eventos disponibles, registrarse, reservar tickets mientras haya capacidad disponible, consultar sus reservas, cancelar tickets y eliminar su cuenta.

Cada evento tiene un nombre, una capacidad máxima y una fecha. El sistema controla la cantidad de tickets reservados para evitar que se supere la capacidad del evento.

Cada ticket representa una reserva asociada a un usuario y a un evento. Los usuarios pueden consultar sus propios tickets y cancelar sus reservas.

### Cuentas y datos iniciales

La base de datos se inicializa con una cuenta administradora, tres usuarios de prueba, cinco eventos y algunas reservas.

**Usuarios iniciales:**

| ID | Usuario | Rol |
|---|---|---|
| 1 | Administrador | `admin` |
| 2 | Juan Pérez | `user` |
| 3 | María López | `user` |
| 4 | Pedro Gómez | `user` |

**Tickets iniciales:**

| Usuario | Evento | Tickets |
|---|---|---:|
| Juan Pérez | Concierto de Rock | 1 |
| Juan Pérez | Festival de Música | 1 |
| María López | Concierto de Rock | 1 |
| María López | Conferencia de Tecnología | 1 |
| Pedro Gómez | Torneo de Videojuegos | 1 |

Los usuarios que se registran desde la aplicación reciben automáticamente el rol `user`. El rol `admin` se asigna a la cuenta administradora creada inicialmente en la base de datos.

## Cambios respecto a la versión anterior

En esta versión se incorporaron las siguientes modificaciones:

- **Separación de handlers:** se dividieron las responsabilidades de los handlers en distintos archivos, organizados según la funcionalidad que gestionan.
- **Eliminación de usuarios:** se añadió la posibilidad de eliminar una cuenta de usuario.
- **Eliminación de eventos:** se añadió la posibilidad de eliminar eventos.
- **Manejo de rutas inexistentes:** se incorporó un handler que responde a las solicitudes de rutas no registradas con una página de error 404.
- **Separación de permisos por rol:** ahora únicamente las cuentas administradoras pueden crear, editar y eliminar eventos.
- **Restricciones para administradores:** no pueden reservar tickets, consultar las reservas de los usuarios, crear usuarios ni eliminar su propia cuenta.
- **Registro de usuarios:** los usuarios registrados desde la aplicación reciben automáticamente el rol `user`, sin posibilidad de seleccionar su rol desde el formulario.

## Stack

- **Go** (`net/http` con `http.ServeMux`) como servidor.
- **[templ](https://templ.guide/):** vistas tipadas que se compilan a Go.
- **PostgreSQL 15** + **[sqlc](https://sqlc.dev/)** para consultas tipadas, con el driver [lib/pq](https://github.com/lib/pq).
- **Docker / Docker Compose** para la base de datos y la aplicación.

Go genera el HTML en el servidor y los formularios se envían mediante `POST` y redirecciones, sin JavaScript en el frontend.

## Estructura

```text
.
├── main.go
├── handlers/
│   ├── home.go             # Página principal
│   ├── eventos.go          # Creación, edición y eliminación de eventos
│   ├── usuarios.go         # Registro y eliminación de usuarios
│   ├── permisos.go         # Verificación de permisos según el rol
│   └── notfound.go         # Manejo de rutas inexistentes (404)
├── views/
│   ├── layout.templ        # Estructura HTML común
│   ├── home.templ          # Página principal
│   ├── notfound.templ      # Página de error 404
│   └── *_templ.go          # Código generado por templ
├── db/
│   ├── schema.sql          # Esquema y datos iniciales de la BD
│   ├── queries/            # Consultas SQL (entrada de sqlc)
│   └── sqlc/               # Código Go generado por sqlc
├── static/
│   └── styles.css          # Estilos
├── tests/                  # Tests de la capa de datos contra Postgres real
├── sqlc.yaml
├── Dockerfile              # Build multi-stage
├── docker-compose.yml      # Servicios: db (Postgres) y app (servidor Go)
├── run.sh                  # Todo en contenedores
├── dev.sh                  # BD en contenedor + Go en tu máquina
└── Makefile                # make test
```

## Rutas

| Método | Ruta | Qué hace | Respuesta |
|---|---|---|---|
| `GET` | `/` | Página principal. Acepta `?usuario_id=N` y `?editar=N` | HTML |
| `POST` | `/usuarios` | Registra un nuevo usuario | `303` → `/?usuario_id=<nuevo>` |
| `POST` | `/usuarios/{id}/eliminar` | Elimina un usuario | `303` → `/` |
| `POST` | `/eventos` | Crea o edita un evento (solo admin) | `303` → `/` |
| `POST` | `/eventos/{id}/eliminar` | Elimina un evento (solo admin) | `303` → `/` |
| `POST` | `/tickets` | Reserva un ticket si hay capacidad disponible (solo user) | `303` → `/?usuario_id=N` |
| `POST` | `/tickets/{id}/cancelar` | Cancela una reserva | `303` → `/?usuario_id=N` |
| `GET` | `/styles.css` | Hoja de estilos | CSS |
| Cualquier otra ruta | — | Página de error 404 | HTML |

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
bash run.sh
```

Construye la imagen y levanta PostgreSQL y la aplicación. Al presionar `Ctrl+C`, se detienen y eliminan los contenedores, pero los datos de PostgreSQL se conservan en un volumen.

Para detener el entorno y borrar también los datos persistidos:

```bash
bash run.sh --down
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
bash dev.sh
```

Cada inicio genera el código de sqlc y las vistas de templ, levanta PostgreSQL y ejecuta el servidor Go localmente. Al presionar `Ctrl+C`, se detiene el entorno y se conservan los datos.

Para detener PostgreSQL y borrar también los datos persistidos:

```bash
bash dev.sh --down
```

### Persistencia de PostgreSQL

La base utiliza el volumen Docker `db_data`. Por eso, detener el entorno con `Ctrl+C` no borra los datos.

La opción `--down` de cualquiera de los dos scripts elimina los contenedores y volúmenes del proyecto. Al iniciar nuevamente, PostgreSQL crea una base vacía y vuelve a aplicar `db/schema.sql`, incluyendo las cuentas, eventos y tickets iniciales.

**Importante:** eliminar el volumen implica perder los datos almacenados en la base de datos.

## Tests

Los tests de [`tests/`](tests/) prueban la capa de datos contra un PostgreSQL real. Requieren Go, sqlc, Docker y Make:

```bash
make test
```