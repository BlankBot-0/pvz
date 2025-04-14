package pvz

import (
	"pvz/internal/app/grpc/mocks"
	"testing"
)

var dateFormat = "2006-01-02T15:04:05Z"

func newService(t *testing.T) (*Service, *mocks.PVZMock, *mocks.AuthMock) {
	t.Parallel()
	pvzMock := mocks.NewPVZMock(t)
	authMock := mocks.NewAuthMock(t)

	return NewService("2006-01-02T15:04:05Z", Deps{
		PVZ:  pvzMock,
		Auth: authMock,
	}), pvzMock, authMock
}
