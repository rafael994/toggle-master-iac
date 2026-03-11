

FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/app .

CMD ["./app"]
