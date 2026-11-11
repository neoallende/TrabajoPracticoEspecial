# ---------- Compilación ----------

# Una imagen de alpine con Go pre-instalado para compilar la aplicación.
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Herramientas de generación de código. Van ANTES de copiar el código fuente
# para que Docker las cachee y no se reinstalen en cada build.
# Pineá SQLC_VERSION con la versión que uses en tu máquina (`sqlc version`)
# si querés builds 100% reproducibles.
ARG SQLC_VERSION=latest
RUN CGO_ENABLED=0 go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}

# Dependencias del proyecto.
COPY go.mod go.sum ./
RUN go mod download

# templ se instala en la MISMA versión que figura en go.mod, así el generador
# y el runtime nunca quedan desfasados.
RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{.Version}}' github.com/a-h/templ)

# Se copia el código fuente.
COPY . .

# Se genera el código que NO viaja en el repo (db/sqlc está en el .gitignore)
# y se regeneran las vistas, para que la imagen siempre refleje los .templ y
# los .sql actuales. Por eso no hace falta tener Go ni sqlc en tu máquina.
RUN sqlc generate && templ generate

# Se compila en un binario llamado server.
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ---------- Ejecución ----------
FROM alpine:latest

WORKDIR /app

# Nos quedamos solo con el binario compilado.
COPY --from=builder /app/server .

# El servidor sirve el CSS desde ./static, así que tiene que estar en la imagen.
COPY --from=builder /app/static ./static

# No corremos como root dentro del contenedor.
USER nobody

# Escuchando en el puerto 8080.
EXPOSE 8080

CMD ["./server"]
