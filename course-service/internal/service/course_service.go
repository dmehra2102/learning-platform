package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dmehra2102/learning-platform/course-service/internal/domain"
	"github.com/dmehra2102/learning-platform/course-service/internal/repository"
	"github.com/dmehra2102/learning-platform/shared/pkg/kafka"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreateCourseRequest struct {
	Title        string
	Description  string
	ThumbnailURL string
	Level        domain.CourseLevel
	Price        float64
	Category     string
	Tags         []string
}

type UpdateCourseRequest struct {
	Title        *string
	Description  *string
	ThumbnailURL *string
	Level        *domain.CourseLevel
	Price        *float64
	Category     *string
	Tags         []string
}

type CourseFilter struct {
	Page     int
	PageSize int
	Category *string
	Status   *domain.CourseStatus
	Level    *domain.CourseLevel
	Search   *string
}

type AddModuleRequest struct {
	Title       string
	Description string
}

type AddLessonRequest struct {
	Title           string
	Description     string
	VideoID         string
	DurationSeconds int
	IsPreview       bool
}

type UpdateLessonRequest struct {
	Title           *string
	Description     *string
	IsPreview       *bool
}

type CourseService interface {
	CreateCourse(ctx context.Context, instructorID string, req CreateCourseRequest) (*domain.Course, error)
	PublishCourse(ctx context.Context, courseID, instructorID string) (*domain.Course, error)
	GetCourse(ctx context.Context, courseID string) (*domain.Course, error)
	UpdateCourse(ctx context.Context, courseID, instructorID string, req UpdateCourseRequest) (*domain.Course, error)
	DeleteCourse(ctx context.Context, courseID, instructorID string) error
	ListCourses(ctx context.Context, filter CourseFilter) ([]*domain.Course, int, error)
	GetInstructorCourses(ctx context.Context, instructorID string, page, pageSize int) ([]*domain.Course, int, error)
	AddModule(ctx context.Context, courseID, instructorID string, req AddModuleRequest) (*domain.Module, error)
	UpdateModule(ctx context.Context, moduleID, courseID, instructorID string, title, description string) (*domain.Module, error)
	DeleteModule(ctx context.Context, moduleID, courseID, instructorID string) error
	GetModules(ctx context.Context, courseID string) ([]*domain.Module, error)
	AddLesson(ctx context.Context, moduleID, courseID, instructorID string, req AddLessonRequest) (*domain.Lesson, error)
	UpdateLesson(ctx context.Context, lessonID, moduleID, courseID, instructorID string, req UpdateLessonRequest) (*domain.Lesson, error)
	DeleteLesson(ctx context.Context, lessonID, moduleID, courseID, instructorID string) error
	GetLessons(ctx context.Context, moduleID string) ([]*domain.Lesson, error)
}

type courseService struct {
	courseRepo    repository.CourseRepository
	moduleRepo    repository.ModuleRepository
	lessonRepo    repository.LessonRepository
	kafkaProducer *kafka.Producer
	logger        *zap.Logger
}

func NewCourseService(
	courseRepo repository.CourseRepository,
	moduleRepo repository.ModuleRepository,
	lessonRepo repository.LessonRepository,
	producer *kafka.Producer,
	logger *zap.Logger,
) CourseService {
	return &courseService{
		courseRepo:    courseRepo,
		moduleRepo:    moduleRepo,
		lessonRepo:    lessonRepo,
		kafkaProducer: producer,
		logger:        logger,
	}
}

