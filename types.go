package frame

import (
	"net/http"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	Get HttpMethod = iota + 1
	Post
	Head
	Put
	Delete
	Options
)

type (
	HttpResponse struct {
		Data []byte
		Code int
	}

	HttpMethod int
	HandlerFn  func(request *http.Request) (*HttpResponse, error)

	GrpcRegistrar interface {
		RegisterGrpcHandler(grpcServer grpc.ServiceRegistrar, mux *runtime.ServeMux) error
	}
)
