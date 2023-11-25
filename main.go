package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/dig"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

func handleError(c *fiber.Ctx, err error) error {
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	return nil
}

func userHandler(uc *UserUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := uc.GetUserByAuth0UID("auth0-uid-123")
		if handleError(c, err) != nil {
			return err
		}
		return c.JSON(user)
	}
}

func (uc *UserUseCase) GetUserByAuth0UID(auth0UID string) (*User, error) {
	return uc.UserRepo.FindByAuth0UID(auth0UID)
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

func validateToken(tokenString string, audience string, issuer string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your-256-bit-secret"), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["aud"] != audience || claims["iss"] != issuer {
			return nil, fmt.Errorf("invalid token claims")
		}
		return token, nil
	} else {
		return nil, err
	}
}

func jwtMiddleware(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) == 2 {
			token, err := validateToken(bearerToken[1], "YOUR_AUTH0_AUDIENCE", "https://YOUR_AUTH0_DOMAIN/")
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
			}

			// トークンが有効な場合の処理
			fmt.Println(token)
			return next(c)
		} else {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}
	}
}

func setupRoutes(app *fiber.App, container *dig.Container) error {
	return container.Invoke(func(uc *UserUseCase) {
		//app.Get("/me", jwtMiddleware(userHandler(uc)))
		app.Get("/me", userHandler(uc))
	})
}

func startServer(app *fiber.App) error {
	if err := app.Listen(":8080"); err != nil && err != http.ErrServerClosed {
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
