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

const tickerInterval = 1000 * time.Millisecond

var timerStart = time.Now()
var ticker = time.NewTicker(tickerInterval)
var showResult = false

func broadcastTicker(um *internal.Manager) {
	for range ticker.C {
		timePassed := fmt.Sprintf("%.0fs", time.Since(timerStart).Seconds())
		message := components.Time(timePassed)
		buffer := &bytes.Buffer{}
		message.Render(context.Background(), buffer)
		um.Broadcast(buffer.Bytes())
	}
}

func sendUserState(um *internal.Manager, user internal.User) {
	buffer := &bytes.Buffer{}
	components.Voter(user.Vote, showResult).Render(context.Background(), buffer)
	components.Votes(internal.GetVotedUsers(um)).Render(context.Background(), buffer)
	components.Result(internal.CalculateMinAvgMax(um, showResult)).Render(context.Background(), buffer)

	if user.Connection != nil {
		if err := user.Connection.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
			log.Printf("Failed to write message to user %s: %v", user.Id, err)
		}
	}
}

func main() {
	um := internal.CreateUserManager()
	go broadcastTicker(um)
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		isAdmin := c.Query("root") == "true" || c.Cookies("root") == "true"
		sessionID := c.Cookies("session", strconv.FormatInt(time.Now().UnixNano(), 10))
		user, ok := um.Get(sessionID)

		if !ok {
			user = um.New(internal.User{Id: sessionID})
		}

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

	app.Post("/name", func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session")
		name := c.FormValue("name")

		if _, ok := um.Update(sessionID, internal.WithName(name)); !ok {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		c.WriteString(name)
		return c.SendStatus(fiber.StatusOK)
	})

	app.Post("/vote/:n", func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session")
		vote, _ := strconv.Atoi(c.Params("n"))
		um.Update(sessionID, internal.WithVote(vote))
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

		for _, user := range um.GetAll() {
			sendUserState(um, user)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Post("/reset", func(c *fiber.Ctx) error {
		timerStart = time.Now()
		ticker.Reset(tickerInterval)
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
		sessionID := c.Cookies("session")

		if _, ok := um.Update(sessionID, internal.WithConnection(c)); !ok {
			return
		}

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				log.Printf("Websocket connection closed for user %s: %v", sessionID, err)
				um.Update(sessionID, internal.WithConnection(nil))
				return
			}
		}
	}))

	app.Static("/static", "./static")
	log.Fatal(app.Listen(":8080"))
}
