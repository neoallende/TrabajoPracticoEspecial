CAPA DE DATOS

1. Inicialización de sqlc (main.go)
1. sql.Open("postgres", connStr): crea la conexión con PostgreSQL.
2. conn.Ping(): comprueba que la conexión funcione.
3. db: es el alias del paquete generado por sqlc (db/sqlc).
4. db.New(conn): crea una instancia de la estructura Queries, usando la conexión conn.
5. queries: es la instancia que después utilizás para ejecutar los métodos generados por sqlc.
2. Estructuras generadas por sqlc
1. Queries: estructura que guarda la conexión a la base de datos.
2. New(conn): función que construye y devuelve un *Queries.
3. ListarTicketsPorUsuarioRow: estructura que representa una fila del resultado de la consulta SQL ListarTicketsPorUsuario.
4. ListarTicketsPorUsuarioParams: estructura que agrupa los parámetros de entrada si la consulta necesita más de uno. En tu consulta solo hay un parámetro, por lo que sqlc puede recibirlo directamente como idUsuario.
Dentro del método generado ocurre lo siguiente:
1. q *Queries: es el receptor del método; representa la instancia queries que contiene la conexión.
2. ctx context.Context: recibe el contexto de la petición HTTP, para permitir cancelar la consulta si corresponde.
3. idUsuario int32: recibe el ID del usuario que se quiere consultar.
4. QueryContext(ctx, query, idUsuario): ejecuta el SQL y obtiene las filas devueltas por PostgreSQL.
5. rows: contiene el resultado de la consulta, que puede tener varias filas.
6. rows.Next(): avanza por cada fila del resultado.
7. var i ListarTicketsPorUsuarioRow: crea una variable de la estructura que representa una fila.
8. rows.Scan(...): copia los valores de la fila actual a los campos de i, en el mismo orden en que aparecen las columnas del SELECT.
9. append(items, i): agrega la fila ya cargada al slice items.
10. rows.Err(): permite comprobar si ocurrió algún error durante el recorrido de las filas.
11. return items, nil: devuelve el slice con todas las filas obtenidas.
4. Operaciones según la anotación SQL
| Anotación | Método de database/sql | Resultado |
| --- | --- | --- |
| :many | QueryContext | Varias filas. |
| :one | QueryRowContext | Una fila. |
| :exec | ExecContext | Ejecuta sin devolver filas. |
| :execrows | ExecContext + RowsAffected | Devuelve cantidad de filas afectadas. |
tickets, err = queries.ListarTicketsPorUsuario(...)


- tickets: contiene todas las filas, cada una representada por ListarTicketsPorUsuarioRow.
- err: contiene un error si falló la consulta o el procesamiento del resultado.

CAPA LOGICA

| database/sql | Librería estándar de Go para trabajar con bases de datos SQL. Proporciona conexiones, consultas y transacciones. |
| --- | --- |
| net/http | Librería estándar de Go para crear servidores HTTP, recibir peticiones, responder y gestionar rutas. |
| _ "github.com/lib/pq" | Importa el driver de PostgreSQL para que database/sql pueda comunicarse con PostgreSQL. El _ indica que se importa por sus efectos de inicialización, sin utilizar directamente sus identificadores. |
| db | Alias del paquete generado por sqlc. Contiene New, Queries, estructuras y métodos para ejecutar consultas. |
| handlers | Paquete propio del proyecto que contiene las funciones que procesan las peticiones HTTP. |

Cadena de conexión:
```
postgres://root:secretpassword@localhost:5432/db?sslmode=disable
```


Contiene:
- postgres: protocolo/esquema de conexión.
- root: usuario de PostgreSQL.
- secretpassword: contraseña.
- localhost: dirección del servidor de base de datos.
- 5432: puerto de PostgreSQL.
- db: nombre de la base de datos.
- sslmode=disable: desactiva SSL para esa conexión.



```
sql.Open("postgres", connStr): crea un objeto *sql.DB, que representa el conjunto de conexiones que Go administrará para comunicarse con la base de datos.
```

"postgres": identifica el driver que debe utilizar database/sql. Ese driver fue registrado por lib/pq al importarlo.

conn: es la instancia de *sql.DB que utilizará el resto de la aplicación para acceder a PostgreSQL.

defer conn.Close(): programa el cierre del conjunto de conexiones cuando finalice main.

conn.Ping(): comprueba que se pueda establecer comunicación con la base de datos.

```
log.Fatalf(...): si ocurre un error crítico, muestra el mensaje y termina la ejecución del programa.
```

```
queries := db.New(conn)
```

Esto conecta las dos partes que vimos:
- conn: proporciona el acceso a PostgreSQL.
- db.New(conn): utiliza esa conexión para construir la instancia queries.
- queries: permite ejecutar los métodos generados por sqlc desde los handlers.

```
fmt.Println("Servidor corriendo en http://localhost:8080")
```

```
if err := http.ListenAndServe(":8080", mux); err != nil {
    log.Fatalf("Error al iniciar el servidor: %v", err)
}
```



```
mux := http.NewServeMux()
```


