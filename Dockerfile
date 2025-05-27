FROM golang:1.24.2-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    cmd/server/main.go

FROM alpine:3.19

RUN apk add --no-cache \
    chromium \
    chromium-chromedriver \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    xvfb \
    && ln -sf /usr/bin/chromium-browser /usr/bin/google-chrome

RUN adduser -D -u 1000 appuser

COPY --from=builder /app/server /app/server
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 5000

ENV PATH="/app:${PATH}" \
    TZ=UTC \
    ENV=production \
    CHROME_LOG_FILE=/dev/null \
    CHROMIUM_FLAGS="--disable-logging --log-level=3 --silent"

ENTRYPOINT ["/app/server"]
