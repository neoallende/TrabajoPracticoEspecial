.PHONY: test

# sqlc: el del PATH, o el de ~/go/bin si no está en el PATH.
SQLC := $(shell command -v sqlc 2>/dev/null || echo $(HOME)/go/bin/sqlc)

test:
	@echo "=== 1. Limpiando contenedores y volúmenes previos ==="
	docker compose down -v --remove-orphans

	@echo "=== 2. Generando código Go con sqlc ==="
	$(SQLC) generate

	@echo "=== 3. Levantando contenedor de PostgreSQL ==="
	docker compose up -d db

	@echo "=== 4. Esperando a que PostgreSQL acepte conexiones ==="
	@until docker compose exec -T db pg_isready -h 127.0.0.1 -U root -d db > /dev/null 2>&1; do \
		echo "Postgres se está inicializando... esperando 1s"; \
		sleep 1; \
	done

	@echo "=== 5. Ejecutando la suite de pruebas ==="
	go test -v ./tests/...

	@echo "=== 6. Limpieza posterior (destruyendo contenedor) ==="
	docker compose down -v