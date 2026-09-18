FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .


FROM python:3.12-slim

WORKDIR /app

COPY ml/requirements.txt ./ml/requirements.txt

RUN pip install \
    --no-cache-dir \
    -r ml/requirements.txt

COPY --from=builder /app/main ./main

COPY ml ./ml

EXPOSE 8080

CMD ["./main"]