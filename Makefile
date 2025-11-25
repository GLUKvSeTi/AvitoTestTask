.PHONY: build run test docker-build docker-compose-up

build:
	go build -o main ./cmd/app

run:
	go run ./main

docker-build:
	docker build -f deployments/Dockerfile -t AvitoTestTask:local .

docker-compose-up:
	docker compose up --build