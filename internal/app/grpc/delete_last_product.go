package merch_store

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) DeleteLastProduct(ctx context.Context, request *pvzpb.DeleteLastProductRequest) (*pvzpb.DeleteLastProductResponse, error) {
	res, err := s.PVZ.DeleteLastProduct(ctx, request.PvzId)
	if errors.Is(err, pvz.ErrPVZNotFound) {
		return nil, status.Error(codes.InvalidArgument, "no pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrNoReceptionsInProgress) {
		return nil, status.Error(codes.FailedPrecondition, "no receptions in progress in pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrNoProducts) {
		return nil, status.Error(codes.FailedPrecondition, "no products in last reception in pvz with id "+request.PvzId)
	} else if err != nil {
		return nil, status.Error(codes.Internal, "unexpected error while deleting last product")
	}

	return &pvzpb.DeleteLastProductResponse{Message: res}, nil
}