1. http: paquete estándar de Go para trabajar con HTTP.
2. NewServeMux(): crea un nuevo multiplexer.
3. mux: variable que almacena ese multiplexer.
¿Pero qué es un multiplexer?

Es el componente encargado de recibir las peticiones HTTP y determinar qué handler debe ejecutarse según la ruta y el método HTTP.

```
mux.HandleFunc("GET /", handlers.Home(queries))
```

También se registra el servidor de archivos estáticos:
```
mux.Handle("GET /styles.css", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
```


- http.Dir("./static"): indica el directorio donde están los archivos.
- http.FileServer: crea un servidor capaz de entregar esos archivos.
- http.StripPrefix: elimina el prefijo / de la ruta antes de buscar el archivo.
- mux.Handle: registra ese servidor de archivos en el multiplexer.

Esta es la parte que pone en funcionamiento el servidor.
1. http.ListenAndServe(":8080", mux): inicia un servidor HTTP que escucha peticiones en el puerto 8080.
2. :8080: indica el puerto en el que el servidor aceptará conexiones. Al no especificar una IP, escucha en las interfaces disponibles.
3. mux: indica qué multiplexer se utilizará para procesar las peticiones recibidas.
4. ListenAndServe: queda esperando conexiones y procesando las peticiones HTTP que llegan.
5. if err != nil: comprueba si el servidor terminó debido a un error.
6. log.Fatalf(...): informa el error y finaliza el programa.


AHORA LOS HANDLERS:

```
mux.HandleFunc("GET /", handlers.Home(queries))
```

handlers: paquete donde están implementados los handlers.
Home(queries): invoca la función externa Home, pasándole la instancia de sqlc.
```
mux.HandleFunc: recibe esa función y la registra en el multiplexer.
```
"GET /": patrón de ruta que indica que esa función atenderá las peticiones GET a /.

```
func Home(queries *db.Queries) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ...
    }
}
```


1. func Home: declara la función externa que inicializa el handler.
2. queries *db.Queries: recibe la instancia de sqlc creada en main.go, para poder ejecutar consultas a PostgreSQL.
3. http.HandlerFunc: tipo de dato que indica que la función devuelve un handler HTTP.
4. return: devuelve la función interna.
5. func(w, r): función anónima que procesará las peticiones HTTP.
6. w http.ResponseWriter: permite escribir la respuesta HTTP que recibirá el navegador.
7. r *http.Request: contiene los datos de la petición HTTP recibida.
8. Closure: mecanismo que permite que la función interna conserve acceso a queries, aunque no lo reciba como parámetro.

2. Tipo http.HandlerFunc
Su definición en Go es:
```
type HandlerFunc func(ResponseWriter, *Request)
```


1. Es un tipo definido por la librería estándar net/http.
2. Representa una función que recibe un ResponseWriter y un Request.
3. Establece la firma que deben tener las funciones que se utilizan como handlers HTTP.
4. Permite que Go las utilice para procesar peticiones y generar respuestas.

HASTA AHORA SE ENTIENDE COMO FUNCIONA LA BASE DE DATOS + CAPA DE DATOS + CONEXION A LA BD + RUN DEL SERVIDOR CON UN MULTIPLEXER PARA ASOCIAR LOS ENDPOINT HTTP UN HANDLER + LA FIRMA DE LOS HANDLERS

DENTRO DEL HANDLER SE ENTIENDE COMO SE UTILIZAN LAS FUNCIONES GENERADAS CON SQLC PARA HACER CONSULTAS

AHORA HAY QUE ENTENDER COMO FUNCIONA HTML y TEMPL PARA TERMINAR DE ENTENDER LA CAPA LOGICA

ESTÁ BUENO OLVIDARSE DE TODO LO ANTERIOR PARA ENTENDER MEJOR COMO FUNCIONA ESTA PARTE

PRIMERO HTML:

Estructura básica de un documento

| Elemento | Función |
| --- | --- |
| <!DOCTYPE html> | Indica que el documento utiliza HTML moderno (HTML5). |
| <html> | Es el elemento raíz del documento. Contiene todo el HTML. |
| lang="es" | Indica el idioma principal del documento. |
| <head> | Contiene metadatos y recursos asociados al documento. |
| <meta charset="UTF-8"> | Define la codificación de caracteres. |
| <title> | Define el título que aparece en la pestaña del navegador. |
| <link> | Permite asociar recursos externos, como una hoja CSS. |
| <body> | Contiene el contenido visible de la página. |

4. Elementos HTML más utilizados
HTML tiene muchas etiquetas, pero podemos clasificarlas según su función.
4.1. Estructura y contenido
| Etiqueta | Función |
| --- | --- |
| <h1> a <h6> | Encabezados de distintos niveles. |
| <p> | Párrafos. |
| <div> | Contenedor genérico de bloques. |
| <span> | Contenedor genérico en línea. |
| <br> | Salto de línea. |
| <hr> | Separación temática. |

4.2. Listas
HTML permite representar listas ordenadas y desordenadas.
```
<ul>
    <li>Primer elemento</li>
    <li>Segundo elemento</li>
</ul>
```


