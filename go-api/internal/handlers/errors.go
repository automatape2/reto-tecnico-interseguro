package handlers

import (
	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/dto"
)

// writeError writes the shared { "error": { "code", "message" } } envelope
// used by every non-2xx response this API returns.
func writeError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(dto.ErrorResponse{
		Error: dto.ErrorBody{Code: code, Message: message},
	})
}
