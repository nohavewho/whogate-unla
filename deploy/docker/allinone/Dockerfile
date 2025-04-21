# Build stage for Go services
FROM golang:1.24.1 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go services
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/apiserver ./cmd/apiserver
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mcp-gateway ./cmd/mcp-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mock-user-svc ./cmd/mock-user-svc

# Build stage for web
FROM node:20.18.0 AS web-builder

WORKDIR /app/web

# Copy web package files
COPY web/package*.json ./

# Install dependencies
RUN npm install

# Copy web source code
COPY web/ .

# Build web application
RUN npm run build

# Final stage
FROM ubuntu:22.04

# Install systemd and other required packages
RUN apt-get update && apt-get install -y \
    systemd \
    nginx \
    && rm -rf /var/lib/apt/lists/* \
    && rm -f /lib/systemd/system/multi-user.target.wants/* \
    && rm -f /etc/systemd/system/*.wants/* \
    && rm -f /lib/systemd/system/local-fs.target.wants/* \
    && rm -f /lib/systemd/system/sockets.target.wants/*udev* \
    && rm -f /lib/systemd/system/sockets.target.wants/*initctl* \
    && rm -f /lib/systemd/system/sysinit.target.wants/systemd-tmpfiles-setup* \
    && rm -f /lib/systemd/system/systemd-update-utmp*

# Copy systemd service files
COPY deploy/docker/allinone/mcp-gateway.service /etc/systemd/system/
COPY deploy/docker/allinone/mock-user-svc.service /etc/systemd/system/
COPY deploy/docker/allinone/apiserver.service /etc/systemd/system/
COPY deploy/docker/allinone/nginx.conf /etc/nginx/nginx.conf

# Enable services
RUN systemctl enable mcp-gateway.service && \
    systemctl enable mock-user-svc.service && \
    systemctl enable apiserver.service

# Copy built binaries
COPY --from=builder /app/bin/mcp-gateway /usr/local/bin/
COPY --from=builder /app/bin/mock-user-svc /usr/local/bin/
COPY --from=builder /app/bin/apiserver /usr/local/bin/

# Copy web build
COPY --from=web-builder /app/web/dist /usr/share/nginx/html

# Expose ports
EXPOSE 80 8080 8081

# Set up systemd
CMD ["/sbin/init"] 