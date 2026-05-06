FROM golang:1.26 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o /out/proxy-lite-wg ./cmd/proxy-lite-wg

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/proxy-lite-wg /usr/local/bin/proxy-lite-wg

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/proxy-lite-wg"]

