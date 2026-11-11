#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

if [ "$#" -gt 1 ] || { [ "$#" -eq 1 ] && [ "$1" != "--down" ]; }; then
    echo "Uso: ./run.sh [--down]"
    exit 1
fi

command -v docker >/dev/null || { echo "ERROR: no encuentro Docker."; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "ERROR: falta el plugin 'docker compose'."; exit 1; }

if [ "${1:-}" = "--down" ]; then
    echo "=== Deteniendo contenedores y eliminando los datos ==="
    docker compose down -v --remove-orphans
    exit 0
fi

trap 'echo; echo "=== Deteniendo el entorno (se conservan los datos) ==="; docker compose down' EXIT

echo "=== Construyendo y levantando la aplicación ==="
echo "    App: http://localhost:8080"
docker compose up --build
