# build
FROM golang:1.21.7-alpine3.19 AS build

RUN apk add --no-cache make

COPY . /app
WORKDIR /app
RUN make build

# run
FROM alpine:3.19

EXPOSE 8080

COPY --from=build /app/bin/app /app
COPY --from=build /app/configs /app/config

WORKDIR /app

ENTRYPOINT ["/app/app", "-config", "/app/config/app.json"]
