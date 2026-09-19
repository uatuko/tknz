package google

import (
	"context"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	"google.golang.org/api/option"
)

func NewClient(ctx context.Context) (*client, error) {
	quotaProject := os.Getenv("GCLOUD_KMS_QUOTA_PROJECT")
	c, err := kms.NewKeyManagementClient(ctx, option.WithQuotaProject(quotaProject))
	if err != nil {
		return nil, err
	}

	return &client{api: c}, nil
}
