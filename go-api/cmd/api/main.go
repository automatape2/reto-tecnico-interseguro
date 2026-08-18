// Command api starts the go-api HTTP server: it wires configuration,
// handlers, and middleware together and starts listening.
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"interseguro/go-api/internal/client"
	"interseguro/go-api/internal/config"
	"interseguro/go-api/internal/handlers"
	"interseguro/go-api/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	nodeClient := client.NewNodeClient(cfg.NodeAPIBaseURL, cfg.JWTSecret, cfg.NodeAPITimeout)
	qrHandler := handlers.NewQRHandler(nodeClient)
	authHandler := handlers.NewAuthHandler(cfg.GoAPIPresharedKey, cfg.JWTSecret, cfg.JWTExpiry)

	app := fiber.New(fiber.Config{
		AppName: "go-api",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOrigins,
	}))

	app.Get("/health", handlers.Health)

	v1 := app.Group("/api/v1")
	v1.Post("/auth/token", authHandler.IssueToken)
	v1.Post("/matrix/qr", middleware.JWTProtected(cfg.JWTSecret), qrHandler.ComputeQR)

	log.Printf("go-api listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
