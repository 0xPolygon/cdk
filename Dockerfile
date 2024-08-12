# CONTAINER FOR BUILDING BINARY
FROM golang:1.22.5-alpine3.20 AS build

WORKDIR $GOPATH/src/github.com/0xPolygon/cdk

RUN apk update && apk add --no-cache make build-base git
# INSTALL DEPENDENCIES
COPY go.mod go.sum /src/
RUN cd /src && go mod download

# BUILD BINARY
COPY . /src
RUN cd /src && make build

# BUILD RUST BIN
FROM --platform=${BUILDPLATFORM} rust:slim-bullseye AS chef
USER root
RUN cargo install cargo-chef
WORKDIR /app

FROM chef AS planner

COPY --link crates crates
COPY --link Cargo.toml Cargo.toml
COPY --link Cargo.lock Cargo.lock

RUN cargo chef prepare --recipe-path recipe.json --bin cdk

FROM chef AS builder

COPY --from=planner /app/recipe.json recipe.json
# Notice that we are specifying the --target flag!
RUN cargo chef cook --release --recipe-path recipe.json

COPY --link crates crates
COPY --link Cargo.toml Cargo.toml
COPY --link Cargo.lock Cargo.lock

ENV BUILD_SCRIPT_DISABLED=1
RUN cargo build --release --bin cdk

# CONTAINER FOR RUNNING BINARY
FROM --platform=${BUILDPLATFORM} debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates postgresql-client
COPY --from=builder /app/target/release/cdk /usr/local/bin/
COPY --from=build /src/target/cdk-node /usr/local/bin/

CMD ["/bin/sh", "-c", "cdk"]
