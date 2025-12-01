package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmehra2102/learning-platform/course-service/internal/domain"
	"github.com/dmehra2102/learning-platform/shared/pkg/database"
	"github.com/lib/pq"
)

type CourseRepository interface {
	Create(ctx context.Context, course *domain.Course) error
	GetByID(ctx context.Context, id string) (*domain.Course, error)
	Update(ctx context.Context, course *domain.Course) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int, category *string, status *domain.CourseStatus, search *string, level *domain.CourseLevel) ([]*domain.Course, int, error)
	GetByInstructor(ctx context.Context, instructorID string, page, pageSize int) ([]*domain.Course, int, error)
	UpdateEnrolledCount(ctx context.Context, courseID string, increment int) error
	UpdateAverageRating(ctx context.Context, courseID string, rating float64) error
}

type courseRepository struct {
	db *database.DB
}

func NewCourseRepository(db *database.DB) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) Create(ctx context.Context, course *domain.Course) error {
	query := `
		INSERT INTO courses (id, title, description, instructor_id, thumbnail_url, status, level, price, category, tags, duration_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		course.ID, course.Title, course.Description, course.InstructorID,
		course.ThumbnailURL, course.Status, course.Level, course.Price,
		course.Category, pq.Array(course.Tags), course.DurationMinutes,
		course.CreatedAt, course.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create course: %w", err)
	}

	return nil
}

func (r *courseRepository) GetByID(ctx context.Context, id string) (*domain.Course, error) {
	query := `
		SELECT id, title, description, instructor_id, thumbnail_url, status, level, price, category, tags, duration_minutes, created_at, updated_at, enrolled_count, average_rating FROM courses WHERE id = $1
	`

	var course domain.Course
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		course.ID, &course.Title, &course.Description, &course.InstructorID,
		&course.ThumbnailURL, &course.Status, &course.Level, &course.Price,
		&course.Category, pq.Array(&course.Tags), &course.DurationMinutes,
		&course.CreatedAt, &course.UpdatedAt, &course.EnrolledCount, &course.AverageRating,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCourseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	return &course, nil
}

func (r *courseRepository) Update(ctx context.Context, course *domain.Course) error {
	query := `
		UPDATE courses
		SET title = $1, description = $2, thumbnail_url = $3, status = $4, level = $5, price = $6, category = $7, tags = $8, updated_at = $9
		WHERE id = $10
	`

	result, err := r.db.ExecContext(ctx, query,
		course.Title, course.Description, course.ThumbnailURL, course.Status,
		course.Level, course.Price, course.Category, pq.Array(course.Tags),
		course.UpdatedAt, course.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update course: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}

func (r *courseRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM courses WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}

func (r *courseRepository) List(ctx context.Context, page, pageSize int, category *string, status *domain.CourseStatus, search *string, level *domain.CourseLevel) ([]*domain.Course, int, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT id,title,description, instructor_id,thumbnail_url, status, level, price, category, tags, duration_minutes, created_at, updated_at, enrolled_count, average_rating FROM courses WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) fROM courses where 1-1`
	args := []any{}
	argCount := 1

	if category != nil {
		query += fmt.Sprintf(" AND category = %d", argCount)
		countQuery += fmt.Sprintf(" AND category = %d", argCount)
		args = append(args, *category)
		argCount++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = %d", argCount)
		countQuery += fmt.Sprintf(" AND status = %d", argCount)
		args = append(args, *status)
		argCount++
	}
	if level != nil {
		query += fmt.Sprintf(" AND level = $%d", argCount)
		countQuery += fmt.Sprintf(" AND level = $%d", argCount)
		args = append(args, *level)
		argCount++
	}
	if search != nil {
		query += fmt.Sprintf("AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf("AND (title ILIKE $%d OR description ILIKR $%d)", argCount, argCount)
		searchItem := "%" + *search + "%"
		args = append(args, searchItem)
		argCount++
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %w", err)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list courses: %w", err)
	}
	defer rows.Close()

	var courses []*domain.Course
	for rows.Next() {
		var course domain.Course
		if err := rows.Scan(
			&course.ID, &course.Title, &course.Description, &course.InstructorID,
			&course.ThumbnailURL, &course.Status, &course.Level, &course.Price,
			&course.Category, pq.Array(&course.Tags), &course.DurationMinutes,
			&course.CreatedAt, &course.UpdatedAt, &course.EnrolledCount, &course.AverageRating,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan course: %w", err)
		}
		courses = append(courses, &course)
	}

	return courses, total, nil
}

func (r *courseRepository) GetByInstructor(ctx context.Context, instructorID string, page, pageSize int) ([]*domain.Course, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM courses WHERE instructor_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, instructorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count courses: %w", err)
	}

	query := `
		SELECT id, title, description, instructor_id, thumbnail_url, status, level, price,
		    category, tags, duration_minutes, created_at, updated_at, enrolled_count, average_rating
		FROM courses WHERE instructor_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, instructorID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list courses: %w", err)
	}
	defer rows.Close()

	var courses []*domain.Course
	for rows.Next() {
		var course domain.Course
		if err := rows.Scan(
			&course.ID, &course.Title, &course.Description, &course.InstructorID,
			&course.ThumbnailURL, &course.Status, &course.Level, &course.Price,
			&course.Category, pq.Array(&course.Tags), &course.DurationMinutes,
			&course.CreatedAt, &course.UpdatedAt, &course.EnrolledCount, &course.AverageRating,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan course: %w", err)
		}
		courses = append(courses, &course)
	}

	return courses, total, nil
}

func (r *courseRepository) UpdateEnrolledCount(ctx context.Context, courseID string, increment int) error {
	query := `
		UPDATE courses
		SET enrolled_count = enrolled_count + $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, increment, courseID)
	if err != nil {
		return fmt.Errorf("failed to update enrolled count: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}

func (r *courseRepository) UpdateAverageRating(ctx context.Context, courseID string, rating float64) error {
	query := `
		UPDATE courses
		SET average_rating = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, rating, courseID)
	if err != nil {
		return fmt.Errorf("failed to update average rating: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCourseNotFound
	}

	return nil
}
