package db

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"errors"
	"math/big"

	"github.com/jackc/pgx/v5"
)

const (
	JwkAlgES256 JwkAlg = "ES256"
	JwkAlgRS256 JwkAlg = "RS256"

	JwkEcCrvP256 JwkEcCrv = "P-256"

	JwkKtyEc  JwkKty = "EC"
	JwkKtyRsa JwkKty = "RSA"

	JwkUseSig JwkUse = "sig"

	JwkKeyOpSign   JwkKeyOp = "sign"
	JwkKeyOpVerify JwkKeyOp = "verify"
)

type JwkAlg string
type JwkEcCrv string
type JwkKty string
type JwkUse string
type JwkKeyOp string

type Jwk struct {
	attrs   JwkAttrs
	id      string
	params  JwkParams
	spaceId string
}

func (k *Jwk) Attrs() JwkAttrs {
	return k.attrs
}

func (k *Jwk) Kid() string {
	return k.id
}

func (k *Jwk) Params() JwkParams {
	return k.params
}

func (k *Jwk) SpaceId() string {
	return k.spaceId
}

func (k *Jwk) insert(ctx context.Context) error {
	k.id = rands(26)
	k.params.Kid = k.id

	qry := `
		insert into jwks (
			id,
			space_id,
			attrs,
			params
		) values (
			$1::text,
			$2::text,
			$3::jsonb,
			$4::jsonb
		);
	`

	if _, err := pg.Exec(ctx, qry, k.id, k.spaceId, k.attrs, k.params); err != nil {
		return err
	}

	return nil
}

type JwkAttrs struct {
	KmsKey        string `json:"kms_key,omitempty"`
	KmsKeyVersion string `json:"kms_key_version,omitempty"`
}

type JwkParams struct {
	Alg    JwkAlg     `json:"alg,omitempty"`
	Kid    string     `json:"kid,omitempty"`
	Kty    JwkKty     `json:"kty,omitempty"`
	Use    JwkUse     `json:"use,omitempty"`
	KeyOps []JwkKeyOp `json:"key_ops,omitempty"`

	// Elliptic curve (EC) params
	Crv JwkEcCrv `json:"crv,omitempty"`
	X   string   `json:"x,omitempty"`
	Y   string   `json:"y,omitempty"`

	// RSA (RSA) params
	N string `json:"n,omitempty"` // modulus
	E string `json:"e,omitempty"` // exponent
}

func (params *JwkParams) EcPublicKey() (*ecdsa.PublicKey, error) {
	if params.Kty != JwkKtyEc || params.Crv != JwkEcCrvP256 {
		return nil, ErrInvalidData
	}

	var x, y big.Int
	b, err := base64.RawURLEncoding.DecodeString(params.X)
	if err != nil {
		return nil, wrapError(err, ErrInvalidData)
	}
	x.SetBytes(b)

	b, err = base64.RawURLEncoding.DecodeString(params.Y)
	if err != nil {
		return nil, wrapError(err, ErrInvalidData)
	}
	y.SetBytes(b)

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     &x,
		Y:     &y,
	}, nil
}

func ListJwks(ctx context.Context, spaceId string) ([]Jwk, error) {
	qry := `
		select
			id,
			space_id,
			attrs,
			params
		from jwks
		where space_id = $1::text
		limit 100;
	`

	jwks := []Jwk{}
	rows, err := pg.Query(ctx, qry, spaceId)
	if err != nil {
		return nil, wrapError(err, ErrUnknown)
	}

	for rows.Next() {
		jwk := Jwk{}
		err := rows.Scan(
			&jwk.id,
			&jwk.spaceId,
			&jwk.attrs,
			&jwk.params,
		)
		if err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		jwks = append(jwks, jwk)
	}

	return jwks, nil
}

func RetrieveJwk(ctx context.Context, spaceId string, kid string) (*Jwk, error) {
	qry := `
		select
			id,
			space_id,
			attrs,
			params
		from jwks
		where space_id = $1::text and id = $2::text
	`

	var jwk Jwk
	if err := pg.QueryRow(ctx, qry, spaceId, kid).Scan(&jwk.id, &jwk.spaceId, &jwk.attrs, &jwk.params); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, wrapError(err, ErrUnknown)
	}

	return &jwk, nil
}

func NewJwk(ctx context.Context, spaceId string, attrs JwkAttrs, params JwkParams) (*Jwk, error) {
	jwk := &Jwk{
		attrs:   attrs,
		params:  params,
		spaceId: spaceId,
	}

	if err := jwk.insert(ctx); err != nil {
		return nil, err
	}

	return jwk, nil
}
