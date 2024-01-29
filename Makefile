
# Run golangci-lint on code
lint:
	golangci-lint run

# Run tests
test:
	go test -v ./...

# Clean
clean:
	rm -rf ./bin
	rm -rf ./web/dist

# Migrate
# https://github.com/golang-migrate/migrate
migrate-up:
	migrate -path ./migrations/sql -database "postgres://postgres:postgres@localhost:5432/clipboard-share?sslmode=disable" up

migrate-down:
	migrate -path ./migrations/sql -database "postgres://postgres:postgres@localhost:5432/clipboard-share?sslmode=disable" down

# Build web
build-web:
	rm -rf ./bin/web
	mkdir -p ./bin/web
	cd ./web && npm install && vite build --outDir ../bin/web

# Build api
build-app:
	rm -rf ./bin/app
	go mod download
	go build -o bin/app/app ./cmd/app/main.go

# Build all
build-all: build-web build-app

# Run web
run-web:
	cd ./web && npm install && vite

# Run api
run-app:
	go run ./cmd/app/main.go --config ./configs/app.json
