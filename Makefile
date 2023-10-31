
# Run golangci-lint on code
lint:
	golangci-lint run

# Run tests
test:
	go test -v ./...

# Clean
clean:
	rm -rf ./bin

# Build web
build-web:
	rm -rf ./bin/web
	go build -o bin/web ./cmd/web/main.go

# Build api
build-api:
	rm -rf ./bin/api
	go build -o bin/api ./cmd/api/main.go

# Build dev
build-dev:
	rm -rf ./bin/dev
	go build -o bin/dev ./cmd/dev/main.go

# Build all
build-all: build-web build-api build-dev

# Run web
run-web: build-web
	./bin/web -config ./configs/web.json

# Run api
run-api: build-api
	./bin/api -config ./configs/api.json

# Run dev
run-dev: build-dev
	./bin/dev -api-config ./configs/api.json -web-config ./configs/web.json