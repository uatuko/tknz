package jwt

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAudienceMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   registeredClaims
		want string
	}{
		{name: "single", in: registeredClaims{Aud: audience{"a"}}, want: `{"aud":"a"}`},
		{name: "multiple", in: registeredClaims{Aud: audience{"a", "b"}}, want: `{"aud":["a","b"]}`},
		{name: "nil", in: registeredClaims{}, want: `{}`},
		{name: "empty", in: registeredClaims{Aud: audience{}}, want: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.in)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(b))
		})
	}
}

func TestAudienceUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want audience
		err  bool
	}{
		{name: "string", in: `{"aud":"a"}`, want: audience{"a"}},
		{name: "array", in: `{"aud":["a","b"]}`, want: audience{"a", "b"}},
		{name: "single element array", in: `{"aud":["a"]}`, want: audience{"a"}},
		{name: "empty string", in: `{"aud":""}`, want: nil},
		{name: "empty array", in: `{"aud":[]}`, want: audience{}},
		{name: "null", in: `{"aud":null}`, want: nil},
		{name: "missing", in: `{}`, want: nil},
		{name: "number", in: `{"aud":1}`, err: true},
		{name: "object", in: `{"aud":{}}`, err: true},
		{name: "mixed array", in: `{"aud":["a",1]}`, err: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c registeredClaims
			err := json.Unmarshal([]byte(tt.in), &c)
			if tt.err {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, c.Aud)
		})
	}
}
