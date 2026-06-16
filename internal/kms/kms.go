package kms

import (
	"context"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	"google.golang.org/api/option"
)

var client *kms.KeyManagementClient

func Init(ctx context.Context) error {
	var err error
	quotaProject := os.Getenv("GCLOUD_KMS_QUOTA_PROJECT")
	client, err = kms.NewKeyManagementClient(ctx, option.WithQuotaProject(quotaProject))
	if err != nil {
		return err
	}

	return nil
}
