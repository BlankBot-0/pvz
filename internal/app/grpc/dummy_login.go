package merch_store

import (
	"context"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) DummyLogin(ctx context.Context, request *pvzpb.DummyLoginRequest) (*pvzpb.DummyLoginResponse, error) {
	res, err := s.Auth.DummyLogin(ctx, request.Role)
	if err != nil {
	}

	return &pvzpb.DummyLoginResponse{Token: res}, nil
}
