package grpc

import (
	"context"

	pb "github.com/dmehra2102/learning-platform/shared/proto/user"
	"github.com/dmehra2102/learning-platform/user-service/internal/domain"
	"github.com/dmehra2102/learning-platform/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, accessToken, refreshToken, err := h.service.Register(
		ctx,
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		roleFromProto(req.Role),
	)

	if err != nil {
		if err == domain.ErrEmailAlreadyExists {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		User:         userToProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, accessToken, refreshToken, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{
		User:         userToProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := h.service.GetUser(ctx, req.Id)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Error(codes.Internal, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UserResponse{
		User: userToProto(user),
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	user, err := h.service.UpdateUser(
		ctx,
		req.Id,
		req.FirstName,
		req.LastName,
		req.AvatarUrl,
		req.Bio,
	)

	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UserResponse{
		User: userToProto(user),
	}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := h.service.DeleteUser(ctx, req.Id); err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	var role *domain.UserRole
	if req.Role != nil {
		r := roleFromProto(*req.Role)
		role = &r
	}

	var statusVal *domain.UserStatus
	if req.Status != nil {
		s := statusFromProto(*req.Status)
		statusVal = &s
	}

	users, total, err := h.service.LisUsers(
		ctx,
		int(req.Page),
		int(req.PageSize),
		role,
		statusVal,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = userToProto(user)
	}

	return &pb.ListUsersResponse{
		Users:    pbUsers,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	valid, userID, role, err := h.service.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:  valid,
		UserId: userID,
		Role:   roleToProto(role),
	}, nil
}

func (h *UserHandler) GetUsersByIds(ctx context.Context, req *pb.GetUsersByIdsRequest) (*pb.GetUsersByIdsResponse, error) {
	users, err := h.service.GetUserByIDs(ctx, req.Ids)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = userToProto(user)
	}

	return &pb.GetUsersByIdsResponse{Users: pbUsers}, nil
}

func (h *UserHandler) ChangeUserRole(ctx context.Context, req *pb.ChangeUserRoleRequest) (*pb.UserResponse, error) {
	user, err := h.service.ChangeUserRole(ctx, req.Id, roleFromProto(req.Role))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UserResponse{
		User: userToProto(user),
	}, nil
}

func userToProto(user *domain.User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      roleToProto(user.Role),
		Status:    statusToProto(user.Status),
		AvatarUrl: user.AvatarURL,
		Bio:       user.Bio,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func roleToProto(role domain.UserRole) pb.UserRole {
	switch role {
	case domain.RoleStudent:
		return pb.UserRole_STUDENT
	case domain.RoleInstructor:
		return pb.UserRole_INSTRUCTOR
	case domain.RoleAdmin:
		return pb.UserRole_ADMIN
	default:
		return pb.UserRole_STUDENT
	}
}

func roleFromProto(role pb.UserRole) domain.UserRole {
	switch role {
	case pb.UserRole_ADMIN:
		return domain.RoleAdmin
	case pb.UserRole_STUDENT:
		return domain.RoleStudent
	case pb.UserRole_INSTRUCTOR:
		return domain.RoleInstructor
	default:
		return domain.RoleStudent
	}
}

func statusToProto(status domain.UserStatus) pb.UserStatus {
	switch status {
	case domain.StatusActive:
		return pb.UserStatus_ACTIVE
	case domain.StatusInactive:
		return pb.UserStatus_INACTIVE
	case domain.StatusSuspended:
		return pb.UserStatus_SUSPENDED
	default:
		return pb.UserStatus_ACTIVE
	}
}

func statusFromProto(status pb.UserStatus) domain.UserStatus {
	switch status {
	case pb.UserStatus_ACTIVE:
		return domain.StatusActive
	case pb.UserStatus_INACTIVE:
		return domain.StatusInactive
	case pb.UserStatus_SUSPENDED:
		return domain.StatusSuspended
	default:
		return domain.StatusActive
	}
}
