package merch_store

import (
	"context"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) Register(ctx context.Context, request *pvzpb.RegisterRequest) (*pvzpb.RegisterResponse, error) {
	res, err := s.Auth.Register(ctx, request.Email, request.Password, request.Role)
	if err != nil {
	}

	return &pvzpb.RegisterResponse{
		Id:    res.Id,
		Email: res.Email,
		Role:  res.Role,
	}, nil
}
