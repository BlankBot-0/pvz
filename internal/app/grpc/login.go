package merch_store

import (
	"context"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) Login(ctx context.Context, request *pvzpb.LoginRequest) (*pvzpb.LoginResponse, error) {
	res, err := s.Auth.UserToken(ctx, request.Email, request.Password)
	if err != nil {
	}

	return &pvzpb.LoginResponse{Token: res}, err
}
