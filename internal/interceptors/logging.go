package interceptors

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type LoggingInterceptorConfig struct {
	LogRequests  bool
	LogResponses bool
}

func marshallDebug(x any) {
	var outBytes []byte
	var err error
	if protoMsg, ok := x.(protoreflect.ProtoMessage); ok {
		opts := protojson.MarshalOptions{
			Indent: "    ",
		}
		outBytes, err = opts.Marshal(protoMsg)
	} else {
		outBytes, err = json.MarshalIndent(x, "", "   ")
	}
	if err != nil {
		fmt.Printf("Unable to marshall object for debugging: %v\n", err)
		panic("unable to marshall")
	}
	fmt.Println(string(outBytes))
}

func NewLoggingInterceptor(logger zerolog.Logger, conf LoggingInterceptorConfig) connect.UnaryInterceptorFunc {
	if conf.LogRequests {
		logger.Warn().Msg("request logging is on, this could potentially leak sensitive information and is only intended for debugging purposes")
	}
	if conf.LogResponses {
		logger.Warn().Msg("response logging is on, this could potentially leak sensitive information and is only intended for debugging purposes")
	}
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if conf.LogRequests {
				marshallDebug(req.Any())
			}

			logger.Info().Msgf("request recieved for %v", req.Spec().Procedure)

			resp, err := next(ctx, req)

			if conf.LogResponses && err == nil {
				marshallDebug(resp.Any())
			}

			return resp, err
		})
	})
}
