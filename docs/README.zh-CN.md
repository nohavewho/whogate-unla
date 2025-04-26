# MCP Gateway

> 🚀 将现有 API 快速转化为 [MCP](https://modelcontextprotocol.io/) 服务，无需改动任何一行代码。

[![English](https://img.shields.io/badge/English-Click-yellow)](../README.md)
[![简体中文](https://img.shields.io/badge/简体中文-点击查看-orange)](README.zh-CN.md)
[![Release](https://img.shields.io/github/v/release/mcp-ecosystem/mcp-gateway)](https://github.com/mcp-ecosystem/mcp-gateway/releases)
[![文档](https://img.shields.io/badge/文档-在线阅读-blue)](https://mcp.ifuryst.com)
[![Discord](https://img.shields.io/badge/Discord-加入讨论-5865F2?logo=discord&logoColor=white)](https://discord.gg/udf69cT9TY)

---

> ⚡ **注意**：MCP Gateway 正在快速迭代中！我们会尽力保证向下兼容，但无法百分百承诺兼容性。升级版本时一定要留意版本变更情况️。同时由于更新频繁，文档内容可能存在一定延迟，如遇问题欢迎通过 [Discord](https://discord.gg/udf69cT9TY) 或 [Issue](https://github.com/mcp-ecosystem/mcp-gateway/issues) 搜索和求助❤️

---

## ✨ MCP Gateway 是什么？

**MCP Gateway** 是一个用 Go 编写的轻量高可用网关服务，帮助个人与企业将已有的 API 通过配置方式转换为符合 [MCP 协议](https://modelcontextprotocol.io/) 的服务，无需改动任何代码。

https://github.com/user-attachments/assets/2a812a14-85cf-45d6-9f37-cc08d8579b33

### 🔧 核心设计理念

- ✅ 零侵入：平台中立，适配物理机、虚拟机、ECS、K8s 等环境，无需改动现有基础设施
- 🔄 配置驱动：通过 YAML 配置即可将存量 API 转换为 MCP Server，无需改代码
- 🪶 轻量高效：架构极致轻量，拒绝在性能与高可用性上妥协
- 🧭 内置管理界面：开箱即用的 Web UI，降低学习与运维成本

---

## 🚀 快速开始

MCP Gateway 提供开箱即用的 Docker 启动方式。完整部署与配置说明请参考 [文档](https://mcp.ifuryst.com/getting-started/quick-start)。

### Docker 方式运行

```bash
mkdir -p mcp-gateway/{configs,data}
cd mcp-gateway/
curl -sL https://raw.githubusercontent.com/mcp-ecosystem/mcp-gateway/refs/heads/main/configs/apiserver.yaml -o configs/apiserver.yaml
curl -sL https://raw.githubusercontent.com/mcp-ecosystem/mcp-gateway/refs/heads/main/configs/mcp-gateway.yaml -o configs/mcp-gateway.yaml
curl -sL https://raw.githubusercontent.com/mcp-ecosystem/mcp-gateway/refs/heads/main/.env.example -o .env.allinone

docker run -d \
           --name mcp-gateway \
           -p 8080:80 \
           -p 5234:5234 \
           -p 5235:5235 \
           -p 5335:5335 \
           -p 5236:5236 \
           -e ENV=production \
           -v $(pwd)/configs:/app/configs \
           -v $(pwd)/data:/app/data \
           -v $(pwd)/.env.allinone:/app/.env \
           --restart unless-stopped \
           ghcr.io/mcp-ecosystem/mcp-gateway/allinone:latest
```

> 在中国境内的设备可以拉阿里云仓库的镜像
>
> ```bash
> registry.ap-southeast-1.aliyuncs.com/mcp-ecosystem/mcp-gateway-allinone:latest
> ```

访问 http://localhost:8080/ 开始配置使用

📖 查看完整指南 → [快速开始文档 »](https://mcp.ifuryst.com/getting-started/quick-start)

---

## 📋 待办事项

- [x] RESTful API 到 MCP-Server 的转换
- [ ] gRPC 到 MCP-Server 的转换
- [x] 请求/响应体转换
- [x] 管理界面
- [ ] 会话持久化
- [x] MCP SSE支持
- [x] MCP Streamable HTTP支持
- [ ] 可迁移可恢复会话
- [ ] 前置认证
- [ ] 配置版本控制
- [ ] 分布式配置持久化支持
- [ ] 服务多副本支持
- [x] Docker 支持
- [ ] Kubernetes 集成
- [ ] Helm 支持

---

## 📚 文档

更多使用方式、配置示例、集成说明请访问文档站点：

👉 **https://mcp.ifuryst.com**

---

## 📄 许可证

本项目采用 [MIT 协议](../LICENSE)。

