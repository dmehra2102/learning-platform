package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmehra2102/learning-platform/course-service/internal/domain"
	"github.com/dmehra2102/learning-platform/shared/pkg/database"
)

type LessonRepository interface {
	Create(ctx context.Context, lesson *domain.Lesson) error
	GetByID(ctx context.Context, id string) (*domain.Lesson, error)
	GetByModuleID(ctx context.Context, moduleID string) ([]*domain.Lesson, error)
	Update(ctx context.Context, lesson *domain.Lesson) error
	Delete(ctx context.Context, id string) error
	GetMaxOrderIndex(ctx context.Context, moduleID string) (int, error)
}

type lessonRepository struct {
	db *database.DB
}

func NewLessonRepository(db *database.DB) LessonRepository {
	return &lessonRepository{db: db}
}

func (r *lessonRepository) Create(ctx context.Context, lesson *domain.Lesson) error {
	query := `
		INSERT INTO lessons (id, module_id, title,description, video_id, duration_seconds, order_index, is_preview, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		lesson.ID, lesson.ModuleID, lesson.Title, lesson.Description,
		lesson.VideoID, lesson.DurationSeconds, lesson.OrderIndex, lesson.IsPreview, lesson.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create lesson: %w", err)
	}

	return nil
}

func (r *lessonRepository) GetByID(ctx context.Context, id string) (*domain.Lesson, error) {
	query := `SELECT id, module_id, title, description, video_id, duration_seconds, 	 order_index, is_preview, created_at FROM lessons WHERE id = $1`

	var lesson domain.Lesson
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lesson.ID, &lesson.ModuleID, &lesson.Title, &lesson.Description,
		&lesson.VideoID, &lesson.DurationSeconds, &lesson.OrderIndex, &lesson.IsPreview, &lesson.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCourseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	return &lesson, nil
}

func (r *lessonRepository) GetByModuleID(ctx context.Context, moduleID string) ([]*domain.Lesson, error) {
	query := `
		SELECT id, module_id, title, description, video_id, duration_seconds, order_index, is_preview, created_at FROM lessons WHERE module_id = $1 ORDER BY order_index
	`

	rows, err := r.db.QueryContext(ctx, query, moduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list lessons: %w", err)
	}
	defer rows.Close()

	var lessons []*domain.Lesson
	for rows.Next() {
		var lesson domain.Lesson
		if err := rows.Scan(
			&lesson.ID, &lesson.ModuleID, &lesson.Title, &lesson.Description,
			&lesson.VideoID, &lesson.DurationSeconds, &lesson.OrderIndex, &lesson.IsPreview, &lesson.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lesson: %w", err)
		}

		lessons = append(lessons, &lesson)
	}

	return lessons, nil
}

func (r *lessonRepository) Update(ctx context.Context, lesson *domain.Lesson) error {
	query := `UPDATE lessons SET title = $1, description = $2, order_index = $3, is_preview = $4 WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query, lesson.Title, lesson.Description, lesson.OrderIndex, lesson.IsPreview, lesson.ID)
	if err != nil {
		return fmt.Errorf("failed to update lesson: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}

func (r *lessonRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM lessons WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lesson: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}

func (r *lessonRepository) GetMaxOrderIndex(ctx context.Context, moduleID string) (int, error) {
	query := `SELECT COALESCE(MAX(order_index), -1) FROM lessons WHERE module_id = $1`

	var maxIndex int
	err := r.db.QueryRowContext(ctx, query, moduleID).Scan(&maxIndex)
	if err != nil {
		return 0, fmt.Errorf("failed to get max order index: %w", err)
	}

	return maxIndex, nil
}
