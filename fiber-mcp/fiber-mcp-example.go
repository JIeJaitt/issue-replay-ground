// fiber-mcp 使用示例

package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/mcp"
)

func main() {
	app := fiber.New()

	// 初始化 MCP 服务器
	mcpConfig := mcp.Config{
		ServerName: "fiber-mcp-example",
		Version:    "1.0.0",
		Transport:  mcp.HTTP,
		Endpoint:   "/mcp",
	}

	mcpServer := mcp.New(mcpConfig)

	// 示例 1: 注册数据库查询工具
	mcpServer.RegisterTool("query_users", func(ctx *mcp.Context) (*mcp.ToolResult, error) {
		// 从参数中获取查询条件
		limit := 10
		if val, ok := ctx.Params["limit"].(float64); ok {
			limit = int(val)
		}

		// 模拟数据库查询
		users := []map[string]interface{}{
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
		}

		return &mcp.ToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Found %d users", len(users)),
				},
				{
					Type: "json",
					Data: map[string]interface{}{"users": users},
				},
			},
		}, nil
	}, mcp.ToolMetadata{
		Name:        "query_users",
		Description: "查询用户列表，支持limit参数限制返回数量",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回的用户数量限制",
					"default":     10,
				},
			},
		},
	})

	// 示例 2: 注册文件操作工具
	mcpServer.RegisterTool("read_file", func(ctx *mcp.Context) (*mcp.ToolResult, error) {
		filepath, ok := ctx.Params["filepath"].(string)
		if !ok {
			return &mcp.ToolResult{
				IsError: true,
				Content: []mcp.Content{
					{Type: "text", Text: "filepath parameter is required"},
				},
			}, nil
		}

		// 这里可以读取文件内容
		content := fmt.Sprintf("Content of %s", filepath)

		return &mcp.ToolResult{
			Content: []mcp.Content{
				{Type: "text", Text: content},
			},
		}, nil
	}, mcp.ToolMetadata{
		Name:        "read_file",
		Description: "读取文件内容",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filepath": map[string]interface{}{
					"type":        "string",
					"description": "要读取的文件路径",
				},
			},
			"required": []string{"filepath"},
		},
	})

	// 示例 3: 注册资源
	mcpServer.RegisterResource("database://stats", &mcp.Resource{
		URI:      "database://stats",
		Name:     "应用统计信息",
		MimeType: "application/json",
		Text:     `{"total_users": 1000, "active_users": 500}`,
	})

	// 示例 4: 注册提示模板
	mcpServer.RegisterPrompt("user_summary", &mcp.Prompt{
		Name:        "user_summary",
		Description: "生成用户摘要",
		Arguments: []mcp.PromptArgument{
			{Name: "user_id", Description: "用户ID", Required: true},
			{Name: "format", Description: "输出格式 (markdown/json)", Required: false},
		},
		Template: `请为用户 {{user_id}} 生成摘要，格式：{{format}}`,
	})

	// 注册 MCP 路由
	app.Use("/mcp", mcpServer)

	// 普通 API 路由（可与 MCP 共存）
	app.Get("/api/users", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "This is a regular API endpoint",
		})
	})

	// 健康检查
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"mcp":    "enabled",
		})
	})

	log.Println("Fiber app with MCP server starting on :3000")
	log.Println("MCP endpoint: http://localhost:3000/mcp")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

// 高级用法：自动将 Fiber 路由暴露为 MCP 工具
func setupAutoMCPTools(app *fiber.App, mcpServer *mcp.Server) {
	// 获取所有路由
	routes := app.GetRoutes()

	for _, route := range routes {
		// 为每个路由自动创建 MCP 工具
		mcpServer.RegisterRoute(route.Path, mcp.RouteConfig{
			ToolName:      fmt.Sprintf("api_%s", route.Method),
			Description:   fmt.Sprintf("调用 %s %s API", route.Method, route.Path),
			HTTPMethod:    route.Method,
			ConvertParams: true,
		})
	}
}
