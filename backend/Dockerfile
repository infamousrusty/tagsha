# =============================================================================
# Stage 1 — Builder
# =============================================================================
# Using GHCR mirror to avoid Docker Hub rate limits / outages
FROM ghcr.io/infamousrusty/golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /build

# Cache dependency layer before copying source.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build the main binary with hardening flags.
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o tagsha-api \
    ./cmd/server

# Build a minimal healthcheck binary.
RUN printf 'package main\nimport("fmt";"net/http";"os")\nfunc main(){r,e:=http.Get("http://localhost:"+os.Getenv("TAGSHA_PORT")+"/health");if e!=nil||r.StatusCode!=200{fmt.Fprintln(os.Stderr,"unhealthy");os.Exit(1)};fmt.Println("healthy")}' > /build/hc.go && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o healthcheck /build/hc.go

# =============================================================================
# Stage 2 — Runtime (scratch)
# =============================================================================
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /build/tagsha-api    /usr/local/bin/tagsha-api
COPY --from=builder /build/healthcheck   /usr/local/bin/healthcheck

USER 1000:1000

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/tagsha-api"]
