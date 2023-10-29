
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
	go build -o bin/web/web ./cmd/web/web.go

# Build api
build-api:
	rm -rf ./bin/api
	go build -o bin/api/api ./cmd/api/api.go

# Build dev
build-dev:
	rm -rf ./bin/dev
	go build -o bin/dev/dev ./cmd/dev/dev.go

# Build all
build-all: build-web build-api build-dev

# Run web
run-web: build-web
	./bin/web/web -config ./configs/web.json

# Run api
run-api: build-api
	./bin/api/api -config ./configs/api.json

# Run dev
run-dev: build-dev
	./bin/dev/dev -api-config ./configs/api.json -web-config ./configs/web.json