package main

import (
	"log" // 添加日志包以便观察

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func main() {
	// 启动目标服务器
	go func() {
		targetApp := fiber.New()
		targetApp.Get("/", func(c fiber.Ctx) error {
			log.Println("Target server received request")
			// 在目标服务器上设置一个自定义头，用于验证
			c.Set("X-Target-Server", "true")
			return c.SendString("Hello from target server!")
		})
		log.Println("Starting target server on :7000")
		if err := targetApp.Listen(":7000"); err != nil {
			log.Fatalf("Failed to start target server: %v", err)
		}
	}()

	// 主 Fiber 应用
	app := fiber.New()

	// Session 中间件
	_, sessionStore := session.NewWithStore()
	app.Use(session.New(session.Config{
		Store: sessionStore,
	}))
	log.Println("Using Session middleware")

	// CSRF 中间件
	app.Use(csrf.New(csrf.Config{
		Session: sessionStore,
	}))
	log.Println("Using CSRF middleware")

	// 添加一个自定义中间件来检查和记录响应头
	app.Use(func(c fiber.Ctx) error {
		// 在代理前设置一个自定义头
		c.Set("X-Before-Proxy", "true")

		// 继续处理请求
		err := c.Next()

		// 在响应返回前记录所有头信息
		log.Println("=== Response Headers After Processing ===")
		c.Response().Header.VisitAll(func(key, value []byte) {
			log.Printf("%s: %s", string(key), string(value))
		})

		return err
	})

	// 代理路由
	app.Get("/", func(c fiber.Ctx) error {
		log.Println("Proxying request to localhost:7000")

		// 在代理前记录头信息
		log.Println("=== Headers Before Proxy ===")
		c.Response().Header.VisitAll(func(key, value []byte) {
			log.Printf("%s: %s", string(key), string(value))
		})

		err := proxy.Do(c, "http://localhost:7000")

		// 在代理后记录头信息
		log.Println("=== Headers After Proxy ===")
		c.Response().Header.VisitAll(func(key, value []byte) {
			log.Printf("%s: %s", string(key), string(value))
		})

		if err != nil {
			log.Printf("Proxy error: %v", err)
			return err
		}

		// 在代理后设置一个自定义头，看是否会保留
		c.Set("X-After-Proxy", "true")

		log.Println("Proxy finished, returning response")
		return nil
	})

	// 添加一个不使用代理的路由，用于比较
	app.Get("/direct", func(c fiber.Ctx) error {
		return c.SendString("Hello directly from main server!")
	})

	log.Println("Starting main server on :8000")
	if err := app.Listen(":8000"); err != nil {
		log.Fatalf("Failed to start main server: %v", err)
	}
}

/*
运行步骤:
1. go mod init proxycsrf // 或你的项目名
2. go mod tidy
3. go run main.go
4. 在另一个终端执行: curl -v http://localhost:8000/

观察 curl 的输出，特别是响应头部分，看是否有 Set-Cookie: csrf_=...
预期：没有 csrf_ cookie。
*/
