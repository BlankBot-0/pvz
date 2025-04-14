package pvz

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/auth"
	"pvz/internal/usecase/pvz"
	pvzpb "pvz/pkg/api/v1"
	"testing"
)

func TestDeleteLastProduct_UnexpectedRole(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), moderatorRole)

	service, _, _ := newService(t)

	_, grpcErr := service.DeleteLastProduct(ctx, &pvzpb.DeleteLastProductRequest{})
	st, _ := status.FromError(grpcErr)
	if c := st.Code(); c != codes.PermissionDenied {
		t.Fatalf("got unexpected err %s, want %s", c, codes.PermissionDenied)
	}
}

func TestDeleteLastProduct_Ok(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), employeeRole)

	request := &pvzpb.DeleteLastProductRequest{PvzId: "1"}

	service, pvzMock, _ := newService(t)

	pvzMock.DeleteLastProductMock.Set(func(_ context.Context, _ string) (_ error) {
		return nil
	})

	response, err := service.DeleteLastProduct(ctx, request)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if response.Message != deleteLastProductMessage {
		t.Fatalf("got unexpected message: %s", response.Message)
	}
}

func TestDeleteLastProduct_Fail(t *testing.T) {
	t.Parallel()

	request := &pvzpb.DeleteLastProductRequest{PvzId: "1"}

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
			code:    codes.FailedPrecondition,
		},
		{
			name:    "no products",
			mockErr: pvz.ErrNoProducts,
			code:    codes.FailedPrecondition,
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

			pvzMock.DeleteLastProductMock.Set(func(_ context.Context, _ string) (_ error) {
				return tt.mockErr
			})

			_, grpcErr := service.DeleteLastProduct(ctx, request)
			st, _ := status.FromError(grpcErr)
			if c := st.Code(); c != tt.code {
				t.Fatalf("got unexpected err %s, want %s", c, tt.code)
			}
		})
	}
}
