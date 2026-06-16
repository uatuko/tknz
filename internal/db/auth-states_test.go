package db

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthStateNew(t *testing.T) {
	attrs := AuthStateAttrs{
		ClientId:    fmt.Sprintf("%s-%s", t.Name(), rands(4)),
		RedirectUri: fmt.Sprintf("https://%s.test/cb", t.Name()),
	}
	now := time.Now()
	ttl := 5 * time.Second

	state, err := NewAuthState(t.Context(), ttl, attrs)
	require.NoError(t, err)
	require.Greater(t, state.Rev(), int32(0))
	require.LessOrEqual(t, now.Add(ttl), state.Exp())
	require.NotEmpty(t, state.Id())
	require.True(t, strings.HasPrefix(state.id, authStatePrefix))
	require.False(t, strings.HasPrefix(state.Id(), authStatePrefix))

	attrs.Rev = state.Rev()
	require.Equal(t, attrs, state.Attrs())
}

func TestAuthStateRetrieve(t *testing.T) {
	id := authStatePrefix + rands(26)
	exp := time.UnixMicro(time.Now().Add(5 * time.Second).UnixMicro()).UTC()
	attrs := AuthStateAttrs{
		ClientId:    fmt.Sprintf("%s-%s", t.Name(), rands(4)),
		RedirectUri: fmt.Sprintf("https://%s.test/cb", t.Name()),
	}

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

	state, err := RetrieveAuthState(t.Context(), strings.TrimPrefix(id, authStatePrefix))
	require.NoError(t, err)
	require.Equal(t, id, authStatePrefix+state.Id())
	require.Equal(t, exp, state.Exp())
	require.Equal(t, attrs, state.Attrs())
}
