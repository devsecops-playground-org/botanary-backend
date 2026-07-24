# syntax=docker/dockerfile:1

# ---- build -----------------------------------------------------------------
FROM golang:1.24.4-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

# Static binary: the runtime stage needs no libc, no shell and no package manager.
ARG GIT_SHA=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.revision=${GIT_SHA}" \
    -o /out/server ./cmd/server

# ---- runtime ---------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

ARG GIT_SHA=unknown
LABEL org.opencontainers.image.source="https://github.com/devsecops-playground-org/botanary-backend" \
      org.opencontainers.image.revision="${GIT_SHA}"

COPY --from=builder /out/server /server

USER nonroot:nonroot
EXPOSE 8080

# Distroless has no shell, so the healthcheck is the binary itself.
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/server", "-healthcheck"]

ENTRYPOINT ["/server"]
