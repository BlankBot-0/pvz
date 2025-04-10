package merch_store

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) CreateReception(ctx context.Context, request *pvzpb.CreateReceptionRequest) (*pvzpb.CreateReceptionResponse, error) {
	res, err := s.PVZ.CreateReception(ctx, request.PvzId)
	if err != nil {
	}

	return &pvzpb.CreateReceptionResponse{
		Reception: &pvzpb.Reception{
			Id:       res.Id,
			DateTime: timestamppb.New(res.DateTime),
			PvzId:    res.PvzId,
			Status:   res.ReceptionStatus,
		},
	}, nil
}
