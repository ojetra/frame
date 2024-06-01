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
		Headers Headers
		Data    []byte
		Code    int
	}

	HttpMethod int
	HandlerFn  func(request *http.Request) (*HttpResponse, error)

	GrpcRegistrar interface {
		RegisterGrpcHandler(grpcServer grpc.ServiceRegistrar, mux *runtime.ServeMux) error
	}
)

type Headers struct {
	setHeaderEntryMap   map[string]string
	addHeaderEntrySlice []addHeaderEntry
}

type addHeaderEntry struct {
	name  string
	value string
}

func (x *Headers) Set(name, value string) {
	if x.setHeaderEntryMap == nil {
		x.setHeaderEntryMap = make(map[string]string)
	}

	x.setHeaderEntryMap[name] = value
}

func (x *Headers) Add(name, value string) {
	x.addHeaderEntrySlice = append(
		x.addHeaderEntrySlice,
		addHeaderEntry{
			name:  name,
			value: value,
		},
	)
}

func (x *Headers) GetSetEntryMap() map[string]string {
	return x.setHeaderEntryMap
}

func (x *Headers) GetAddEntrySlice() []addHeaderEntry {
	return x.addHeaderEntrySlice
}
