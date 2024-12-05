package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		// handler := adaptor.HTTPHandler(templ.Handler(index))

		// return handler(c)
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Listen(":3000"))
}
