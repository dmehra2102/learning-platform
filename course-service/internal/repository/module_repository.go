package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmehra2102/learning-platform/course-service/internal/domain"
	"github.com/dmehra2102/learning-platform/shared/pkg/database"
)

type ModuleRepository interface {
	Create(ctx context.Context, module *domain.Module) error
	GetByID(ctx context.Context, id string) (*domain.Module, error)
	GetByCourseID(ctx context.Context, courseID string) ([]*domain.Module, error)
	Update(ctx context.Context, module *domain.Module) error
	Delete(ctx context.Context, id string) error
	GetMaxOrderIndex(ctx context.Context, courseID string) (int, error)
}

type moduleRepository struct {
	db *database.DB
}

func NewModuleRepository(db *database.DB) ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) Create(ctx context.Context, module *domain.Module) error {
	query := `
		INSERT INTO modules (id, course_id, title, description, order_index, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		module.ID, module.CourseID, module.Title, module.Description, module.OrderIndex, module.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}

	return nil
}

func (r *moduleRepository) GetByID(ctx context.Context, id string) (*domain.Module, error) {
	query := `SELECT id, course_id, title, description, order_index, created_at
		FROM modules WHERE id = $1`

	var module domain.Module
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&module.ID, &module.CourseID, &module.Title, &module.Description, &module.OrderIndex, &module.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCourseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}

	return &module, nil
}

func (r *moduleRepository) GetByCourseID(ctx context.Context, courseID string) ([]*domain.Module, error) {
	query := `
		SELECT id, course_id, title, description, order_index, created_at 
		FROM modules WHERE course_id = $1 ORDER BY order_index
	`

	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}
	defer rows.Close()

	var modules []*domain.Module
	for rows.Next() {
		var module domain.Module
		if err := rows.Scan(
			&module.ID,
			&module.CourseID,
			&module.Title,
			&module.Description,
			&module.OrderIndex,
			&module.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}

		modules = append(modules, &module)
	}

	return modules, nil
}

func (r *moduleRepository) Update(ctx context.Context, module *domain.Module) error {
	qyery := `UPDATE modules SET title = $1, description = $2, order_index = $3 WHERE id = $4`

	result, err := r.db.ExecContext(ctx, qyery, module.Title, module.Description, module.OrderIndex, module.ID)
	if err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}

func (r *moduleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM modules WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}

func (r *moduleRepository) GetMaxOrderIndex(ctx context.Context, courseID string) (int, error) {
	query := `SELECT COALESCE(MAX(order_index), -1) FROM modules WHERE course_id = $1`

	var maxIndex int
	err := r.db.QueryRowContext(ctx, query, courseID).Scan(&maxIndex)
	if err != nil {
		return 0, fmt.Errorf("failed to get max order index: %w", err)
	}

	return maxIndex, nil
}
