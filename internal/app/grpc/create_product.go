package pvz

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"pvz/internal/auth"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) CreateProduct(ctx context.Context, request *pvzpb.CreateProductRequest) (*pvzpb.CreateProductResponse, error) {
	if auth.GetUserRoleFromCtx(ctx) != employeeRole {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	res, err := s.PVZ.CreateProduct(ctx, request.Type, request.PvzId)
	if errors.Is(err, pvz.ErrPVZNotFound) {
		return nil, status.Error(codes.NotFound, "no pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrNoReceptionsInProgress) {
		return nil, status.Error(codes.FailedPrecondition, "no receptions in progress in pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrProductTypeNotFound) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product type %s", request.Type)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while creating pvz: %s", err.Error())
	}

	SetHTTPCode(ctx, http.StatusCreated)
	observeCreatedProducts(request.Type)
	return &pvzpb.CreateProductResponse{Product: &pvzpb.Product{
		Id:          res.ID,
		DateTime:    timestamppb.New(res.DateTime),
		Type:        res.Type,
		ReceptionId: res.ReceptionID,
	}}, nil
}
