package pvz

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/usecase/auth"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) Register(ctx context.Context, request *pvzpb.RegisterRequest) (*pvzpb.RegisterResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	res, err := s.Auth.Register(ctx, request.Email, request.Password, request.Role)
	if errors.Is(err, auth.ErrUserAlreadyExists) {
		return nil, status.Error(codes.FailedPrecondition, "user with such email already exists")
	} else if errors.Is(err, auth.ErrInvalidRole) {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while registering user: %s", err.Error())
	}

	return &pvzpb.RegisterResponse{
		Id:    res.ID,
		Email: res.Email,
		Role:  res.Role,
	}, nil
}
