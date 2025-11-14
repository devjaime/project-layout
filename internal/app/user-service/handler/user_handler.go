package handler

import (
	"context"
	"errors"

	"github.com/golang-standards/project-layout/internal/app/user-service/model"
	"github.com/golang-standards/project-layout/internal/app/user-service/repository"
	"github.com/golang-standards/project-layout/internal/app/user-service/service"
	"github.com/golang-standards/project-layout/internal/pkg/logger"
	pb "github.com/golang-standards/project-layout/pkg/api/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserHandler implements the gRPC user service
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service service.UserService
	logger  logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(service service.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	h.logger.Info("CreateUser request received", "email", req.Email)

	user, err := h.service.CreateUser(ctx, req.Email, req.Password, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		h.logger.Error("Failed to create user", "error", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.CreateUserResponse{
		User: h.modelToProto(user),
	}, nil
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	h.logger.Debug("GetUser request received", "user_id", req.Id)

	user, err := h.service.GetUser(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		h.logger.Error("Failed to get user", "error", err)
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &pb.GetUserResponse{
		User: h.modelToProto(user),
	}, nil
}

// GetUserByEmail retrieves a user by email
func (h *UserHandler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserResponse, error) {
	h.logger.Debug("GetUserByEmail request received", "email", req.Email)

	user, err := h.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		h.logger.Error("Failed to get user by email", "error", err)
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &pb.GetUserResponse{
		User: h.modelToProto(user),
	}, nil
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	h.logger.Info("UpdateUser request received", "user_id", req.Id)

	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Status != nil {
		updates["status"] = h.protoStatusToModel(*req.Status)
	}

	user, err := h.service.UpdateUser(ctx, req.Id, updates)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		h.logger.Error("Failed to update user", "error", err)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return &pb.UpdateUserResponse{
		User: h.modelToProto(user),
	}, nil
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	h.logger.Info("DeleteUser request received", "user_id", req.Id)

	if err := h.service.DeleteUser(ctx, req.Id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		h.logger.Error("Failed to delete user", "error", err)
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

// ListUsers retrieves a paginated list of users
func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	h.logger.Debug("ListUsers request received", "page", req.Page, "page_size", req.PageSize)

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	users, total, err := h.service.ListUsers(ctx, page, pageSize, req.Filter)
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = h.modelToProto(user)
	}

	return &pb.ListUsersResponse{
		Users:    pbUsers,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// modelToProto converts model.User to pb.User
func (h *UserHandler) modelToProto(user *model.User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Status:    h.modelStatusToProto(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

// modelStatusToProto converts model status to proto status
func (h *UserHandler) modelStatusToProto(status model.UserStatus) pb.UserStatus {
	switch status {
	case model.UserStatusActive:
		return pb.UserStatus_USER_STATUS_ACTIVE
	case model.UserStatusInactive:
		return pb.UserStatus_USER_STATUS_INACTIVE
	case model.UserStatusSuspended:
		return pb.UserStatus_USER_STATUS_SUSPENDED
	default:
		return pb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

// protoStatusToModel converts proto status to model status
func (h *UserHandler) protoStatusToModel(status pb.UserStatus) model.UserStatus {
	switch status {
	case pb.UserStatus_USER_STATUS_ACTIVE:
		return model.UserStatusActive
	case pb.UserStatus_USER_STATUS_INACTIVE:
		return model.UserStatusInactive
	case pb.UserStatus_USER_STATUS_SUSPENDED:
		return model.UserStatusSuspended
	default:
		return model.UserStatusActive
	}
}
