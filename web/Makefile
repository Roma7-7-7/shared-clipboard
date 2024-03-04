
VERSION=0.1.0

# Clean
clean:
	rm -rf ./dist

# Build
build: clean
	npm install && vite build

# Run
run:
	npm install && vite

# Docker
build-docker:
	docker build -t clipboard-share-web:$(VERSION) -f ./Dockerfile .

run-docker:
	docker run -p 80:80 --name clipboard-share-web clipboard-share-web:$(VERSION)