package mw

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"pvz/internal/auth"
	pvzpb "pvz/pkg/api/v1"
	"strings"
)

type tokenVerifier interface {
	Verify(token string) (auth.Claims, error)
}

func extractToken(ctx context.Context) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AuthInterceptor/extractToken")
	defer span.Finish()

	values := metadata.ValueFromIncomingContext(ctx, "Authorization")
	if len(values) < 1 {
		return "", fmt.Errorf("missing authorization header")
	}
	token := values[0]

	parts := strings.Split(token, " ")
	if len(parts) < 2 {
		return "", fmt.Errorf("incorrect authorization header value")
	}

	if strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("incorrect authorization token type, expected bearer, got: %s", parts[0])
	}

	return parts[1], nil
}

func AuthInterceptor(verifier tokenVerifier) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "AuthInterceptor")
		defer span.Finish()

		if _, ok := handlerWhitelistAuth[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		token, err := extractToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		claims, err := verifier.Verify(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return handler(auth.SetUserRoleToCtx(ctx, claims.UserRole()), req)
	}
}

var handlerWhitelistAuth = map[string]struct{}{
	pvzpb.PVZService_DummyLogin_FullMethodName: {},
	pvzpb.PVZService_Register_FullMethodName:   {},
	pvzpb.PVZService_Login_FullMethodName:      {},
	pvzpb.PVZService_GetPVZList_FullMethodName: {},
}
