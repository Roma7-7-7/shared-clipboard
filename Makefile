
# Run golangci-lint on code
lint:
	golangci-lint run

# Run tests
test:
	go test -v ./...

# Build
build:
	go build -o bin/app ./cmd/app/app.go