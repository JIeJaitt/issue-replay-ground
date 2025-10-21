package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Fiber Great",
		Prefork: true,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		time.Sleep(time.Second * 5)
		return c.SendString("Hello")
	})

	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutdown Server Beginning...")
	_ = app.ShutdownWithContext(ctx)
	log.Info("Running cleanup tasks...")
	log.Info("Exited Server...")
}
