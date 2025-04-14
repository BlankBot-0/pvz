package pvz

import (
	"context"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) GetPVZList(ctx context.Context, _ *emptypb.Empty) (*pvzpb.GetPVZListResponse, error) {
	res, err := s.PVZ.ListPVZ(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while getting pvz list: %s", err.Error())
	}

	return &pvzpb.GetPVZListResponse{
		Pvzs: lo.Map(res, func(pvz models.PVZ, _ int) *pvzpb.PVZ {

			return &pvzpb.PVZ{
				Id:               &pvz.ID,
				RegistrationDate: lo.ToPtr(pvz.RegistrationDate.Format(s.datetimeFormat)),
				City:             pvz.City,
			}
		}),
	}, nil
}
