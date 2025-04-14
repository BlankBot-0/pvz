package pvz

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pvz/internal/usecase/auth"
	pvzpb "pvz/pkg/api/v1"
	"testing"
)

func TestDummyLogin_Ok(t *testing.T) {
	t.Parallel()

	service, _, authMock := newService(t)

	authMock.DummyLoginMock.Set(func(_ context.Context, _ string) (_ string, _ error) {
		return "token", nil
	})

	response, err := service.DummyLogin(context.Background(), &pvzpb.DummyLoginRequest{Role: "mocherator"})
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if response.Token != "token" {
		t.Fatalf("got unexpected token: %s", response.Token)
	}
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	t.Parallel()

	service, _, authMock := newService(t)

	authMock.DummyLoginMock.Set(func(_ context.Context, _ string) (_ string, _ error) {
		return "", auth.ErrInvalidRole
	})

	_, grpcErr := service.DummyLogin(context.Background(), &pvzpb.DummyLoginRequest{Role: "crocodile"})
	st, _ := status.FromError(grpcErr)
	if c := st.Code(); c != codes.InvalidArgument {
		t.Fatalf("got unexpected err %s, want %s", c, codes.InvalidArgument)
	}
}

func TestDummyLogin_UnexpectedError(t *testing.T) {
	t.Parallel()

	service, _, authMock := newService(t)

	authMock.DummyLoginMock.Set(func(_ context.Context, _ string) (_ string, _ error) {
		return "", fmt.Errorf("world is collapsed")
	})

	_, grpcErr := service.DummyLogin(context.Background(), &pvzpb.DummyLoginRequest{})
	st, _ := status.FromError(grpcErr)
	if c := st.Code(); c != codes.Internal {
		t.Fatalf("got unexpected err %s, want %s", c, codes.Internal)
	}
}
