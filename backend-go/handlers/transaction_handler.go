package handlers

import (
	"cuan-backend/models"
	"cuan-backend/database"
	"github.com/gofiber/fiber/v2"
)

func GetTransactions(c *fiber.Ctx) error {
	var transactions []models.Transaction
	
	database.DB.Order("created_at desc").Find(&transactions)

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   transactions,
	})
}

func WebhookReceiver(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "message": "Webhook receiver is ready!",
    })
}