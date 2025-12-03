package grpc

import (
	"context"

	"github.com/dmehra2102/learning-platform/course-service/internal/domain"
	"github.com/dmehra2102/learning-platform/course-service/internal/service"
	"github.com/dmehra2102/learning-platform/shared/pkg/interceptor"
	pb "github.com/dmehra2102/learning-platform/shared/proto/course"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CourseHandler struct {
	pb.UnimplementedCourseServiceServer
	service service.CourseService
}

func NewCourseHandler(service service.CourseService) *CourseHandler {
	return &CourseHandler{service: service}
}

func (h *CourseHandler) CreateCourse(ctx context.Context, req *pb.CreateCourseRequest) (*pb.CourseResponse, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	course, err := h.service.CreateCourse(ctx, instructorID, service.CreateCourseRequest{
		Title:        req.Title,
		Description:  req.Description,
		ThumbnailURL: req.ThumbnailUrl,
		Level:        levelFromProto(req.Level),
		Price:        req.Price,
		Category:     req.Category,
		Tags:         req.Tags,
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CourseResponse{Course: courseToProto(course)}, nil
}

func (h *CourseHandler) GetCourse(ctx context.Context, req *pb.GetCourseRequest) (*pb.CourseResponse, error) {
	course, err := h.service.GetCourse(ctx, req.Id)
	if err != nil {
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "course not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CourseResponse{Course: courseToProto(course)}, nil
}

func (h *CourseHandler) UpdateCourse(ctx context.Context, req *pb.UpdateCourseRequest) (*pb.CourseResponse, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	updateReq := service.UpdateCourseRequest{
		Title:        req.Title,
		Description:  req.Description,
		ThumbnailURL: req.ThumbnailUrl,
		Category:     req.Category,
		Tags:         req.Tags,
	}

	if req.Level != nil {
		level := levelFromProto(*req.Level)
		updateReq.Level = &level
	}

	if req.Price != nil {
		updateReq.Price = req.Price
	}

	course, err := h.service.UpdateCourse(ctx, req.Id, instructorID, updateReq)
	if err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "course not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CourseResponse{Course: courseToProto(course)}, nil
}

func (h *CourseHandler) DeleteCourse(ctx context.Context, req *pb.DeleteCourseRequest) (*emptypb.Empty, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := h.service.DeleteCourse(ctx, req.Id, instructorID); err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "course not found")
		}
	}

	return &emptypb.Empty{}, nil
}

func (h *CourseHandler) ListCourses(ctx context.Context, req *pb.ListCoursesRequest) (*pb.ListCoursesResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := service.CourseFilter{
		Page:     page,
		PageSize: pageSize,
		Category: req.Category,
		Search:   req.Search,
	}

	if req.Level != nil {
		level := levelFromProto(*req.Level)
		filter.Level = &level
	}

	courses, total, err := h.service.ListCourses(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCourses := make([]*pb.Course, len(courses))
	for i, course := range courses {
		pbCourses[i] = courseToProto(course)
	}

	return &pb.ListCoursesResponse{
		Courses:  pbCourses,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (h *CourseHandler) PublishCourse(ctx context.Context, req *pb.PublishCourseRequest) (*pb.CourseResponse, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	course, err := h.service.PublishCourse(ctx, req.Id, instructorID)
	if err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "course not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CourseResponse{Course: courseToProto(course)}, nil
}

func (h *CourseHandler) GetCoursesByInstructor(ctx context.Context, req *pb.GetCoursesByInstructorRequest) (*pb.ListCoursesResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	courses, total, err := h.service.GetInstructorCourses(ctx, req.InstructorId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCourses := make([]*pb.Course, len(courses))
	for i, course := range courses {
		pbCourses[i] = courseToProto(course)
	}

	return &pb.ListCoursesResponse{
		Courses:  pbCourses,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (h *CourseHandler) AddModule(ctx context.Context, req *pb.AddModuleRequest) (*pb.ModuleResponse, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	module, err := h.service.AddModule(ctx, req.CourseId, instructorID, service.AddModuleRequest{
		Title:       req.Title,
		Description: req.Description,
	})

	if err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "course not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ModuleResponse{Module: moduleToProto(module)}, nil
}

func (h *CourseHandler) UpdateModule(ctx context.Context, req *pb.UpdateModuleRequest) (*pb.ModuleResponse, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	module, err := h.service.UpdateModule(ctx, req.Id, *req.CourseId, instructorID, *req.Title, *req.Description)
	if err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ModuleResponse{Module: moduleToProto(module)}, nil
}

func (h *CourseHandler) DeleteModule(ctx context.Context, req *pb.DeleteModuleRequest) (*emptypb.Empty, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := h.service.DeleteModule(ctx, req.Id, req.CourseId, instructorID); err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *CourseHandler) GetModules(ctx context.Context, req *pb.GetModulesRequest) (*pb.ListModulesResponse, error) {
	modules, err := h.service.GetModules(ctx, req.CourseId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbModules := make([]*pb.Module, len(modules))
	for i, module := range modules {
		pbModules[i] = moduleToProto(module)
	}

	return &pb.ListModulesResponse{Modules: pbModules}, nil
}

func (h *CourseHandler) AddLesson(ctx context.Context, req *pb.AddLessonRequest) (*pb.LessonResponse, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	lesson, err := h.service.AddLesson(ctx, req.ModuleId, req.CourseId, instructorID, service.AddLessonRequest{
		Title:           req.Title,
		Description:     req.Description,
		VideoID:         req.VideoId,
		DurationSeconds: int(req.DurationSeconds),
		IsPreview:       req.IsPreview,
	})

	if err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LessonResponse{Lesson: lessonToProto(lesson)}, nil
}

func (h *CourseHandler) UpdateLesson(ctx context.Context, req *pb.UpdateLessonRequest) (*pb.LessonResponse, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	lesson, err := h.service.UpdateLesson(ctx, req.Id, req.ModuleId, req.CourseId, instructorID, service.UpdateLessonRequest{
		Title:           req.Title,
		Description:     req.Description,
		IsPreview:       req.IsPreview,
	})

	if err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LessonResponse{Lesson: lessonToProto(lesson)}, nil
}

func (h *CourseHandler) DeleteLesson(ctx context.Context, req *pb.DeleteLessonRequest) (*emptypb.Empty, error) {
	instructorID, err := interceptor.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err := h.service.DeleteLesson(ctx, req.Id, req.ModuleId, req.CourseId, instructorID); err != nil {
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "unauthorized")
		}
		if err == domain.ErrCourseNotFound {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *CourseHandler) GetLessons(ctx context.Context, req *pb.GetLessonsRequest) (*pb.ListLessonsResponse, error) {
	lessons, err := h.service.GetLessons(ctx, req.ModuleId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbLessons := make([]*pb.Lesson, len(lessons))
	for i, lesson := range lessons {
		pbLessons[i] = lessonToProto(lesson)
	}

	return &pb.ListLessonsResponse{Lessons: pbLessons}, nil
}

func courseToProto(course *domain.Course) *pb.Course {
	return &pb.Course{
		Id:              course.ID,
		Title:           course.Title,
		Description:     course.Description,
		InstructorId:    course.InstructorID,
		ThumbnailUrl:    course.ThumbnailURL,
		Status:          statusToProto(course.Status),
		Level:           levelToProto(course.Level),
		Price:           course.Price,
		Category:        course.Category,
		Tags:            course.Tags,
		DurationMinutes: int32(course.DurationMinutes),
		CreatedAt:       timestamppb.New(course.CreatedAt),
		UpdatedAt:       timestamppb.New(course.UpdatedAt),
		EnrolledCount:   int32(course.EnrolledCount),
		AverageRating:   course.AverageRating,
	}
}

func moduleToProto(module *domain.Module) *pb.Module {
	return &pb.Module{
		Id:          module.ID,
		CourseId:    module.CourseID,
		Title:       module.Title,
		Description: module.Description,
		OrderIndex:  int32(module.OrderIndex),
		CreatedAt:   timestamppb.New(module.CreatedAt),
	}
}

func lessonToProto(lesson *domain.Lesson) *pb.Lesson {
	return &pb.Lesson{
		Id:              lesson.ID,
		ModuleId:        lesson.ModuleID,
		Title:           lesson.Title,
		Description:     lesson.Description,
		VideoId:         lesson.VideoID,
		DurationSeconds: int32(lesson.DurationSeconds),
		OrderIndex:      int32(lesson.OrderIndex),
		IsPreview:       lesson.IsPreview,
		CreatedAt:       timestamppb.New(lesson.CreatedAt),
	}
}

func statusToProto(status domain.CourseStatus) pb.CourseStatus {
	switch status {
	case domain.StatusPublished:
		return pb.CourseStatus_PUBLISHED
	case domain.StatusArchived:
		return pb.CourseStatus_ARCHIVED
	default:
		return pb.CourseStatus_DRAFT
	}
}

func levelToProto(level domain.CourseLevel) pb.CourseLevel {
	switch level {
	case domain.LevelIntermediate:
		return pb.CourseLevel_INTERMEDIATE
	case domain.LevelAdvanced:
		return pb.CourseLevel_ADVANCED
	default:
		return pb.CourseLevel_BEGINNER
	}
}

func levelFromProto(level pb.CourseLevel) domain.CourseLevel {
	switch level {
	case pb.CourseLevel_ADVANCED:
		return domain.LevelAdvanced
	case pb.CourseLevel_INTERMEDIATE:
		return domain.LevelIntermediate
	default:
		return domain.LevelBeginner
	}
}
