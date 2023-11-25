package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type User struct {
	ID        string
	FirstName string
	LastName  string
	Auth0UID  string
}

type IUserRepository interface {
	FindByAuth0UID(auth0UID string) (*User, error)
}

type UserRepository struct {
}

func (repo *UserRepository) FindByAuth0UID(auth0UID string) (*User, error) {
	return &User{ID: "123", FirstName: "Sample", LastName: "User", Auth0UID: auth0UID}, nil
}

type UserUseCase struct {
	UserRepo IUserRepository
}

func main() {
	container := setupDIContainer()
	app := fiber.New()

	if err := setupRoutes(app, container); err != nil {
		log.Fatalf("Route setup failed: %v", err)
	}

	go handleSignals(app)

	if err := startServer(app); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func userHandler(uc *UserUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := uc.UserRepo.FindByAuth0UID("auth0-uid-123")
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}
		return c.JSON(user)
	}
}

func setupDIContainer() *dig.Container {
	container := dig.New()
	container.Provide(func() IUserRepository {
		return &UserRepository{}
	})
	container.Provide(func(repo IUserRepository) *UserUseCase {
		return &UserUseCase{UserRepo: repo}
	})
	return container
}

func setupRoutes(app *fiber.App, container *dig.Container) error {
	return container.Invoke(func(uc *UserUseCase) {
		app.Get("/me", userHandler(uc))
	})
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
