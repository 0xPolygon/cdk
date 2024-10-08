# CONTAINER FOR BUILDING BINARY
FROM --platform=${BUILDPLATFORM} golang:1.22.4 AS build

WORKDIR $GOPATH/src/github.com/0xPolygon/cdk

# INSTALL DEPENDENCIES
COPY go.mod go.sum ./
RUN go mod download

# BUILD BINARY
COPY . .
RUN make build-go

# BUILD RUST BIN
FROM --platform=${BUILDPLATFORM} rust:slim-bookworm AS chef
USER root
RUN apt-get update && apt-get install -y openssl pkg-config libssl-dev
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

RUN apt-get update && apt-get install -y ca-certificates postgresql-client libssl-dev && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/cdk /usr/local/bin/
COPY --from=build /go/src/github.com/0xPolygon/cdk/target/cdk-node /usr/local/bin/

CMD ["/bin/sh", "-c", "cdk"]
