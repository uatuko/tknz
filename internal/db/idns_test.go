package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdnNew(t *testing.T) {
	app := TouchApp(t)

	t.Run("success", func(t *testing.T) {
		idn, err := NewIdn(
			t.Context(),
			app.Id(),
			fmt.Sprintf("%s-%s", t.Name(), rands(4)),
			IdnAttrs{
				Email: fmt.Sprintf("%s@db.test", strings.ReplaceAll(t.Name(), "/", "-")),
			},
		)
		require.NoError(t, err)
		require.Greater(t, idn.Rev(), int32(0))
		require.Len(t, idn.Id(), 26)
		require.Equal(t, app.Id(), idn.AppId())
		require.Equal(t, IdnStatusPending, idn.Status())
		require.Equal(t, "", idn.SpaceId())
		require.Equal(t, "", idn.FederationId())

		qry := `
			select
				id,
				app_id,
				login,
				federation_id,
				attrs
			from idns
			where id = $1::text;
		`

		var id string
		var appId string
		var login string
		var federationId sql.NullString
		var attrs IdnAttrs
		require.NoError(t, pg.QueryRow(t.Context(), qry, idn.id).Scan(&id, &appId, &login, &federationId, &attrs))

		require.Equal(t, idn.Id(), id)
		require.Equal(t, idn.AppId(), appId)
		require.Equal(t, idn.Login(), login)
		require.Equal(t, idn.FederationId(), federationId.String)
		require.Equal(t, idn.Attrs(), attrs)
	})

	t.Run("error: missing app id", func(t *testing.T) {
		idn, err := NewIdn(t.Context(), "", t.Name(), IdnAttrs{})
		require.Nil(t, idn)
		require.ErrorIs(t, err, ErrInvalidData)
		require.Equal(t, "invalid data", err.Error())
	})
}

func TestIdnRetrieveByLogin(t *testing.T) {
	app := TouchApp(t)

	t.Run("success", func(t *testing.T) {
		id := fmt.Sprintf("id:%s.%s", t.Name(), rands(4))
		login := fmt.Sprintf("login:%s.%s", t.Name(), rands(4))
		attrs := IdnAttrs{
			Email:   fmt.Sprintf("%s@db.test", t.Name()),
			Rev:     rand.Int32(),
			SpaceId: app.SpaceId(),
		}

		qry := `
			insert into idns (
				id,
				app_id,
				login,
				attrs
			) values (
				$1::text,
				$2::text,
				$3::text,
				$4::jsonb
			);
		`

		_, err := pg.Exec(t.Context(), qry, id, app.Id(), login, attrs)
		require.NoError(t, err)

		idn, err := RetrieveIdnByLogin(t.Context(), app.Id(), login)
		require.NoError(t, err)

		require.Equal(t, id, idn.Id())
		require.Equal(t, app.Id(), idn.AppId())
		require.Equal(t, app.SpaceId(), idn.SpaceId())
		require.Equal(t, login, idn.Login())
		require.Equal(t, attrs.Rev, idn.Rev())
		require.Equal(t, attrs, idn.Attrs())
	})

	t.Run("not found", func(t *testing.T) {
		idn, err := RetrieveIdnByLogin(t.Context(), app.Id(), "dummy")
		require.Nil(t, idn)
		require.ErrorIs(t, err, ErrNotFound)
		require.Equal(t, "resource not found", err.Error())
	})
}

func TestIdnRetrieveBySrc(t *testing.T) {
	app := TouchApp(t)
	p := TouchProvider(t)

	t.Run("success", func(t *testing.T) {
		idn, err := NewIdn(t.Context(), app.Id(), fmt.Sprintf("%s-%s", t.Name(), rands(4)), IdnAttrs{})
		require.NoError(t, err)

		sub := fmt.Sprintf("sub:%s.%s", t.Name(), rands(4))
		_, err = NewIdnSrc(t.Context(), idn.Id(), p.Id(), sub)
		require.NoError(t, err)

		i, err := RetrieveIdnBySrc(t.Context(), p.Id(), sub)
		require.NoError(t, err)
		require.Equal(t, idn, i)
	})

	t.Run("not found", func(t *testing.T) {
		idn, err := RetrieveIdnBySrc(t.Context(), p.Id(), "dummy")
		require.Nil(t, idn)
		require.ErrorIs(t, err, ErrNotFound)
		require.Equal(t, "resource not found", err.Error())
	})
}

func TestIdnSrcNew(t *testing.T) {
	idn := TouchIdn(t)
	app := TouchApp(t)
	sub := fmt.Sprintf("sub:%s.%s", t.Name(), rands(4))

	p, err := NewProvider(t.Context(), app.Id(), ProviderSlug(fmt.Sprintf("test-%s", rands(4))), ProviderAttrs{})
	require.NoError(t, err)

	src, err := NewIdnSrc(t.Context(), idn.Id(), p.Id(), sub)
	require.NoError(t, err)
	require.Greater(t, idn.Rev(), int32(0))
	require.Equal(t, idn.Id(), src.IdnId())
	require.Equal(t, p.Id(), src.ProviderId())
	require.Equal(t, sub, src.Sub())

	qry := `
		select
			idn_id,
			provider_id,
			sub,
			attrs
		from idn_srcs
		where idn_id = $1::text and provider_id = $2::text
	`

	var idnId string
	var providerId string
	var s string
	var attrs IdnSrcAttrs
	require.NoError(t, pg.QueryRow(t.Context(), qry, idn.Id(), p.Id()).Scan(&idnId, &providerId, &s, &attrs))

	require.Equal(t, src.IdnId(), idnId)
	require.Equal(t, src.ProviderId(), providerId)
	require.Equal(t, src.Sub(), s)
	require.Equal(t, src.attrs, attrs)
}
