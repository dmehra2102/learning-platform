package domain

import (
	"errors"
	"time"
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidInput   = errors.New("invalid input")
)

type CourseStatus string

const (
	StatusDraft     CourseStatus = "DRAFT"
	StatusPublished CourseStatus = "PUBLISHED"
	StatusArchived  CourseStatus = "ARCHIVED"
)

type CourseLevel string

const (
	LevelBeginner     CourseLevel = "BEGINNER"
	LevelIntermediate CourseLevel = "INTERMEDIATE"
	LevelAdvanced     CourseLevel = "ADVANCED"
)

type Course struct {
	ID              string
	Title           string
	Description     string
	InstructorID    string
	ThumbnailURL    string
	Status          CourseStatus
	Level           CourseLevel
	Price           float64
	Category        string
	Tags            []string
	DurationMinutes int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	EnrolledCount   int
	AverageRating   float64
}

type Module struct {
	ID          string
	CourseID    string
	Title       string
	Description string
	OrderIndex  int
	CreatedAt   time.Time
}

type Lesson struct {
	ID              string
	ModuleID        string
	Title           string
	Description     string
	VideoID         string
	DurationSeconds int
	OrderIndex      int
	IsPreview       bool
	CreatedAt       time.Time
}

func (c *Course) Validate() error {
	if c.Title == "" || len(c.Title) > 255 {
		return ErrInvalidInput
	}
	if c.Description == "" {
		return ErrInvalidInput
	}
	if c.Price < 0 {
		return ErrInvalidInput
	}
	if c.Category == "" {
		return ErrInvalidInput
	}

	return nil
}

func (m *Module) Validate() error {
	if m.Title == "" || len(m.Title) > 255 {
		return ErrInvalidInput
	}
	if m.OrderIndex < 0 {
		return ErrInvalidInput
	}
	return nil
}

func (l *Lesson) Validate() error {
	if l.Title == "" || len(l.Title) > 255 {
		return ErrInvalidInput
	}
	if l.OrderIndex < 0 {
		return ErrInvalidInput
	}
	if l.DurationSeconds < 0 {
		return ErrInvalidInput
	}
	return nil
}
