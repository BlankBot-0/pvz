package pvz

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/usecase/auth"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) Login(ctx context.Context, request *pvzpb.LoginRequest) (*pvzpb.LoginResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	res, err := s.Auth.UserToken(ctx, request.Email, request.Password)
	if errors.Is(err, auth.ErrUserNotFound) {
		return nil, status.Error(codes.Unauthenticated, "user with such email not found")
	} else if errors.Is(err, auth.ErrIncorrectPassword) {
		return nil, status.Error(codes.Unauthenticated, "incorrect email or password")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while logging in: %s", err.Error())
	}

	return &pvzpb.LoginResponse{Token: res}, err
}
