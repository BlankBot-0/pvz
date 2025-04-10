package merch_store

import (
	"context"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) DeleteLastProduct(ctx context.Context, request *pvzpb.DeleteLastProductRequest) (*pvzpb.DeleteLastProductResponse, error) {
	res, err := s.PVZ.DeleteLastProduct(ctx, request.PvzId)
	if err != nil {
	}

	return &pvzpb.DeleteLastProductResponse{Message: res}, nil
}
