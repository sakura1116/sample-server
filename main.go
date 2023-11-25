package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	app := fiber.New()

	if err := setupRoutes(app); err != nil {
		log.Fatalf("Route setup failed: %v", err)
	}

	go handleSignals(app)

	if err := startServer(app); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func setupRoutes(app *fiber.App) error {
	return nil
}

func startServer(app *fiber.App) error {
	if err := app.Listen(":3000"); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func handleSignals(app *fiber.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Gracefully shutting down...")
	if err := shutdownServer(app); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}

func shutdownServer(app *fiber.App) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.ShutdownWithContext(ctx)
}
