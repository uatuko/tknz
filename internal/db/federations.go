package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Federation struct {
	appId string
	attrs FederationAttrs
	id    string
	iss   string
}

func (f *Federation) AppId() string {
	return f.appId
}

func (f *Federation) Attrs() FederationAttrs {
	return f.attrs
}

func (f *Federation) Aud() string {
	return f.attrs.Aud
}

func (f *Federation) Id() string {
	return f.id
}

func (f *Federation) Iss() string {
	return f.iss
}

func (f *Federation) JwksUri() string {
	return f.attrs.JwksUri
}

func (f *Federation) Rev() int32 {
	return f.attrs.Rev
}

func (f *Federation) insert(ctx context.Context) error {
	if f.appId == "" {
		return wrapError(fmt.Errorf("missing app id"), ErrInvalidData)
	}

	f.id = uuidv7()
	f.attrs.Rev = rand.Int32()

	qry := `
		insert into federations (
			id,
			app_id,
			iss,
			attrs
		) values (
			$1::text,
			$2::text,
			$3::text,
			$4::jsonb
		);
	`

	if _, err := pg.Exec(ctx, qry, f.id, f.appId, f.iss, f.attrs); err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			switch pgxErr.Code {
			case "23505": // unique violation
				return wrapError(err, ErrConflict)
			default:
				return wrapError(err, ErrUnknown)
			}
		}

		return wrapError(err, ErrUnknown)
	}

	return nil
}

type FederationAttrs struct {
	Aud     string `json:"aud,omitempty"`
	JwksUri string `json:"jwks_uri,omitempty"`

	Rev int32 `json:"_rev,omitempty"`
}

func NewFederation(ctx context.Context, appId string, iss string, attrs FederationAttrs) (*Federation, error) {
	f := &Federation{
		attrs: attrs,
		appId: appId,
		iss:   iss,
	}

	if err := f.insert(ctx); err != nil {
		return nil, err
	}

	return f, nil
}

func RetrieveFederationByIss(ctx context.Context, appId string, iss string) (*Federation, error) {
	qry := `
		select
			id,
			app_id,
			iss,
			attrs
		from federations
		where
			app_id = $1::text
			and iss = $2::text
		;
	`

	var f Federation
	err := pg.QueryRow(ctx, qry, appId, iss).Scan(&f.id, &f.appId, &f.iss, &f.attrs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &f, nil
}
