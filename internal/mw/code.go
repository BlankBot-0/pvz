package mw

import (
	"context"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/proto"
)

const HttpCodeHeader = "X-Http-Code"

func ForwardResponseFunc(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	smd, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}
	if vals := smd.HeaderMD.Get(HttpCodeHeader); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}
		w.WriteHeader(code)
	}
	return nil
}
