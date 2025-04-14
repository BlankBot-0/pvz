package pvz

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/usecase/auth"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) DummyLogin(ctx context.Context, request *pvzpb.DummyLoginRequest) (*pvzpb.DummyLoginResponse, error) {
	res, err := s.Auth.DummyLogin(ctx, request.Role)
	if errors.Is(err, auth.ErrInvalidRole) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while getting token through dummy login: %s", err.Error())
	}

	return &pvzpb.DummyLoginResponse{Token: res}, nil
}
