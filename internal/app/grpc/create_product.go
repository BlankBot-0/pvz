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

func (s *Service) CreateProduct(ctx context.Context, request *pvzpb.CreateProductRequest) (*pvzpb.CreateProductResponse, error) {
	res, err := s.PVZ.CreateProduct(ctx, request.Type, request.PvzId)
	if errors.Is(err, pvz.ErrPVZNotFound) {
		return nil, status.Error(codes.InvalidArgument, "no pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrNoReceptionsInProgress) {
		return nil, status.Error(codes.InvalidArgument, "no receptions in progress in pvz with id "+request.PvzId)
	} else if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error while creating pvz")
	}

	return &pvzpb.CreateProductResponse{Product: &pvzpb.Product{
		Id:          res.Id,
		DateTime:    timestamppb.New(res.DateTime),
		Type:        res.Type,
		ReceptionId: res.ReceptionId,
	}}, nil
}
