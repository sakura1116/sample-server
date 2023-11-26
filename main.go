package main

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/dig"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sample-server/ent"
	"sample-server/ent/user"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func NewConfig() *Config {
	return &Config{
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),
		DBName: os.Getenv("DB_NAME"),
	}
}

type User struct {
	ID        int
	FirstName string
	LastName  string
	Auth0UID  string
}

type IUserRepository interface {
	FindByAuth0UID(auth0UID string) (*User, error)
}

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (repo *UserRepository) FindByAuth0UID(auth0UID string) (*User, error) {
	u, err := repo.client.User.
		Query().
		Where(user.Auth0UIDEQ(auth0UID)).
		Only(context.Background())
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Auth0UID:  u.Auth0UID,
	}, nil
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
		u, err := uc.GetUserByAuth0UID("abc123")
		if handleError(c, err) != nil {
			return err
		}
		return c.JSON(u)
	}
}

func (uc *UserUseCase) GetUserByAuth0UID(auth0UID string) (*User, error) {
	return uc.UserRepo.FindByAuth0UID(auth0UID)
}

func main() {
	config := NewConfig()
	client := setupDatabase(config)
	defer client.Close()
	container := setupDIContainer(client)

	app := fiber.New()

	if err := setupRoutes(app, container); err != nil {
		log.Fatalf("Route setup failed: %v", err)
	}

	go handleSignals(app)

	if err := startServer(app); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func setupDatabase(config *Config) *ent.Client {
	dsn := config.DBUser + ":" + config.DBPass + "@tcp(" + config.DBHost + ":" + config.DBPort + ")/" + config.DBName + "?parseTime=True"
	client, err := ent.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	} else {
	}
	return client
}
func setupDIContainer(client *ent.Client) *dig.Container {
	container := dig.New()
	container.Provide(func() IUserRepository {
		return NewUserRepository(client)
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

			// TODO トークンが有効な場合の処理
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
