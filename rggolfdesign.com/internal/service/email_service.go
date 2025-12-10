package service

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client *resend.Client
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	client := resend.NewClient(apiKey)
	return &EmailService{
		client: client,
	}
}

func (s *EmailService) SendConsultationFollowUp(ctx context.Context, name, email string) error {
	params := &resend.SendEmailRequest{
		From:    "Ryan @ RG Golf Design <ryan@rggolfdesign.com>",
		To:      []string{email},
		Subject: "Your RG Golf Design Consultation Request",
		Html: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
	<div style="background: linear-gradient(135deg, #1a1a1a 0%%, #2a2a2a 100%%); padding: 40px 30px; text-align: center; border-radius: 8px 8px 0 0;">
		<h1 style="color: #00ff88; margin: 0; font-size: 28px; font-weight: 700; text-transform: uppercase; letter-spacing: 2px;">RG Golf Design</h1>
	</div>

	<div style="background: #ffffff; padding: 40px 30px; border-radius: 0 0 8px 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
		<p style="font-size: 18px; margin: 0 0 20px 0;">Hi %s,</p>

		<p style="font-size: 16px; margin: 0 0 20px 0; line-height: 1.8;">
			Thank you for reaching out about your golf simulator project! I'm Ryan, the founder of RG Golf Design, and I'm excited to learn more about your vision.
		</p>

		<p style="font-size: 16px; margin: 0 0 20px 0; line-height: 1.8;">
			I've received your consultation request and will be reviewing the details you provided. I'll personally follow up with you within the next 24 hours to discuss your space, requirements, and how we can create the perfect golf sanctuary for you.
		</p>

		<p style="font-size: 16px; margin: 0 0 20px 0; line-height: 1.8;">
			In the meantime, feel free to reply to this email if you have any additional questions or details you'd like to share.
		</p>

		<p style="font-size: 16px; margin: 0 0 10px 0;">Looking forward to connecting soon,</p>
		<p style="font-size: 16px; margin: 0; font-weight: 600;">Ryan</p>
		<p style="font-size: 14px; margin: 5px 0 0 0; color: #666;">Founder, RG Golf Design</p>

		<hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

		<p style="font-size: 13px; color: #888; margin: 0; text-align: center;">
			RG Golf Design | Premium Golf Simulators | Raleigh, NC<br>
			<a href="tel:+19193084133" style="color: #00ff88; text-decoration: none;">(919) 308-4133</a> |
			<a href="mailto:ryan@rggolfdesign.com" style="color: #00ff88; text-decoration: none;">ryan@rggolfdesign.com</a>
		</p>
	</div>
</body>
</html>
		`, name),
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if sent == nil {
		return fmt.Errorf("no response from email service")
	}

	return nil
}
