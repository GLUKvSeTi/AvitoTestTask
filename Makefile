.PHONY: build run test docker-build docker-compose-up

#не рекомендованный способ
build:
	go build -o bin/main ./cmd/app

db-up:
	docker run -d --name postgres-db \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=AvitoTestTask \
		-p 5432:5432 \
		postgres:15-alpine
db-down:
	docker stop postgres-db || true
	docker rm postgres-db || true

run: build
	./bin/main
#рекомендованный
docker-compose-up:
	docker compose up --build