- <ul>: lista sin orden numérico.
- <ol>: lista ordenada.
- <li>: elemento de una lista.
4.3. Enlaces
```
<a href="https://google.com">Ir a Google</a>
```


- <a>: representa un enlace.
- href: atributo que indica el destino.
Cuando el usuario hace clic, el navegador navega hacia la dirección indicada.
4.4. Imágenes
```
<img src="/imagen.png" alt="Descripción de la imagen">
```


- src: ubicación de la imagen.
- alt: descripción alternativa.
5. Formularios: la parte más importante para tu TP5

Por ejemplo:
```
<form action="/usuarios" method="POST">
```

```
    <label for="nombre">Nombre:</label>
    <input type="text" id="nombre" name="nombre">
```

```
    <label for="email">Email:</label>
    <input type="email" id="email" name="email">
```

```
    <button type="submit">Crear usuario</button>
```

```
</form>
```


Veamos cada componente.
5.1. <form>
Es el contenedor que agrupa los campos que el usuario puede completar.

Tiene dos atributos fundamentales:
- action: indica a qué URL se enviarán los datos.
- method: indica qué método HTTP se utilizará.
En nuestro ejemplo:
```
<form action="/usuarios" method="POST">
```


Significa que, cuando se envíe el formulario, el navegador realizará una petición HTTP POST a /usuarios.

Esto coincide con la ruta que registraste en main.go:
```
mux.HandleFunc("POST /usuarios", handlers.CrearUsuario(queries))
```


5.2. <input>
Representa un campo donde el usuario puede introducir o seleccionar información.
```
<input type="text" name="nombre">
```


Sus atributos principales:
| Atributo | Función |
| --- | --- |
| type | Determina el tipo de campo. |
| name | Identifica el dato que se enviará al servidor. |
| id | Identifica el elemento dentro del documento HTML. |
| value | Define su valor inicial. |
| required | Indica que el campo es obligatorio. |
| placeholder | Muestra un texto orientativo. |


Cuando se envía el formulario, el navegador incluye un dato llamado nombre, cuyo valor es Juan.

Conceptualmente, el contenido enviado sería:
nombre=Juan

¿Y cómo lo recibe Go?

En tu handler:
```
nombre := r.FormValue("nombre")
```

Go obtiene el valor asociado al campo nombre.



5.4. <label>
Permite asociar una descripción a un campo.
```
<label for="nombre">Nombre:</label>
<input id="nombre" name="nombre" type="text">
```

5.5. <button>
Permite ejecutar acciones.
```
<button type="submit">Crear usuario</button>
```

Cuando el botón tiene type="submit" y está dentro de un formulario, envía ese formulario.

Otros tipos:
- submit: envía el formulario.
- button: botón genérico, sin envío automático.
- reset: restablece los campos del formulario.


7. Otros elementos importantes para entender tu TP5
Hay algunos conceptos que aparecen frecuentemente en aplicaciones web.
Campos ocultos
```
<input type="hidden" name="usuario_id" value="12">
```


El campo no se muestra visualmente, pero su valor se incluye cuando se envía el formulario.


AHORA VEMOS COMO FUNCIONA TEMPL Y TERMINAMOS CON LA CAPA LÓGICA

```
templ Saludo(nombre string) {
    <h1>Hola { nombre }</h1>
}
```

1. templ: palabra reservada que declara un componente.
2. Saludo: nombre del componente.
3. (nombre string): parámetros que recibe el componente. En este caso, una cadena de texto de Go.
4. { ... }: contiene la estructura que se va a generar.
5. <h1>: etiqueta HTML.
6. { nombre }: inserta el valor de la variable Go dentro del HTML.


LO BUENO: SE PUEDE USAR CONDICIONALES E ITERADORES

```
templ EstadoUsuario(activo bool) {
    if activo {
        <p>Usuario activo</p>
    } else {
        <p>Usuario inactivo</p>
    }
}
```


```
templ ListaUsuarios(usuarios []string) {
    <ul>
        for _, usuario := range usuarios {
            <li>{ usuario }</li>
        }
    </ul>
}
```


INVOCADOS EN LOS HANDLERS:

```
err := Saludo("Juan").Render(r.Context(), w)
```

1. Saludo("Juan"): crea el componente con los datos recibidos.
2. .Render(...): genera el HTML y lo escribe en el destino indicado.
3. r.Context(): proporciona el contexto de la petición HTTP.
4. w: es el http.ResponseWriter, donde se escribe el HTML que se enviará al navegador.

Composición de componentes:

```
templ Layout(titulo string, contenido templ.Component) {
    <!DOCTYPE html>
    <html lang="es">
        <head>
            <title>{ titulo }</title>
        </head>
        <body>
            @contenido
        </body>
    </html>
}
```



```
templ HomeContent() {
    <h1>Bienvenido</h1>
    <p>Esta es la página principal.</p>
}
```



```
templ Home() {
    @Layout("Mi aplicación", HomeContent())
}
```
