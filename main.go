package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/ottigan/planning-poker/templates"
	"github.com/ottigan/planning-poker/templates/components"
	"github.com/ottigan/planning-poker/types"
)

type User = types.User

var timerStart = time.Now()
var ticker = time.NewTicker(100 * time.Millisecond)
var users = make(map[string]User)
var showResult = false

func getVotedUsers() int {
	votedUsers := 0
	for _, user := range users {
		if user.Vote != 0 {
			votedUsers++
		}
	}

	return votedUsers
}

func calculateMinAvgMax() types.Stats {
	votes := make([]float64, 0)

	if len(users) == 0 || !showResult || getVotedUsers() == 0 {
		return types.Stats{}
	}

	log.Println("Calculating min, avg, max", len(votes))

	for _, user := range users {
		if user.Vote != 0 {
			votes = append(votes, float64(user.Vote))
		}
	}

	min := math.Inf(1)
	max := math.Inf(-1)
	sum := 0.0

	for _, vote := range votes {
		if vote < min {
			min = vote
		}

		if vote > max {
			max = vote
		}

		sum += vote
	}

	avg := sum / float64(len(votes))

	return types.Stats{
		Min: strconv.Itoa(int(min)),
		Avg: strconv.Itoa(int(avg)),
		Max: strconv.Itoa(int(max)),
	}
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		isAdmin := c.Query("admin") == "true"
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

			// Long cookie
			c.Cookie(&fiber.Cookie{
				Name:   "session",
				Value:  sessionID,
				MaxAge: 60 * 60 * 24,
			})
		}

		stats := calculateMinAvgMax()
		votedUsers := getVotedUsers()
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

	app.Get("/ws/poker", websocket.New(func(c *websocket.Conn) {
		log.Println("Websocket connection established")
		sessionID := c.Cookies("session")
		user, ok := users[sessionID]

		if ok {
			user.Connection = c
			users[sessionID] = user
		}

		for range ticker.C {
			timePassed := fmt.Sprintf("%.0fs", time.Since(timerStart).Seconds())
			message := components.Time(timePassed)
			buffer := &bytes.Buffer{}
			message.Render(context.Background(), buffer)

			if err := c.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
				log.Println("Failed to write message")
				return
			}
		}
	}))

	app.Post("/vote/:n", func(c *fiber.Ctx) error {
		log.Println("Vote received")
		sessionID := c.Cookies("session")
		user, ok := users[sessionID]
		vote, _ := strconv.Atoi(c.Params("n"))

		if ok {
			user.Vote = vote
			users[sessionID] = user
		}

		for _, u := range users {
			if u.Connection != nil {
				buffer := &bytes.Buffer{}
				votedUsers := getVotedUsers()
				components.Votes(votedUsers).Render(context.Background(), buffer)

				if err := u.Connection.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
					log.Println("Failed to write message")
				}
			}
		}

		handler := adaptor.HTTPHandler(templ.Handler(components.Voter(vote, showResult)))
		return handler(c)
	})

	app.Post("/show", func(c *fiber.Ctx) error {
		log.Println("Showing results")
		showResult = true
		ticker.Stop()

		for _, u := range users {
			if u.Connection != nil {
				votedUsers := getVotedUsers()
				buffer := &bytes.Buffer{}
				components.Votes(votedUsers).Render(context.Background(), buffer)
				components.Voter(u.Vote, showResult).Render(context.Background(), buffer)
				components.Result(calculateMinAvgMax()).Render(context.Background(), buffer)

				if err := u.Connection.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
					log.Println("Failed to write message")
				}
			}
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Post("/reset", func(c *fiber.Ctx) error {
		log.Println("Resetting votes")
		timerStart = time.Now()
		showResult = false

		for _, user := range users {
			user.Vote = 0
			users[user.ID] = user
		}

		for _, u := range users {
			log.Println("Sending reset message")

			if u.Connection != nil {
				voter := components.Voter(0, showResult)
				message := components.Votes(0)
				buffer := &bytes.Buffer{}
				message.Render(context.Background(), buffer)
				voter.Render(context.Background(), buffer)
				components.Time("0s").Render(context.Background(), buffer)
				components.Result(calculateMinAvgMax()).Render(context.Background(), buffer)

				if err := u.Connection.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
					log.Println("Failed to write message")
				}
			}
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Static("/static", "./static")
	log.Fatal(app.Listen(":8080"))
}
