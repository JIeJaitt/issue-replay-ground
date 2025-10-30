# Fiber-MCP 设计文档

## 概述

Fiber-MCP 是将 Model Context Protocol (MCP) 集成到 Fiber 框架中的中间件，使 Fiber 应用能够作为 MCP 服务器与 AI 代理交互。

## 核心组件

### 1. MCP 服务器 (`server.go`)

```go
package mcp

import (
    "github.com/gofiber/fiber/v3"
)

type Server struct {
    config   Config
    tools    map[string]*Tool
    resources map[string]*Resource
    prompts  map[string]*Prompt
}

type Config struct {
    ServerName string
    Version    string
    Transport  TransportType // SSE, WebSocket, HTTP
    Endpoint   string        // 默认 /mcp
}

func New(config Config) fiber.Handler {
    server := &Server{
        config:    config,
        tools:     make(map[string]*Tool),
        resources: make(map[string]*Resource),
        prompts:   make(map[string]*Prompt),
    }
    return server.handle
}
```

### 2. 工具注册 (`tools.go`)

```go
package mcp

type Tool struct {
    Name        string
    Description string
    Handler     func(*Context) (*ToolResult, error)
    Metadata    ToolMetadata
}

type ToolMetadata struct {
    Name        string
    Description string
    InputSchema map[string]interface{}
}

type Context struct {
    Params map[string]interface{}
    Ctx    fiber.Ctx
}

type ToolResult struct {
    Content []Content
    IsError bool
}

type Content struct {
    Type string
    Text string
    Data map[string]interface{}
}

func (s *Server) RegisterTool(name string, handler func(*Context) (*ToolResult, error), metadata ToolMetadata) {
    s.tools[name] = &Tool{
        Name:        name,
        Description: metadata.Description,
        Handler:     handler,
        Metadata:    metadata,
    }
}
```

### 3. 资源管理 (`resources.go`)

```go
package mcp

type Resource struct {
    URI      string
    Name     string
    MimeType string
    Text     string
    Data     []byte
}

func (s *Server) RegisterResource(uri string, resource *Resource) {
    s.resources[uri] = resource
}
```

### 4. 提示模板 (`prompts.go`)

```go
package mcp

type Prompt struct {
    Name        string
    Description string
    Arguments   []PromptArgument
    Template    string
}

type PromptArgument struct {
    Name        string
    Description string
    Required    bool
}

func (s *Server) RegisterPrompt(name string, prompt *Prompt) {
    s.prompts[name] = prompt
}
```

### 5. 协议处理 (`protocol.go`)

```go
package mcp

import (
    "encoding/json"
    "github.com/gofiber/fiber/v3"
)

type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
    ID      interface{} `json:"id"`
}

type MCPResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    Result  interface{} `json:"result,omitempty"`
    Error   *MCPError   `json:"error,omitempty"`
    ID      interface{} `json:"id"`
}

type MCPError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func (s *Server) handle(c fiber.Ctx) error {
    var req MCPRequest
    if err := c.Bind().Body(&req); err != nil {
        return c.JSON(MCPResponse{
            JSONRPC: "2.0",
            Error: &MCPError{
                Code:    -32700,
                Message: "Parse error",
            },
            ID: req.ID,
        })
    }

    switch req.Method {
    case "initialize":
        return s.handleInitialize(c, req)
    case "tools/list":
        return s.handleToolsList(c, req)
    case "tools/call":
        return s.handleToolCall(c, req)
    case "resources/list":
        return s.handleResourcesList(c, req)
    case "resources/read":
        return s.handleResourceRead(c, req)
    case "prompts/list":
        return s.handlePromptsList(c, req)
    default:
        return c.JSON(MCPResponse{
            JSONRPC: "2.0",
            Error: &MCPError{
                Code:    -32601,
                Message: "Method not found",
            },
            ID: req.ID,
        })
    }
}
```

## 使用示例

### 基础用法

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/mcp"
)

func main() {
    app := fiber.New()

    // 初始化 MCP 服务器
    mcpServer := mcp.New(mcp.Config{
        ServerName: "my-fiber-app",
        Version:    "1.0.0",
        Transport:  mcp.HTTP,
        Endpoint:   "/mcp",
    })

    // 注册工具
    mcpServer.RegisterTool("get_user", getUserHandler, mcp.ToolMetadata{
        Name:        "get_user",
        Description: "根据ID获取用户信息",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "user_id": map[string]interface{}{
                    "type":        "string",
                    "description": "用户ID",
                },
            },
            "required": []string{"user_id"},
        },
    })

    // 注册资源
    mcpServer.RegisterResource("database://users", &mcp.Resource{
        URI:      "database://users",
        Name:     "用户数据库",
        MimeType: "application/json",
        Text:     `{"users": [...]}`,
    })

    app.Use("/mcp", mcpServer)

    app.Listen(":3000")
}

func getUserHandler(ctx *mcp.Context) (*mcp.ToolResult, error) {
    userID := ctx.Params["user_id"].(string)
    // 实现获取用户逻辑
    return &mcp.ToolResult{
        Content: []mcp.Content{
            {Type: "text", Text: "User found"},
        },
    }, nil
}
```

### 高级用法：与 Fiber 路由集成

```go
// 将 Fiber 路由自动暴露为 MCP 工具
mcpServer.AutoRegisterRoutes(app)

// 或选择性注册
mcpServer.RegisterRoute("/api/users/:id", mcp.RouteConfig{
    ToolName:        "get_user",
    Description:     "获取用户信息",
    ConvertParams:   true, // 自动转换路径参数
})
```

## 特性

### 1. 多种传输协议支持
- HTTP/JSON-RPC
- Server-Sent Events (SSE)
- WebSocket

### 2. 与 Fiber 深度集成
- 自动路由注册
- Context 共享
- 中间件链支持

### 3. 类型安全
- Go 类型系统
- 输入验证
- 错误处理

### 4. 性能优化
- 连接池
- 批量请求处理
- 响应缓存

## 实现优先级

### Phase 1: 核心功能
- [ ] MCP 协议基础实现
- [ ] 工具注册和调用
- [ ] 资源管理
- [ ] HTTP 传输支持

### Phase 2: 高级功能
- [ ] WebSocket 支持
- [ ] SSE 支持
- [ ] 自动路由注册
- [ ] 工具链组合

### Phase 3: 生态集成
- [ ] OpenAI 兼容
- [ ] Claude 兼容
- [ ] 本地模型支持
- [ ] 工具市场

## 对比 Flask-MCP

| 特性 | Flask-MCP | Fiber-MCP (建议) |
|------|-----------|------------------|
| 性能 | 中等 | 高（基于 fasthttp）|
| 并发模型 | 多线程 | Goroutine（更高效）|
| 类型安全 | 动态类型 | 静态类型（编译时检查）|
| 内存占用 | 较高 | 较低 |
| 部署 | 需要 WSGI | 可直接部署 |

## 总结

Fiber-MCP 不仅可以跟进 Flask-MCP，还有潜力提供更好的性能和类型安全。建议将其作为 Fiber 框架的官方中间件实现。

