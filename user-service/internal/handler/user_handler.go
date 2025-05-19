package handler

import (
	"context"
	"log"

	pb "proto/user"
	"user-service/internal/domain"
	"user-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGrpcHandler struct {
	pb.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserGrpcHandler(userService service.UserService) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
	}
}

func (h *UserGrpcHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.UserResponse, error) {
	log.Printf("Received RegisterUser request for email: %s", req.Email)

	user := domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	createdUser, err := h.userService.RegisterUser(ctx, user)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &pb.UserResponse{
		Id:        createdUser.ID,
		Username:  createdUser.Username,
		Email:     createdUser.Email,
		CreatedAt: timestamppb.New(createdUser.CreatedAt),
	}, nil
}

func (h *UserGrpcHandler) AuthenticateUser(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	log.Printf("Received AuthenticateUser request for email: %s", req.Email)

	token, user, err := h.userService.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
	}

	return &pb.AuthResponse{
		Token:    token,
		UserId:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (h *UserGrpcHandler) GetUserProfile(ctx context.Context, req *pb.UserIDRequest) (*pb.UserProfile, error) {
	log.Printf("Received GetUserProfile request for user ID: %s", req.UserId)

	user, err := h.userService.GetUserProfile(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.UserProfile{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}
