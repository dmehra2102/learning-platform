package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmehra2102/learning-platform/enrollment-service/internal/domain"
	"github.com/dmehra2102/learning-platform/shared/pkg/database"
)

type EnrollmentRepository interface {
	Create(ctx context.Context, enrollment *domain.Enrollment) error
	GetByID(ctx context.Context, id string) (*domain.Enrollment, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID string) (*domain.Enrollment, error)
	Update(ctx context.Context, enrollment *domain.Enrollment) error
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*domain.Enrollment, int, error)
	ListByCourse(ctx context.Context, courseID string, page, pageSize int) ([]*domain.Enrollment, int, error)
	ListByStatus(ctx context.Context, status domain.EnrollmentStatus, page, pageSize int) ([]*domain.Enrollment, int, error)
	CountByUser(ctx context.Context, userID string) (int, error)
	CountByCourse(ctx context.Context, courseID string) (int, error)
}

type enrollmentRepository struct {
	db *database.DB
}

func NewEnrollmentRepository(db *database.DB) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) Create(ctx context.Context, enrollment *domain.Enrollment) error {
	query := `
		INSERT INTO enrollments (id, user_id, course_id, status, amount_paid, payment_id, enrolled_at, progress_percentage) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`

	_, err := r.db.ExecContext(ctx, query,
		enrollment.ID, enrollment.UserID, enrollment.CourseID, enrollment.Status,
		enrollment.AmountPaid, enrollment.PaymentID, enrollment.EnrolledAt, enrollment.ProgressPercentage,
	)

	if err != nil {
		return fmt.Errorf("failed to create enrollment: %w", err)
	}

	return nil
}

func (r *enrollmentRepository) GetByID(ctx context.Context, id string) (*domain.Enrollment, error) {
	query := `
		SELECT id, user_id, course_id, status, amount_paid, payment_id, enrolled_at, completed_at, progress_percentage FROM enrollments WHERE id = $1
	`

	var enrollment domain.Enrollment
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&enrollment.ID, &enrollment.UserID, &enrollment.CourseID, &enrollment.Status,
		&enrollment.AmountPaid, &enrollment.PaymentID, &enrollment.EnrolledAt,
		&completedAt, &enrollment.ProgressPercentage,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrEnrollmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get enrollment: %w", err)
	}

	if completedAt.Valid {
		enrollment.CompletedAt = &completedAt.Time
	}

	return &enrollment, nil
}

func (r *enrollmentRepository) GetByUserAndCourse(ctx context.Context, userID, courseID string) (*domain.Enrollment, error) {
	query := `
		SELECT id, user_id, course_id, status, amount_paid, payment_id, enrolled_at, completed_at, progress_percentage FROM enrollments WHERE user_id = $1 AND course_id = $2
	`

	var enrollment domain.Enrollment
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, courseID).Scan(
		&enrollment.ID, &enrollment.UserID, &enrollment.CourseID, &enrollment.Status,
		&enrollment.AmountPaid, &enrollment.PaymentID, &enrollment.EnrolledAt,
		&completedAt, &enrollment.ProgressPercentage,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrEnrollmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get enrollment: %w", err)
	}

	if completedAt.Valid {
		enrollment.CompletedAt = &completedAt.Time
	}

	return &enrollment, nil
}

func (r *enrollmentRepository) Update(ctx context.Context, enrollment *domain.Enrollment) error {
	query := `
		UPDATE enrollments
		SET status = $1, payment_id = $2, completed_at = $3, progress_percentage = $4, amount_paid = $5 WHERE id = $6
	`

	var completedAt any
	if enrollment.CompletedAt != nil {
		completedAt = *enrollment.CompletedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		enrollment.Status, enrollment.PaymentID, completedAt,
		enrollment.ProgressPercentage, enrollment.AmountPaid, enrollment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update enrollment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrEnrollmentNotFound
	}

	return nil
}

func (r *enrollmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM enrollments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete enrollment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrEnrollmentNotFound
	}

	return nil
}

func (r *enrollmentRepository) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*domain.Enrollment, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM enrollments WHERE user_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count enrollments: %w", err)
	}

	query := `
		SELECT id, user_id, course_id, status, amount_paid, payment_id, enrolled_at, completed_at, progress_percentage FROM enrollments WHERE user_id = $1 ORDER BY enrolled_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*domain.Enrollment
	for rows.Next() {
		var enrollment domain.Enrollment
		var completedAt sql.NullTime

		if err := rows.Scan(
			&enrollment.ID, &enrollment.UserID, &enrollment.CourseID, &enrollment.Status,
			&enrollment.AmountPaid, &enrollment.PaymentID, &enrollment.EnrolledAt,
			&completedAt, &enrollment.ProgressPercentage,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan enrollment: %w", err)
		}

		if completedAt.Valid {
			enrollment.CompletedAt = &completedAt.Time
		}

		enrollments = append(enrollments, &enrollment)
	}

	return enrollments, total, nil
}

func (r *enrollmentRepository) ListByCourse(ctx context.Context, courseID string, page, pageSize int) ([]*domain.Enrollment, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, courseID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count enrollments: %w", err)
	}

	query := `
		SELECT id, user_id, course_id, status, amount_paid, payment_id, enrolled_at, completed_at, progress_percentage
		FROM enrollments WHERE course_id = $1
		ORDER BY enrolled_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, courseID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*domain.Enrollment
	for rows.Next() {
		var enrollment domain.Enrollment
		var completedAt sql.NullTime

		if err := rows.Scan(
			&enrollment.ID, &enrollment.UserID, &enrollment.CourseID, &enrollment.Status,
			&enrollment.AmountPaid, &enrollment.PaymentID, &enrollment.EnrolledAt,
			&completedAt, &enrollment.ProgressPercentage,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan enrollment: %w", err)
		}

		if completedAt.Valid {
			enrollment.CompletedAt = &completedAt.Time
		}

		enrollments = append(enrollments, &enrollment)
	}

	return enrollments, total, nil
}

func (r *enrollmentRepository) ListByStatus(ctx context.Context, status domain.EnrollmentStatus, page, pageSize int) ([]*domain.Enrollment, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM enrollments WHERE status = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count enrollments: %w", err)
	}

	query := `
		SELECT id, user_id, course_id, status, amount_paid, payment_id, enrolled_at, completed_at, progress_percentage
		FROM enrollments WHERE status = $1
		ORDER BY enrolled_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*domain.Enrollment
	for rows.Next() {
		var enrollment domain.Enrollment
		var completedAt sql.NullTime

		if err := rows.Scan(
			&enrollment.ID, &enrollment.UserID, &enrollment.CourseID, &enrollment.Status,
			&enrollment.AmountPaid, &enrollment.PaymentID, &enrollment.EnrolledAt,
			&completedAt, &enrollment.ProgressPercentage,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan enrollment: %w", err)
		}

		if completedAt.Valid {
			enrollment.CompletedAt = &completedAt.Time
		}

		enrollments = append(enrollments, &enrollment)
	}

	return enrollments, total, nil
}

func (r *enrollmentRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND status = $2`
	err := r.db.QueryRowContext(ctx, query, userID, domain.StatusActive).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count enrollments: %w", err)
	}

	return count, nil
}

func (r *enrollmentRepository) CountByCourse(ctx context.Context, courseID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1 AND status = $2`
	err := r.db.QueryRowContext(ctx, query, courseID, domain.StatusActive).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count enrollments: %w", err)
	}

	return count, nil
}
