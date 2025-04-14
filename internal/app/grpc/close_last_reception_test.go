package pvz

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/auth"
	"pvz/internal/models"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
	"testing"
	"time"
)

func TestCloseLastReception_UnexpectedRole(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), moderatorRole)

	request := &pvzpb.CloseLastReceptionRequest{PvzId: "1"}

	service, _, _ := newService(t)

	_, grpcErr := service.CloseLastReception(ctx, request)
	st, _ := status.FromError(grpcErr)
	if c := st.Code(); c != codes.PermissionDenied {
		t.Fatalf("got unexpected err %s, want %s", c, codes.PermissionDenied)
	}
}

func TestCloseLastReception_Ok(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), employeeRole)

	request := &pvzpb.CloseLastReceptionRequest{PvzId: "1"}

	service, pvzMock, _ := newService(t)

	reception := models.Reception{
		ID:              "10",
		DateTime:        time.Now(),
		PvzID:           "1",
		ReceptionStatus: "closed",
	}

	pvzMock.CloseLastReceptionMock.Set(func(_ context.Context, _ string) (_ models.Reception, _ error) {
		return reception, nil
	})

	response, err := service.CloseLastReception(ctx, request)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if id := response.GetReception().GetId(); id != reception.ID {
		t.Fatalf("got unexpected id `%s`, expected `%s`", id, reception.ID)
	} else if dt := response.GetReception().GetDateTime().AsTime().UTC(); dt != reception.DateTime.UTC() {
		t.Fatalf("got unexpected dateTime `%s`, expected `%s`", dt, reception.DateTime.UTC())
	} else if pvzID := response.GetReception().GetPvzId(); pvzID != reception.PvzID {
		t.Fatalf("got unexpected pvz id `%s`, expected `%s`", pvzID, reception.PvzID)
	} else if st := response.GetReception().GetStatus(); st != reception.ReceptionStatus {
		t.Fatalf("got unexpected status `%s`, expected `%s`", st, reception.ReceptionStatus)
	}
}

func TestCloseLastReception_Fail(t *testing.T) {
	t.Parallel()

	request := &pvzpb.CloseLastReceptionRequest{PvzId: "1"}

	testcases := []struct {
		name    string
		mockErr error
		code    codes.Code
	}{
		{
			name:    "pvz not found",
			mockErr: pvz.ErrPVZNotFound,
			code:    codes.NotFound,
		},
		{
			name:    "no reception in progress",
			mockErr: pvz.ErrNoReceptionsInProgress,
			code:    codes.NotFound,
		},
		{
			name:    "unexpected error",
			mockErr: fmt.Errorf("unexpected"),
			code:    codes.Internal,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := auth.SetUserRoleToCtx(context.Background(), employeeRole)
			service, pvzMock, _ := newService(t)

			pvzMock.CloseLastReceptionMock.Set(func(_ context.Context, _ string) (_ models.Reception, _ error) {
				return models.Reception{}, tt.mockErr
			})

			_, grpcErr := service.CloseLastReception(ctx, request)
			st, _ := status.FromError(grpcErr)
			if c := st.Code(); c != tt.code {
				t.Fatalf("got unexpected err %s, want %s", c, tt.code)
			}
		})
	}
}
