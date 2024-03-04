
VERSION=0.1.0

# Run golangci-lint on code
lint:
	golangci-lint run

# Run tests
test:
	go test -v ./...

# Clean
clean:
	rm -rf ./bin

# Migrate
# https://github.com/golang-migrate/migrate
migrate-up:
	migrate -path ./migrations/sql -database "postgres://postgres:postgres@localhost:5432/clipboard-share?sslmode=disable" up

migrate-down:
	migrate -path ./migrations/sql -database "postgres://postgres:postgres@localhost:5432/clipboard-share?sslmode=disable" down

# Build api
build: clean
	go mod download
	CGO_ENABLED=0 go build -o bin/app/app ./cmd/app/main.go

# Run
run:
	go run ./cmd/app/main.go --config ./configs/app.json

# Docker
build-docker:
	docker build -t clipboard-share:$(VERSION) .

run-docker:
	docker run -p 8080:8080 -v $(shell pwd)/db:/app/db -v $(shell pwd)/configs:/app/config --name clipboard-share clipboard-share:$(VERSION)