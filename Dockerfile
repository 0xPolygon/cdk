# CONTAINER FOR BUILDING BINARY
FROM golang:1.22.5-alpine3.20 AS build

WORKDIR $GOPATH/src/github.com/0xPolygon/cdk

RUN apk update && apk add --no-cache make build-base git

# INSTALL DEPENDENCIES
COPY go.mod go.sum ./
RUN go mod download

# BUILD BINARY
COPY . .
RUN make build

# CONTAINER FOR RUNNING BINARY
FROM alpine:3.20

COPY --from=build /go/src/github.com/0xPolygon/cdk/dist/cdk /app/cdk

RUN mkdir /app/data && apk update && apk add postgresql15-client

EXPOSE 8123

CMD ["/bin/sh", "-c", "/app/cdk run"]