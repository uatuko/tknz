package srv

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.tknz.dev/internal/srv/auth"
	"go.tknz.dev/internal/srv/common"
	"go.tknz.dev/internal/srv/oidc"
	"go.tknz.dev/pb"
)

const (
	authSchemeBearer  = "Bearer"
	bearerTokenPrefix = authSchemeBearer + " "

	metaAuthorizationKey = "authorization"
)

func New() http.Handler {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			unaryLoggerInterceptor,
			unaryErrorInterceptor,
		),
	)
	pb.RegisterAuthnServer(grpcServer, &authn{})
	pb.RegisterSpacesServer(grpcServer, &spaces{})

	mux := http.NewServeMux()
	mux.Handle(common.AuthPathPattern(), http.StripPrefix(common.AuthPathPrefix(), auth.New()))
	mux.Handle(common.OidcPathPattern(), http.StripPrefix(common.OidcPathPrefix(), oidc.New()))
	mux.HandleFunc("/", filesHandler)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		defer func() {
			log.Debug().
				Str("protocol", r.Proto).
				Str("method", r.Method).
				Str("uri", r.URL.RequestURI()).
				Str("user_agent", r.UserAgent()).
				Str("referrer", r.Referer()).
				Uint32("status", rw.Status()).
				Uint64("response_size", rw.Size()).
				Dur("latency", time.Since(start)).
				Msg("http request")
		}()

		r = r.WithContext(log.Logger.WithContext(r.Context()))

		// Serve gRPC
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(rw, r)
			return
		}

		// Serve http
		mux.ServeHTTP(rw, r)
	})

	return handler
}

func unaryErrorInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	log := zerolog.Ctx(ctx)
	log.Debug().Err(err).Msg("")

	var c errorCode
	if errors.As(err, &c) {
		return resp, NewError(c, nil)
	}

	var e *Error
	if errors.As(err, &e) {
		return resp, err
	}

	if status.Convert(err).Code() == codes.Unimplemented {
		return resp, err
	}

	log.Warn().Err(err).Msg("unknown error")
	return resp, NewError(ErrUnknown, err)
}

func unaryLoggerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	log := zerolog.Ctx(ctx)
	start := time.Now()
	resp, err := handler(ctx, req)
	latency := time.Since(start)

	log.Debug().
		Dur("latency", latency).
		Str("rpc", info.FullMethod).
		Str("status", status.Code(err).String()).
		Msg("grpc request")

	return resp, err
}
