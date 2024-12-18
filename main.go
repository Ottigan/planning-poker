package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/ottigan/planning-poker/internal"
	"github.com/ottigan/planning-poker/templates"
	"github.com/ottigan/planning-poker/templates/components"
)

var timerStart = time.Now()
var ticker = time.NewTicker(1000 * time.Millisecond)
var showResult = false

func BroadcastTicker(um *internal.Manager) {
	for range ticker.C {
		timePassed := fmt.Sprintf("%.0fs", time.Since(timerStart).Seconds())
		message := components.Time(timePassed)
		buffer := &bytes.Buffer{}
		message.Render(context.Background(), buffer)
		um.Broadcast(buffer.Bytes())
	}
}

func main() {
	um := internal.CreateUserManager()
	go BroadcastTicker(um)
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		isAdmin := c.Query("root") == "true" || c.Cookies("root") == "true"
		sessionID := c.Cookies("session")
		user, ok := um.Get(sessionID)

		if !ok {
			log.Println("User not found, creating new user")
			sessionID = um.New()

			c.Cookie(&fiber.Cookie{
				Name:   "session",
				Value:  sessionID,
				MaxAge: 60 * 60 * 24,
			})

			if isAdmin {
				c.Cookie(&fiber.Cookie{
					Name:   "root",
					Value:  "true",
					MaxAge: 60 * 60 * 24,
				})
			}
		}

		stats := internal.CalculateMinAvgMax(um, showResult)
		votedUsers := internal.GetVotedUsers(um)
		handler := adaptor.HTTPHandler(templ.Handler(templates.Index(user, votedUsers, isAdmin, showResult, stats)))
		return handler(c)
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			log.Println("Upgrading to websocket")
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Post("/vote/:n", func(c *fiber.Ctx) error {
		log.Println("Vote received")
		sessionID := c.Cookies("session")
		vote, _ := strconv.Atoi(c.Params("n"))
		um.SetVote(sessionID, vote)
		buffer := &bytes.Buffer{}
		votedUsers := internal.GetVotedUsers(um)
		components.Votes(votedUsers).Render(context.Background(), buffer)

		um.Broadcast(buffer.Bytes())

		handler := adaptor.HTTPHandler(templ.Handler(components.Voter(vote, showResult)))
		return handler(c)
	})

	app.Post("/show", func(c *fiber.Ctx) error {
		log.Println("Showing results")
		showResult = true
		ticker.Stop()
		users := um.GetAll()

		for _, user := range users {
			if user.Connection != nil {
				buffer := &bytes.Buffer{}
				components.Voter(user.Vote, showResult).Render(context.Background(), buffer)
				components.Votes(internal.GetVotedUsers(um)).Render(context.Background(), buffer)
				components.Result(internal.CalculateMinAvgMax(um, showResult)).Render(context.Background(), buffer)

				if err := user.Connection.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
					log.Printf("Failed to write message to user %s: %v", user.ID, err)
				}
			}
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Post("/reset", func(c *fiber.Ctx) error {
		log.Println("Resetting votes")
		timerStart = time.Now()
		ticker.Reset(1000 * time.Millisecond)
		showResult = false
		um.ResetVotes()
		voter := components.Voter(0, showResult)
		message := components.Votes(0)
		buffer := &bytes.Buffer{}
		message.Render(context.Background(), buffer)
		voter.Render(context.Background(), buffer)
		components.Time("0s").Render(context.Background(), buffer)
		components.Result(internal.CalculateMinAvgMax(um, showResult)).Render(context.Background(), buffer)

		um.Broadcast(buffer.Bytes())
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/ws/poker", websocket.New(func(c *websocket.Conn) {
		log.Println("Websocket connection established")
		sessionID := c.Cookies("session")
		um.SetConnection(sessionID, c)

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				log.Printf("Websocket connection closed for user %s: %v", sessionID, err)
				um.RemoveConnection(sessionID)
				return
			}
		}
	}))

	app.Static("/static", "./static")
	log.Fatal(app.Listen(":8080"))
}
