package pvz

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/auth"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
)

const deleteLastProductMessage = "product is deleted from last opened reception"

func (s *Service) DeleteLastProduct(ctx context.Context, request *pvzpb.DeleteLastProductRequest) (*pvzpb.DeleteLastProductResponse, error) {
	if auth.GetUserRoleFromCtx(ctx) != employeeRole {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	err := s.PVZ.DeleteLastProduct(ctx, request.PvzId)
	if errors.Is(err, pvz.ErrPVZNotFound) {
		return nil, status.Error(codes.NotFound, "no pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrNoReceptionsInProgress) {
		return nil, status.Error(codes.FailedPrecondition, "no receptions in progress in pvz with id "+request.PvzId)
	} else if errors.Is(err, pvz.ErrNoProducts) {
		return nil, status.Error(codes.FailedPrecondition, "no products in last reception in pvz with id "+request.PvzId)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error while deleting last product: %s", err.Error())
	}

	return &pvzpb.DeleteLastProductResponse{Message: deleteLastProductMessage}, nil
}
