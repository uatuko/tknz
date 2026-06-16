//go:build test

package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func Setup() error {
	dbname := os.Getenv("PGDATABASE")
	if !strings.HasPrefix(dbname, "test-") {
		os.Setenv("PGDATABASE", "test-"+dbname)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := Init(ctx); err != nil {
		return err
	}

	return nil
}

func Teardown() {
	Shutdown()
}

func TouchApp(t *testing.T) *App {
	space := TouchSpace(t)
	aud := fmt.Sprintf("https://%s.test", strings.ToLower(t.Name()))

	var app *App
	apps, err := ListApps(t.Context(), space.Id())
	if err != nil {
		t.Fatal(err)
	}

	for _, a := range apps {
		if a.Attrs().Aud == aud {
			app = &a
			break
		}
	}

	if app != nil {
		return app
	}

	app, err = NewOAuthApp(t.Context(), space.Id(), AppAttrs{Aud: aud})
	if err != nil {
		t.Fatal(err)
	}

	return app
}

func TouchIdn(t *testing.T) *Idn {
	app := TouchApp(t)
	login := strings.ToLower(t.Name())
	idn, err := RetrieveIdnByLogin(t.Context(), app.Id(), login)
	if err == nil {
		return idn
	}

	idn, err = NewIdn(t.Context(), app.Id(), login, IdnAttrs{})
	if err != nil {
		t.Fatal(err)
	}

	return idn
}

func TouchProvider(t *testing.T) *Provider {
	app := TouchApp(t)
	slug := ProviderSlug(strings.ToLower(t.Name()))

	var p *Provider
	pvs, err := ListProviders(t.Context(), app.Id())
	if err != nil {
		t.Fatal(err)
	}

	for _, pv := range pvs {
		if pv.Slug() == slug {
			p = &pv
			break
		}
	}

	if p != nil {
		return p
	}

	p, err = NewProvider(t.Context(), app.Id(), slug, ProviderAttrs{})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func TouchSpace(t *testing.T) *Space {
	slug := strings.ToLower(t.Name())
	space, err := RetrieveSpaceBySlug(t.Context(), slug)
	if err == nil {
		return space
	}

	space, err = NewSpace(t.Context(), slug, SpaceAttrs{})
	if err != nil {
		t.Fatal(err)
	}

	return space
}
