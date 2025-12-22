package main

import (
	"log"
	"cuan-backend/database"
	"cuan-backend/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: No .env file found, relying on system env")
	}

	database.Connect()

	app := fiber.New()

	app.Use(cors.New())

	api := app.Group("/api")
	api.Get("/transactions", handlers.GetTransactions)
	api.Post("/webhook", handlers.WebhookReceiver)

	log.Fatal(app.Listen(":8080"))
}