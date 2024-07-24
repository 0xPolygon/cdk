# CONTAINER FOR BUILDING BINARY
FROM golang:1.22.4 AS build

WORKDIR $GOPATH/src/github.com/0xPolygon/cdk

# INSTALL DEPENDENCIES
COPY go.mod go.sum ./
RUN go mod download

# BUILD BINARY
COPY . .
RUN make build

# CONTAINER FOR RUNNING BINARY
FROM alpine:3.18.4
COPY --from=build /src/dist/cdk /app/cdk
RUN mkdir /app/data &&  apk update && apk add postgresql15-client
EXPOSE 8123

CMD ["/bin/sh", "-c", "/app/cdk run"]
