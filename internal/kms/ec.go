package kms

import (
	"context"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
	"os"

	"cloud.google.com/go/kms/apiv1/kmspb"
)

type EcSig struct {
	R *big.Int
	S *big.Int
}

type EcKey struct {
	kid           string
	kmsKey        string
	kmsKeyVersion string
}

func (k *EcKey) Kid() string {
	return k.kid
}

func (k *EcKey) Sign(ctx context.Context, data []byte) (*EcSig, error) {
	keyName := fmt.Sprintf("%s/cryptoKeys/%s/cryptoKeyVersions/%s",
		os.Getenv("GCLOUD_KMS_KEYRING"), k.kmsKey, k.kmsKeyVersion)

	digest := sha256.Sum256(data)
	result, err := client.AsymmetricSign(ctx, &kmspb.AsymmetricSignRequest{
		Name: keyName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: digest[:],
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var sig EcSig
	if _, err = asn1.Unmarshal(result.Signature, &sig); err != nil {
		return nil, err
	}

	return &sig, nil
}

func NewEcKey(kid string, kmsKey string, kmsKeyVersion string) *EcKey {
	return &EcKey{
		kid:           kid,
		kmsKey:        kmsKey,
		kmsKeyVersion: kmsKeyVersion,
	}
}
