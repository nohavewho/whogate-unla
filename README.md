# MCP Gateway

> 🚀 Instantly transform your existing APIs into [MCP](https://modelcontextprotocol.io/) servers — without changing a line of code.

[![English](https://img.shields.io/badge/English-Click-yellow)](./README.md)
[![简体中文](https://img.shields.io/badge/简体中文-点击查看-orange)](docs/README.zh-CN.md)
[![Release](https://img.shields.io/github/v/release/mcp-ecosystem/mcp-gateway)](https://github.com/mcp-ecosystem/mcp-gateway/releases)
[![Docs](https://img.shields.io/badge/Docs-View%20Online-blue)](https://mcp.ifuryst.com)
[![Discord](https://img.shields.io/badge/Discord-Join%20our%20Discord-5865F2?logo=discord&logoColor=white)](https://discord.gg/udf69cT9TY)

---

## 🎯 Support Us on Product Hunt

We just launched **MCP Gateway** on Product Hunt! 🚀  
If you find this project interesting or helpful, we'd love your support.  
Every upvote helps more developers discover it! ❤️

<a href="https://www.producthunt.com/posts/mcp-gateway?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-mcp&#0045;gateway" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=958310&theme=light&t=1745767484477" alt="MCP&#0032;Gateway - Turn&#0032;APIs&#0032;into&#0032;MCP&#0032;endpoints&#0044;without&#0032;changing&#0032;a&#0032;line&#0032;of&#0032;code | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

---

> ⚡ **Note**: MCP Gateway is under rapid development! We strive to maintain backward compatibility, but it cannot be 100% guaranteed. Please make sure to check version changes carefully when upgrading. Due to the fast iteration, documentation updates may sometimes lag behind. If you encounter any issues, feel free to search or ask for help via [Discord](https://discord.gg/udf69cT9TY) or [Issues](https://github.com/mcp-ecosystem/mcp-gateway/issues) ❤️

---

## ✨ What is MCP Gateway?

**MCP Gateway** is a lightweight and highly available gateway service written in Go. It enables individuals and organizations to convert their existing APIs into services compliant with the [MCP Protocol](https://modelcontextprotocol.io/) — all through configuration, with **zero code changes**.

https://github.com/user-attachments/assets/2a812a14-85cf-45d6-9f37-cc08d8579b33

### 🔧 Core Design Principles

- ✅ Zero Intrusion: Platform-agnostic, supports deployment on bare metal, VMs, ECS, Kubernetes, etc., without modifying existing infrastructure
- 🔄 Configuration-Driven: Convert legacy APIs to MCP Servers using YAML configuration — no code required
- 🪶 Lightweight & Efficient: Designed for minimal resource usage without compromising on performance or availability
- 🧭 Built-in Management UI: Ready-to-use web interface to simplify setup and reduce operational overhead

---

## 🚀 Getting Started

MCP Gateway supports a ready-to-run Docker deployment. Full deployment and configuration instructions are available in the [docs](https://mcp.ifuryst.com/getting-started/quick-start).

### Run with Docker

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

> For users in China, you can pull the image from Alibaba Cloud registry:
>
> ```bash
> registry.ap-southeast-1.aliyuncs.com/mcp-ecosystem/mcp-gateway-allinone:latest
> ```

Visit http://localhost:8080/ to start configuring.

📖 Read the full guide → [Quick Start »](https://mcp.ifuryst.com/getting-started/quick-start)

---

## 📋 TODOs

- [x] Convert RESTful API to MCP-Server
- [ ] Convert gRPC to MCP-Server
- [x] Request/Response body transformation
- [x] Management interface
- [x] Session persistence
- [x] MCP SSE support
- [x] MCP Streamable HTTP support
- [ ] Migratable and restorable sessions
- [ ] Pre-request authentication
- [ ] Configuration versioning
- [ ] Distributed configuration persistence
- [ ] Multi-replica service support
- [x] Docker support
- [ ] Kubernetes integration
- [ ] Helm chart support

---

## 📚 Documentation

For more usage patterns, configuration examples, and integration guides, please visit:

👉 **https://mcp.ifuryst.com**

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

## 💬 Join Our WeChat Community

Scan the QR code below to add us on WeChat. Please include a note: `mcp-gateway` or `mcpgw`.

<img src="web/public/wechat-qrcode.png" alt="WeChat QR Code" width="350" height="350" />

## 📈 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=mcp-ecosystem/mcp-gateway&type=Date)](https://star-history.com/#mcp-ecosystem/mcp-gateway&Date)