func (s *courseService) CreateCourse(ctx context.Context, instructorID string, req CreateCourseRequest) (*domain.Course, error) {
	if err := validateCreateCourseRequest(req); err != nil {
		return nil, err
	}

	course := &domain.Course{
		ID:           uuid.New().String(),
		Title:        req.Title,
		Description:  req.Description,
		InstructorID: instructorID,
		ThumbnailURL: req.ThumbnailURL,
		Status:       domain.StatusDraft,
		Level:        req.Level,
		Price:        req.Price,
		Category:     req.Category,
		Tags:         req.Tags,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.courseRepo.Create(ctx, course); err != nil {
		return nil, err
	}

	event := kafka.CourseCreatedEvent{
		CourseID:     course.ID,
		Title:        course.Title,
		InstructorID: course.InstructorID,
		Timestamp:    time.Now(),
	}

	_ = s.kafkaProducer.PublishMessage(ctx, course.ID, event)

	s.logger.Info("course created", zap.String("course_id", course.ID), zap.String("instructor_id", instructorID))

	return course, nil
}

func (s *courseService) PublishCourse(ctx context.Context, courseID, instructorID string) (*domain.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if course.InstructorID != instructorID {
		return nil, domain.ErrUnauthorized
	}

	course.Status = domain.StatusPublished
	course.UpdatedAt = time.Now()

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	event := kafka.CoursePublishedEvent{
		CourseID:  course.ID,
		Title:     course.Title,
		Timestamp: time.Now(),
	}

	_ = s.kafkaProducer.PublishMessage(ctx, course.ID, event)

	s.logger.Info("course published", zap.String("course_id", courseID))

	return course, nil
}

func (s *courseService) GetCourse(ctx context.Context, courseID string) (*domain.Course, error) {
	return s.courseRepo.GetByID(ctx, courseID)
}

func (s *courseService) UpdateCourse(ctx context.Context, courseID, instructorID string, req UpdateCourseRequest) (*domain.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if course.InstructorID != instructorID {
		return nil, domain.ErrUnauthorized
	}

	if req.Title != nil {
		course.Title = *req.Title
	}
	if req.Description != nil {
		course.Description = *req.Description
	}
	if req.ThumbnailURL != nil {
		course.ThumbnailURL = *req.ThumbnailURL
	}
	if req.Level != nil {
		course.Level = *req.Level
	}
	if req.Price != nil {
		course.Price = *req.Price
	}
	if req.Category != nil {
		course.Category = *req.Category
	}
	if len(req.Tags) > 0 {
		course.Tags = req.Tags
	}

	course.UpdatedAt = time.Now()

	if err := s.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	s.logger.Info("course updated", zap.String("course_id", courseID))
	return course, nil
}

func (s *courseService) DeleteCourse(ctx context.Context, courseID, instructorID string) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return err
	}

	if course.InstructorID != instructorID {
		return domain.ErrUnauthorized
	}

	if err := s.courseRepo.Delete(ctx, courseID); err != nil {
		return err
	}

	s.logger.Info("course deleted", zap.String("course_id", courseID))
	return nil
}

func (s *courseService) ListCourses(ctx context.Context, filter CourseFilter) ([]*domain.Course, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}

	return s.courseRepo.List(ctx, filter.Page, filter.PageSize, filter.Category, filter.Status, filter.Search, filter.Level)
}

func (s *courseService) GetInstructorCourses(ctx context.Context, instructorID string, page, pageSize int) ([]*domain.Course, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.courseRepo.GetByInstructor(ctx, instructorID, page, pageSize)
}

func (s *courseService) AddModule(ctx context.Context, courseID, instructorID string, req AddModuleRequest) (*domain.Module, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if course.InstructorID != instructorID {
		return nil, domain.ErrUnauthorized
	}

	maxIndex, err := s.moduleRepo.GetMaxOrderIndex(ctx, courseID)
	if err != nil {
		return nil, err
	}

	module := &domain.Module{
		ID:          uuid.New().String(),
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		OrderIndex:  maxIndex + 1,
		CreatedAt:   time.Now(),
	}

	if err := module.Validate(); err != nil {
		return nil, err
	}

	if err := s.moduleRepo.Create(ctx, module); err != nil {
		return nil, err
	}

	s.logger.Info("module created", zap.String("module_id", module.ID), zap.String("course_id", courseID))

	return module, nil
}

func (s *courseService) UpdateModule(ctx context.Context, moduleID, courseID, instructorID string, title, description string) (*domain.Module, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if instructorID != course.InstructorID {
		return nil, domain.ErrUnauthorized
	}

	module, err := s.moduleRepo.GetByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	if module.CourseID != courseID {
		return nil, fmt.Errorf("module does not belong to course")
	}

	if title != "" {
		module.Title = title
	}
	if description != "" {
		module.Description = description
	}

	if err := s.moduleRepo.Update(ctx, module); err != nil {
		return nil, err
	}

	s.logger.Info("module updated", zap.String("module_id", moduleID))
	return module, nil
}

