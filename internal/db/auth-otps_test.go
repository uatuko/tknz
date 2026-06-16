package db

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthOtpNew(t *testing.T) {
	attrs := AuthOtpAttrs{
		Login: fmt.Sprintf("%s-%s", t.Name(), rands(4)),
	}
	now := time.Now()
	ttl := 5 * time.Second

	otp, err := NewAuthOtp(t.Context(), ttl, attrs)
	require.NoError(t, err)
	require.Equal(t, attrs, otp.Attrs())
	require.LessOrEqual(t, now.Add(ttl), otp.Exp())
	require.NotEmpty(t, otp.Code())
	require.True(t, strings.HasPrefix(otp.id, authOtpPrefix))
	require.False(t, strings.HasPrefix(otp.Code(), authOtpPrefix))
}

func TestAuthOtpRetrieve(t *testing.T) {
	id := authOtpPrefix + rands(26)
	exp := time.UnixMicro(time.Now().Add(5 * time.Second).UnixMicro()).UTC()
	attrs := AuthOtpAttrs{
		Login: fmt.Sprintf("%s-%s", t.Name(), rands(4)),
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

	otp, err := RetrieveAuthOtp(t.Context(), strings.TrimPrefix(id, authOtpPrefix))
	require.NoError(t, err)
	require.Equal(t, id, authOtpPrefix+otp.Code())
	require.Equal(t, exp, otp.Exp())
	require.Equal(t, attrs, otp.Attrs())
}
