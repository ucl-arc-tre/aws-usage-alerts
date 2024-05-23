FROM golang:1.22.3-alpine AS builder

RUN adduser --uid 1000 --disabled-password user && \
    apk add -U --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -v -o main cmd/main.go && \
    chmod +x main

# --------------------------------------------------------
FROM scratch as release

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main main

ENV DEBUG="false"

USER user
CMD ["./main"]
