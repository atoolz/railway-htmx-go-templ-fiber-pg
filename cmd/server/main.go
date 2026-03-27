package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/atoolz/railway-htmx-go-templ-fiber-pg/internal/database"
	"github.com/atoolz/railway-htmx-go-templ-fiber-pg/internal/handlers"
)

func main() {
	ctx := context.Background()

	pool, err := database.Connect(ctx)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}

	h := handlers.New(pool)

	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", h.Home)
	app.Post("/todos", h.CreateTodo)
	app.Patch("/todos/:id/toggle", h.ToggleTodo)
	app.Delete("/todos/:id", h.DeleteTodo)
	app.Get("/health", h.HealthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
