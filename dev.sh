#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

if [ "$#" -gt 1 ] || { [ "$#" -eq 1 ] && [ "$1" != "--down" ]; }; then
    echo "Uso: ./dev.sh [--down]"
    exit 1
fi

command -v docker >/dev/null || { echo "ERROR: no encuentro Docker."; exit 1; }
command -v go >/dev/null || { echo "ERROR: no encuentro Go."; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "ERROR: falta el plugin 'docker compose'."; exit 1; }

SQLC="$(command -v sqlc || true)"
[ -z "$SQLC" ] && [ -x "$HOME/go/bin/sqlc" ] && SQLC="$HOME/go/bin/sqlc"
[ -n "$SQLC" ] || {
    echo "ERROR: no encuentro sqlc. Instalalo con:"
    echo "  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
    exit 1
}

TEMPL="$(command -v templ || true)"
[ -z "$TEMPL" ] && [ -x "$HOME/go/bin/templ" ] && TEMPL="$HOME/go/bin/templ"
[ -n "$TEMPL" ] || {
    echo "ERROR: no encuentro templ. Instalalo con:"
    echo "  go install github.com/a-h/templ/cmd/templ@latest"
    exit 1
}

echo "=== Generando código con sqlc ==="
"$SQLC" generate

echo "=== Generando vistas con templ ==="
"$TEMPL" generate

if [ "${1:-}" = "--down" ]; then
    echo "=== Deteniendo PostgreSQL y eliminando los datos ==="
    docker compose down -v --remove-orphans
    exit 0
fi

echo "=== Levantando PostgreSQL ==="
docker compose up -d --wait db

trap 'echo; echo "=== Deteniendo el entorno (se conservan los datos) ==="; docker compose down' EXIT

echo "=== Iniciando servidor en http://localhost:8080 ==="
go run .
