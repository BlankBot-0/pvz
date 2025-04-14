package pvz

import (
	"context"
	"errors"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/auth"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
	"time"
)

func (s *Service) CreatePVZ(ctx context.Context, req *pvzpb.CreatePVZRequest) (*pvzpb.CreatePVZResponse, error) {
	if auth.GetUserRoleFromCtx(ctx) != moderatorRole {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	var customRegistrationDate *time.Time
	if req.RegistrationDate != nil {
		parsedDatetime, err := time.Parse(s.datetimeFormat, *req.RegistrationDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid registration date format: %s", err)
		}
		customRegistrationDate = &parsedDatetime
	}

	res, err := s.PVZ.CreatePVZ(ctx, req.City, req.Id, customRegistrationDate)
	if errors.Is(err, pvz.ErrIDIsOccupied) {
		return nil, status.Error(codes.FailedPrecondition, "pvz with such id already exists")
	} else if errors.Is(err, pvz.ErrCityNotFound) {
		return nil, status.Errorf(codes.FailedPrecondition, "city '%s' is not suppoted", req.City)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while creating pvz: %s", err.Error())
	}

	observeCreatedPVZ(req.City)
	return &pvzpb.CreatePVZResponse{
		Pvz: &pvzpb.PVZ{
			Id:               &res.ID,
			RegistrationDate: lo.ToPtr(res.RegistrationDate.Format(s.datetimeFormat)),
			City:             res.City,
		},
	}, nil
}
