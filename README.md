# MCP Gateway

[![English](https://img.shields.io/badge/English-Click-yellow)](README.md)
[![简体中文](https://img.shields.io/badge/简体中文-点击查看-orange)](./docs/README.zh-CN.md)

MCP Gateway is a lightweight yet highly available gateway service developed in Go, designed to help individuals and businesses easily convert their existing API services (RESTful, gRPC, etc.) into MCP-Servers through configuration under the wave of MCP (Model Control Protocol).

Clear Purpose and Key Features:
- 🌐 **Platform Agnostic**: Can be integrated easily in any environment—bare metal, virtual machines, ECS, K8s—without touching the infrastructure.
- 🔁 **Multi-protocol Conversion**: Supports converting RESTful and gRPC APIs to MCP-Server through configuration.
- ⚡️ **High Performance and Replication-Friendly**: Lightweight without compromising availability and performance.
- 🧭 **User-Friendly Management UI**: Minimizes learning and maintenance costs.

https://github.com/user-attachments/assets/2a812a14-85cf-45d6-9f37-cc08d8579b33

## Features

- ⚙️ Zero-intrusion integration
- 🪶 Lightweight and easy to deploy
- 💡 Transparent pass-through for headers, parameters, body, and response
- 🧭 Intuitive admin interface

## Quick Start

### Backend Service

#### Gateway Service

1. Clone the project
    ```bash
    git clone https://github.com/mcp-ecosystem/mcp-gateway.git
    cd mcp-gateway
    ```

2. Install dependencies
    ```bash
    go mod download
    ```

3. Run the service
    ```bash
    go run ./cmd/mcp-gateway/main.go
    ```

#### Management Service

1. Clone the project
    ```bash
    git clone https://github.com/mcp-ecosystem/mcp-gateway.git
    cd mcp-gateway
    ```

2. Install dependencies
    ```bash
    go mod download
    ```

3. Run the service
    ```bash
    go run cmd/apiserver/main.go
    ```

### Frontend Development

1. Navigate to the frontend directory
    ```bash
    cd web
    ```

2. Install dependencies
    ```bash
    npm install
    ```

3. Start the development server
    ```bash
    npm run dev
    ```

## Project Structure

```
.
├── cmd/            # Backend service entry points
├── configs/        # Configuration files
├── internal/       # Internal packages
├── pkg/            # Shared packages
├── web/            # Frontend code
└── docs/           # Project documentation
```

## Configuration

Configuration files are located in the `configs` directory and support YAML format. Key configuration items include:

- 🖥️ Server settings
- 🔀 Routing rules
- 🔐 Tool permissions
- ⚙️ System parameters

## Contribution Guide

1. Fork the project
2. Create a feature branch
3. Commit your changes
4. Push to your branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.  
