package pvz

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/auth"
	"pvz/internal/models"
	pvzpb "pvz/pkg/api/v1"
	"testing"
	"time"
)

func TestListPVZ_UnexpectedRole(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), "user")

	request := &pvzpb.ListPVZRequest{}

	service, _, _ := newService(t)

	_, grpcErr := service.ListPVZ(ctx, request)
	st, _ := status.FromError(grpcErr)
	if c := st.Code(); c != codes.PermissionDenied {
		t.Fatalf("got unexpected err %s, want %s", c, codes.PermissionDenied)
	}
}
func TestListPVZ_Ok(t *testing.T) {
	ctx := auth.SetUserRoleToCtx(context.Background(), moderatorRole)

	request := &pvzpb.ListPVZRequest{
		StartDate: time.Now().Add(-time.Hour * 4).Format(dateFormat),
		EndDate:   time.Now().Format(dateFormat),
		Limit:     10,
	}

	service, pvzMock, _ := newService(t)

	oneHourAgo := time.Now().Add(-time.Hour)
	twoHourAgo := time.Now().Add(-time.Hour * 2)
	threeHourAgo := time.Now().Add(-time.Hour * 3)

	pvzs := []models.PVZInfo{
		{
			PVZ: models.PVZ{
				ID:               "1",
				RegistrationDate: oneHourAgo,
				City:             DefaultCity,
			},
			Receptions: []models.ReceptionInfo{
				{
					Reception: models.Reception{
						ID:              "1",
						DateTime:        oneHourAgo,
						PvzID:           "1",
						ReceptionStatus: "closed",
					},
					Products: []models.Product{
						{
							ID:          "1",
							DateTime:    oneHourAgo,
							Type:        "голубцы",
							ReceptionID: "1",
						},
					},
				},
			},
		},
		{
			PVZ: models.PVZ{
				ID:               "2",
				RegistrationDate: twoHourAgo,
				City:             "Санкт-Петербург",
			},
			Receptions: []models.ReceptionInfo{
				{
					Reception: models.Reception{
						ID:              "2",
						DateTime:        twoHourAgo,
						PvzID:           "2",
						ReceptionStatus: "closed",
					},
					Products: []models.Product{
						{
							ID:          "2",
							DateTime:    twoHourAgo,
							Type:        "голубцы2",
							ReceptionID: "2",
						},
					},
				},
			},
		},
		{
			PVZ: models.PVZ{
				ID:               "3",
				RegistrationDate: threeHourAgo,
				City:             "Казань",
			},
			Receptions: []models.ReceptionInfo{
				{
					Reception: models.Reception{
						ID:              "3",
						DateTime:        threeHourAgo,
						PvzID:           "3",
						ReceptionStatus: "in_progress",
					},
					Products: []models.Product{
						{
							ID:          "3",
							DateTime:    threeHourAgo,
							Type:        "голубцы3",
							ReceptionID: "3",
						},
					},
				},
			},
		},
	}

	pvzMock.ListPVZPaginatedMock.Set(func(_ context.Context, _ time.Time, _ time.Time, _ uint32, _ uint32) (_ []models.PVZInfo, _ error) {
		return pvzs, nil
	})

	response, err := service.ListPVZ(ctx, request)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	for i, pvzInfo := range response.Pvzs {
		pvz := pvzs[i]
		if pvzInfo.Pvz.GetId() != pvz.PVZ.ID {
			t.Fatalf("got unexpected id `%s`, want `%s`", pvzInfo.Pvz.GetId(), pvz.PVZ.ID)
		} else if pvzInfo.Pvz.City != pvz.PVZ.City {
			t.Fatalf("got unexpected city `%s`, want `%s`", pvzInfo.Pvz.City, pvz.PVZ.City)
		}
	}
}
