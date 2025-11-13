package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"auth-service/internal/config"
)

// Service defines the interface for email operations
type Service interface {
	SendPasswordResetEmail(toEmail, resetToken string) error
}

// SendGridPayload represents the payload for SendGrid API
type SendGridPayload struct {
	Personalizations []SendGridPersonalization `json:"personalizations"`
	From             SendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []SendGridContent         `json:"content"`
}

type SendGridPersonalization struct {
	To []SendGridEmail `json:"to"`
}

type SendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type SendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// service implements Service interface
type service struct {
	config *config.EmailConfig
	client *http.Client
}

// NewService creates a new email service
func NewService(cfg *config.EmailConfig) Service {
	return &service{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendPasswordResetEmail sends a password reset email
func (s *service) SendPasswordResetEmail(toEmail, resetToken string) error {
	switch s.config.Provider {
	case "sendgrid":
		return s.sendViaSendGrid(toEmail, resetToken)
	default:
		// For development, just log the email
		fmt.Printf("Password reset email would be sent to %s with token %s\n", toEmail, resetToken)
		return nil
	}
}

// sendViaSendGrid sends email via SendGrid API
func (s *service) sendViaSendGrid(toEmail, resetToken string) error {
	if s.config.APIKey == "" {
		return fmt.Errorf("SendGrid API key not configured")
	}

	resetURL := fmt.Sprintf("%s?token=%s", s.config.ResetURL, resetToken)

	// Generate HTML email content
	htmlContent, err := s.generateResetEmailHTML(resetURL)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	payload := SendGridPayload{
		Personalizations: []SendGridPersonalization{
			{
				To: []SendGridEmail{
					{Email: toEmail},
				},
			},
		},
		From: SendGridEmail{
			Email: s.config.FromEmail,
			Name:  s.config.FromName,
		},
		Subject: "Password Reset Request",
		Content: []SendGridContent{
			{
				Type:  "text/html",
				Value: htmlContent,
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("SendGrid API returned status %d", resp.StatusCode)
	}

	return nil
}

// generateResetEmailHTML generates HTML content for password reset email
func (s *service) generateResetEmailHTML(resetURL string) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2>Password Reset Request</h2>
    <p>You have requested to reset your password. Click the link below to reset your password:</p>
    <p style="margin: 30px 0;">
        <a href="{{.ResetURL}}" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a>
    </p>
    <p>If you didn't request this password reset, please ignore this email.</p>
    <p>This link will expire in 15 minutes for security reasons.</p>
    <p>If the button doesn't work, copy and paste this URL into your browser:</p>
    <p style="word-break: break-all; color: #666;">{{.ResetURL}}</p>
    <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
    <p style="color: #666; font-size: 12px;">This email was sent by {{.FromName}}. If you have any questions, please contact support.</p>
</body>
</html>`

	t, err := template.New("resetEmail").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		ResetURL string
		FromName string
	}{
		ResetURL: resetURL,
		FromName: s.config.FromName,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}