up:
	docker compose up --build

down:
	docker compose down -v

restart: down up

oapi-gen:
	oapi-codegen -generate types,chi-server -package generated -o internal/generated/api.gen.go api/openapi.yml

loadtest:
	k6 run loadtest/k6_loadtest.js

lint:
	golangci-lint run

fmt:
	golangci-lint fmt
