package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/jackc/pgx/v5"
)

const (
	ProviderSlugGoogleOAuth ProviderSlug = "google-oauth"
	ProviderSlugPasskey     ProviderSlug = "passkey"
	ProviderSlugPassword    ProviderSlug = "password"
)

type ProviderSlug string

type Provider struct {
	appId string
	attrs ProviderAttrs
	id    string
	slug  ProviderSlug
}

func (p *Provider) AppId() string {
	return p.appId
}

func (p *Provider) Attrs() ProviderAttrs {
	return p.attrs
}

func (p *Provider) ClientId() string {
	return p.attrs.ClientId
}

func (p *Provider) ClientSeceret() []byte {
	return p.attrs.ClientSeceret
}

func (p *Provider) Id() string {
	return p.id
}

func (p *Provider) Rev() int32 {
	return p.attrs.Rev
}

func (p *Provider) Slug() ProviderSlug {
	return p.slug
}

func (p *Provider) insert(ctx context.Context) error {
	if p.appId == "" {
		return wrapError(fmt.Errorf("missing app id"), ErrInvalidData)
	}

	p.id = uuidv7()
	p.attrs.Rev = rand.Int32()

	qry := `
		insert into providers (
			id,
			app_id,
			slug,
			attrs
		) values (
			$1::text,
			$2::text,
			$3::text,
			$4::jsonb
		);
	`

	if _, err := pg.Exec(ctx, qry, p.id, p.appId, p.slug, p.attrs); err != nil {
		return wrapError(err, ErrUnknown)
	}

	return nil
}

type ProviderAttrs struct {
	ClientId      string `json:"client_id,omitempty"`
	ClientSeceret []byte `json:"client_secret,omitempty"`

	Rev int32 `json:"_rev,omitempty"`
}

func ListProviders(ctx context.Context, appId string) ([]Provider, error) {
	qry := `
		select
			id,
			app_id,
			slug,
			attrs
		from providers
		where app_id = $1::text
		limit 100;
	`

	rows, err := pg.Query(ctx, qry, appId)
	if err != nil {
		return nil, wrapError(err, ErrUnknown)
	}

	providers := make([]Provider, 0)
	for rows.Next() {
		p := Provider{}
		if err := rows.Scan(&p.id, &p.appId, &p.slug, &p.attrs); err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		providers = append(providers, p)
	}

	return providers, nil
}

func NewProvider(ctx context.Context, appId string, slug ProviderSlug, attrs ProviderAttrs) (*Provider, error) {
	p := &Provider{
		attrs: attrs,
		appId: appId,
		slug:  slug,
	}

	if err := p.insert(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func ListProvidersByLogin(ctx context.Context, appId string, login string) ([]Provider, error) {
	qry := `
		select distinct
			p.id,
			p.app_id,
			p.slug,
			p.attrs
		from providers p
			inner join idn_srcs s on p.id = s.provider_id
			inner join idns i on s.idn_id = i.id
		where
			i.app_id = $1::text
			and i.login = $2::text
		;
	`

	rows, err := pg.Query(ctx, qry, appId, login)
	if err != nil {
		return nil, wrapError(err, ErrUnknown)
	}

	providers := make([]Provider, 0)
	for rows.Next() {
		p := Provider{}
		if err := rows.Scan(&p.id, &p.appId, &p.slug, &p.attrs); err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		providers = append(providers, p)
	}

	return providers, nil
}

func RetrieveProvider(ctx context.Context, id string) (*Provider, error) {
	qry := `
		select
			id,
			app_id,
			slug,
			attrs
		from providers
		where id = $1::text;
	`

	p := Provider{}
	if err := pg.QueryRow(ctx, qry, id).Scan(&p.id, &p.appId, &p.slug, &p.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &p, nil
}

func RetrieveProviderBySlub(ctx context.Context, appId string, slug ProviderSlug) (*Provider, error) {
	qry := `
		select
			id,
			app_id,
			slug,
			attrs
		from providers
		where
			app_id = $1::text
			and slug = $2::text
		;
	`

	p := Provider{}
	if err := pg.QueryRow(ctx, qry, appId, slug).Scan(&p.id, &p.appId, &p.slug, &p.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &p, nil
}
