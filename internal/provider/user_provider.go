package provider

import (
	userDomain "booking-system/internal/user/domain"
	userpb "booking-system/proto/user/v1"
	"context"

	"github.com/google/uuid"
)

type UserProvider struct {
	client userpb.UserServiceClient
}

func NewUserProvider(client userpb.UserServiceClient) *UserProvider {
	return &UserProvider{client: client}
}

func (p *UserProvider) FindByPhone(ctx context.Context, phone string) (*userDomain.User, error) {
	resp, err := p.client.FindByPhone(ctx, &userpb.FindByPhoneRequest{
		Phone: phone,
	})
	if err != nil {
		return nil, err
	}
	return &userDomain.User{
		ID:          uuid.MustParse(resp.User.Id),
		PhoneNumber: resp.User.PhoneNumber,
		Name:        resp.User.Name,
	}, nil
}

func (p *UserProvider) Create(ctx context.Context, user *userDomain.User) error {
	_, err := p.client.CreateUser(ctx, &userpb.CreateUserRequest{
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Email:       user.Email,
	})
	return err
}

func (p *UserProvider) Update(ctx context.Context, user *userDomain.User) error {
	_, err := p.client.UpdateUser(ctx, &userpb.UpdateUserRequest{
		User: &userpb.User{
			Id:          user.ID.String(),
			PhoneNumber: user.PhoneNumber,
			Name:        user.Name,
			Email:       user.Email,
		},
	})
	return err
}
