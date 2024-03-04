# build
FROM node:21-alpine3.18 as build

RUN apk add --no-cache make
RUN npm install -g vite

COPY . /web
WORKDIR /web
RUN make build

# run
FROM nginx:1.25.4-alpine3.18

EXPOSE 80

COPY --from=build /web/dist/ /usr/share/nginx/html/
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

ENTRYPOINT ["nginx", "-g", "daemon off;"]