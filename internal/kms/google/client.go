package google

import (
	"context"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
	"os"
	"slices"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"go.tknz.dev/internal/pb"
)

// ecCoordinateSize is the byte size of an ec signature coordinate.
//
// The signing requests below use a sha256 digest which Cloud KMS only accepts
// for 256 bit curves (EC_SIGN_P256_SHA256 and EC_SIGN_SECP256K1_SHA256), so the
// coordinates are P-256 sized.
var ecCoordinateSize = (elliptic.P256().Params().BitSize + 7) / 8

type client struct {
	api *kms.KeyManagementClient
}

func (c *client) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	var pbData pb.Cipher
	if err := proto.Unmarshal(data, &pbData); err != nil {
		return nil, err
	}

	keyName := fmt.Sprintf("%s/cryptoKeys/%s", os.Getenv("GCLOUD_KMS_KEYRING"), pbData.GetKmsKey())
	ciphertextCRC32C := crc32c(pbData.GetCiphertext())
	result, err := c.api.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:             keyName,
		Ciphertext:       pbData.GetCiphertext(),
		CiphertextCrc32C: wrapperspb.Int64(int64(ciphertextCRC32C)),
	})

	if err != nil {
		return nil, err
	}

	if int64(crc32c(result.GetPlaintext())) != result.GetPlaintextCrc32C().GetValue() {
		// FIXME: errors
		return nil, fmt.Errorf("decrypt: response corrupted in-transit")
	}

	return result.GetPlaintext(), nil
}

func (c *client) Sign(ctx context.Context, key string, keyVersion string, data []byte) ([]byte, error) {
	keyName := fmt.Sprintf("%s/cryptoKeys/%s/cryptoKeyVersions/%s",
		os.Getenv("GCLOUD_KMS_KEYRING"), key, keyVersion)

	digest := sha256.Sum256(data)
	result, err := c.api.AsymmetricSign(ctx, &kmspb.AsymmetricSignRequest{
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

	var sig struct {
		R *big.Int
		S *big.Int
	}
	if _, err = asn1.Unmarshal(result.Signature, &sig); err != nil {
		return nil, err
	}

	return slices.Concat(
		sig.R.FillBytes(make([]byte, ecCoordinateSize)),
		sig.S.FillBytes(make([]byte, ecCoordinateSize)),
	), nil
}
