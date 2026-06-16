package mail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/felk-ai/idaas/internal/klarapb"
)

func SendSignUpVerify(ctx context.Context, toEmail string, link string, ttl time.Duration) error {
	data := struct {
		VerifyLink string
		TtlMinutes float64
	}{
		VerifyLink: link,
		TtlMinutes: ttl.Minutes(),
	}

	log := zerolog.Ctx(ctx)
	tmpl := signUpVerityTemplate()

	var out strings.Builder
	err := tmpl.Execute(&out, data)
	if err != nil {
		log.Error().
			Err(err).
			Str("name", tmpl.Name()).
			Msg("failed to execute template")

		return err
	}

	req := klarapb.MailSendRequest{
		From: &klarapb.Address{
			Email: fmt.Sprintf("no-reply@%s", domain),
			Name:  "Felk",
		},
		Subject: "Verify your email for Felk",
		To:      []*klarapb.Address{{Email: toEmail}},
		Content: []*klarapb.Message{{Data: out.String(), Type: "text/html"}},
	}

	_, err = mailClient.Send(ctx, &req)
	if err != nil {
		log.Error().
			Err(err).
			Msg("mail send grpc request failed")

		return err
	}

	return nil
}
