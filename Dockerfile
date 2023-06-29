# CONTAINER FOR BUILDING BINARY
FROM golang:1.19 AS build

# INSTALL DEPENDENCIES
RUN go install github.com/gobuffalo/packr/v2/packr2@v2.8.3
COPY go.mod go.sum /src/
RUN cd /src && go mod download

# BUILD BINARY
COPY . /src
RUN cd /src/db && packr2
RUN cd /src && make build

# CONTAINER FOR RUNNING BINARY
# postgresql-client 15 available for alpine>=3.18
FROM alpine:3.18.0
RUN apk add --no-cache postgresql-client
COPY --from=build /src/dist/zkevm-node /app/zkevm-node
COPY --from=build /src/config/environments/public/public.node.config.toml /app/example.config.toml
EXPOSE 8123
CMD ["/bin/sh", "-c", "/app/zkevm-node run"]
