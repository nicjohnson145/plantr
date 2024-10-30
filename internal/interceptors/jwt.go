package interceptors

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/nicjohnson145/hlp/set"
	"github.com/nicjohnson145/plantr/internal/token"
)

const tokenHeader = "authorization"

func NewAuthInterceptor(signingKey []byte, excludedMethods *set.Set[string]) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if excludedMethods.Contains(req.Spec().Procedure) {
				return next(ctx, req)
			}

			tokenStr := req.Header().Get(tokenHeader)
			if tokenStr == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("no token provided"))
			}
			if _, err := token.ParseJWT(tokenStr, signingKey); err != nil {
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}

			return next(ctx, req)
		})
	})
}
