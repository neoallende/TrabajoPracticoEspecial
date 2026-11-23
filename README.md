# IAccess — rama `tp6`

# Trabajo Práctico Especial: aplicación web para la gestión de eventos y reservas de tickets

IAccess permite administrar eventos y gestionar las reservas de tickets de los usuarios.

## Dominio de la aplicación

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
- **Integración de HTMX:** se incorporó HTMX para gestionar las interacciones con el servidor y actualizar partes de la interfaz sin recargar la página completa.

## Stack

- **Go** (`net/http` con `http.ServeMux`) como servidor.
- **[templ](https://templ.guide/):** vistas tipadas que se compilan a Go.
- **PostgreSQL 15** + **[sqlc](https://sqlc.dev/)** para consultas tipadas, con el driver [lib/pq](https://github.com/lib/pq).
- **[HTMX 2](https://htmx.org/):** para gestionar solicitudes HTTP y actualizaciones parciales de la interfaz.
- **Docker / Docker Compose** para la base de datos y la aplicación.

Go renderiza los componentes del servidor mediante `templ`, mientras que HTMX permite actualizar partes de la interfaz a partir de las respuestas HTML generadas por el servidor, sin implementar una SPA JavaScript.

## Estructura

```text
.
├── main.go
├── handlers/
│   ├── home.go             # Página principal
│   ├── eventos.go          # Creación, edición y eliminación de eventos
│   ├── usuarios.go         # Registro de usuarios
│   ├── tickets.go          # Reserva y cancelación de reservas de eventos
│   ├── htmx.go             # Funcionalidad relacionada con HTMX
│   ├── partials.go         # Preparación de datos para actualizaciones parciales
│   ├── esadmin.go          # Verificación del rol administrador
│   └── notfound.go         # Manejo de rutas inexistentes (404)
├── views/
│   ├── app.templ           # Página inicial y formulario de eventos
│   ├── fragments.templ     # Componentes y fragmentos HTMX
│   ├── layout.templ        # Layout HTML compartido
│   ├── helpers.go          # Adaptadores entre handlers y componentes
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
| `GET` | `/` | Página principal. Acepta `?usuario_id=N` y `?editar=N` | Documento inicial o fragmento del formulario para HTMX |
| `GET` | `/fragmentos/sesion` | Cambia el usuario seleccionado | Fragmentos OOB de los paneles dependientes del usuario |
| `POST` | `/usuarios` | Registra un nuevo usuario | Fragmentos OOB del selector, sesión, eventos y tickets |
| `POST` | `/eventos` | Crea o edita un evento (solo admin) | Fragmentos OOB del evento afectado |
| `POST` | `/eventos/{id}/eliminar` | Elimina un evento (solo admin) | Fragmentos OOB del evento afectado |
| `POST` | `/tickets` | Reserva un ticket si hay capacidad disponible (solo user) | Reemplaza el evento y agrega el ticket reservado |
| `POST` | `/tickets/{id}/cancelar` | Cancela una reserva | Elimina el ticket y actualiza el evento afectado |
| `GET` | `/styles.css` | Hoja de estilos | CSS |
| Cualquier otra ruta | — | Página de error 404 | HTML |

### Actualizaciones parciales con HTMX

- No se usa `hx-boost` global ni se reemplaza `#app-content` después de cada operación.
- Los formularios de creación, edición, reserva y cancelación envían peticiones HTMX con `hx-post` y `hx-swap="none"`. Las respuestas contienen únicamente los elementos afectados mediante `hx-swap-oob`: un evento o ticket por operación, o los componentes dependientes al cambiar de usuario.
- Crear un evento agrega un único elemento a la lista; editarlo reemplaza ese elemento y restablece el formulario. Eliminarlo quita ese elemento y los tickets asociados al usuario seleccionado.
- Reservar un ticket reemplaza el evento para reflejar los cupos y agrega el ticket. Cancelarlo elimina ese ticket y reemplaza el evento afectado.
- Cambiar el usuario seleccionado actualiza el selector, el estado de sesión, el panel de eventos, los tickets y el formulario de administración cuando corresponda.
- Editar un evento carga solamente `#panel-form-evento`; cancelar la edición reemplaza ese mismo panel.
- Los endpoints de interacción exigen `HX-Request`; no hay redirecciones PRG ni formularios HTML alternativos. Los errores de validación se devuelven como notificaciones OOB para que también sean visibles en la interfaz.

## Cómo ejecutarlo

Clonar la rama:

```bash
git clone --branch tp5-revision-controlada https://github.com/neoallende/TrabajoPracticoEspecial.git
cd TrabajoPracticoEspecial
```

La aplicación queda disponible en http://localhost:8080. Los dos modos usan los puertos `8080` y `5432`: no ejecutes ambos al mismo tiempo.

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
