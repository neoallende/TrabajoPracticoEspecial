# Del CSR al SSR: teoría, arquitectura y aplicación en IAccess

> Guía de estudio basada en la implementación de la rama
> `tp5-revision-controlada` del proyecto Trabajo Práctico Especial.
> Explica el cambio de una aplicación donde el navegador construye la
> interfaz a una aplicación donde Go genera el HTML en el servidor, y
> desarrolla los conceptos que aparecieron al revisar y mejorar el
> proyecto.

------------------------------------------------------------------------

## Índice

1.  [Objetivo y contexto del cambio](#1-objetivo-y-contexto-del-cambio)
2.  [Cómo funciona una aplicación
    CSR](#2-cómo-funciona-una-aplicación-csr)
3.  [Cómo funciona una aplicación
    SSR](#3-cómo-funciona-una-aplicación-ssr)
4.  [CSR y SSR comparados](#4-csr-y-ssr-comparados)
5.  [Qué significa SSR en nuestro
    proyecto](#5-qué-significa-ssr-en-nuestro-proyecto)
6.  [El recorrido completo de una
    petición](#6-el-recorrido-completo-de-una-petición)
7.  [Go, `net/http` y los handlers](#7-go-nethttp-y-los-handlers)
8.  [Templ: generación de HTML desde
    Go](#8-templ-generación-de-html-desde-go)
9.  [SQL, sqlc y acceso a datos](#9-sql-sqlc-y-acceso-a-datos)
10. [Formularios HTML y navegación sin
    JavaScript](#10-formularios-html-y-navegación-sin-javascript)
11. [Patrón Post/Redirect/Get y códigos
    HTTP](#11-patrón-postredirectget-y-códigos-http)
12. [Transacciones y concurrencia en las
    reservas](#12-transacciones-y-concurrencia-en-las-reservas)
13. [Bloqueo de filas con
    `SELECT ... FOR UPDATE`](#13-bloqueo-de-filas-con-select--for-update)
14. [Validación de capacidad al editar
    eventos](#14-validación-de-capacidad-al-editar-eventos)
15. [Autorización, propiedad de los recursos y límites
    actuales](#15-autorización-propiedad-de-los-recursos-y-límites-actuales)
16. [Consultas `:execrows` y filas
    afectadas](#16-consultas-execrows-y-filas-afectadas)
17. [Generación de código: templ y
    sqlc](#17-generación-de-código-templ-y-sqlc)
18. [Ejecución, Docker y
    persistencia](#18-ejecución-docker-y-persistencia)
19. [Qué responsabilidades tiene cada
    parte](#19-qué-responsabilidades-tiene-cada-parte)
20. [Ventajas, costos y límites de la
    arquitectura](#20-ventajas-costos-y-límites-de-la-arquitectura)
21. [Glosario de conceptos](#21-glosario-de-conceptos)
22. [Resumen mental para recordar el
    flujo](#22-resumen-mental-para-recordar-el-flujo)

------------------------------------------------------------------------

## 1. Objetivo y contexto del cambio

En una aplicación web siempre hay, como mínimo, dos entornos que
colaboran:

-   **Cliente:** normalmente el navegador de la persona. Muestra la
    interfaz y permite interactuar con ella.
-   **Servidor:** recibe peticiones, ejecuta lógica, consulta o modifica
    datos y devuelve respuestas.

La diferencia central entre CSR y SSR es **dónde se construye el HTML
que representa la interfaz**.

-   En **CSR (Client-Side Rendering)**, el navegador recibe normalmente
    una estructura HTML inicial y código JavaScript. Ese JavaScript
    obtiene los datos y construye o actualiza la interfaz.
-   En **SSR (Server-Side Rendering)**, el servidor obtiene los datos
    necesarios y genera el HTML antes de enviarlo. El navegador recibe
    una página ya construida y la muestra.

En IAccess se adoptó un flujo SSR con Go, `net/http`, PostgreSQL, `sqlc`
y `templ`. La interfaz se construye en el servidor y las interacciones
principales se realizan mediante formularios HTML tradicionales. No se
necesita que JavaScript arme la página a partir de respuestas JSON.

Esto no significa que el navegador deje de participar: sigue mostrando
la página, enviando formularios, navegando y descargando CSS. Lo que
cambia es quién produce el HTML.

------------------------------------------------------------------------

## 2. Cómo funciona una aplicación CSR

CSR significa *Client-Side Rendering*, o renderizado del lado del
cliente.

En una arquitectura CSR típica, el servidor entrega al navegador los
archivos necesarios para ejecutar la aplicación:

-   Un HTML inicial, que puede ser muy pequeño.
-   Archivos JavaScript.
-   Archivos CSS.
-   Eventualmente otros recursos.

Luego el JavaScript se ejecuta en el navegador, consulta una API y
utiliza los datos recibidos para construir la interfaz.

### 2.1. Flujo general

Imaginemos que el usuario entra a una aplicación de eventos:

1.  El navegador solicita `GET /`.
2.  El servidor devuelve un HTML inicial y referencias a los archivos
    JavaScript y CSS.
3.  El navegador descarga y ejecuta JavaScript.
4.  El JavaScript realiza una petición a una API, por ejemplo
    `GET /api/eventos`.
5.  El servidor consulta la base de datos y responde con JSON.
6.  JavaScript recibe ese JSON.
7.  JavaScript crea los elementos HTML correspondientes y los incorpora
    al DOM.
8.  El navegador muestra los eventos.

El flujo se puede representar así:

``` text
NAVEGADOR                         SERVIDOR                     BASE DE DATOS
    |                                |                               |
    |--------- GET / --------------->|                               |
    |<------ HTML + JS + CSS --------|                               |
    |                                |                               |
    | Ejecuta JavaScript              |                               |
    |                                |                               |
    |------ GET /api/eventos -------->|                               |
    |                                |--------- SELECT -------------->|
    |                                |<-------- filas ----------------|
    |<--------- JSON ----------------|                               |
    |                                |                               |
    | JavaScript interpreta JSON     |                               |
    | JavaScript construye HTML      |                               |
    | Navegador actualiza el DOM     |                               |
```

### 2.2. Qué es JSON en este escenario

JSON es un formato de intercambio de datos. No es HTML ni una página
web.

Por ejemplo, una API podría responder:

``` json
[
  {
    "id": 1,
    "nombre": "Concierto",
    "capacidad": 100
  },
  {
    "id": 2,
    "nombre": "Conferencia",
    "capacidad": 50
  }
]
```

El servidor está entregando información estructurada. No está entregando
necesariamente la representación visual de esa información.

El JavaScript debe interpretar esos datos y decidir qué elementos crear:
títulos, párrafos, botones, listas, formularios, etc.

### 2.3. El DOM

DOM significa *Document Object Model*. Es la representación estructurada
del documento HTML que mantiene el navegador.

Cuando JavaScript crea un elemento, por ejemplo:

``` javascript
const titulo = document.createElement("h2");
titulo.textContent = evento.nombre;
contenedor.appendChild(titulo);
```

está modificando el DOM. El navegador detecta esos cambios y actualiza
lo que se ve en pantalla.

Por eso, en CSR, una parte importante de la responsabilidad de
presentación está en el JavaScript del navegador.

### 2.4. Responsabilidades habituales en CSR

En un CSR tradicional:

-   El servidor expone endpoints de API.
-   La API devuelve datos, frecuentemente JSON.
-   JavaScript realiza las peticiones.
-   JavaScript administra estados de carga, errores y datos.
-   JavaScript crea y actualiza el DOM.
-   El navegador mantiene buena parte del estado de la interfaz.

Esto no quiere decir que el servidor no tenga lógica de negocio. La
validación, las reglas y el acceso a la base de datos siguen
correspondiendo al backend. CSR no significa trasladar toda la
aplicación al navegador.

### 2.5. Ejemplo conceptual de CSR

Supongamos que el navegador recibe un objeto:

``` javascript
const evento = {
  nombre: "Concierto",
  capacidad: 100
};
```

El código cliente puede construir la vista:

``` javascript
const tarjeta = document.createElement("article");

const titulo = document.createElement("h2");
titulo.textContent = evento.nombre;

const capacidad = document.createElement("p");
capacidad.textContent = `Capacidad: ${evento.capacidad}`;

tarjeta.append(titulo, capacidad);
document.querySelector("#eventos").appendChild(tarjeta);
```

El servidor entregó los datos; el navegador construyó el HTML a partir
de ellos.

------------------------------------------------------------------------

## 3. Cómo funciona una aplicación SSR

SSR significa *Server-Side Rendering*, o renderizado del lado del
servidor.

En SSR, el servidor no se limita a entregar datos para que el navegador
construya la página. El servidor obtiene los datos y genera el HTML
final.

### 3.1. Flujo general

1.  El navegador solicita `GET /`.
2.  Go recibe la petición.
3.  El handler consulta los datos necesarios en PostgreSQL.
4.  El handler entrega esos datos a una plantilla.
5.  La plantilla genera HTML.
6.  Go envía ese HTML como respuesta HTTP.
7.  El navegador interpreta el HTML y lo muestra.

``` text
NAVEGADOR                         SERVIDOR                     BASE DE DATOS
    |                                |                               |
    |----------- GET / ------------->|                               |
    |                                |--------- SELECT -------------->|
    |                                |<-------- filas ----------------|
    |                                |                               |
    |                                | Handler prepara los datos      |
    |                                | Templ genera el HTML           |
    |<--------- HTML completo -------|                               |
    |                                |                               |
    | Navegador interpreta HTML      |                               |
    | Muestra la página              |                               |
```

La diferencia fundamental con el flujo CSR anterior es que el HTML ya
llega construido. El navegador no necesita recibir un JSON de eventos
para crear las tarjetas.

### 3.2. ¿El navegador sigue recibiendo datos?

Sí. El HTML contiene información que el servidor incorporó al documento.

Por ejemplo, el servidor puede generar:

``` html
<h2>Concierto</h2>
<p>Capacidad: 100</p>
```

El nombre y la capacidad provienen de la base de datos, pero el
navegador los recibe dentro del HTML.

Por lo tanto, SSR no significa que los datos no viajen al navegador.
Significa que **viajan representados en HTML**, en lugar de depender de
que JavaScript reciba JSON y construya la vista.

### 3.3. ¿SSR necesita JavaScript?

No necesariamente.

Una página SSR puede funcionar con HTML y CSS, sin JavaScript propio.
Los formularios HTML pueden enviar peticiones HTTP por sí mismos.

JavaScript puede seguir utilizándose para interacciones adicionales,
pero no es un requisito para generar la página.

En IAccess, las operaciones principales se resuelven con formularios
HTML y navegación normal del navegador.

### 3.4. SSR no es lo mismo que tener una API

Una aplicación SSR puede tener APIs, y una aplicación CSR puede
comunicarse con servidores que también generan HTML. Son decisiones
arquitectónicas que pueden combinarse.

En nuestro caso, el flujo principal no usa una API JSON para alimentar
una interfaz JavaScript. Los handlers procesan las peticiones y
devuelven HTML o redirecciones.

------------------------------------------------------------------------

## 4. CSR y SSR comparados

  -----------------------------------------------------------------------
  Aspecto                 CSR                     SSR
  ----------------------- ----------------------- -----------------------
  Quién construye el HTML JavaScript en el        Servidor
  de la vista             navegador               

  Respuesta inicial       HTML inicial + JS       HTML ya renderizado
  habitual                                        

  Intercambio habitual de JSON                    Datos consultados por
  datos para actualizar                           el servidor e incluidos
  la UI                                           en HTML

  Quién modifica el DOM   Principalmente          El navegador interpreta
  de la aplicación        JavaScript              el HTML recibido

  Necesidad de JavaScript Habitualmente sí        No necesariamente
  para la vista principal                         

  Formularios             Suelen interceptarse    Pueden ser formularios
                          con JS, aunque no es    HTML tradicionales
                          obligatorio             

  Navegación              Puede ser gestionada    Puede realizarse
                          por JavaScript          mediante peticiones y
                                                  respuestas HTTP
                                                  normales
  -----------------------------------------------------------------------

### 4.1. Ventajas habituales de CSR

-   Permite interfaces muy interactivas sin recargar toda la página.
-   Puede ofrecer transiciones y actualizaciones parciales.
-   Una vez cargada la aplicación, puede reducir la necesidad de
    solicitar documentos HTML completos.
-   Es útil cuando la interfaz tiene mucho estado local o comportamiento
    complejo.

### 4.2. Costos habituales de CSR

-   Hay que desarrollar y mantener código cliente.
-   Hay que administrar estados de carga, errores y sincronización.
-   Se necesita coordinar contratos entre API y frontend.
-   El navegador debe ejecutar JavaScript para construir la interfaz.
-   Puede aumentar la complejidad de la arquitectura.

### 4.3. Ventajas habituales de SSR

-   El HTML llega preparado.
-   La lógica de presentación puede permanecer en el servidor.
-   Los formularios HTML permiten operaciones sin JavaScript.
-   Hay menos necesidad de mantener una API JSON dedicada exclusivamente
    a alimentar la vista.
-   La primera representación puede ser más sencilla de entender y
    depurar.

### 4.4. Costos habituales de SSR

-   Una navegación o envío de formulario suele producir una nueva
    petición y una nueva respuesta de página o redirección.
-   El servidor realiza el trabajo de renderizado.
-   Las interacciones muy dinámicas pueden requerir JavaScript adicional
    o más peticiones.
-   Hay que diseñar cuidadosamente qué datos se consultan para cada
    página.

No existe una arquitectura universalmente superior. La elección depende
de la interacción que se necesita, el equipo, el producto y la
complejidad que se quiere administrar.

------------------------------------------------------------------------

## 5. Qué significa SSR en nuestro proyecto

La aplicación IAccess utiliza:

-   **Go:** lenguaje y runtime del servidor.
-   **`net/http`:** servidor HTTP, rutas y handlers.
-   **PostgreSQL:** almacenamiento persistente.
-   **sqlc:** generación de funciones Go tipadas a partir de consultas
    SQL.
-   **templ:** plantillas tipadas que generan HTML y se compilan a Go.
-   **HTML y CSS:** estructura y presentación que recibe el navegador.
-   **Docker Compose:** ejecución coordinada de PostgreSQL y, en el modo
    contenedorizado, de la aplicación.

El flujo principal no consiste en:

``` text
Navegador -> API JSON -> JavaScript -> construcción de HTML
```

Consiste en:

``` text
Navegador -> handler Go -> consultas sqlc -> PostgreSQL
                         |
                         v
                    datos en Go
                         |
                         v
                    plantilla templ
                         |
                         v
                    HTML generado
                         |
                         v
                       navegador
```

### 5.1. Dónde se ve cada responsabilidad

-   `main.go`: prepara la conexión a PostgreSQL y registra las rutas.
-   `handlers/ssr.go`: procesa las peticiones HTTP, valida entradas,
    consulta la base y decide qué respuesta enviar.
-   `views/home.templ`: define cómo se representa la página principal.
-   `db/queries/*.sql`: contiene las consultas SQL escritas por el
    proyecto.
-   `db/sqlc/`: contiene código Go generado por sqlc.
-   `views/*_templ.go`: contiene código Go generado por templ.
-   `static/styles.css`: define los estilos.
-   `docker-compose.yml`: describe los servicios y el volumen de
    PostgreSQL.

El nombre `handlers/ssr.go` se conserva como nombre de archivo, pero los
nombres de las funciones ya no llevan el sufijo `SSR`. Por ejemplo, se
llaman `Home`, `CrearUsuario`, `GuardarEvento`, `ReservarTicket` y
`CancelarTicket`.

------------------------------------------------------------------------

## 6. El recorrido completo de una petición

Para comprender SSR conviene seguir una operación concreta desde el
navegador hasta la base de datos y de vuelta.

### 6.1. El usuario abre la página principal

El navegador solicita:

``` http
GET /
```

En `main.go` se registra:

``` go
mux.HandleFunc("GET /", handlers.Home(queries))
```

`http.ServeMux` asocia el método y la ruta con el handler
correspondiente.

El handler `Home`:

1.  Consulta los eventos mediante `queries.ListarEventos`.
2.  Consulta la cantidad de reservas de cada evento.
3.  Lee `usuario_id` y `mensaje` de la URL, si existen.
4.  Si hay un usuario indicado, consulta sus tickets.
5.  Si se pidió editar un evento, obtiene ese evento.
6.  Llama a `views.Home(...)`.
7.  Renderiza el resultado en la respuesta HTTP.

La llamada de renderizado es:

``` go
err = views.Home(
    eventos,
    reservas,
    usuarioID,
    tickets,
    eventoEditar,
    mensaje,
).Render(r.Context(), w)
```

Los argumentos son los datos que necesita la plantilla para construir la
página.

### 6.2. ¿Qué son `r` y `w`?

En un handler de `net/http`:

-   `r *http.Request` representa la petición recibida. Contiene el
    método, la URL, los encabezados, el cuerpo y el contexto.
-   `w http.ResponseWriter` permite construir la respuesta: escribir el
    HTML, establecer encabezados o indicar un código de estado.

La plantilla escribe el HTML en `w` mediante `Render`.

### 6.3. La plantilla construye el documento

En `views/home.templ` se define:

``` templ
templ Home(
    eventos []db.Evento,
    reservas map[int32]int64,
    usuarioID string,
    tickets []db.ListarTicketsPorUsuarioRow,
    eventoEditar *db.Evento,
    mensaje string,
) {
    @Layout("IAccess", HomeContent(
        eventos,
        reservas,
        usuarioID,
        tickets,
        eventoEditar,
        mensaje,
    ))
}
```

La plantilla recibe datos tipados de Go. No necesita realizar una
petición a una API para obtener los eventos: el handler ya se los pasó.

`Home` delega parte de la estructura en `Layout` y `HomeContent`. Esto
permite reutilizar un layout general y separar la estructura común del
contenido específico.

### 6.4. El navegador muestra el resultado

Cuando `Render` termina, el servidor devuelve el documento HTML. El
navegador lo interpreta, construye su DOM y muestra la página.

El navegador no ejecuta `Home` ni conoce los objetos Go. Esos objetos
existen en el servidor. Lo que cruza la red es la respuesta HTTP
generada.

------------------------------------------------------------------------

## 7. Go, `net/http` y los handlers

### 7.1. Qué es un handler

Un handler es una función o estructura que sabe responder a una petición
HTTP.

En Go, la forma habitual de una función handler es:

``` go
func(w http.ResponseWriter, r *http.Request)
```

En el proyecto, muchas funciones devuelven `http.HandlerFunc`:

``` go
func Home(queries *db.Queries) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Procesamiento de la petición.
    }
}
```

Esto es un patrón de **inyección de dependencias mediante cierre
(closure)**.

La función externa recibe dependencias ---en este caso `queries`--- y
devuelve una función que `net/http` puede ejecutar cuando llega una
petición.

Así, el handler tiene acceso a la capa de datos sin crear una conexión
nueva en cada solicitud.

### 7.2. Por qué `Home` recibe `queries`

En `main.go` se crea una instancia:

``` go
queries := db.New(conn)
```

Después se pasa a los handlers:

``` go
mux.HandleFunc("GET /", handlers.Home(queries))
```

Esto permite que todos los handlers utilicen el acceso a datos asociado
a la conexión configurada al iniciar la aplicación.

### 7.3. Contexto de la petición

En las consultas aparece:

``` go
queries.ListarEventos(r.Context())
```

El contexto permite propagar cancelaciones, plazos y otros valores
asociados a la petición. Si la petición deja de ser válida o se cancela,
las operaciones que respetan ese contexto pueden interrumpirse.

No es el lugar donde se guardan los datos de la página ni reemplaza a
los parámetros de la consulta.

### 7.4. Validar antes de operar

Los handlers convierten y validan entradas externas antes de pasarlas a
operaciones que esperan tipos concretos.

Por ejemplo:

``` go
eventoID, err := strconv.ParseInt(
    r.FormValue("evento_id"),
    10,
    32,
)
if err != nil || eventoID < 1 {
    http.Error(w, "Evento inválido", http.StatusBadRequest)
    return
}
```

El formulario envía texto. El handler convierte ese texto a un entero y
comprueba que sea válido.

La validación del handler mejora la calidad de los datos y permite
responder con errores claros. La base de datos, además, mantiene
restricciones propias como claves primarias y claves foráneas.

------------------------------------------------------------------------

## 8. Templ: generación de HTML desde Go

### 8.1. Qué es templ

`templ` es una herramienta para escribir plantillas HTML con sintaxis
integrada al ecosistema Go.

En lugar de construir HTML concatenando strings, se escribe una
plantilla en archivos `.templ`. templ genera código Go a partir de esos
archivos.

Por ejemplo:

``` templ
templ Saludo(nombre string) {
    <h1>Hola, { nombre }</h1>
}
```

La plantilla recibe un `string` y genera un elemento HTML con ese valor.

### 8.2. Las plantillas son funciones tipadas

Una plantilla declara los parámetros que necesita:

``` templ
templ Home(eventos []db.Evento) {
    <main>
        for _, evento := range eventos {
            <h2>{ evento.Nombre }</h2>
        }
    </main>
}
```

Esto expresa varias cosas:

-   La plantilla recibe una lista de eventos.
-   Puede recorrerla.
-   Puede acceder a los campos de cada evento.
-   Produce HTML a partir de esos valores.

El tipo de los datos se conoce en compilación. No se depende de recibir
un objeto JSON arbitrario y acceder a propiedades dinámicamente en
JavaScript.

### 8.3. Código generado

El archivo `.templ` no es ejecutado directamente por el servidor. El
comando:

``` bash
templ generate
```

genera archivos Go, habitualmente con nombres como `home_templ.go`.

Esos archivos contienen la implementación necesaria para producir el
HTML. Luego Go compila y ejecuta ese código.

Por eso, el ciclo de trabajo es:

``` text
home.templ
    |
    | templ generate
    v
home_templ.go
    |
    | go build / go run
    v
Programa Go
    |
    | petición HTTP
    v
HTML generado en tiempo de ejecución
```

Hay dos momentos diferentes:

1.  **Generación/compilación:** templ transforma la plantilla en código
    Go.
2.  **Renderizado:** durante una petición, el código generado produce el
    HTML con los datos concretos de esa petición.

No hay que confundir generar el archivo Go con renderizar la página para
un usuario.

### 8.4. Condiciones y recorridos

En `home.templ` se usan condiciones y bucles:

``` templ
if len(eventos) == 0 {
    <p class="vacio">Todavía no hay eventos disponibles.</p>
} else {
    <ul class="lista">
        for _, evento := range eventos {
            <li>
                <h3>{ evento.Nombre }</h3>
            </li>
        }
    </ul>
}
```

El servidor decide qué estructura HTML generar según los datos.

Por ejemplo, si no hay eventos, genera el mensaje vacío. Si hay eventos,
genera una lista con una tarjeta por cada uno.

### 8.5. Escape de contenido

Al insertar valores en una plantilla, el motor debe tratarlos como
contenido y no como HTML arbitrario ejecutable. Esto es importante para
evitar que un dato ingresado por un usuario se convierta accidentalmente
en código HTML o JavaScript.

La regla práctica es utilizar las expresiones y mecanismos normales de
templ, evitando concatenar HTML sin control.

### 8.6. Layout reutilizable

La vista principal utiliza:

``` templ
@Layout("IAccess", HomeContent(...))
```

El layout concentra la estructura compartida ---por ejemplo, documento,
encabezados, CSS y contenido--- y la página aporta su contenido
particular.

Esto es composición de vistas: una plantilla utiliza otra para evitar
repetir la estructura general en cada página.

------------------------------------------------------------------------

## 9. SQL, sqlc y acceso a datos

### 9.1. Separación entre lógica de aplicación y consultas

El handler no necesita construir manualmente cada sentencia SQL. Las
consultas se escriben en archivos `.sql`, y sqlc genera funciones Go
para ejecutarlas.

Por ejemplo, en `db/queries/eventos.sql`:

``` sql
-- name: ListarEventos :many
SELECT * FROM eventos
ORDER BY fecha ASC;
```

La anotación indica a sqlc:

-   Nombre de la función: `ListarEventos`.
-   Tipo de resultado: `:many`, porque puede devolver varias filas.

Luego se ejecuta:

``` bash
sqlc generate
```

sqlc analiza el esquema y las consultas y genera código Go.

### 9.2. Qué aporta sqlc

sqlc permite escribir SQL directamente y generar funciones con
parámetros y resultados tipados.

En lugar de escribir manualmente el código repetitivo para preparar
consultas, ejecutar sentencias, recorrer filas y convertir valores, se
utiliza el código generado.

En el proyecto se configura:

``` yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "db/schema.sql"
    queries: "db/queries"
    gen:
      go:
        package: "db"
        out: "db/sqlc"
        sql_package: "database/sql"
```

Esto le indica a sqlc:

-   Que el motor es PostgreSQL.
-   Que el esquema está en `db/schema.sql`.
-   Que las consultas están en `db/queries`.
-   Que genere un paquete Go llamado `db`.
-   Que escriba el resultado en `db/sqlc`.
-   Que utilice la integración `database/sql`.

### 9.3. Tipos de consultas sqlc que aparecen

Las anotaciones más relevantes del proyecto son:

-   `:one`: se espera una fila.
-   `:many`: se esperan cero o más filas.
-   `:execrows`: se ejecuta una sentencia y se devuelve la cantidad de
    filas afectadas.

Ejemplos:

``` sql
-- name: ObtenerEventoPorID :one
SELECT * FROM eventos
WHERE id = $1
LIMIT 1;
```

``` sql
-- name: ListarEventos :many
SELECT * FROM eventos
ORDER BY fecha ASC;
```

``` sql
-- name: EliminarTicket :execrows
DELETE FROM tickets
WHERE id = $1 AND id_usuario = $2;
```

La anotación no reemplaza las reglas de SQL. Describe a sqlc cómo debe
generar la función Go correspondiente.

### 9.4. Parámetros SQL

En PostgreSQL se utilizan parámetros como `$1`, `$2`, etc.

Por ejemplo:

``` sql
INSERT INTO tickets (id_usuario, id_evento)
VALUES ($1, $2)
RETURNING *;
```

Los valores se envían por separado de la sentencia. Esto evita tener que
concatenar datos del usuario dentro del SQL y ayuda a prevenir inyección
SQL.

### 9.5. `RETURNING`

PostgreSQL permite devolver filas afectadas por una operación:

``` sql
INSERT INTO eventos (nombre, capacidad, fecha)
VALUES ($1, $2, $3)
RETURNING *;
```

Esto permite que la función generada por sqlc devuelva el registro
creado, incluyendo el ID asignado por la base.

### 9.6. La conexión y `db.Queries`

En `main.go`:

``` go
conn, err := sql.Open("postgres", connStr)
```

`sql.Open` prepara un objeto de acceso a la base de datos. Luego:

``` go
if err := conn.Ping(); err != nil {
    log.Fatalf("No se pudo conectar a la BD: %v", err)
}
```

`Ping` comprueba que la conexión pueda comunicarse con el servidor de
base de datos.

Finalmente:

``` go
queries := db.New(conn)
```

Se crea el objeto generado por sqlc que ofrece los métodos para ejecutar
las consultas.

------------------------------------------------------------------------

## 10. Formularios HTML y navegación sin JavaScript

Una de las consecuencias más visibles del SSR elegido es que las
operaciones se realizan con formularios HTML tradicionales.

### 10.1. Formulario de creación de usuario

En `home.templ`:

``` html
<form action="/usuarios" method="POST">
    <input type="text" name="nombre" required>
    <input type="email" name="email" required>
    <button type="submit">Crear usuario</button>
</form>
```

Cuando el usuario envía el formulario:

1.  El navegador reúne los campos que tienen atributo `name`.
2.  Envía una petición `POST /usuarios`.
3.  Los valores se incluyen en el cuerpo de la petición.
4.  Go recibe la petición.
5.  El handler lee los valores.
6.  El servidor crea el usuario.
7.  El servidor responde con una redirección.

El navegador ya sabe enviar formularios. No hace falta escribir
JavaScript para interceptar el botón y construir una petición manual.

### 10.2. `action`, `method` y `name`

-   `action`: URL a la que se envía el formulario.
-   `method`: método HTTP utilizado.
-   `name`: nombre con el que se transmite el valor de cada campo.
-   `required`: validación básica del navegador que exige completar el
    campo.

El atributo `id` sirve principalmente para identificar el elemento
dentro del documento y asociarlo con etiquetas, estilos o código. No es
el nombre con el que el formulario envía el valor.

### 10.3. Cómo lee Go el formulario

En el handler:

``` go
if err := r.ParseForm(); err != nil {
    http.Error(w, "Formulario inválido", http.StatusBadRequest)
    return
}

nombre := r.FormValue("nombre")
email := r.FormValue("email")
```

`ParseForm` procesa los datos del formulario. `FormValue` obtiene el
valor asociado al nombre indicado.

Después, el handler valida esos valores y llama a sqlc.

### 10.4. Campos ocultos

En la vista aparecen campos como:

``` html
<input type="hidden" name="usuario_id" value="...">
<input type="hidden" name="evento_id" value="...">
```

Un campo oculto no se muestra visualmente, pero se envía junto con el
formulario.

Se utiliza para transportar datos que la operación necesita, como el ID
del evento que se quiere reservar.

**Importante:** oculto no significa secreto ni confiable. El usuario
puede inspeccionar y modificar los campos desde las herramientas del
navegador o enviando una petición manual. Todo valor que llega del
cliente debe validarse en el servidor.

### 10.5. Query parameters y path parameters

El proyecto utiliza ambos:

-   Query parameter: `/?usuario_id=15`.
-   Path parameter: `/tickets/42/cancelar`.

Se leen de forma diferente:

``` go
usuarioID := r.URL.Query().Get("usuario_id")
```

``` go
ticketID := r.PathValue("id")
```

El primero está en la parte de consulta de la URL. El segundo forma
parte del patrón de ruta registrado en `ServeMux`.

------------------------------------------------------------------------

## 11. Patrón Post/Redirect/Get y códigos HTTP

### 11.1. El problema de reenviar un formulario

Si el servidor responde directamente con una página después de procesar
un `POST`, al actualizar el navegador puede aparecer una advertencia
para reenviar el formulario.

Esto puede provocar que una operación se intente repetir, por ejemplo,
crear dos veces un registro o volver a reservar.

### 11.2. Post/Redirect/Get (PRG)

PRG es un patrón de navegación:

1.  **POST:** el navegador envía el formulario.
2.  **Redirect:** el servidor responde indicando que debe navegar a otra
    URL.
3.  **GET:** el navegador solicita esa URL y recibe la página.

``` text
Navegador                 Servidor
    |                         |
    |------ POST /tickets ---->|
    |                         | Procesa la reserva
    |<----- 303 Location ------|
    |                         |
    |------- GET / ---------->|
    |<------ HTML -------------|
```

En el proyecto se usa:

``` go
http.Redirect(
    w,
    r,
    fmt.Sprintf("/?usuario_id=%d", usuarioID),
    http.StatusSeeOther,
)
```

### 11.3. ¿Por qué `303 See Other`?

El código HTTP `303 See Other` indica que el resultado de la operación
debe consultarse en otra URL mediante una petición `GET`.

Es especialmente apropiado después de un `POST`, porque evita que el
navegador vuelva a enviar el cuerpo del formulario al seguir la
redirección.

Otros códigos relacionados:

-   `200 OK`: la petición se procesó correctamente.
-   `302 Found`: redirección temporal, con comportamiento histórico que
    puede transformar el método en ciertos clientes.
-   `303 See Other`: redirigir a una consulta `GET`.
-   `400 Bad Request`: datos o petición inválidos.
-   `404 Not Found`: recurso no encontrado o no accesible bajo la
    condición consultada.
-   `500 Internal Server Error`: error interno del servidor.

### 11.4. Mensajes después de una operación

Cuando no quedan lugares, el handler redirige a:

``` text
/?usuario_id=15&mensaje=El%20evento%20no%20tiene%20más%20lugares%20disponibles.
```

El handler `Home` lee el parámetro:

``` go
mensaje := r.URL.Query().Get("mensaje")
```

Y se lo pasa a la plantilla, que muestra el mensaje si no está vacío.

Así se conserva el patrón PRG: la operación se realiza con `POST`, pero
el mensaje se muestra en una página solicitada con `GET`.

El mensaje de la URL sirve para informar al usuario; no es una prueba de
que una operación haya ocurrido ni debe usarse como dato de seguridad.

------------------------------------------------------------------------

## 12. Transacciones y concurrencia en las reservas

Este fue uno de los conceptos importantes incorporados al proyecto.

### 12.1. El problema de concurrencia

Supongamos un evento con capacidad para 10 personas y 9 reservas
realizadas.

Dos usuarios intentan reservar casi al mismo tiempo.

Sin coordinación, puede ocurrir:

1.  La petición A consulta la cantidad de reservas: 9.
2.  La petición B consulta la cantidad de reservas: 9.
3.  A comprueba que 9 es menor que 10.
4.  B también comprueba que 9 es menor que 10.
5.  A inserta una reserva.
6.  B inserta otra reserva.

El resultado sería 11 reservas para una capacidad de 10.

El problema no está en que la comparación esté mal escrita. El problema
es que las dos peticiones pueden ejecutar la comparación sobre el mismo
estado antes de que alguna inserte su reserva.

Esto es una **condición de carrera (race condition)**: el resultado
depende del orden y la superposición temporal de operaciones
concurrentes.

### 12.2. Qué es una transacción

Una transacción agrupa varias operaciones de base de datos para
tratarlas como una unidad lógica.

En el proyecto, reservar un ticket implica:

1.  Iniciar una transacción.
2.  Bloquear la fila del evento.
3.  Consultar cuántas reservas existen.
4.  Comparar la cantidad con la capacidad.
5.  Insertar el ticket si hay lugar.
6.  Confirmar la transacción.

La idea es que la comprobación y la inserción no queden separadas por
una ventana en la que otra petición pueda realizar la misma comprobación
sin coordinación.

### 12.3. ACID

Las transacciones relacionales suelen describirse mediante las
propiedades ACID:

-   **Atomicidad:** la transacción se confirma completa o no se
    confirma.
-   **Consistencia:** las operaciones deben preservar las reglas e
    invariantes de los datos.
-   **Aislamiento:** las transacciones concurrentes no deben interferir
    de formas que rompan el comportamiento esperado.
-   **Durabilidad:** una vez confirmados los cambios, la base debe
    conservarlos según sus garantías de persistencia.

En este caso interesa especialmente el aislamiento y la coordinación
entre reservas.

### 12.4. Cómo inicia la transacción el código

En `ReservarTicket`:

``` go
tx, err := conn.BeginTx(r.Context(), nil)
if err != nil {
    http.Error(w, "No se pudo iniciar la reserva", http.StatusInternalServerError)
    return
}
defer tx.Rollback()
```

`BeginTx` inicia una transacción.

`defer tx.Rollback()` programa un rollback al salir de la función. Esto
sirve como protección para los caminos de error: si la transacción no
fue confirmada, se intenta deshacer.

Si luego se ejecuta `Commit`, el rollback diferido ya no deshace los
cambios confirmados; la transacción ya terminó.

### 12.5. Asociar sqlc a la transacción

La aplicación tiene un objeto `queries` asociado a la conexión general.
Para que las consultas de esta reserva se ejecuten dentro de la
transacción se crea:

``` go
txQueries := queries.WithTx(tx)
```

Este paso es fundamental.

Si se iniciara una transacción pero después se usara el objeto `queries`
original, algunas consultas podrían ejecutarse fuera de ella, utilizando
la conexión general.

Con `WithTx(tx)`, las consultas realizadas mediante `txQueries` utilizan
la transacción.

### 12.6. Confirmar la transacción

Cuando la inserción terminó correctamente:

``` go
if err := tx.Commit(); err != nil {
    http.Error(w, "No se pudo confirmar la reserva", http.StatusInternalServerError)
    return
}
```

`Commit` confirma los cambios.

Si ocurre un error antes del commit, la función sale y el rollback
diferido se ocupa de cerrar la transacción sin confirmar los cambios
pendientes.

------------------------------------------------------------------------

## 13. Bloqueo de filas con `SELECT ... FOR UPDATE`

La transacción agrupa operaciones, pero para coordinar las reservas
concurrentes el proyecto incorpora un bloqueo explícito de la fila del
evento.

### 13.1. La consulta nueva

En `db/queries/eventos.sql`:

``` sql
-- name: BloquearEvento :one
-- Bloquea la fila del evento durante la transacción para coordinar reservas concurrentes.
SELECT * FROM eventos
WHERE id = $1
FOR UPDATE;
```

`FOR UPDATE` es una cláusula de PostgreSQL que solicita un bloqueo de
fila sobre las filas seleccionadas.

El bloqueo se mantiene hasta que la transacción termina, mediante
`COMMIT` o `ROLLBACK`.

### 13.2. Qué pasa con dos reservas simultáneas

Supongamos que A y B quieren reservar el mismo evento:

1.  A inicia una transacción.
2.  B inicia otra transacción.
3.  A ejecuta `BloquearEvento` y obtiene el bloqueo de la fila del
    evento.
4.  B ejecuta `BloquearEvento` sobre esa misma fila y debe esperar.
5.  A cuenta las reservas.
6.  A inserta el ticket.
7.  A confirma con `Commit`.
8.  Se libera el bloqueo.
9.  B puede continuar y obtiene el bloqueo.
10. B cuenta las reservas nuevamente, ahora incluyendo la reserva de A.
11. Si se alcanzó la capacidad, B no inserta otro ticket y muestra el
    mensaje de que no quedan lugares.

La segunda petición no puede realizar la comprobación simultáneamente
sobre el mismo evento mientras la primera mantiene el bloqueo.

### 13.3. ¿Por qué bloquear el evento y no el ticket?

La capacidad es una propiedad del evento. Las reservas son filas
relacionadas con ese evento, pero la fila que representa el recurso cuya
capacidad se está administrando es la del evento.

Por eso, las operaciones de reserva para un mismo evento se coordinan
bloqueando esa fila.

Dos reservas para eventos distintos pueden bloquear filas distintas y no
necesitan esperar entre sí por este mecanismo.

### 13.4. El bloqueo no reemplaza la transacción

`FOR UPDATE` debe utilizarse dentro de una transacción para que el
bloqueo se mantenga durante el conjunto de operaciones.

La consulta aislada no resuelve el problema si se ejecuta fuera de una
transacción y luego se cuenta e inserta por separado.

En el proyecto, el orden correcto es:

``` text
BeginTx
  |
  v
WithTx(tx)
  |
  v
BloquearEvento (FOR UPDATE)
  |
  v
ObtenerCapacidadReservada
  |
  v
¿Hay lugar?
  |             |
  | Sí          | No
  v             v
INSERT          No insertar
  |             |
  v             v
Commit       Rollback al salir
```

### 13.5. Alcance de la solución

El bloqueo coordina las reservas que siguen este mismo protocolo:
iniciar una transacción, bloquear la fila del evento, consultar la
capacidad e insertar antes de confirmar.

No es una protección mágica contra cualquier escritura posible. Si otra
parte del sistema insertara tickets sin respetar ese protocolo, podría
volver a introducir inconsistencias. Las reglas de concurrencia deben
aplicarse de forma coherente en todos los caminos que modifican ese
recurso.

------------------------------------------------------------------------

## 14. Validación de capacidad al editar eventos

Además de impedir reservas por encima de la capacidad, hay que evitar
que una edición deje al evento con una capacidad inferior a la cantidad
de tickets ya reservados.

Ejemplo:

-   Capacidad actual: 100.
-   Reservas existentes: 80.
-   El organizador intenta editar la capacidad a 60.

Si se permitiera, el evento quedaría con 80 reservas y capacidad 60.

En `GuardarEvento`, cuando se edita un evento, se consulta:

``` go
reservadas, err := queries.ObtenerCapacidadReservada(
    r.Context(),
    int32(eventoID),
)
```

Luego se compara:

``` go
if int64(capacidad) < reservadas {
    http.Error(
        w,
        fmt.Sprintf(
            "La capacidad no puede ser menor que las reservas existentes (%d).",
            reservadas,
        ),
        http.StatusBadRequest,
    )
    return
}
```

La regla es:

``` text
nueva capacidad >= cantidad de reservas existentes
```

La validación se hace en el servidor, no solamente en el formulario,
porque un usuario puede modificar los datos enviados desde el navegador.

Es una regla de negocio: la capacidad no puede reducirse por debajo de
los compromisos ya registrados.

**Nota de alcance:** esta validación se incorporó al flujo de edición
acordado. No equivale a un mecanismo completo de coordinación entre
ediciones y reservas simultáneas. Si en el futuro se necesitara
garantizar también ese escenario concurrente, habría que coordinar la
edición con el mismo recurso y definir la estrategia transaccional
correspondiente.

------------------------------------------------------------------------

## 15. Autorización, propiedad de los recursos y límites actuales

Este tema es importante porque una URL o un formulario no demuestran
quién es el usuario.

### 15.1. El problema de cancelar un ticket ajeno

La ruta para cancelar un ticket es:

``` text
POST /tickets/{id}/cancelar?usuario_id=N
```

El ID del ticket identifica el recurso. El `usuario_id` indica para qué
usuario se intenta realizar la operación.

Si la consulta SQL fuera:

``` sql
DELETE FROM tickets
WHERE id = $1;
```

el servidor borraría el ticket indicado sin comprobar a qué usuario
pertenece.

La modificación realizada agrega la condición de propiedad:

``` sql
-- name: EliminarTicket :execrows
DELETE FROM tickets
WHERE id = $1 AND id_usuario = $2;
```

Ahora la eliminación solo afecta una fila si coinciden ambas
condiciones:

-   El ID del ticket.
-   El ID del usuario asociado a ese ticket.

### 15.2. Por qué validar en SQL

La condición de propiedad queda integrada en la propia operación de
borrado.

No se hace primero una consulta independiente para comprobar el
propietario y después un borrado que solo filtre por ticket. La
sentencia de borrado contiene ambas condiciones.

Esto reduce el riesgo de que la comprobación y la operación se separen y
hace explícita la regla en la base de datos.

### 15.3. Qué significa `usuario_id` en esta versión

En la etapa actual, `usuario_id` se obtiene de parámetros de URL o
formularios. No se implementó autenticación real mediante credenciales y
sesiones.

Por lo tanto, alguien puede modificar el `usuario_id` enviado por el
navegador.

La consulta impide borrar un ticket cuyo `id_usuario` no coincida con el
valor recibido, pero **no demuestra que quien envió ese valor sea
realmente ese usuario**.

Esta distinción es fundamental:

-   **Comprobar propiedad respecto de un ID recibido:** la operación
    verifica que el recurso pertenece al ID indicado.
-   **Autenticar a una persona:** el servidor verifica la identidad
    mediante un mecanismo confiable, por ejemplo una sesión autenticada.
-   **Autorizar una operación:** el servidor decide si esa identidad
    tiene permiso para realizarla.

La implementación actual agrega una condición de propiedad a la
operación de borrado, pero no implementa autenticación real ni una
identidad confiable.

La autenticación y autorización reales mediante credenciales y sesiones
quedan fuera del alcance de esta etapa.

### 15.4. `404 Not Found` cuando no se eliminó ninguna fila

El handler recibe la cantidad de filas afectadas por `EliminarTicket`.

Si es cero, responde:

``` go
http.Error(w, "No se encontró el ticket", http.StatusNotFound)
return
```

Desde el punto de vista de la operación, no se encontró un ticket que
cumpla simultáneamente con el ID y el usuario indicados.

La respuesta evita confirmar una cancelación que en realidad no ocurrió.

------------------------------------------------------------------------

## 16. Consultas `:execrows` y filas afectadas

Antes, una operación de borrado podía modelarse como una ejecución sin
resultado de filas. En este cambio se utiliza `:execrows`:

``` sql
-- name: EliminarTicket :execrows
DELETE FROM tickets
WHERE id = $1 AND id_usuario = $2;
```

Esto hace que sqlc genere un método que devuelve la cantidad de filas
afectadas.

En el handler:

``` go
filas, err := queries.EliminarTicket(
    r.Context(),
    db.EliminarTicketParams{
        ID:        int32(ticketID),
        IDUsuario: int32(usuarioID),
    },
)
```

Después se evalúa:

``` go
if filas == 0 {
    http.Error(w, "No se encontró el ticket", http.StatusNotFound)
    return
}
```

La cantidad de filas permite distinguir:

-   Una fila afectada: se eliminó el ticket que coincidía con ambas
    condiciones.
-   Cero filas afectadas: no había un ticket que coincidiera.
-   Error de ejecución: la base de datos no pudo completar la operación.

Es una manera simple de convertir el resultado de SQL en una decisión
HTTP.

------------------------------------------------------------------------

## 17. Generación de código: templ y sqlc

En el proyecto hay dos generadores y ambos cumplen funciones distintas.

  -----------------------------------------------------------------------
  Herramienta       Entrada           Salida            Para qué sirve
  ----------------- ----------------- ----------------- -----------------
  templ             Archivos `.templ` Código Go         Generar funciones
                                                        que producen HTML

  sqlc              Esquema SQL y     Código Go         Generar métodos
                    consultas SQL                       tipados para
                                                        acceder a la base
  -----------------------------------------------------------------------

### 17.1. Por qué se genera código

Los generadores evitan escribir manualmente código repetitivo y permiten
mantener una fuente declarativa:

-   La estructura de la vista se escribe en `.templ`.
-   Las consultas se escriben en `.sql`.
-   El código Go correspondiente se genera automáticamente.

Los archivos generados son parte del proceso de construcción, no la
fuente principal que se edita a mano.

En este repositorio, `db/sqlc` no se versiona. Por eso debe generarse
antes de compilar la aplicación.

### 17.2. Generación en `dev.sh`

El script de desarrollo comprueba que existan Docker, Go, sqlc y templ.
Luego ejecuta:

``` bash
sqlc generate
templ generate
```

Después levanta PostgreSQL y ejecuta Go localmente:

``` bash
docker compose up -d --wait db
go run .
```

Esto permite desarrollar con el servidor Go ejecutándose en la máquina
local y PostgreSQL dentro de Docker.

### 17.3. Generación en el Dockerfile

En el modo contenedorizado, el Dockerfile instala las herramientas,
copia el código, genera los archivos y compila:

``` dockerfile
RUN sqlc generate && templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o server .
```

Por eso `run.sh` no necesita que Go, sqlc y templ estén instalados en la
máquina del usuario: el proceso de construcción de la imagen se ocupa de
ellos.

------------------------------------------------------------------------

## 18. Ejecución, Docker y persistencia

Estos conceptos no definen SSR, pero forman parte de cómo se ejecuta y
desarrolla la aplicación.

### 18.1. Docker y contenedores

Docker permite ejecutar servicios en entornos aislados llamados
contenedores.

En el proyecto hay dos servicios principales:

-   `db`: PostgreSQL.
-   `app`: aplicación Go.

Docker Compose coordina su configuración, redes, puertos, dependencias y
volúmenes.

### 18.2. Comunicación entre servicios

Dentro de Docker Compose, la aplicación se conecta a PostgreSQL usando
el nombre del servicio como host:

``` text
postgres://root:secretpassword@db:5432/db?sslmode=disable
```

`db` es el nombre del servicio de PostgreSQL en `docker-compose.yml`.

Cuando Go se ejecuta directamente en la máquina local mediante `dev.sh`,
la conexión utiliza `localhost`:

``` text
postgres://root:secretpassword@localhost:5432/db?sslmode=disable
```

La diferencia existe porque, desde un contenedor, `localhost` se refiere
a ese mismo contenedor, no al contenedor vecino.

### 18.3. Volúmenes y persistencia

El servicio PostgreSQL monta:

``` yaml
volumes:
  - db_data:/var/lib/postgresql/data
```

Un volumen nombrado permite conservar los datos fuera del ciclo de vida
del contenedor.

Por eso, detener y eliminar el contenedor no implica necesariamente
eliminar la base de datos.

### 18.4. `docker compose down` y `down -v`

-   `docker compose down`: detiene y elimina los contenedores y redes
    del proyecto, pero conserva los volúmenes nombrados.
-   `docker compose down -v`: también elimina los volúmenes nombrados
    del proyecto.

En los scripts:

-   `Ctrl+C` detiene el entorno y conserva los datos.
-   `./run.sh --down` o `./dev.sh --down` elimina también los volúmenes
    y, por lo tanto, los datos persistidos en `db_data`.

### 18.5. Los dos modos de ejecución

**`./dev.sh`:**

``` text
Máquina local:
  Go
  sqlc
  templ
  código fuente

Docker:
  PostgreSQL
```

Es útil para desarrollar porque Go se ejecuta directamente en la máquina
local.

**`./run.sh`:**

``` text
Docker:
  PostgreSQL
  Aplicación Go compilada
```

Es útil para ejecutar el proyecto en un entorno más cercano al
despliegue y no requiere instalar las herramientas de desarrollo en la
máquina.

------------------------------------------------------------------------

## 19. Qué responsabilidades tiene cada parte

Una forma clara de entender la arquitectura es seguir las
responsabilidades de cada componente.

  -----------------------------------------------------------------------
  Componente                          Responsabilidad
  ----------------------------------- -----------------------------------
  Navegador                           Mostrar HTML/CSS, permitir
                                      interacción y enviar formularios

  HTML                                Estructura de la página y
                                      formularios

  CSS                                 Presentación visual

  `main.go`                           Configurar conexión, crear
                                      dependencias, registrar rutas e
                                      iniciar el servidor

  `http.ServeMux`                     Asociar método y ruta con un
                                      handler

  Handler                             Leer petición, validar entradas,
                                      coordinar lógica, consultar datos y
                                      responder

  `templ`                             Generar HTML a partir de plantillas
                                      y datos

  `sqlc`                              Generar métodos Go a partir de
                                      consultas SQL

  Consultas SQL                       Definir las operaciones concretas
                                      sobre la base

  PostgreSQL                          Persistir datos y aplicar
                                      restricciones y bloqueos

  Docker Compose                      Coordinar los servicios de
                                      ejecución
  -----------------------------------------------------------------------

### 19.1. Ejemplo: reservar un ticket

La distribución de responsabilidades queda así:

1.  El usuario pulsa **Reservar**.
2.  El formulario HTML envía `usuario_id` y `evento_id` a
    `POST /tickets`.
3.  `http.ServeMux` dirige la petición a `handlers.ReservarTicket`.
4.  El handler parsea y valida los IDs.
5.  El handler inicia una transacción.
6.  `queries.WithTx(tx)` prepara las consultas para ejecutarse dentro de
    esa transacción.
7.  `BloquearEvento` coordina las reservas concurrentes sobre ese
    evento.
8.  `ObtenerCapacidadReservada` consulta cuántos tickets existen.
9.  El handler compara las reservas con la capacidad.
10. Si hay lugar, `ReservarTicket` inserta el ticket.
11. `Commit` confirma la operación.
12. El handler responde con una redirección `303`.
13. El navegador realiza un `GET /?usuario_id=...`.
14. `Home` vuelve a consultar los datos.
15. `templ` genera el HTML actualizado.
16. El navegador muestra la página con el nuevo estado.

El navegador no calcula la disponibilidad ni decide si la reserva es
válida. Esa regla se aplica en el servidor, que consulta el estado real
de la base de datos.

------------------------------------------------------------------------

## 20. Ventajas, costos y límites de la arquitectura

### 20.1. Qué ganamos con el enfoque elegido

-   El flujo de una petición se puede seguir desde una ruta hasta un
    handler y una plantilla.
-   La generación de HTML está centralizada en el servidor.
-   Las plantillas reciben tipos de Go y datos ya preparados.
-   Los formularios HTML resuelven interacciones básicas sin JavaScript.
-   El backend conserva el control de las validaciones y operaciones.
-   Se reduce la necesidad de construir una API JSON dedicada únicamente
    a la página principal.

### 20.2. Qué no resuelve SSR por sí solo

SSR no garantiza:

-   Seguridad o autenticación.
-   Autorización correcta.
-   Ausencia de condiciones de carrera.
-   Validación de todos los datos.
-   Integridad de la base de datos.
-   Buen rendimiento automáticamente.
-   Persistencia de datos por sí mismo.

Esas propiedades dependen de decisiones adicionales. En el proyecto
aparecieron precisamente varios de esos temas: validación,
transacciones, bloqueos, restricciones y comprobación de propiedad.

### 20.3. Por qué no alcanza con cambiar la forma de renderizar

Pasar de CSR a SSR cambia el recorrido de presentación, pero las reglas
del negocio siguen existiendo.

Por ejemplo, mostrar un botón de reservar solo cuando parece haber cupo
no garantiza que haya lugar cuando el usuario lo pulsa. Otro usuario
podría reservar antes.

La decisión definitiva debe tomarse en el servidor y coordinarse con la
base de datos. Por eso el cambio de renderizado y las mejoras de
concurrencia son conceptos distintos, aunque convivan en la misma
aplicación.

------------------------------------------------------------------------

## 21. Glosario de conceptos

-   **API:** interfaz mediante la cual un sistema ofrece operaciones o
    datos a otro.
-   **Backend:** parte de la aplicación que se ejecuta en el servidor.
-   **Cliente:** programa que consume el servicio; en este caso, el
    navegador.
-   **CSR:** renderizado de la interfaz del lado del cliente.
-   **CSS:** lenguaje que define la presentación visual de documentos
    HTML.
-   **DOM:** representación estructurada del documento que mantiene el
    navegador.
-   **Endpoint:** combinación de método HTTP y ruta que identifica una
    operación.
-   **Handler:** función que procesa una petición HTTP y produce una
    respuesta.
-   **HTML:** lenguaje de marcado que estructura el contenido de una
    página.
-   **HTTP:** protocolo de comunicación utilizado para las peticiones y
    respuestas web.
-   **JSON:** formato de intercambio de datos estructurados.
-   **PostgreSQL:** sistema de gestión de bases de datos relacional
    utilizado por el proyecto.
-   **PRG:** patrón Post/Redirect/Get para evitar repetir el envío de
    formularios al actualizar una página.
-   **Query parameter:** dato incluido en la parte `?clave=valor` de una
    URL.
-   **Race condition:** situación en la que el resultado depende del
    orden o solapamiento de operaciones concurrentes.
-   **Renderizado:** proceso de producir la representación visual o el
    documento que se mostrará.
-   **Sesión:** mecanismo del servidor para mantener asociada una
    identidad o estado entre distintas peticiones.
-   **SQL:** lenguaje utilizado para consultar y modificar bases de
    datos relacionales.
-   **SSR:** renderizado de la interfaz del lado del servidor.
-   **Templ:** herramienta utilizada para definir plantillas tipadas y
    generar código Go que produce HTML.
-   **Transacción:** conjunto de operaciones de base de datos que se
    confirman o revierten como unidad.
-   **sqlc:** generador que convierte consultas SQL en métodos Go
    tipados.
-   **Volumen Docker:** almacenamiento persistente administrado por
    Docker.
-   **`FOR UPDATE`:** cláusula SQL que bloquea las filas seleccionadas
    para coordinar modificaciones concurrentes.
-   **`Commit`:** confirmación de una transacción.
-   **`Rollback`:** reversión de una transacción no confirmada.
-   **`ResponseWriter`:** interfaz de Go utilizada para construir la
    respuesta HTTP.
-   **`Request`:** estructura de Go que representa la petición HTTP
    recibida.

------------------------------------------------------------------------

## 22. Resumen mental para recordar el flujo

La idea principal del paso de CSR a SSR es:

**En CSR, el servidor suele enviar datos y JavaScript construye la
interfaz. En SSR, el servidor obtiene los datos, genera el HTML y lo
envía al navegador.**

En IAccess:

``` text
                 PETICIÓN DEL NAVEGADOR
                           |
                           v
                    Go / net/http
                           |
                           v
                        Handler
                   /       |       \
                  /        |        \
             Validar   Consultar   Coordinar
                       datos       operación
                         |            |
                         v            v
                      sqlc       Transacción
                         |        (cuando hace
                         v         falta)
                    PostgreSQL
                         |
                         v
                  Datos en estructuras Go
                         |
                         v
                       templ
                         |
                         v
                    HTML generado
                         |
                         v
                  ResponseWriter
                         |
                         v
                 RESPUESTA AL NAVEGADOR
```

Y para una reserva, además:

``` text
BeginTx
  -> WithTx(tx)
  -> BloquearEvento (FOR UPDATE)
  -> ObtenerCapacidadReservada
  -> Validar capacidad
  -> Insertar ticket si hay lugar
  -> Commit
  -> Redirect 303
  -> GET de la página principal
  -> Renderizado con templ
```

La separación más importante para recordar es:

-   **El handler coordina la petición y las reglas de la aplicación.**
-   **sqlc ejecuta las consultas SQL definidas por el proyecto.**
-   **PostgreSQL almacena los datos y ofrece transacciones y bloqueos.**
-   **templ genera el HTML a partir de datos de Go.**
-   **El navegador interpreta el HTML y permite que el usuario
    interactúe.**

Ese es el recorrido completo de la arquitectura SSR implementada en
IAccess.
