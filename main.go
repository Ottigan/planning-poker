package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/ottigan/planning-poker/templates"
	"github.com/ottigan/planning-poker/templates/components"
)

var connectionCount int

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		handler := adaptor.HTTPHandler(templ.Handler(templates.Index()))
		return handler(c)
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/poker", websocket.New(func(c *websocket.Conn) {
		connectionCount++
		log.Printf("Current connections: %d\n", connectionCount)
		defer func() {
			connectionCount--
			log.Printf("Current connections: %d\n", connectionCount)
		}()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// range over ticker.C
		for range ticker.C {
			message := components.Time(time.Now().Format("15:04:05"))
			buffer := &bytes.Buffer{}
			message.Render(context.Background(), buffer)

			if err := c.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
				log.Println("Failed to write message")
				return
			}
		}
	}))

	app.Static("/static", "./static")
	log.Fatal(app.Listen(":3000"))
}
