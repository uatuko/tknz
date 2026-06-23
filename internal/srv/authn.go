package srv

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/rs/zerolog"
	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/jose/jwk"
	"go.tknz.dev/internal/jose/jwt"
	pbi "go.tknz.dev/internal/pb"
	"go.tknz.dev/internal/srv/auth"
	"go.tknz.dev/internal/srv/common"
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
	idn, err := checkAccessToken(ctx, token)
	if err != nil {
		if errors.Is(err, errInvalidAccessToken) {
			return resp, nil
		}

		if errors.Is(err, errInvalidAccessTokenPrefix) {
			appId := req.GetAppId()
			if appId == "" {
				// No app id, can't validate federated tokens
				return resp, nil
			}

			idn, err = checkJwt(ctx, appId, token)
			if err != nil {
				if errors.Is(err, errInvalidJwt) {
					return resp, nil
				}

				return nil, err
			}
		} else {
			return nil, err
		}
	}

	resp.Ok = true
	resp.Idn = mapIdnToPb(idn)

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

func checkAccessToken(ctx context.Context, token string) (*db.Idn, error) {
	if !strings.HasPrefix(token, auth.AccessTokenPrefix) {
		return nil, errInvalidAccessTokenPrefix
	}

	b, err := base64.RawURLEncoding.DecodeString(token[len(auth.AccessTokenPrefix):])
	if err != nil {
		fmt.Printf("[error] failed to decode, err: %v\n", err)
		return nil, errInvalidAccessToken
	}

	var pbToken pbi.Token
	err = proto.Unmarshal(b, &pbToken)
	if err != nil {
		fmt.Printf("[error] failed to unmarshal, err: %v\n", err)
		return nil, errInvalidAccessToken
	}

	jwk, err := db.RetrieveJwk(ctx, common.SysSpaceId, pbToken.GetKid())
	if err != nil {
		fmt.Printf("[error] failed to retrieve jwk, err: %v\n", err)
		return nil, errInvalidAccessToken
	}

	digest := sha256.Sum256(pbToken.GetToken())
	sig := pbToken.GetSignature()
	var r, s big.Int
	r.SetBytes(sig[:32])
	s.SetBytes(sig[32:])

	params := jwk.Params()
	key, err := params.EcPublicKey()
	if err != nil {
		fmt.Printf("[error] failed to get ec public key, err: %v\n", err)
		return nil, ErrInternal
	}

	if !ecdsa.Verify(key, digest[:], &r, &s) {
		fmt.Println("[error] invalid signature")
		return nil, errInvalidAccessToken
	}

	idn, err := db.RetrieveIdnByAuthToken(ctx, pbToken.GetToken())
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			fmt.Printf("[error] failed to retrieve idn, err: %v\n", err)
		}

		return nil, errInvalidAccessToken
	}

	return idn, nil
}

func checkJwt(ctx context.Context, appId string, s string) (*db.Idn, error) {
	log := zerolog.Ctx(ctx)
	token, err := jwt.Validate(ctx, s, mkJwkFunc(appId))
	if err != nil {
		log.Debug().Err(err).Msg("failed to validate jwt")
		return nil, errInvalidJwt
	}

	claims := token.Claims()
	idn, err := db.RetrieveIdnByFederatedLogin(ctx, appId, claims.Sub(), claims.Iss(), claims.Aud())
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, errInvalidJwt
		}

		return nil, err
	}

	return idn, nil
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
