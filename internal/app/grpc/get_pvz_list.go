package merch_store

import (
	"context"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) GetPVZList(ctx context.Context, _ *emptypb.Empty) (*pvzpb.GetPVZListResponse, error) {
	res, err := s.PVZ.GetPVZList(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error while getting pvz list")
	}

	return &pvzpb.GetPVZListResponse{
		Pvzs: lo.Map(res, func(pvz models.PVZ, _ int) *pvzpb.PVZ {
			return &pvzpb.PVZ{
				Id:               &pvz.Id,
				RegistrationDate: timestamppb.New(pvz.RegistrationDate),
				City:             pvz.City,
			}
		}),
	}, nil
}
