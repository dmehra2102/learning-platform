package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dmehra2102/learning-platform/shared/pkg/jwt"
	"github.com/dmehra2102/learning-platform/shared/pkg/kafka"
	"github.com/dmehra2102/learning-platform/user-service/internal/domain"
	"github.com/dmehra2102/learning-platform/user-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, email, password, firstName, lastName string, role domain.UserRole) (*domain.User, string, string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, string, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUserByIDs(ctx context.Context, ids []string) ([]*domain.User, error)
	UpdateUser(ctx context.Context, id string, firstName, lastName, avatarURL, bio *string) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	LisUsers(ctx context.Context, page, pageSize int, role *domain.UserRole, status *domain.UserStatus) ([]*domain.User, int, error)
	ValidateToken(ctx context.Context, token string) (bool, string, domain.UserRole, error)
	ChangeUserRole(ctx context.Context, id string, role domain.UserRole) (*domain.User, error)
}

type userService struct {
	repo          repository.UserRepository
	jwtManager    *jwt.Manager
	kafkaProducer *kafka.Producer
	logger        *zap.Logger
}

func NewUserService(
	repo repository.UserRepository,
	jwtManager *jwt.Manager,
	kafkaProducer *kafka.Producer,
	logger *zap.Logger,
) UserService {
	return &userService{
		repo:          repo,
		jwtManager:    jwtManager,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

func (s *userService) Register(ctx context.Context, email, password, firstName, lastName string, role domain.UserRole) (*domain.User, string, string, error) {
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, "", "", fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, "", "", domain.ErrEmailAlreadyExists
	}

	// hashing the password to store it in DB
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		Status:       domain.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed tp generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	event := kafka.UserRegisteredEvent{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
		Timestamp: time.Now(),
	}

	if err := s.kafkaProducer.PublishMessage(ctx, user.ID, event); err != nil {
		s.logger.Error("failed to publish user registered event", zap.Error(err))
	}

	s.logger.Info("user registered successfully", zap.String("user_id", user.ID))

	return user, accessToken, refreshToken, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, "", "", domain.ErrInvalidCredentials
		}
		return nil, "", "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", domain.ErrInvalidCredentials
	}

	if user.Status != domain.StatusActive {
		return nil, "", "", fmt.Errorf("user account is %s", user.Status)
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.logger.Info("user logged in successfully", zap.String("user_id", user.ID))

	return user, accessToken, refreshToken, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetUserByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
	return s.repo.GetByIDs(ctx, ids)
}

func (s *userService) UpdateUser(ctx context.Context, id string, firstName, lastName, avatarURL, bio *string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.UpdateProfile(firstName, lastName, avatarURL, bio)

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("user updated successfully", zap.String("user_id", user.ID))

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Info("user deleted successfully", zap.String("user_id", id))

	return nil
}

func (s *userService) LisUsers(ctx context.Context, page, pageSize int, role *domain.UserRole, status *domain.UserStatus) ([]*domain.User, int, error) {
	return s.repo.List(ctx, page, pageSize, role, status)
}

func (s *userService) ValidateToken(ctx context.Context, token string) (bool, string, domain.UserRole, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return false, "", "", err
	}

	return true, claims.UserID, domain.UserRole(claims.Role), nil
}

func (s *userService) ChangeUserRole(ctx context.Context, id string, role domain.UserRole) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.ChangeRole(role)

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	s.logger.Info("user role changed successfully",
		zap.String("user_id", user.ID),
		zap.String("new_role", string(role)),
	)

	return user, nil
}
