
VERSION=0.3.0

DB_URL ?= postgres://postgres:postgres@localhost:5432/clipboard-share?sslmode=disable

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
	migrate -path ./migrations/sql -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations/sql -database "$(DB_URL)" down

# Build api
build: clean
	go mod download
	CGO_ENABLED=0 go build -o bin/app/app ./cmd/app/main.go

# Run
run:
	go run ./cmd/app/main.go --config ./configs/app.json

# Docker
build-docker:
	docker build -t clipboard-share-api:$(VERSION) .

build-docker-release:
	make build-docker TAG_SUFFIX=

run-docker:
	docker run -p 8080:8080 --name clipboard-share-api clipboard-share-api:$(VERSION)