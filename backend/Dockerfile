ARG ARCH=amd64
FROM --platform=$BUILDPLATFORM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=${ARCH} go build -o backend .

FROM --platform=$BUILDPLATFORM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y ca-certificates libc6 libstdc++6 curl && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY --from=builder /app/backend /backend

EXPOSE 8000

ENTRYPOINT ["/backend"]

