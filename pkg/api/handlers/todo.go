package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm"
	"strconv"
	"we-know/pkg/domain/models"
	"we-know/pkg/domain/repositories"
)

type TodoHandler struct {
	repository *repositories.TodoRepository
}

func (handler *TodoHandler) GetAll(c *fiber.Ctx) error {
	var todos []models.Todo = handler.repository.FindAll()
	return c.JSON(todos)
}

func (handler *TodoHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	todo, err := handler.repository.Find(id)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status": 404,
			"error":  err,
		})
	}

	return c.JSON(todo)
}

func (handler *TodoHandler) Create(c *fiber.Ctx) error {
	data := new(models.Todo)

	if err := c.BodyParser(data); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "error": err})
	}

	item, err := handler.repository.Create(*data)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "Failed creating item",
			"error":   err,
		})
	}

	return c.JSON(item)
}

func (handler *TodoHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "Item not found",
			"error":   err,
		})
	}

	todo, err := handler.repository.Find(id)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"message": "Item not found",
		})
	}

	todoData := new(models.Todo)

	if err := c.BodyParser(todoData); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}

	todo.Name = todoData.Name
	todo.Description = todoData.Description
	todo.Status = todoData.Status

	item, err := handler.repository.Save(todo)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "Error updating todo",
			"error":   err,
		})
	}

	return c.JSON(item)
}

func (handler *TodoHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "Failed deleting todo",
			"err":     err,
		})
	}
	RowsAffected := handler.repository.Delete(id)
	statusCode := 204
	if RowsAffected == 0 {
		statusCode = 400
	}
	return c.Status(statusCode).JSON(nil)
}

func NewTodoHandler(repository *repositories.TodoRepository) *TodoHandler {
	return &TodoHandler{
		repository: repository,
	}
}

func Register(router fiber.Router, database *gorm.DB) {
	database.AutoMigrate(&models.Todo{})
	todoRepository := repositories.NewTodoRepository(database)
	todoHandler := NewTodoHandler(todoRepository)

	movieRouter := router.Group("/todo")
	movieRouter.Get("/", todoHandler.GetAll)
	movieRouter.Get("/:id", todoHandler.Get)
	movieRouter.Put("/:id", todoHandler.Update)
	movieRouter.Post("/", todoHandler.Create)
	movieRouter.Delete("/:id", todoHandler.Delete)
}
