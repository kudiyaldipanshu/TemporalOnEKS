FROM golang:1.26.2-alpine3.22 AS builder

WORKDIR /app

COPY . ./starter

WORKDIR /app/starter

RUN go mod tidy

RUN go build -o starter ./cmd/main.go


# ----------- Runtime Stage -----------
FROM alpine:latest

WORKDIR /app

# Copy only binary (small image)
COPY --from=builder /app/starter/starter .

# Expose port (if API server)
EXPOSE 8080

# Run app
CMD ["./starter"]