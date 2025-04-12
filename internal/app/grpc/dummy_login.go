package merch_store

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) DummyLogin(ctx context.Context, request *pvzpb.DummyLoginRequest) (*pvzpb.DummyLoginResponse, error) {
	res, err := s.Auth.DummyLogin(ctx, request.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error while getting token through dummy login")
	}

	return &pvzpb.DummyLoginResponse{Token: res}, nil
}
