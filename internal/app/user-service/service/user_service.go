package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-standards/project-layout/internal/app/user-service/model"
	"github.com/golang-standards/project-layout/internal/app/user-service/repository"
	"github.com/golang-standards/project-layout/internal/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidEmail    = errors.New("invalid email")
)

// UserService defines the business logic interface for user operations
type UserService interface {
	CreateUser(ctx context.Context, email, password, firstName, lastName, phone string) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, pageSize int, filter string) ([]*model.User, int64, error)
	ValidatePassword(ctx context.Context, email, password string) (*model.User, error)
}

type userService struct {
	repo   repository.UserRepository
	logger logger.Logger
}

// NewUserService creates a new instance of UserService
func NewUserService(repo repository.UserRepository, logger logger.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

// CreateUser creates a new user with encrypted password
func (s *userService) CreateUser(ctx context.Context, email, password, firstName, lastName, phone string) (*model.User, error) {
	s.logger.Info("Creating new user", "email", email)

	// Validate input
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if password == "" || len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Status:    model.UserStatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", "error", err, "email", email)
		return nil, err
	}

	s.logger.Info("User created successfully", "user_id", user.ID, "email", email)
	return user, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
	s.logger.Debug("Getting user", "user_id", id)

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err, "user_id", id)
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	s.logger.Debug("Getting user by email", "email", email)

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to get user by email", "error", err, "email", email)
		return nil, err
	}

	return user, nil
}

// UpdateUser updates user information
func (s *userService) UpdateUser(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
	s.logger.Info("Updating user", "user_id", id)

	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if firstName, ok := updates["first_name"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := updates["last_name"].(string); ok {
		user.LastName = lastName
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
	}
	if status, ok := updates["status"].(model.UserStatus); ok {
		user.Status = status
	}

	// Update in repository
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user", "error", err, "user_id", id)
		return nil, err
	}

	s.logger.Info("User updated successfully", "user_id", id)
	return user, nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Deleting user", "user_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete user", "error", err, "user_id", id)
		return err
	}

	s.logger.Info("User deleted successfully", "user_id", id)
	return nil
}

// ListUsers retrieves a paginated list of users
func (s *userService) ListUsers(ctx context.Context, page, pageSize int, filter string) ([]*model.User, int64, error) {
	s.logger.Debug("Listing users", "page", page, "page_size", pageSize, "filter", filter)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := s.repo.List(ctx, page, pageSize, filter)
	if err != nil {
		s.logger.Error("Failed to list users", "error", err)
		return nil, 0, err
	}

	return users, total, nil
}

// ValidatePassword validates user credentials
func (s *userService) ValidatePassword(ctx context.Context, email, password string) (*model.User, error) {
	s.logger.Debug("Validating user password", "email", email)

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Warn("Invalid password attempt", "email", email)
		return nil, ErrInvalidPassword
	}

	return user, nil
}
