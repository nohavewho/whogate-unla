# MCP Gateway 开发需求文档

## 🎯 项目目标

开发一个用于将传统 RESTful 接口接入 MCP 协议的 Gateway 服务，支持通过配置 YAML 文件自动注册工具（tools），对外暴露标准 MCP API，供大语言模型（LLMs）进行调用。

---

## 📁 配置结构（参考 `all-in-one.yaml`）

配置文件分为四个block部分：

- `global`: 命名空间与前缀设置
- `routers`: 路由 -> 服务server的映射关系，比如定义前缀
- `servers`: 服务配置（认证、工具白名单等）
- `tools`: 工具注册，包含参数定义、请求模版与响应模板

---

## ✅ 工具定义格式（标准化）

每个 Tool 支持以下字段：

- `name`: 工具名
- `description`: 工具说明
- `method`: 请求方法（GET/POST/PUT 等）
- `endpoint`: 请求地址（可包含变量模版）
- `headers`: 请求头配置（支持模版）
- `args`: 工具参数定义（供 LLMs 使用）
- `requestBody`: 请求体模版（支持 `{{args.xxx}}` 和 `{{request.headers.xxx}}` 等）
- `responseBody`: 响应体展示模版（支持 `{{response.xxx}}`）

---

## 🔁 参数模板变量说明

| 类型              | 模板变量                        | 来源说明             |
|-------------------|----------------------------------|----------------------|
| 工具参数          | `{{args.xxx}}`                   | LLM 调用时提供的参数 |
| 请求原始 Header   | `{{request.headers.xxx}}`        | HTTP 请求头          |
| 请求原始 Query    | `{{request.query.xxx}}`          | URL 查询参数         |
| 请求原始 Path     | `{{request.path.xxx}}`           | 路径参数             |
| 请求原始 Body     | `{{request.body.xxx}}`           | 请求体字段           |
| 响应字段          | `{{response.data.xxx}}`          | 响应体 JSON 数据     |

---

## ✅ 示例模板用法

```yaml
headers:
  Authorization: "{{args.accessToken}}"
  X-Trace-Id: "{{request.headers.X-Trace-Id}}"

requestBody: |-
  {
    "username": "{{args.username}}",
    "email": "{{args.email}}",
    "ua": "{{request.headers.User-Agent}}"
  }

responseBody: |-
  注册成功 🎉
  - 用户名：{{response.data.username}}
  - 邮箱：{{response.data.email}}
```

---

## 🧱 模板上下文结构体（Go建议）

```go
type TemplateContext struct {
  Args    map[string]any
  Request struct {
    Headers map[string]string
    Query   map[string]string
    Path    map[string]string
    Body    map[string]any
  }
  Response map[string]any
}
```

---

## 🛠️ MVP 实现任务拆解

### Step 1: 配置解析
- 支持加载 `all-in-one.yaml` 并解析为结构体
- 校验 tool 名、server 名唯一性，prefix 不冲突

### Step 2: 路由注册
- 根据 router.prefix + tool.name 注册 handler
- 生成路径如：`/api/v1/user/user_register`

### Step 3: 参数提取与绑定
- 提取 headers/query/path/body → request map
- 映射 args → 合并为 `TemplateContext`

### Step 4: 模板渲染与请求转发
- 渲染 endpoint / headers / requestBody
- 发送实际 HTTP 请求并获取响应
- 使用 responseBody 模板生成用户可读响应

### Step 5: 认证机制（MVP 支持）
- `none`：默认开放
- `bearer`：Authorization 头认证
- `apikey`：自定义 Header 认证

---

## 📦 示例 Tool 路径规划（自动注册）

| Tool 名称             | 路径                           |
|------------------------|--------------------------------|
| `user_register`        | `/api/v1/user/user_register`   |
| `user_location_get`    | `/api/v1/map/user_location_get`|
