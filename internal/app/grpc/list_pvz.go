package merch_store

import (
	"context"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) ListPVZ(ctx context.Context, request *pvzpb.ListPVZRequest) (*pvzpb.ListPVZResponse, error) {
	res, err := s.PVZ.ListPVZPaginated(
		ctx,
		request.StartDate.AsTime(), request.EndDate.AsTime(),
		request.Page, request.Limit,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "Unexpected error while listing pvz")
	}

	list := &pvzpb.ListPVZResponse{
		Pvzs: lo.Map(res, func(pvzInfo models.PVZInfo, _ int) *pvzpb.ListPVZResponsePvzInfo {
			return &pvzpb.ListPVZResponsePvzInfo{
				Pvz: &pvzpb.PVZ{
					Id:               &pvzInfo.PVZ.ID,
					RegistrationDate: timestamppb.New(pvzInfo.PVZ.RegistrationDate),
					City:             pvzInfo.PVZ.City,
				},
			}
		}),
	}

	for i, pvzInfo := range res {
		list.Pvzs[i].Receptions = lo.Map(pvzInfo.Receptions, func(reception models.ReceptionInfo, _ int) *pvzpb.ListPVZResponseReceptionInfo {
			return &pvzpb.ListPVZResponseReceptionInfo{
				Reception: &pvzpb.Reception{
					Id:       reception.Reception.ID,
					DateTime: timestamppb.New(reception.Reception.DateTime),
					PvzId:    reception.Reception.PvzID,
					Status:   reception.Reception.ReceptionStatus,
				},
			}
		})
	}

	for i, pvzInfo := range res {
		for j, reception := range pvzInfo.Receptions {
			list.Pvzs[i].Receptions[j].Products = lo.Map(reception.Products, func(product models.Product, _ int) *pvzpb.Product {

				return &pvzpb.Product{
					Id:          product.ID,
					DateTime:    timestamppb.New(product.DateTime),
					Type:        product.Type,
					ReceptionId: product.ReceptionID,
				}
			})
		}
	}

	return list, nil
}
