// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: plantr/v1/service.proto

package plantrv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/nicjohnson145/plantr/gen/plantr/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// ControllerServiceName is the fully-qualified name of the ControllerService service.
	ControllerServiceName = "plantr.v1.ControllerService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ControllerServiceLoginProcedure is the fully-qualified name of the ControllerService's Login RPC.
	ControllerServiceLoginProcedure = "/plantr.v1.ControllerService/Login"
	// ControllerServiceGetSyncDataProcedure is the fully-qualified name of the ControllerService's
	// GetSyncData RPC.
	ControllerServiceGetSyncDataProcedure = "/plantr.v1.ControllerService/GetSyncData"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	controllerServiceServiceDescriptor           = v1.File_plantr_v1_service_proto.Services().ByName("ControllerService")
	controllerServiceLoginMethodDescriptor       = controllerServiceServiceDescriptor.Methods().ByName("Login")
	controllerServiceGetSyncDataMethodDescriptor = controllerServiceServiceDescriptor.Methods().ByName("GetSyncData")
)

// ControllerServiceClient is a client for the plantr.v1.ControllerService service.
type ControllerServiceClient interface {
	Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error)
	GetSyncData(context.Context, *connect.Request[v1.GetSyncDataRequest]) (*connect.Response[v1.GetSyncDataReponse], error)
}

// NewControllerServiceClient constructs a client for the plantr.v1.ControllerService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewControllerServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ControllerServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &controllerServiceClient{
		login: connect.NewClient[v1.LoginRequest, v1.LoginResponse](
			httpClient,
			baseURL+ControllerServiceLoginProcedure,
			connect.WithSchema(controllerServiceLoginMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getSyncData: connect.NewClient[v1.GetSyncDataRequest, v1.GetSyncDataReponse](
			httpClient,
			baseURL+ControllerServiceGetSyncDataProcedure,
			connect.WithSchema(controllerServiceGetSyncDataMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// controllerServiceClient implements ControllerServiceClient.
type controllerServiceClient struct {
	login       *connect.Client[v1.LoginRequest, v1.LoginResponse]
	getSyncData *connect.Client[v1.GetSyncDataRequest, v1.GetSyncDataReponse]
}

// Login calls plantr.v1.ControllerService.Login.
func (c *controllerServiceClient) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	return c.login.CallUnary(ctx, req)
}

// GetSyncData calls plantr.v1.ControllerService.GetSyncData.
func (c *controllerServiceClient) GetSyncData(ctx context.Context, req *connect.Request[v1.GetSyncDataRequest]) (*connect.Response[v1.GetSyncDataReponse], error) {
	return c.getSyncData.CallUnary(ctx, req)
}

// ControllerServiceHandler is an implementation of the plantr.v1.ControllerService service.
type ControllerServiceHandler interface {
	Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error)
	GetSyncData(context.Context, *connect.Request[v1.GetSyncDataRequest]) (*connect.Response[v1.GetSyncDataReponse], error)
}

// NewControllerServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewControllerServiceHandler(svc ControllerServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	controllerServiceLoginHandler := connect.NewUnaryHandler(
		ControllerServiceLoginProcedure,
		svc.Login,
		connect.WithSchema(controllerServiceLoginMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	controllerServiceGetSyncDataHandler := connect.NewUnaryHandler(
		ControllerServiceGetSyncDataProcedure,
		svc.GetSyncData,
		connect.WithSchema(controllerServiceGetSyncDataMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/plantr.v1.ControllerService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ControllerServiceLoginProcedure:
			controllerServiceLoginHandler.ServeHTTP(w, r)
		case ControllerServiceGetSyncDataProcedure:
			controllerServiceGetSyncDataHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedControllerServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedControllerServiceHandler struct{}

func (UnimplementedControllerServiceHandler) Login(context.Context, *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("plantr.v1.ControllerService.Login is not implemented"))
}

func (UnimplementedControllerServiceHandler) GetSyncData(context.Context, *connect.Request[v1.GetSyncDataRequest]) (*connect.Response[v1.GetSyncDataReponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("plantr.v1.ControllerService.GetSyncData is not implemented"))
}
