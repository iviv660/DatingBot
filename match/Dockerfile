# ---------- Build ----------
FROM golang:1.24.4-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY match/ ./match/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -trimpath -ldflags "-s -w" -o /app ./match/cmd/app

# ---------- Runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /app /app
EXPOSE 50051
USER nonroot:nonroot
ENTRYPOINT ["/app"]
