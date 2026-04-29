# API image (Go backend). Build from repo root:
#   docker build --build-arg APP_ENV=stag -t carmanage-api .
#
# APP_ENV selects backend/config/{APP_ENV}.yaml (must match APP_ENV in backend/.env at runtime).
# With docker-compose, set APP_ENV in compose.env / .env (keep in sync with backend/.env).

# Build stage
FROM golang:1.23-alpine AS builder

ARG APP_ENV=stag

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o main .

RUN printf 'APP_ENV=%s\n' "${APP_ENV}" > .env

# Serve stage
FROM alpine:3.20

ARG APP_ENV=stag

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ENV APP_ENV=${APP_ENV}

COPY --from=builder /app/main ./main
COPY --from=builder /app/.env ./.env
RUN mkdir -p /app/config
COPY --from=builder /app/config/${APP_ENV}.yaml /app/config/${APP_ENV}.yaml
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]
