package merch_store

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) CreatePvz(ctx context.Context, req *pvzpb.CreatePVZRequest) (*pvzpb.CreatePVZResponse, error) {
	res, err := s.PVZ.CreatePVZ(ctx, req.City)
	if err != nil {
	}

	return &pvzpb.CreatePVZResponse{
		Pvz: &pvzpb.PVZ{
			Id:               res.Id,
			RegistrationDate: timestamppb.New(res.RegistrationDate),
			City:             res.City,
		},
	}, nil
}
