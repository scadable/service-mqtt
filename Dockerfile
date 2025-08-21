# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" \
        -o /out/service-mqtt ./cmd/service-mqtt

# Runtime stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /out/service-mqtt /app/service-mqtt

USER nonroot:nonroot

ENTRYPOINT ["/app/service-mqtt"]