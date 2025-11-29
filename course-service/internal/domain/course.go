package domain

import (
	"errors"
	"time"
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrUnauthorized   = errors.New("unauthorized")
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
