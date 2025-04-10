package merch_store

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) CloseLastReception(ctx context.Context, request *pvzpb.CloseLastReceptionRequest) (*pvzpb.CloseLastReceptionResponse, error) {
	res, err := s.PVZ.CloseLastReception(ctx, request.PvzId)
	if err != nil {
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
