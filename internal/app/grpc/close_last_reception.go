package merch_store

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) CloseLastReception(ctx context.Context, request *pvzpb.CloseLastReceptionRequest) (*pvzpb.CloseLastReceptionResponse, error) {
	res, err := s.PVZ.CloseLastReception(ctx, request.PvzId)
	if errors.Is(err, pvz.ErrPVZNotFound) {
		return nil, status.Error(codes.NotFound, "pvz with id "+request.PvzId+" not found")
	} else if errors.Is(err, pvz.ErrNoReceptionsInProgress) {
		return nil, status.Error(codes.NotFound, "no receptions in progress in pvz with id "+request.PvzId)
	} else if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error while closing last reception")
	}

	return &pvzpb.CloseLastReceptionResponse{
		Reception: &pvzpb.Reception{
			Id:       res.Id,
			DateTime: timestamppb.New(res.DateTime),
			PvzId:    res.PvzId,
			Status:   res.ReceptionStatus,
		},
	}, nil
}
