package merch_store

import (
	"context"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	pvzpb "pvz/pkg/api/v1"
	"time"
)

func (s *Service) CreatePvz(ctx context.Context, req *pvzpb.CreatePVZRequest) (*pvzpb.CreatePVZResponse, error) {
	var customRegistrationDate *time.Time
	if req.RegistrationDate != nil {
		customRegistrationDate = lo.ToPtr(req.RegistrationDate.AsTime())
	}

	res, err := s.PVZ.CreatePVZ(ctx, req.City, req.Id, customRegistrationDate)
	if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error while creating pvz")
	}

	return &pvzpb.CreatePVZResponse{
		Pvz: &pvzpb.PVZ{
			Id:               &res.ID,
			RegistrationDate: timestamppb.New(res.RegistrationDate),
			City:             res.City,
		},
	}, nil
}
