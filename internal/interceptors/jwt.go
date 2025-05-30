package interceptors

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/nicjohnson145/hlp/set"
	"github.com/nicjohnson145/plantr/internal/token"
	"github.com/rs/zerolog"
)

var (
	ErrNoClaimsError         = errors.New("no claims in context")
	ErrCannotCastClaimsError = errors.New("context value not *token.Token")
)

type claimsKey struct{}

const (
	tokenHeader = "authorization"
)

func NewAuthInterceptor(logger zerolog.Logger, signingKey []byte, excludedMethods *set.Set[string]) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if excludedMethods.Contains(req.Spec().Procedure) {
				return next(ctx, req)
			}

			// Check the header exists
			tokenStr := req.Header().Get(tokenHeader)
			if tokenStr == "" {
				logger.Error().Msg("no token provided")
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("no token provided"))
			}

			// Parse out the token
			token, err := token.ParseJWT(tokenStr, signingKey)
			if err != nil {
				logger.Err(err).Msg("error parsing token")
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}

			// Add the parsed token to the context
			return next(context.WithValue(ctx, claimsKey{}, token), req)
		})
	})
}

func NewClientAuthInterceptor(token string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set(tokenHeader, token)
			return next(ctx, req)
		})
	})
}

func ClaimsFromCtx(ctx context.Context) (*token.Token, error) {
	val := ctx.Value(claimsKey{})
	if val == nil {
		return nil, ErrNoClaimsError
	}

	tok, ok := val.(*token.Token)
	if !ok {
		return nil, ErrCannotCastClaimsError
	}

	return tok, nil
}

func SetTokenOnContext(ctx context.Context, token *token.Token) context.Context {
	return context.WithValue(ctx, claimsKey{}, token)
}
