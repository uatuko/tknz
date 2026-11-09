package srv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/rs/zerolog"
	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/jose/jwk"
	"go.tknz.dev/internal/jose/jwt"
	"go.tknz.dev/internal/srv/auth"
	"go.tknz.dev/pb"
)

type authn struct {
	pb.UnimplementedAuthnServer
}

func (a *authn) Check(ctx context.Context, req *pb.AuthnCheckRequest) (*pb.AuthnCheckResponse, error) {
	// TODO: check if the "caller" has permissions to do this check
	resp := &pb.AuthnCheckResponse{
		Ok: false,
	}

	token := req.GetToken()
	idn, claims, err := auth.CheckAccessToken(ctx, token)
	if err == nil {
		resp.Ok = true
		resp.Idn = mapIdnToPb(idn)
		resp.Claims = &pb.Claims{
			Aud: claims.Aud,
			Iss: claims.Iss,
		}

		return resp, nil
	}

	if errors.Is(err, auth.ErrInvalidToken) {
		return resp, nil
	}

	if !errors.Is(err, auth.ErrInvalidTokenPrefix) {
		return nil, err
	}

	appId := req.GetAppId()
	if appId == "" {
		// No app id, can't validate federated tokens
		return resp, nil
	}

	idn, jwtClaims, err := checkJwt(ctx, appId, token)
	if err != nil {
		if errors.Is(err, errInvalidJwt) {
			return resp, nil
		}

		return nil, err
	}

	resp.Ok = true
	resp.Idn = mapIdnToPb(idn)
	resp.Claims = newTokenClaims(jwtClaims.Aud(), jwtClaims.Iss())

	return resp, nil
}

func accessTokenFromCtx(ctx context.Context) (string, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		zerolog.Ctx(ctx).Error().Msg("missing metadata in incoming context")
		return "", ErrInternal
	}

	var token string
	if v := meta.Get(metaAuthorizationKey); len(v) > 0 {
		token = v[0]
	}

	if token == "" {
		return "", NewErrorf(ErrUnauthenticated, "missing credentials")
	}

	if !strings.HasPrefix(token, bearerTokenPrefix) {
		return "", NewErrorf(ErrUnauthenticated, "invalid or unsupported credentials")
	}

	token = token[7:] // strip 'Bearer ' prefix
	return token, nil
}

func checkJwt(ctx context.Context, appId string, s string) (*db.Idn, *jwt.Claims, error) {
	log := zerolog.Ctx(ctx)
	token, err := jwt.Validate(ctx, s, mkJwkFunc(appId))
	if err != nil {
		log.Debug().Err(err).Msg("failed to validate jwt")
		return nil, nil, errInvalidJwt
	}

	claims := token.Claims()
	idn, err := db.RetrieveIdnByFederatedLogin(ctx, appId, claims.Sub(), claims.Iss(), claims.Aud())
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, nil, errInvalidJwt
		}

		return nil, nil, err
	}

	return idn, claims, nil
}

func mapIdnToPb(idn *db.Idn) *pb.Idn {
	pbIdn := &pb.Idn{
		Id:      idn.Id(),
		SpaceId: idn.SpaceId(),
		Status:  pb.IdnStatus(idn.Status()),
		Login:   idn.Login(),
	}

	email := idn.Email()
	if email != "" {
		pbIdn.Email = &email
	}

	name := idn.Name()
	if name != "" {
		pbIdn.Name = &name
	}

	picture := idn.Picture()
	if picture != "" {
		pbIdn.Picture = &picture
	}

	return pbIdn
}

func mkJwkFunc(appId string) jwt.JwkFunc {
	return func(ctx context.Context, kid string, iss string) (*jwk.Jwk, error) {
		f, err := db.RetrieveFederationByIss(ctx, appId, iss)
		if err != nil {
			return nil, err
		}

		jwks, err := jwk.Fetch(ctx, f.JwksUri())
		if err != nil {
			return nil, err
		}

		have := make([]string, 0, len(jwks))
		for idx := range jwks {
			have = append(have, jwks[idx].Kid)

			if jwks[idx].Kid == kid {
				return &jwks[idx], nil
			}
		}

		return nil, fmt.Errorf("jwk not found, want: %s, have: %v", kid, have)
	}
}

func newTokenClaims(aud string, iss string) *pb.Claims {
	claims := &pb.Claims{}
	if aud != "" {
		claims.Aud = &aud
	}

	if iss != "" {
		claims.Iss = &iss
	}

	return claims
}
