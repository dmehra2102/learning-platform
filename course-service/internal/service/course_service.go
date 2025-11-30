package service

import (
	"context"

	"github.com/dmehra2102/learning-platform/course-service/internal/domain"
	"github.com/dmehra2102/learning-platform/course-service/internal/repository"
	"github.com/dmehra2102/learning-platform/shared/pkg/kafka"
	pb "github.com/dmehra2102/learning-platform/shared/proto/course"
	"go.uber.org/zap"
)

type CourseService interface {
	CreateCourse(ctx context.Context, instructorID string, req pb.CreateCourseRequest) (*domain.Course, error)
	PublishCourse(ctx context.Context, courseID, instructorID string) (*domain.Course, error)
	GetCourse(ctx context.Context, courseID string) (*domain.Course, error)
	UpdateCourse(ctx context.Context, courseID, instructorID string, req pb.UpdateCourseRequest) (*domain.Course, error)
	DeleteCourse(ctx context.Context, courseID, instructorID string) error
	ListCourses(ctx context.Context, filter domain.CourseFilter) ([]*domain.Course, int, error)
	AddModule(ctx context.Context, courseID string, req pb.AddModuleRequest) (*domain.Module, error)
	AddLesson(ctx context.Context, moduleID string, req pb.AddLessonRequest) (*domain.Lesson, error)
}

type userService struct {
	repo          repository.CourseRepository
	kafkaProducer *kafka.Producer
	logger        *zap.Logger
}
