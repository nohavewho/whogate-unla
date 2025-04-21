FROM golang:1.24.1 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mock-user-svc ./cmd/mock-user-svc

FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/mock-user-svc .

EXPOSE 8082
CMD ["./mock-user-svc"]