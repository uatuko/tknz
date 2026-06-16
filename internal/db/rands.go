package db

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
)

var encoding = base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(-1)

func RandB64Url(byteLen int) string {
	bytes := make([]byte, byteLen)
	rand.Read(bytes)

	return base64.RawURLEncoding.EncodeToString(bytes)
}

func rands(n int) string {
	if n < 4 {
		panic(fmt.Errorf("length too small, want at least 4 but got %d", n))
	}

	bytes := make([]byte, encoding.DecodedLen(n))
	rand.Read(bytes)

	return encoding.EncodeToString(bytes)
}
