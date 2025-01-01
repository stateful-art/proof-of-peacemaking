package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"proofofpeacemaking/internal/core/ports"

	"github.com/mailgun/mailgun-go/v4"
)

type newsletterService struct {
	mailgunClient *mailgun.MailgunImpl
}

func NewNewsletterService(mailgunClient *mailgun.MailgunImpl) ports.NewsletterService {
	return &newsletterService{
		mailgunClient: mailgunClient,
	}
}

func (s *newsletterService) SendContactEmail(ctx context.Context, who string) error {
	s.mailgunClient.SetAPIBase("https://api.eu.mailgun.net/v3")

	sender := os.Getenv("EMAIL_SENDER_ADDRESS")
	recipient := os.Getenv("CONTACT_EMAIL_RECIPIENT_ADDRESS")
	subject := "Newsletter"
	body := generateContactEmailBody(who)
	message := s.mailgunClient.NewMessage(sender, subject, "", recipient)
	message.SetHtml(body)

	_, _, err := s.mailgunClient.Send(ctx, message)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func generateContactEmailBody(who string) string {
	return fmt.Sprintf(`
	<html>
	<head>
		<style>
			.container {
				max-width: 600px;
				margin: 0 auto;
				padding: 20px;
				font-family: Arial, sans-serif;
			}
			

			p {
			margin-top:2px;}	
		</style>
	</head>
	<body>
		<div class="container">
			<p> %s joined to Proof of Peacemaking newsletter</p>
		</div>
	</body>
	</html>
	`, who)
}
