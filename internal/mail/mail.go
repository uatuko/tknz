package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"

	"github.com/felk-ai/idaas/internal/klarapb"
)

var mailClient klarapb.MailClient
var domain string

func Init(ctx context.Context, mailAddr string, mailDomain string) error {
	var opts []grpc.DialOption
	if strings.HasPrefix(mailAddr, "localhost:") {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		ts, err := idtoken.NewTokenSource(ctx, fmt.Sprintf("https://%s", mailAddr))
		if err != nil {
			return err
		}

		opts = append(opts,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
			grpc.WithPerRPCCredentials(oauth.TokenSource{TokenSource: ts}),
		)
	}

	conn, err := grpc.NewClient(mailAddr, opts...)
	if err != nil {
		return err
	}

	mailClient = klarapb.NewMailClient(conn)
	domain = mailDomain
	return nil
}
