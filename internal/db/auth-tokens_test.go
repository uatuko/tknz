package db

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuthTokenNew(t *testing.T) {
	attrs := AuthTokenAttrs{}
	now := time.Now()
	ttl := 5 * time.Second

	token, err := NewAuthToken(t.Context(), ttl, attrs)
	require.NoError(t, err)
	require.Equal(t, attrs, token.Attrs())
	require.LessOrEqual(t, now.Add(ttl), token.Exp())
	require.Len(t, token.Token(), 20)
	require.True(t, strings.HasPrefix(token.id, authTokenPrefix))
}
