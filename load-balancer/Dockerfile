FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o load-balancer .

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/load-balancer .

EXPOSE 9000

CMD ["./load-balancer"]