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

type User struct {
	ID         string
	Name       string
	Active     bool
	connection *websocket.Conn
}

// In memory users
var users = make(map[string]User)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session")
		user, ok := users[sessionID]

		if !ok {
			log.Println("User not found, creating new user")
			sessionID = "session-" + time.Now().Format("20060102150405")
			user := User{
				ID:     sessionID,
				Name:   sessionID,
				Active: false,
			}

			users[sessionID] = user
			c.Cookie(&fiber.Cookie{
				Name:  "session",
				Value: sessionID,
			})
		}

		handler := adaptor.HTTPHandler(templ.Handler(templates.Index(user.Active)))
		return handler(c)
	})

	app.Post("/sit", func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session")
		user, ok := users[sessionID]

		if ok {
			user.Active = true
			users[sessionID] = user

			// for _, u := range users {
			// 	if u.connection != nil {
			// 		message := components.Seat(true)
			// 		buffer := &bytes.Buffer{}
			// 		message.Render(context.Background(), buffer)

			// 		if err := u.connection.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
			// 			log.Println("Failed to write message")
			// 		}
			// 	}
			// }
		}

		handler := adaptor.HTTPHandler(templ.Handler(components.Seat(true)))
		return handler(c)
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/poker", websocket.New(func(c *websocket.Conn) {
		sessionID := c.Cookies("session")
		user, ok := users[sessionID]

		if !ok {
			log.Println("User not found")
		} else {
			user.connection = c
			users[sessionID] = user
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

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
	log.Fatal(app.Listen(":8080"))
}
