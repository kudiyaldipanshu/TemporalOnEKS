FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /app

COPY . ./worker

WORKDIR /app/worker

RUN go mod tidy

RUN go build -o worker ./cmd/main.go

#-----------------   Runtime Stage

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/worker/worker .

CMD ["./worker"]
