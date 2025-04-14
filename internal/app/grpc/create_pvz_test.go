package pvz

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/auth"
	"pvz/internal/models"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
	"testing"
	"time"
)

const DefaultCity = "Москва"

var registrationDate = "2026-01-02T15:04:05Z"

func TestCreatePVZ_UnexpectedRole(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), employeeRole)

	request := &pvzpb.CreatePVZRequest{City: DefaultCity}

	service, _, _ := newService(t)

	_, grpcErr := service.CreatePVZ(ctx, request)
	st, _ := status.FromError(grpcErr)
	if c := st.Code(); c != codes.PermissionDenied {
		t.Fatalf("got unexpected err %s, want %s", c, codes.PermissionDenied)
	}
}

func TestCreatePVZ_Ok(t *testing.T) {
	ctx := auth.SetUserRoleToCtx(context.Background(), moderatorRole)

	request := &pvzpb.CreatePVZRequest{City: DefaultCity}

	service, pvzMock, _ := newService(t)

	pvzMock.CreatePVZMock.Set(func(_ context.Context, _ string, _ *string, _ *time.Time) (pp1 *models.PVZ, err error) {
		return &models.PVZ{City: DefaultCity}, nil
	})

	response, err := service.CreatePVZ(ctx, request)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if response.GetPvz().GetCity() != DefaultCity {
		t.Fatalf("got unexpected city `%s`, expected `%s`", response.Pvz.City, DefaultCity)
	}
}

func TestCreatePVZ_Ok_WithCustomDate(t *testing.T) {
	ctx := auth.SetUserRoleToCtx(context.Background(), moderatorRole)

	service, pvzMock, _ := newService(t)

	request := &pvzpb.CreatePVZRequest{City: DefaultCity, RegistrationDate: lo.ToPtr(registrationDate)}

	pvzMock.CreatePVZMock.Set(func(_ context.Context, _ string, _ *string, _ *time.Time) (pp1 *models.PVZ, err error) {
		dateTime, _ := time.Parse(time.RFC3339, registrationDate)

		return &models.PVZ{City: DefaultCity, RegistrationDate: dateTime}, nil
	})

	response, err := service.CreatePVZ(ctx, request)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if response.GetPvz().GetCity() != DefaultCity {
		t.Fatalf("got unexpected city `%s`, expected `%s`", response.Pvz.City, DefaultCity)
	} else if dt := response.Pvz.GetRegistrationDate(); dt != registrationDate {
		t.Fatalf("got unexpected date `%s`, expected `%s`", dt, registrationDate)
	}
}

func TestCreatePVZ_Fail(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		request *pvzpb.CreatePVZRequest

		mockPVZ *models.PVZ
		mockErr error

		code codes.Code
	}{
		{
			name: "invalid date",
			request: &pvzpb.CreatePVZRequest{
				RegistrationDate: lo.ToPtr("123"),
			},
			code: codes.InvalidArgument,
		},
		{
			name: "id is occupied",
			request: &pvzpb.CreatePVZRequest{
				RegistrationDate: &registrationDate,
			},
			mockErr: pvz.ErrIDIsOccupied,
			code:    codes.FailedPrecondition,
		},
		{
			name: "unexpected city",
			request: &pvzpb.CreatePVZRequest{
				RegistrationDate: &registrationDate,
				City:             "Петуховка",
			},
			mockErr: pvz.ErrCityNotFound,
			code:    codes.FailedPrecondition,
		},
		{
			name: "an internal error",
			request: &pvzpb.CreatePVZRequest{
				RegistrationDate: &registrationDate,
			},
			mockErr: fmt.Errorf("unexpected"),
			code:    codes.Internal,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := auth.SetUserRoleToCtx(context.Background(), moderatorRole)
			service, pvzMock, _ := newService(t)

			pvzMock.CreatePVZMock.Optional().Set(func(_ context.Context, _ string, _ *string, _ *time.Time) (pp1 *models.PVZ, err error) {
				return tt.mockPVZ, tt.mockErr
			})

			_, grpcErr := service.CreatePVZ(ctx, tt.request)
			st, _ := status.FromError(grpcErr)
			if c := st.Code(); c != tt.code {
				t.Fatalf("got unexpected err %s, want %s", c, tt.code)
			}
		})
	}
}
