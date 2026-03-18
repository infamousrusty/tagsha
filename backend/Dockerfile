# =============================================================================
# Stage 1 — Builder
# =============================================================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /build

# Cache dependency layer before copying source.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Use Docker's built-in TARGETOS/TARGETARCH for multi-arch support.
# CGO_ENABLED=0 produces a fully static binary (no libc dependency).
# -trimpath removes build host path from binary.
# -ldflags "-s -w" strips debug symbols.
ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o tagsha-api \
    ./cmd/server

# Build a minimal healthcheck binary.
RUN printf 'package main\nimport("fmt";"net/http";"os")\nfunc main(){r,e:=http.Get("http://localhost:"+os.Getenv("TAGSHA_PORT")+"/health");if e!=nil||r.StatusCode!=200{fmt.Fprintln(os.Stderr,"unhealthy");os.Exit(1)};fmt.Println("healthy")}' > /build/hc.go && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o healthcheck /build/hc.go

# =============================================================================
# Stage 2 — Runtime (scratch — no shell, minimal attack surface)
# =============================================================================
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /build/tagsha-api    /usr/local/bin/tagsha-api
COPY --from=builder /build/healthcheck   /usr/local/bin/healthcheck

USER 1000:1000

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/tagsha-api"]
