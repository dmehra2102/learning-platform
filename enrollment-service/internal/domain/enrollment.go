package domain

import (
	"errors"
	"time"
)

var (
	ErrEnrollmentNotFound      = errors.New("enrollment not found")
	ErrAlreadyEnrolled         = errors.New("already enrolled")
	ErrEnrollmentCancelled     = errors.New("enrollment cancelled")
	ErrInvalidEnrollmentStatus = errors.New("invalid enrollment status")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrInvalidInput            = errors.New("invalid input")
)

type EnrollmentStatus string

const (
	StatusPending   EnrollmentStatus = "PENDING"
	StatusActive    EnrollmentStatus = "ACTIVE"
	StatusCompleted EnrollmentStatus = "COMPLETED"
	StatusCancelled EnrollmentStatus = "CANCELLED"
	StatusRefunded  EnrollmentStatus = "REFUNDED"
)

type Enrollment struct {
	ID                 string
	UserID             string
	CourseID           string
	Status             EnrollmentStatus
	AmountPaid         float64
	PaymentID          string
	EnrolledAt         time.Time
	CompletedAt        time.Time
	ProgressPercentage int
}

type EnrollmentEvent struct {
	EnnrollmentID string
	UserID        string
	CourseID      string
	Status        EnrollmentStatus
	Amount        float64
	Timestamp     time.Time
}

func (e *Enrollment) Validate() error {
	if e.UserID == "" {
		return ErrInvalidInput
	}
	if e.CourseID == "" {
		return ErrInvalidInput
	}
	if e.AmountPaid < 0 {
		return ErrInvalidInput
	}
	if e.ProgressPercentage < 0 || e.ProgressPercentage > 100 {
		return ErrInvalidInput
	}
	return nil
}

func (e *Enrollment) IsActive() bool {
	return e.Status == StatusActive
}

func (e *Enrollment) CanBeCancelled() bool {
	return e.Status == StatusActive || e.Status == StatusPending
}

func IsValidStatus(status string) bool {
	switch EnrollmentStatus(status) {
	case StatusPending, StatusActive, StatusCompleted, StatusCancelled, StatusRefunded:
		return true
	default:
		return false
	}
}
