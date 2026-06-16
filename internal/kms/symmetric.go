package kms

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/felk-ai/idaas/internal/pb"
)

func Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	pbData := pb.Cipher{}
	if err := proto.Unmarshal(data, &pbData); err != nil {
		return nil, err
	}

	keyName := fmt.Sprintf("%s/cryptoKeys/%s", os.Getenv("GCLOUD_KMS_KEYRING"), pbData.GetKmsKey())
	ciphertextCRC32C := crc32c(pbData.GetCiphertext())
	result, err := client.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:             keyName,
		Ciphertext:       pbData.GetCiphertext(),
		CiphertextCrc32C: wrapperspb.Int64(int64(ciphertextCRC32C)),
	})

	if err != nil {
		return nil, err
	}

	if int64(crc32c(result.GetPlaintext())) != result.GetPlaintextCrc32C().GetValue() {
		// FIXME: errors
		return nil, fmt.Errorf("Decrypt: response corrupted in-transit")
	}

	return result.GetPlaintext(), nil
}
