package db

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthCodeNew(t *testing.T) {
	attrs := AuthCodeAttrs{
		ClientId:    fmt.Sprintf("client_id|%s", t.Name()),
		RedirectUri: fmt.Sprintf("https://%s.test/cb", t.Name()),
	}
	now := time.Now()
	ttl := 5 * time.Second

	code, err := NewAuthCode(t.Context(), ttl, attrs)
	require.NoError(t, err)
	require.Equal(t, attrs, code.Attrs())
	require.LessOrEqual(t, now.Add(ttl), code.Exp())
	require.NotEmpty(t, code.Id())
	require.True(t, strings.HasPrefix(code.id, authCodePrefix))
	require.False(t, strings.HasPrefix(code.Id(), authCodePrefix))
}

func TestAuthCodeRetrieve(t *testing.T) {
	id := authCodePrefix + rands(26)
	exp := time.UnixMicro(time.Now().Add(5 * time.Second).UnixMicro()).UTC()
	attrs := AuthCodeAttrs{
		ClientId:    fmt.Sprintf("client_id|%s", t.Name()),
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

	code, err := RetrieveAuthCode(t.Context(), strings.TrimPrefix(id, authCodePrefix))
	require.NoError(t, err)
	require.Equal(t, id, authCodePrefix+code.Id())
	require.Equal(t, exp, code.Exp())
	require.Equal(t, attrs, code.Attrs())
}
