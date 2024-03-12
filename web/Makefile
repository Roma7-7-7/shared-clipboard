
VERSION=0.2.0

TAG_SUFFIX ?= -local
API_HOST ?= http://localhost:8080

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
	docker build --build-arg API_HOST=$(API_HOST) -t clipboard-share-web:$(VERSION)$(TAG_SUFFIX) -f ./Dockerfile .

build-docker-release:
	make build-docker TAG_SUFFIX= API_HOST=http://api.clipboard-share.home

run-docker:
	docker run -p 80:80 --name clipboard-share-web clipboard-share-web:$(VERSION)$(TAG_SUFFIX)