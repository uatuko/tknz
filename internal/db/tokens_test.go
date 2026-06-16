package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenExpire(t *testing.T) {
	attrs := struct{}{}
	tkn := token{
		attrs: attrs,
		exp:   time.Now().Add(5 * time.Second).UTC(),
		id:    fmt.Sprintf("test|%s", rands(26)),
	}
	require.NoError(t, tkn.insert(t.Context()))

	t.Run("success", func(t *testing.T) {
		require.NoError(t, tkn.Expire(t.Context()))
		require.Equal(t, time.Time{}, tkn.Exp())

		qry := `
			select
				exp
			from tokens
			where id = $1::text;
		`

		var exp time.Time
		require.NoError(t, pg.QueryRow(t.Context(), qry, tkn.id).Scan(&exp))
		require.Equal(t, time.Time{}, exp)
	})

	t.Run("error: data conflict", func(t *testing.T) {
		tkn := token{
			attrs: attrs,
			exp:   time.Now().Add(5 * time.Second).UTC(),
			id:    fmt.Sprintf("test|%s", rands(26)),
		}
		require.NoError(t, tkn.insert(t.Context()))

		tkn.exp = time.Now().UTC()
		err := tkn.Expire(t.Context())
		require.ErrorIs(t, err, ErrConflict)
		require.Equal(t, "data conflict", err.Error())
	})
}

func TestTokenStore(t *testing.T) {
	attrs := struct {
		Email string `json:"email,omitempty"`
	}{fmt.Sprintf("%s@db.test", t.Name())}

	tkn := token{
		attrs: attrs,
		exp:   time.UnixMicro(time.Now().Add(5 * time.Second).UnixMicro()).UTC(),
		id:    fmt.Sprintf("test|%s", rands(26)),
	}
	require.NoError(t, tkn.insert(t.Context()))

	qry := `
		select
			id,
			exp,
			attrs
		from tokens
		where id = $1::text;
	`

	var id string
	var exp time.Time
	var attrsStr string
	require.NoError(
		t,
		pg.QueryRow(t.Context(), qry, tkn.id).Scan(&id, &exp, &attrsStr),
	)

	require.Equal(t, tkn.id, id)
	require.Equal(t, tkn.exp, exp)
	require.Equal(t, fmt.Sprintf(`{"email": "%s"}`, fmt.Sprintf("%s@db.test", t.Name())), attrsStr)
}

func TestTokenRetrieve(t *testing.T) {
	id := fmt.Sprintf("test|%s", rands(26))
	exp := time.UnixMicro(time.Now().Add(5 * time.Second).UnixMicro()).UTC()
	attrs := fmt.Sprintf(`{"email": "%s"}`, fmt.Sprintf("%s@db.test", t.Name()))

	qry := `
		insert into tokens (
			id,
			exp,
			attrs
		) values (
			$1::text,
			$2::timestamp,
			$3::jsonb
		);
	`

	_, err := pg.Exec(t.Context(), qry, id, exp, attrs)
	require.NoError(t, err)

	tkn := token{}
	require.NoError(t, retrieveToken(t.Context(), id, &tkn))

	require.Equal(t, id, tkn.id)
	require.Equal(t, exp, tkn.exp)
	require.NotNil(t, tkn.attrs)
	require.Equal(t, []byte(attrs), tkn.attrs.([]byte))
}
