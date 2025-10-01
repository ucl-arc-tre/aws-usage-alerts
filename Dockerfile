FROM golang:1.25.1-alpine AS builder

RUN adduser --uid 1000 --disabled-password user && \
  apk add -U --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
  go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -v -o main cmd/main.go

# --------------------------------------------------------
FROM scratch AS release

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chmod=777 /app/main main

ENV DEBUG="false"
ENV HEALTH_PORT="8080"
ENV UPDATE_DELAY_SECONDS="60"
ENV SNS_TOPIC_ARN=""

USER user
CMD ["./main"]
