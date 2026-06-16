package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJwkNew(t *testing.T) {
	space := TouchSpace(t)
	attrs := JwkAttrs{}
	params := JwkParams{}

	jwk, err := NewJwk(t.Context(), space.Id(), attrs, params)
	require.NoError(t, err)

	require.Len(t, jwk.id, 26)
	require.Equal(t, jwk.id, jwk.Kid())
	require.Equal(t, space.Id(), jwk.SpaceId())
	require.Equal(t, attrs, jwk.Attrs())

	expected := params
	expected.Kid = jwk.Kid()
	require.Equal(t, expected, jwk.Params())
}

func TestJwksList(t *testing.T) {
	space, err := NewSpace(
		t.Context(),
		fmt.Sprintf("%s-%s", strings.ToLower(t.Name()), rands(4)),
		SpaceAttrs{},
	)
	require.NoError(t, err)

	jwk, err := NewJwk(t.Context(), space.Id(), JwkAttrs{}, JwkParams{})
	require.NoError(t, err)

	results, err := ListJwks(t.Context(), space.Id())
	require.NoError(t, err)

	require.Len(t, results, 1)
	require.Equal(t, []Jwk{*jwk}, results)
}
