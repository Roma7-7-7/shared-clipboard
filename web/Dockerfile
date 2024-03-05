# build
FROM node:21-alpine3.18 as build

ARG API_HOST=http://localhost:8080

RUN apk add --no-cache make
RUN npm install -g vite
COPY . /web
WORKDIR /web

RUN echo "VITE_API_BASE_URL=${API_HOST}" > .env.production
RUN make build

# run
FROM nginx:1.25.4-alpine3.18

EXPOSE 80

COPY --from=build /web/dist/ /usr/share/nginx/html/
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

ENTRYPOINT ["nginx", "-g", "daemon off;"]