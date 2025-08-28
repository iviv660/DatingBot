FROM golang:1.24.4-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY match/proto/ ./match/proto/
COPY user/proto/ ./user/proto/
COPY notifier/ ./notifier/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -trimpath -ldflags "-s -w" -o /app ./notifier/cmd/app

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
