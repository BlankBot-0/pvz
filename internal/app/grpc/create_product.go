package merch_store

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) CreateProduct(ctx context.Context, request *pvzpb.CreateProductRequest) (*pvzpb.CreateProductResponse, error) {
	res, err := s.PVZ.CreateProduct(ctx, request.Type, request.PvzId)
	if err != nil {
	}

	return &pvzpb.CreateProductResponse{Product: &pvzpb.Product{
		Id:          res.Id,
		DateTime:    timestamppb.New(res.DateTime),
		Type:        res.Type,
		ReceptionId: res.ReceptionId,
	}}, nil
}
