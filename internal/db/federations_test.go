package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFederationNew(t *testing.T) {
	app := TouchApp(t)

	t.Run("success", func(t *testing.T) {
		iss := fmt.Sprintf("%s-%s", t.Name(), rands(4))
		f, err := NewFederation(t.Context(), app.Id(), iss, FederationAttrs{})
		require.NoError(t, err)
		require.Greater(t, f.Rev(), int32(0))
		require.Len(t, f.Id(), 26)
		require.Equal(t, iss, f.Iss())
		require.Equal(t, FederationAttrs{Rev: f.Rev()}, f.Attrs())

		{
			qry := `
				select
					id,
					app_id,
					iss,
					attrs
				from federations
				where id = $1::text;
			`

			var id string
			var appId string
			var iss string
			var attrs FederationAttrs
			require.NoError(t, pg.QueryRow(t.Context(), qry, f.Id()).Scan(&id, &appId, &iss, &attrs))

			require.Equal(t, f.Id(), id)
			require.Equal(t, f.AppId(), appId)
			require.Equal(t, f.Iss(), iss)
			require.Equal(t, f.Attrs(), attrs)
		}
	})

	t.Run("error: missing app id", func(t *testing.T) {
		f, err := NewFederation(t.Context(), "", t.Name(), FederationAttrs{})
		require.Nil(t, f)
		require.ErrorIs(t, err, ErrInvalidData)
		require.Equal(t, "invalid data", err.Error())

		var e *wrappedError
		require.ErrorAs(t, err, &e)
		require.Equal(t, "missing app id", e.wrapped.Error())
	})

	t.Run("error: duplicate iss", func(t *testing.T) {
		iss := fmt.Sprintf("%s-%s", t.Name(), rands(4))
		_, err := NewFederation(t.Context(), app.Id(), iss, FederationAttrs{})
		require.NoError(t, err)

		f, err := NewFederation(t.Context(), app.Id(), iss, FederationAttrs{})
		require.Nil(t, f)
		require.ErrorIs(t, err, ErrConflict)
		require.Equal(t, "data conflict", err.Error())

		var e *wrappedError
		require.ErrorAs(t, err, &e)
		require.Equal(t, "ERROR: duplicate key value violates unique constraint \"federations.unique\" (SQLSTATE 23505)", e.wrapped.Error())
	})
}
