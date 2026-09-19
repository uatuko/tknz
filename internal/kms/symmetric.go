package kms

import (
	"context"
)

func Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	return client.Decrypt(ctx, data)
}
