package pvz

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz/internal/auth"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) CreateReception(ctx context.Context, request *pvzpb.CreateReceptionRequest) (*pvzpb.CreateReceptionResponse, error) {
	if auth.GetUserRoleFromCtx(ctx) != employeeRole {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	res, err := s.PVZ.CreateReception(ctx, request.PvzId)
	if errors.Is(err, pvz.ErrPVZNotFound) {
		return nil, status.Error(codes.InvalidArgument, "no receptions in progress in pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrAnotherReceptionInProgress) {
		return nil, status.Error(codes.FailedPrecondition, "another reception in progress in pvz with id "+request.PvzId)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while creating reception: %s", err.Error())
	}

	return &pvzpb.CreateReceptionResponse{
		Reception: &pvzpb.Reception{
			Id:       res.ID,
			DateTime: timestamppb.New(res.DateTime),
			PvzId:    res.PvzID,
			Status:   res.ReceptionStatus,
		},
	}, nil
}
