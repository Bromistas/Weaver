package main

import (
	"log"
)

func main() {
	app := fiber.New()

	app.Get("/query/:param", func(c *fiber.Ctx) error {
		param := c.Params("param")
		return c.SendString("Received param: " + param)
	})

	log.Fatal(app.Listen(":3000"))
}
