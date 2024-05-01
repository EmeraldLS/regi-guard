FROM golang:1.22.2-alpine3.18 AS builder

WORKDIR /app

COPY . .

RUN go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 2222

CMD ["./main"]