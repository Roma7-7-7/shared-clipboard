
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
	cp -r ./web ./bin/web/static

# Build api
build-api:
	rm -rf ./bin/api
	go build -o bin/api/api ./cmd/api/api.go

# Build all
build-all: build-web build-api

# Run web
run-web: build-web
	./bin/web/web -api-host $(API_HOST) -static-files-path ./bin/web/static

# Run api
run-api: build-api
	./bin/api/api -data ./data
