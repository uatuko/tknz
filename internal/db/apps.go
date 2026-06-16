package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/url"

	"github.com/jackc/pgx/v5"
)

type App struct {
	attrs    AppAttrs
	clientId *string
	id       string
	spaceId  string
}

func (app *App) Attrs() AppAttrs {
	return app.attrs
}

func (app *App) Id() string {
	return app.id
}

func (app *App) OAuthClientId() string {
	if app.clientId == nil {
		return ""
	}

	return *app.clientId
}

func (app *App) RedirectUris() []string {
	return app.attrs.RedirectUris
}

func (app *App) Rev() int32 {
	return app.attrs.Rev
}

func (app *App) RpId() string {
	return app.attrs.RpId
}

func (app *App) RpName() string {
	return app.attrs.RpName
}

func (app *App) SpaceId() string {
	return app.spaceId
}

func (app *App) insert(ctx context.Context) error {
	if err := app.validate(); err != nil {
		return err
	}

	app.id = uuidv7()
	app.attrs.Rev = rand.Int32()

	qry := `
		insert into apps (
			id,
			space_id,
			client_id,
			attrs
		) values (
			$1::text,
			$2::text,
			$3::text,
			$4::jsonb
		);
	`

	if _, err := pg.Exec(
		ctx,
		qry,
		app.id,
		app.spaceId,
		app.clientId,
		app.attrs,
	); err != nil {
		return wrapError(err, ErrUnknown)
	}

	return nil
}

func (app *App) validate() error {
	return app.attrs.validate()
}

type AppAttrs struct {
	Aud          string      `json:"aud,omitempty"`
	Keys         []JwkParams `json:"keys,omitempty"`
	Name         string      `json:"name,omitempty"`
	RedirectUris []string    `json:"redirect_uris,omitempty"`

	RpId   string `json:"rp_id,omitempty"`
	RpName string `json:"rp_name,omitempty"`

	Rev int32 `json:"_rev,omitempty"`
}

func (attrs *AppAttrs) validate() error {
	for _, str := range attrs.RedirectUris {
		u, err := url.Parse(str)
		if err != nil || !u.IsAbs() {
			return wrapError(fmt.Errorf("invalid redirect uri"), ErrInvalidData)
		}

		addr := net.ParseIP(u.Hostname())
		if addr != nil {
			return wrapError(fmt.Errorf("only hostnames are allowed in redirect uris"), ErrInvalidData)
		}

		if u.Scheme != "https" {
			if u.Scheme != "http" {
				return wrapError(fmt.Errorf("unsupported url scheme for redirect uri"), ErrInvalidData)
			}

			if u.Scheme == "http" && u.Hostname() != "localhost" {
				return wrapError(fmt.Errorf("redirect uri must use tls"), ErrInvalidData)
			}
		}

		if u.Fragment != "" {
			return wrapError(fmt.Errorf("fragments are not allowed in redirect uris"), ErrInvalidData)
		}
	}

	return nil
}

func ListApps(ctx context.Context, spaceId string) ([]App, error) {
	qry := `
		select
			id,
			space_id,
			client_id,
			attrs
		from apps
		where space_id = $1::text
		limit 100;
	`

	rows, err := pg.Query(ctx, qry, spaceId)
	if err != nil {
		return nil, wrapError(err, ErrUnknown)
	}

	apps := make([]App, 0, rows.CommandTag().RowsAffected())
	for rows.Next() {
		var app App
		if err := rows.Scan(&app.id, &app.spaceId, &app.clientId, &app.attrs); err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		apps = append(apps, app)
	}

	return apps, nil
}

func NewOAuthApp(ctx context.Context, spaceId string, attrs AppAttrs) (*App, error) {
	clientId := rands(26)
	app := &App{
		attrs:    attrs,
		clientId: &clientId,
		spaceId:  spaceId,
	}

	if err := app.insert(ctx); err != nil {
		return nil, err
	}

	return app, nil
}

func RetrieveApp(ctx context.Context, id string) (*App, error) {
	qry := `
		select
			id,
			space_id,
			client_id,
			attrs
		from apps
		where id = $1::text
	`

	app := App{}
	err := pg.QueryRow(ctx, qry, id).Scan(&app.id, &app.spaceId, &app.clientId, &app.attrs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &app, nil
}

func RetrieveAppByOAuthClientId(ctx context.Context, clientId string) (*App, error) {
	qry := `
		select
			id,
			space_id,
			client_id,
			attrs
		from apps
		where
			client_id = $1::text
	`

	app := App{}
	err := pg.QueryRow(ctx, qry, clientId).Scan(&app.id, &app.spaceId, &app.clientId, &app.attrs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &app, nil
}
