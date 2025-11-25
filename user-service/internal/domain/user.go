package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password too weak")
)

type UserRole string

const (
	RoleStudent    UserRole = "STUDENT"
	RoleInstructor UserRole = "INSTRUCTOR"
	RoleAdmin      UserRole = "ADMIN"
)

type UserStatus string

const (
	StatusActive    UserStatus = "ACTIVE"
	StatusInactive  UserStatus = "INACTIVE"
	StatusSuspended UserStatus = "SUSPENDED"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         UserRole
	Status       UserStatus
	AvatarURL    string
	Bio          string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, firstname, lastname string, role UserRole) (*User, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	return &User{
		Email:     email,
		FirstName: firstname,
		LastName:  lastname,
		Role:      role,
		Status:    StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (u *User) UpdateProfile(firstName, lastName, avatarURL, bio *string) {
	if firstName != nil {
		u.FirstName = *firstName
	}
	if lastName != nil {
		u.LastName = *lastName
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	if bio != nil {
		u.Bio = *bio
	}

	u.UpdatedAt = time.Now()
}

func (u *User) Activate() {
	u.Status = StatusActive
	u.UpdatedAt = time.Now()
}

func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now()
}

func (u *User) ChangeRole(role UserRole) {
	u.Role = role
	u.UpdatedAt = time.Now()
}

func isValidEmail(email string) bool {
	return len(email) > 3 && len(email) < 255 &&
		contains(email, "@") && contains(email, ".")
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
