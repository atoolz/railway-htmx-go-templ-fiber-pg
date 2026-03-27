package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/atoolz/railway-htmx-go-templ-fiber-pg/internal/models"
	"github.com/atoolz/railway-htmx-go-templ-fiber-pg/templates"
)

type Handler struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func render(c *fiber.Ctx, status int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)
	if err := t.Render(c.Context(), buf); err != nil {
		return err
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Status(status).SendString(buf.String())
}

func (h *Handler) Home(c *fiber.Ctx) error {
	todos, err := h.listTodos(c.Context())
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return render(c, http.StatusOK, templates.Home(todos))
}

func (h *Handler) CreateTodo(c *fiber.Ctx) error {
	title := c.FormValue("title")
	if title == "" {
		return fiber.NewError(http.StatusBadRequest, "title is required")
	}

	var todo models.Todo
	err := h.db.QueryRow(c.Context(),
		"INSERT INTO todos (title) VALUES ($1) RETURNING id, title, completed, created_at",
		title,
	).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return render(c, http.StatusCreated, templates.TodoItem(todo))
}

func (h *Handler) ToggleTodo(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}

	var todo models.Todo
	err = h.db.QueryRow(c.Context(),
		"UPDATE todos SET completed = NOT completed WHERE id = $1 RETURNING id, title, completed, created_at",
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return render(c, http.StatusOK, templates.TodoItem(todo))
}

func (h *Handler) DeleteTodo(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}

	_, err = h.db.Exec(c.Context(), "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(http.StatusOK)
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	if err := h.db.Ping(c.Context()); err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"status": "unhealthy", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "healthy"})
}

func (h *Handler) listTodos(ctx context.Context) ([]models.Todo, error) {
	rows, err := h.db.Query(ctx, "SELECT id, title, completed, created_at FROM todos ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, t)
	}
	return todos, nil
}