func (s *courseService) DeleteModule(ctx context.Context, moduleID, courseID, instructorID string) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return err
	}

	if course.InstructorID != instructorID {
		return domain.ErrUnauthorized
	}

	module, err := s.moduleRepo.GetByID(ctx, moduleID)
	if err != nil {
		return err
	}

	if module.CourseID != courseID {
		return fmt.Errorf("module does not belong to course")
	}

	if err := s.moduleRepo.Delete(ctx, moduleID); err != nil {
		return err
	}

	s.logger.Info("nodule deleted", zap.String("module_id", moduleID))
	return nil
}

func (s *courseService) GetModules(ctx context.Context, courseID string) ([]*domain.Module, error) {
	return s.moduleRepo.GetByCourseID(ctx, courseID)
}

func (s *courseService) AddLesson(ctx context.Context, moduleID, courseID, instructorID string, req AddLessonRequest) (*domain.Lesson, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if course.InstructorID != instructorID {
		return nil, domain.ErrUnauthorized
	}

	// Verify module exists and belongs to course
	module, err := s.moduleRepo.GetByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	if module.CourseID != courseID {
		return nil, fmt.Errorf("module does not belong to course")
	}

	// Get next order index
	maxIndex, err := s.lessonRepo.GetMaxOrderIndex(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	lesson := &domain.Lesson{
		ID:              uuid.New().String(),
		ModuleID:        moduleID,
		Title:           req.Title,
		Description:     req.Description,
		VideoID:         req.VideoID,
		DurationSeconds: req.DurationSeconds,
		OrderIndex:      maxIndex + 1,
		IsPreview:       req.IsPreview,
		CreatedAt:       time.Now(),
	}

	if err := lesson.Validate(); err != nil {
		return nil, err
	}

	if err := s.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, err
	}

	s.logger.Info("lesson created", zap.String("lesson_id", lesson.ID), zap.String("module_id", moduleID))
	return lesson, nil
}

func (s *courseService) UpdateLesson(ctx context.Context, lessonID, moduleID, courseID, instructorID string, req UpdateLessonRequest) (*domain.Lesson, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if course.InstructorID != instructorID {
		return nil, domain.ErrUnauthorized
	}

	// Verify module exists and belongs to course
	module, err := s.moduleRepo.GetByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	if module.CourseID != courseID {
		return nil, fmt.Errorf("module does not belong to course")
	}

	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	if lesson.ModuleID != moduleID {
		return nil, fmt.Errorf("lesson does not belong to module")
	}

	if req.Title != nil {
		lesson.Title = *req.Title
	}
	if req.Description != nil {
		lesson.Description = *req.Description
	}
	if req.IsPreview != nil {
		lesson.IsPreview = *req.IsPreview
	}

	if err := lesson.Validate(); err != nil {
		return nil, err
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, err
	}

	s.logger.Info("lesson updated", zap.String("lesson_id", lessonID))
	return lesson, nil
}

func (s *courseService) DeleteLesson(ctx context.Context, lessonID, moduleID, courseID, instructorID string) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return err
	}

	if course.InstructorID != instructorID {
		return domain.ErrUnauthorized
	}

	module, err := s.moduleRepo.GetByID(ctx, moduleID)
	if err != nil {
		return err
	}

	if module.CourseID != courseID {
		return fmt.Errorf("module does not belong to course")
	}

	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if lesson.ModuleID != moduleID {
		return fmt.Errorf("lesson does not belong to module")
	}

	if err := s.lessonRepo.Delete(ctx, lessonID); err != nil {
		return err
	}
	s.logger.Info("lesson deleted", zap.String("lesson_id", lessonID))
	return nil
}

func (s *courseService) GetLessons(ctx context.Context, moduleID string) ([]*domain.Lesson, error) {
	return s.lessonRepo.GetByModuleID(ctx, moduleID)
}

func validateCreateCourseRequest(req CreateCourseRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.Category == "" {
		return fmt.Errorf("category is required")
	}
	if req.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	return nil
}
