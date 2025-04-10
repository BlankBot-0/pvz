package merch_store

import (
	"context"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) GetPVZList(ctx context.Context) (*pvzpb.GetPVZListResponse, error) {
	res, err := s.PVZ.GetPVZList(ctx)
	if err != nil {
	}

	return &pvzpb.GetPVZListResponse{
		Pvzs: lo.Map(res, func(pvz models.PVZ, _ int) *pvzpb.PVZ {
			return &pvzpb.PVZ{
				Id:               pvz.Id,
				RegistrationDate: timestamppb.New(pvz.RegistrationDate),
				City:             pvz.City,
			}
		}),
	}, nil
}
