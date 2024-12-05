package main

import (
	"log"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/ottigan/planning-poker/templates"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		handler := adaptor.HTTPHandler(templ.Handler(templates.Index()))

		return handler(c)
	})

	app.Get("/css/output.css", static.New("./css/output.css"))
	log.Fatal(app.Listen(":3000"))
}
