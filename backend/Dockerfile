# ---- build stage ----
FROM golang:1.23-alpine AS build
WORKDIR /app

# Cache deps first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Pure-Go build (no cgo — we use pure-Go sqlite and pgx), statically linked.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/server

# ---- run stage ----
FROM alpine:3.20
# ca-certificates is required for outbound HTTPS (e.g. verifying Google tokens).
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server

# Render (and most PaaS) inject PORT; the app reads it and defaults to 8080.
ENV PORT=8080
EXPOSE 8080
CMD ["/server"]
