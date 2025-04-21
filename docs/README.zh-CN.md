# MCP Gateway

[![English](https://img.shields.io/badge/English-Click-yellow)](../README.en.md)
[![简体中文](https://img.shields.io/badge/简体中文-点击查看-orange)](README.zh-CN.md)

MCP Gateway 是一个基于 Go 语言开发的轻量但高可用的网关服务，旨在让个人和企业在MCP(Model Control Protocol)浪潮下可以无痛将存量的API服务（RESTful, gRPC等）通过配置的方式转成MCP-Server

非常存粹的目的和功能特性：
- 平台中立，不管是物理机、虚拟机、ECS、K8s等场景都可以无痛接入，无需对基础设施动手
- 支持多种协议的转换，RESTful、gRPC等都可以通过配置的方式转成MCP-Server
- 追求性能和轻松的多副本高可用，轻量但不对可用及性能妥协
- 简单直观的管理页面，拒绝学习和运维成本

https://github.com/user-attachments/assets/2a812a14-85cf-45d6-9f37-cc08d8579b33

## 功能特性

- ⚙️ 零侵入式接入
- 🪶 轻量设计，易于部署
- 💡 请求头、参数、请求体和响应体等参数透传
- 🧭 管理后台直观易用

## 待办事项

- [x] RESTful API 到 MCP-Server 的转换
- [ ] gRPC 到 MCP-Server 的转换
- [x] 请求/响应体转换
- [x] 管理界面
- [ ] 会话持久化
- [x] Streamable HTTP
- [ ] 可迁移可恢复会话
- [ ] 前置认证
- [ ] 配置版本控制
- [ ] 分布式配置持久化支持
- [ ] 服务多副本支持
- [ ] Docker 支持
- [ ] Kubernetes 集成
- [ ] Helm 支持

## 快速开始

### 后端服务

#### 网关服务

1. 克隆项目
    ```bash
    git clone https://github.com/mcp-ecosystem/mcp-gateway.git
    cd mcp-gateway
    ```

2. 安装依赖
    ```bash
    go mod download
    ```

3. 运行服务
    ```bash
    go run ./cmd/mcp-gateway/main.go
    ```

#### 管理服务

1. 克隆项目
    ```bash
    git clone https://github.com/mcp-ecosystem/mcp-gateway.git
    cd mcp-gateway
    ```

2. 安装依赖
    ```bash
    go mod download
    ```

3. 运行服务
    ```bash
    go run cmd/apiserver/main.go
    ```

### 前端开发

1. 进入前端目录
    ```bash
    cd web
    ```

2. 安装依赖
    ```bash
    npm install
    ```

3. 启动开发服务器
    ```bash
    npm run dev
    ```

## 项目结构

```
.
├── cmd/            # 后端服务入口
├── configs/        # 配置文件
├── internal/       # 内部包
├── pkg/            # 公共包
├── web/            # 前端代码
└── docs/           # 项目文档
```

## 配置说明

配置文件位于 `configs` 目录下，支持 YAML 格式的配置文件。主要配置项包括：

- 🖥️ 服务器配置
- 🔀 路由规则
- 🔐 工具权限
- ⚙️ 系统参数

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件


### 后端服务

#### 网关服务

1. 克隆项目
```bash
git clone https://github.com/mcp-ecosystem/mcp-gateway.git
cd mcp-gateway
```

2. 安装依赖
```bash
go mod download
```

3. 运行服务
```bash
go run ./cmd/mcp-gateway/main.go
```

#### 管理服务

1. 克隆项目
```bash
git clone https://github.com/mcp-ecosystem/mcp-gateway.git
cd mcp-gateway
```

2. 安装依赖
```bash
go mod download
```

3. 运行服务
```bash
go run cmd/apiserver/main.go
```

### 前端开发

1. 进入前端目录
```bash
cd web
```

2. 安装依赖
```bash
npm install
```

3. 启动开发服务器
```bash
npm run dev
```

## 项目结构

```
.
├── cmd/            # 后端服务入口
├── configs/        # 配置文件
├── internal/       # 内部包
├── pkg/            # 公共包
├── web/            # 前端代码
└── docs/           # 项目文档
```

## 配置说明

配置文件位于 `configs` 目录下，支持 YAML 格式的配置文件。主要配置项包括：

- 服务器配置
- 路由规则
- 工具权限
- 系统参数

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](../LICENSE) 文件
