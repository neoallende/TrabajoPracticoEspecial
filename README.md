# IAccess — rama `tp5-revision`

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
│  