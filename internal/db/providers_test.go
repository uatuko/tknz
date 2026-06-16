package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderNew(t *testing.T) {
	app := TouchApp(t)
	t.Run("success", func(t *testing.T) {
		p, err := NewProvider(
			t.Context(),
			app.Id(),
			ProviderSlug(fmt.Sprintf("test-%s", rands(4))),
			ProviderAttrs{
				ClientId:      app.OAuthClientId(),
				ClientSeceret: []byte("secret"),
			},
		)
		require.NoError(t, err)
		require.Greater(t, p.Rev(), int32(0))
		require.Len(t, p.Id(), 26)

		qry := `
			select
				id,
				app_id,
				slug,
				attrs
			from providers
			where id = $1::text;
		`

		var id string
		var appId string
		var slug ProviderSlug
		var attrs ProviderAttrs
		require.NoError(t, pg.QueryRow(t.Context(), qry, p.id).Scan(&id, &appId, &slug, &attrs))

		require.Equal(t, p.Id(), id)
		require.Equal(t, p.AppId(), appId)
		require.Equal(t, p.Slug(), slug)
		require.Equal(t, p.Rev(), attrs.Rev)
		require.Equal(t, p.Attrs(), attrs)
	})

	t.Run("error: missing app id", func(t *testing.T) {
		p, err := NewProvider(t.Context(), "", ProviderSlug(t.Name()), ProviderAttrs{})
		require.Nil(t, p)
		require.ErrorIs(t, err, ErrInvalidData)
		require.Equal(t, "invalid data", err.Error())
	})
}

func TestProvidersList(t *testing.T) {
	space := TouchSpace(t)
	app, err := NewOAuthApp(
		t.Context(),
		space.Id(),
		AppAttrs{Aud: fmt.Sprintf("https://%s.test", t.Name())},
	)
	require.NoError(t, err)

	p, err := NewProvider(t.Context(), app.Id(), ProviderSlug(t.Name()), ProviderAttrs{})
	require.NoError(t, err)

	results, err := ListProviders(t.Context(), app.Id())
	require.NoError(t, err)

	require.Len(t, results, 1)
	require.Equal(t, []Provider{*p}, results)
}
