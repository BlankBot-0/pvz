package merch_store

import (
	"context"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
)

func (s *Service) ListPVZ(ctx context.Context, request *pvzpb.ListPVZRequest) (*pvzpb.ListPVZResponse, error) {
	res, err := s.PVZ.ListPVZ(
		ctx,
		request.StartDate.AsTime(), request.EndDate.AsTime(),
		request.Page, request.Limit,
	)
	if err != nil {
	}

	return &pvzpb.ListPVZResponse{
		Pvzs: lo.Map(res, func(pvzInfo models.PVZInfo, _ int) *pvzpb.ListPVZResponsePvzInfo {

			return &pvzpb.ListPVZResponsePvzInfo{
				Pvz: &pvzpb.PVZ{
					Id:               pvzInfo.PVZ.Id,
					RegistrationDate: timestamppb.New(pvzInfo.PVZ.RegistrationDate),
					City:             pvzInfo.PVZ.City,
				},

				Receptions: lo.Map(pvzInfo.Receptions, func(reception models.ReceptionInfo, _ int) *pvzpb.ListPVZResponseReceptionInfo {

					return &pvzpb.ListPVZResponseReceptionInfo{
						Reception: &pvzpb.Reception{
							Id:       reception.Reception.Id,
							DateTime: timestamppb.New(reception.Reception.DateTime),
							PvzId:    reception.Reception.PvzId,
							Status:   reception.Reception.ReceptionStatus,
						},
						Products: lo.Map(reception.Products, func(product models.Product, _ int) *pvzpb.Product {

							return &pvzpb.Product{
								Id:          product.Id,
								DateTime:    timestamppb.New(product.DateTime),
								Type:        product.Type,
								ReceptionId: product.ReceptionId,
							}
						}),
					}

				}),
			}
		}),
	}, nil
}
