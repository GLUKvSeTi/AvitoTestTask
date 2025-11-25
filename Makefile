build:
    go build -o bin/app ./cmd/app

run:
    go run ./cmd/app

test:
    go test ./... -v

docker-build:
    docker build -f deployments/Dockerfile -t AvitoTestTask:local .

docker-compose-up:
    docker compose up --build
