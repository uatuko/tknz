package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpaceNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		slug := fmt.Sprintf(
			"%s-%s",
			strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-")),
			rands(4),
		)

		space, err := NewSpace(t.Context(), slug, SpaceAttrs{})
		require.NoError(t, err)
		require.Greater(t, space.Rev(), int32(0))
		require.Len(t, space.Id(), 26)
		require.Equal(t, slug, space.Slug())
		require.Equal(t, SpaceAttrs{Rev: space.Rev()}, space.Attrs())

		{
			qry := `
				select
					id,
					slug,
					attrs
				from spaces
				where id = $1::text;
			`

			var id string
			var slug string
			var attrs SpaceAttrs
			require.NoError(t, pg.QueryRow(t.Context(), qry, space.id).Scan(&id, &slug, &attrs))

			require.Equal(t, space.Id(), id)
			require.Equal(t, space.Slug(), slug)
			require.Equal(t, space.Attrs(), attrs)
		}
	})
}

func TestSpaceValidateSlug(t *testing.T) {
	tests := []struct {
		slug  string
		valid bool
		msg   string
	}{
		{"slug", true, ""},
		{"a-co", true, ""},

		{"sys", false, "invalid slug"},
		{"sys-x", false, "reserved words in slug"},
		{"x-sys", true, ""},
		{"x-sys-x", true, ""},
		{"system", false, "reserved words in slug"},
		{"system-x", false, "reserved words in slug"},
		{"x-system", true, ""},
		{"x-system-x", true, ""},

		{"felk", false, "reserved words in slug"},
		{"felk-x", false, "reserved words in slug"},
		{"x-felk", false, "reserved words in slug"},
		{"x-felk-x", false, "reserved words in slug"},
		{"felkx", false, "reserved words in slug"},
		{"xfelk", true, ""},
		{"xfelkx", true, ""},

		{"a--b", false, "invalid slug"},
		{"abc", false, "invalid slug"},
		{"slug-", false, "invalid slug"},
		{"-slug", false, "invalid slug"},
	}

	for _, test := range tests {
		var name string
		if test.valid {
			name = fmt.Sprintf("success: valid slug (%s)", test.slug)
		} else {
			name = fmt.Sprintf("error: invalid slug (%s)", test.slug)
		}

		t.Run(name, func(t *testing.T) {
			space := Space{slug: test.slug}
			err := space.validate()

			if test.valid {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, ErrInvalidData)
			require.Equal(t, "invalid data", err.Error())

			var wrapped *wrappedError
			require.ErrorAs(t, err, &wrapped)
			require.Equal(t, test.msg, wrapped.Unwrap().Error())
		})
	}
}
