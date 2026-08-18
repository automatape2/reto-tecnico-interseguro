package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/dto"
)

// Health reports liveness for container/orchestrator health checks. It is
// intentionally unauthenticated.
func Health(c *fiber.Ctx) error {
	return c.JSON(dto.HealthResponse{
		Status:  "ok",
		Service: "go-api",
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}
