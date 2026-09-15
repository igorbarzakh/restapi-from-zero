package database

import (
	"database/sql"
	"errors"
	"fmt"
	"restapi-tasks/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskStore struct {
	db *sqlx.DB
}

func NewTaskStore(db *sqlx.DB) *TaskStore {
	return &TaskStore{db: db}
}

func (s *TaskStore) GetAll() ([]models.Task, error) {
	var tasks []models.Task

	query :=
		`
		SELECT id, title, description, completed, created_at, updated_at
		FROM tasks 
		order by created_at desc;`

	err := s.db.Select(&tasks, query)

	if err != nil {
		return nil, fmt.Errorf("get all tasks: %w", err)
	}

	return tasks, nil
}

func (s *TaskStore) GetByID(id int) (*models.Task, error) {
	var task models.Task

	query :=
		`
		SELECT id, title, description, completed, created_at, updated_at
		FROM tasks 
		where id = $1;`

	err := s.db.Get(&task, query, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: id=%d", ErrTaskNotFound, id)
	}

	if err != nil {
		return nil, fmt.Errorf("get task by id %d: %w", id, err)
	}

	return &task, nil
}

func (s *TaskStore) Create(input models.CreateTaskInput) (*models.Task, error) {
	var task models.Task

	query :=
		`
		INSERT INTO tasks (title, description, completed, created_at, updated_at)
		values ($1, $2, $3, $4, $5)
		returning id, title, description, completed, created_at, updated_at;
		`

	now := time.Now()

	err := s.db.QueryRowx(query, input.Title, input.Description, input.Completed, now, now).StructScan(&task)

	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return &task, nil
}

func (s *TaskStore) Update(id int, input models.UpdateTaskInput) (*models.Task, error) {
	task, err := s.GetByID(id)

	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		task.Title = *input.Title
	}

	if input.Description != nil {
		task.Description = *input.Description
	}

	if input.Completed != nil {
		task.Completed = *input.Completed
	}

	task.UpdatedAt = time.Now()

	query :=
		`
		UPDATE tasks
		SET title = $1, description = $2, completed = $3, updated_at = $4
		WHERE id = $5
		returning id, title, description, completed, created_at, updated_at;
		`

	var updatedTask models.Task

	err = s.db.QueryRowx(
		query, task.Title,
		task.Description,
		task.Completed,
		task.UpdatedAt,
		task.ID,
	).StructScan(&updatedTask)

	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	return &updatedTask, nil
}

func (s *TaskStore) Delete(id int) error {
	query := `DELETE FROM tasks WHERE id = $1;`

	result, err := s.db.Exec(query, id)

	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: id=%d", ErrTaskNotFound, id)
	}

	return nil
}
