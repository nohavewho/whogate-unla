FROM golang:1.24.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/apiserver ./cmd/apiserver
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mcp-gateway ./cmd/mcp-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mock-user-svc ./cmd/mock-user-svc

FROM node:20.18.0 AS web-builder

WORKDIR /app/web

COPY web/package*.json ./

RUN npm install

COPY web/ .

RUN npm run build

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    supervisor \
    nginx \
    curl \
    vim \
    && rm -rf /var/lib/apt/lists/*

COPY deploy/docker/allinone/supervisord.conf /etc/supervisor/conf.d/

COPY deploy/docker/allinone/nginx.conf /etc/nginx/nginx.conf

COPY --from=builder /app/bin/mcp-gateway /usr/local/bin/
COPY --from=builder /app/bin/mock-user-svc /usr/local/bin/
COPY --from=builder /app/bin/apiserver /usr/local/bin/

COPY --from=web-builder /app/web/dist /usr/share/nginx/html

EXPOSE 80

CMD ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/conf.d/supervisord.conf"]