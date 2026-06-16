package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUuidV7(t *testing.T) {
	str := uuidv7()
	require.Len(t, str, 26)
}
