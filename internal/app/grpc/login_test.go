package pvz

import (
	"context"
	pvzpb "pvz/pkg/api/v1"
	"testing"
)

func TestLogin_Ok(t *testing.T) {
	t.Parallel()

	request := &pvzpb.LoginRequest{
		Email:    "aboba@mail.ru",
		Password: "12345678",
	}

	service, _, authMock := newService(t)

	authMock.UserTokenMock.Set(func(_ context.Context, _ string, _ string) (_ string, _ error) {
		return "token", nil
	})

	response, err := service.Login(context.Background(), request)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	} else if token := response.Token; token != "token" {
		t.Fatalf("got unexpected token `%s`, expected `%s`", token, "token")
	}
}
