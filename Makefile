run-api:
	go run cmd/api/main.go

run-cli:
	go run cmd/cli/main.go

build:
	go build -o bin/api cmd/api/main.go
	go build -o bin/worker cmd/cli/main.go

docker-build:
	docker build -t rpa-template .