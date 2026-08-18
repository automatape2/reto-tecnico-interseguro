# --- build stage -------------------------------------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Copy the whole module before resolving dependencies: this repo ships
# without a committed go.sum, so `go mod tidy` (run below, after the source
# is present) both completes go.mod and generates go.sum in one step. This
# trades away the usual "copy go.mod first" layer-caching optimization for
# simplicity, which is an acceptable trade-off at this project's size.
COPY . .

RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/go-api ./cmd/api

# --- runtime stage -------------------------------------------------------
# alpine (not distroless) is used specifically so the busybox `wget` below
# is available for a self-contained HEALTHCHECK, at the cost of a few MB.
FROM alpine:3.19 AS runtime

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /out/go-api .

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./go-api"]
