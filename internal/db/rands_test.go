package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRands(t *testing.T) {
	for _, n := range []int{4, 5, 7, 15, 26, 32} {
		str := rands(n)
		require.Len(t, str, n)
	}
}

func TestRandsError(t *testing.T) {
	fn := func() {
		rands(3)
	}

	assert.PanicsWithError(t, "length too small, want at least 4 but got 3", fn)
}
