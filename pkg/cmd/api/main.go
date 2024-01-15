package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
	"we-know/pkg/api/handlers"
	"we-know/pkg/infrastructure/database"
)

func main() {
	app := fiber.New()
	app.Use(cors.New())
	database.ConnectDB()
	defer database.DB.Close()

	api := app.Group("/api")
	handlers.Register(api, database.DB)

	log.Fatal(app.Listen(":5000"))
}
