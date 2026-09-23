package kms

import (
	"context"

	"go.tknz.dev/internal/kms/google"
	"go.tknz.dev/internal/kms/local"
)

var (
	client DecryptSigner
)

type DecryptSigner interface {
	Decrypt(ctx context.Context, data []byte) ([]byte, error)

	Sign(ctx context.Context, keyName string, keyVersion string, data []byte) ([]byte, error)
}

// Init initialises the key management service.
//
// Google Cloud KMS is used when keyFile is empty, otherwise the pem encoded key
// in keyFile is used to sign and decrypt locally.
func Init(ctx context.Context, keyFile string) error {
	var err error
	if keyFile != "" {
		client, err = local.NewClient(ctx, keyFile)
	} else {
		client, err = google.NewClient(ctx)
	}

	if err != nil {
		return err
	}

	return nil
}
