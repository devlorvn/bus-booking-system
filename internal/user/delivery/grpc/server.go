package grpc

import (
	"booking-system/internal/user/domain"
	"booking-system/internal/user/usecase"
	userpb "booking-system/proto/user/v1"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGRPCServer struct {
	userpb.UnimplementedUserServiceServer
	userUsecase *usecase.UserUsecase
}

func NewUserGRPCServer(userUsecase *usecase.UserUsecase) *UserGRPCServer {
	return &UserGRPCServer{
		userUsecase: userUsecase,
	}
}

func (s *UserGRPCServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	u := &domain.User{
		ID:          uuid.New(),
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	created, err := s.userUsecase.Create(ctx, u)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &userpb.CreateUserResponse{
		User: &userpb.User{
			Id:          created.ID.String(),
			Name:        created.Name,
			Email:       created.Email,
			PhoneNumber: created.PhoneNumber,
		},
	}, nil
}

func (s *UserGRPCServer) GetUserByID(ctx context.Context, req *userpb.GetUserByIDRequest) (*userpb.GetUserByIDResponse, error) {
	userUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	u, err := s.userUsecase.GetByID(ctx, userUUID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &userpb.GetUserByIDResponse{
		User: &userpb.User{
			Id:          u.ID.String(),
			Name:        u.Name,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber,
		},
	}, nil
}

func (s *UserGRPCServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	userUUID, err := uuid.Parse(req.User.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	u := &domain.User{
		ID:          userUUID,
		Name:        req.User.Name,
		Email:       req.User.Email,
		PhoneNumber: req.User.PhoneNumber,
	}

	updated, err := s.userUsecase.Update(ctx, u)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &userpb.UpdateUserResponse{
		User: &userpb.User{
			Id:          updated.ID.String(),
			Name:        updated.Name,
			Email:       updated.Email,
			PhoneNumber: updated.PhoneNumber,
		},
	}, nil
}

func (s *UserGRPCServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	userUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	err = s.userUsecase.Delete(ctx, userUUID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &userpb.DeleteUserResponse{Success: true}, nil
}

func (s *UserGRPCServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	users, err := s.userUsecase.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	pbUsers := make([]*userpb.User, 0, len(users))
	for _, u := range users {
		pbUsers = append(pbUsers, &userpb.User{
			Id:          u.ID.String(),
			Name:        u.Name,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber,
		})
	}

	return &userpb.ListUsersResponse{Users: pbUsers}, nil
}

func (s *UserGRPCServer) FindByPhone(ctx context.Context, req *userpb.FindByPhoneRequest) (*userpb.FindByPhoneResponse, error) {
	u, err := s.userUsecase.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find user by phone: %v", err)
	}

	if u == nil {
		return &userpb.FindByPhoneResponse{User: nil}, nil
	}

	return &userpb.FindByPhoneResponse{
		User: &userpb.User{
			Id:          u.ID.String(),
			Name:        u.Name,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber,
		},
	}, nil
}
