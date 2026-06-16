package db

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppAttrsValidate(t *testing.T) {
	tests := []struct {
		attrs AppAttrs
		err   error
		msg   string
		name  string
	}{
		{
			AppAttrs{RedirectUris: []string{"https://redirect.test/callback"}},
			nil, "",
			"success: valid redirect uri",
		},
		{
			AppAttrs{RedirectUris: []string{"https://redirect.test/callback?source=test"}},
			nil, "",
			"success: redirect uri with query string",
		},
		{
			AppAttrs{RedirectUris: []string{"http://localhost:8080/callback"}},
			nil, "",
			"success: non tls localhost redirect uri",
		},
		{
			AppAttrs{RedirectUris: []string{"invalid"}},
			ErrInvalidData, "invalid redirect uri",
			"error: invalid redirect uri",
		},
		{
			AppAttrs{RedirectUris: []string{"https://127.0.0.1/callback"}},
			ErrInvalidData, "only hostnames are allowed in redirect uris",
			"error: ipv4 redirect uri",
		},
		{
			AppAttrs{RedirectUris: []string{"https://[::1]:8080/callback"}},
			ErrInvalidData, "only hostnames are allowed in redirect uris",
			"error: ipv6 redirect uri",
		},
		{
			AppAttrs{RedirectUris: []string{"ftp://redirect.test:8080/callback"}},
			ErrInvalidData, "unsupported url scheme for redirect uri",
			"error: ftp redirect uri",
		},
		{
			AppAttrs{RedirectUris: []string{"http://redirect.test:8080/callback"}},
			ErrInvalidData, "redirect uri must use tls",
			"error: non tls redirect uri",
		},
		{
			AppAttrs{RedirectUris: []string{"https://redirect.test/callback#fragment"}},
			ErrInvalidData, "fragments are not allowed in redirect uris",
			"error: fragments in redirect uri",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.attrs.validate()
			if test.err == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, test.err)
			require.Equal(t, test.err.Error(), err.Error())

			var wrapped *wrappedError
			require.ErrorAs(t, err, &wrapped)
			require.Equal(t, test.msg, wrapped.Unwrap().Error())
		})
	}
}

func TestAppNew(t *testing.T) {
	space := TouchSpace(t)

	app, err := NewOAuthApp(
		t.Context(),
		space.Id(),
		AppAttrs{
			Aud:  fmt.Sprintf("https://%s.test", t.Name()),
			Name: t.Name(),
		},
	)
	require.NoError(t, err)
	require.Greater(t, app.Rev(), int32(0))
	require.Len(t, app.Id(), 26)
	require.Len(t, app.OAuthClientId(), 26)

	qry := `
		select
			id,
			space_id,
			client_id,
			attrs
		from apps
		where id = $1::text;
	`

	var id string
	var spaceId string
	var clientId string
	var attrs AppAttrs
	require.NoError(t, pg.QueryRow(t.Context(), qry, app.id).Scan(&id, &spaceId, &clientId, &attrs))

	require.Equal(t, app.Id(), id)
	require.Equal(t, app.SpaceId(), spaceId)
	require.Equal(t, app.OAuthClientId(), clientId)
	require.Equal(t, app.Rev(), attrs.Rev)
	require.Equal(t, app.Attrs(), attrs)
}

func TestAppRetrieveByOAuthClientId(t *testing.T) {
	space := TouchSpace(t)

	t.Run("success", func(t *testing.T) {
		id := fmt.Sprintf("id:%s.%s", t.Name(), rands(4))
		clientId := fmt.Sprintf("client_id:%s.%s", t.Name(), rands(4))
		attrs := AppAttrs{
			Aud: fmt.Sprintf("https://%s.test", t.Name()),
			Rev: rand.Int32(),
		}

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

		_, err := pg.Exec(t.Context(), qry, id, space.Id(), clientId, attrs)
		require.NoError(t, err)

		app, err := RetrieveAppByOAuthClientId(t.Context(), clientId)
		require.NoError(t, err)

		require.Equal(t, id, app.Id())
		require.Equal(t, space.Id(), app.SpaceId())
		require.Equal(t, clientId, app.OAuthClientId())
		require.Equal(t, attrs.Rev, app.Rev())
		require.Equal(t, attrs, app.Attrs())
	})

	t.Run("not found", func(t *testing.T) {
		app, err := RetrieveAppByOAuthClientId(t.Context(), "dummy")
		require.Nil(t, app)
		require.ErrorIs(t, err, ErrNotFound)
		require.Equal(t, "resource not found", err.Error())
	})
}
