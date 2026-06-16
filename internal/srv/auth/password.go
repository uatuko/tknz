package auth

import (
	"bytes"
	"fmt"

	"github.com/felk-ai/idaas/internal/db"
	"golang.org/x/crypto/argon2"
)

func verifyCredentials(password string, pwd db.IdnSrcAttrsPwd) error {
	if pwd.Typ != db.IdnSrcAttrsPwdTypArgon2 {
		return fmt.Errorf("unsupported credentials")
	}

	key := argon2.IDKey([]byte(password), pwd.Salt, pwd.Time, pwd.Memory, pwd.Threads, uint32(len(pwd.Key)))

	if !bytes.Equal(pwd.Key, key) {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}
