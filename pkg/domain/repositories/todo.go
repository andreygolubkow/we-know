package repositories

import (
	"errors"
	"github.com/jinzhu/gorm"
	"we-know/pkg/domain/models"
)

type TodoRepository struct {
	database *gorm.DB
}

func (repository *TodoRepository) FindAll() []models.Todo {
	var todos []models.Todo
	repository.database.Find(&todos)
	return todos
}

func (repository *TodoRepository) Find(id int) (models.Todo, error) {
	var todo models.Todo
	err := repository.database.Find(&todo, id).Error
	if todo.Name == "" {
		err = errors.New("Todo not found")
	}
	return todo, err
}

func (repository *TodoRepository) Create(todo models.Todo) (models.Todo, error) {
	err := repository.database.Create(&todo).Error
	if err != nil {
		return todo, err
	}

	return todo, nil
}

func (repository *TodoRepository) Save(user models.Todo) (models.Todo, error) {
	err := repository.database.Save(user).Error
	return user, err
}

func (repository *TodoRepository) Delete(id int) int64 {
	count := repository.database.Delete(&models.Todo{}, id).RowsAffected
	return count
}

func NewTodoRepository(database *gorm.DB) *TodoRepository {
	return &TodoRepository{
		database: database,
	}
}
