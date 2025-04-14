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

func TestCreateProduct_UnexpectedRole(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), moderatorRole)

	service, _, _ := newService(t)

	_, grpcErr := service.CreateProduct(ctx, &pvzpb.CreateProductRequest{})
	st, _ := status.FromError(grpcErr)
	if c := st.Code(); c != codes.PermissionDenied {
		t.Fatalf("got unexpected err %s, want %s", c, codes.PermissionDenied)
	}
}

func TestCreateProduct_Ok(t *testing.T) {
	t.Parallel()

	ctx := auth.SetUserRoleToCtx(context.Background(), employeeRole)

	const productType = "голубцы"

	request := &pvzpb.CreateProductRequest{
		Type:  productType,
		PvzId: "1",
	}

	product := models.Product{
		ID:          "1",
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: "1",
	}

	service, pvzMock, _ := newService(t)

	pvzMock.CreateProductMock.Set(func(_ context.Context, _ string, _ string) (_ models.Product, _ error) {
		return product, nil
	})

	response, err := service.CreateProduct(ctx, request)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if prType := response.GetProduct().GetType(); prType != product.Type {
		t.Fatalf("got unexpected productType type `%s`, expected `%s`", prType, product.Type)
	} else if dtTime := response.GetProduct().GetDateTime().AsTime().UTC(); dtTime != product.DateTime.UTC() {
		t.Fatalf("got unexpected dateTime `%s`, expected `%s`", dtTime, product.DateTime.UTC())
	}
}

func TestCreateProduct_Fail(t *testing.T) {
	t.Parallel()

	request := &pvzpb.CreateProductRequest{
		Type:  "x",
		PvzId: "1",
	}

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
			name:    "got unexpected product type",
			mockErr: pvz.ErrProductTypeNotFound,
			code:    codes.InvalidArgument,
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

			pvzMock.CreateProductMock.Optional().Set(func(_ context.Context, _ string, _ string) (_ models.Product, _ error) {
				return models.Product{}, tt.mockErr
			})

			_, grpcErr := service.CreateProduct(ctx, request)
			st, _ := status.FromError(grpcErr)
			if c := st.Code(); c != tt.code {
				t.Fatalf("got unexpected err %s, want %s", c, tt.code)
			}
		})
	}
}
