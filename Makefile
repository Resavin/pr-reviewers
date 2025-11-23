up:
	docker compose up

down:
	docker compose down

restart: down up

build:
	docker compose up --build

delete:
	docker compose down -v

rebuild: delete build

oapi-gen:
	oapi-codegen -generate types,chi-server -package generated -o internal/generated/api.gen.go api/openapi.yml

loadtest:
	k6 run loadtest/k6_loadtest.js

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

e2e:
	docker compose -f docker-compose.e2e.yaml up --build --abort-on-container-exit --exit-code-from e2e
