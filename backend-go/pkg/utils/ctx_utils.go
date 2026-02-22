package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// GetUserIDFromContext safely extracts the userID from fiber.Ctx Locals
// and returns it as a uint, avoiding panics from incorrect type assertions.
func GetUserIDFromContext(c *fiber.Ctx) (uint, error) {
	val := c.Locals("userID")
	if val == nil {
		return 0, fmt.Errorf("userID not found in context")
	}

	switch v := val.(type) {
	case uint:
		return v, nil
	case float64:
		return uint(v), nil
	case int:
		return uint(v), nil
	default:
		return 0, fmt.Errorf("unexpected type for userID: %T", val)
	}
}
