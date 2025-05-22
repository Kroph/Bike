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

	// Map proto role to domain role
	var role domain.UserRole
	switch req.Role {
	case pb.UserRole_ADMIN:
		role = domain.UserRoleAdmin
	case pb.UserRole_USER:
		fallthrough
	default:
		role = domain.UserRoleUser
	}

	user := domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     role,
	}

	createdUser, err := h.userService.RegisterUser(ctx, user)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	// Map domain role to proto role
	var protoRole pb.UserRole
	switch createdUser.Role {
	case domain.UserRoleAdmin:
		protoRole = pb.UserRole_ADMIN
	case domain.UserRoleUser:
		fallthrough
	default:
		protoRole = pb.UserRole_USER
	}

	return &pb.UserResponse{
		Id:        createdUser.ID,
		Username:  createdUser.Username,
		Email:     createdUser.Email,
		Role:      protoRole,
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

	// Map domain role to proto role
	var protoRole pb.UserRole
	switch user.Role {
	case domain.UserRoleAdmin:
		protoRole = pb.UserRole_ADMIN
	case domain.UserRoleUser:
		fallthrough
	default:
		protoRole = pb.UserRole_USER
	}

	return &pb.AuthResponse{
		Token:    token,
		UserId:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     protoRole,
	}, nil
}

func (h *UserGrpcHandler) GetUserProfile(ctx context.Context, req *pb.UserIDRequest) (*pb.UserProfile, error) {
	log.Printf("Received GetUserProfile request for user ID: %s", req.UserId)

	user, err := h.userService.GetUserProfile(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Map domain role to proto role
	var protoRole pb.UserRole
	switch user.Role {
	case domain.UserRoleAdmin:
		protoRole = pb.UserRole_ADMIN
	case domain.UserRoleUser:
		fallthrough
	default:
		protoRole = pb.UserRole_USER
	}

	return &pb.UserProfile{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      protoRole,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (h *UserGrpcHandler) GenerateVerificationCode(ctx context.Context, req *pb.GenerateCodeRequest) (*pb.GenerateCodeResponse, error) {
	log.Printf("Received GenerateVerificationCode request for user ID: %s", req.UserId)

	code, err := h.userService.GenerateVerificationCode(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to generate verification code: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to generate verification code: %v", err)
	}

	return &pb.GenerateCodeResponse{
		Success: true,
		Code:    code,
		Message: "Verification code generated successfully",
	}, nil
}

func (h *UserGrpcHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	log.Printf("Received VerifyEmail request for user ID: %s with code: %s", req.UserId, req.Code)

	err := h.userService.VerifyEmailCode(ctx, req.UserId, req.Code)
	if err != nil {
		log.Printf("Failed to verify email: %v", err)
		return &pb.VerifyEmailResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.VerifyEmailResponse{
		Success: true,
		Message: "Email verified successfully",
	}, nil
}
