package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"net/url"

	"auth-service/internal/config"

	"github.com/resend/resend-go/v2"
)

// Service defines the interface for email operations
type Service interface {
	SendPasswordResetEmail(toEmail, resetToken string) error
	SendInvitationEmail(toEmail, inviterName, organizationName, invitationToken string) error
	SendVerificationEmail(toEmail, verificationToken string) error
}

// service implements Service interface
type service struct {
	config *config.EmailConfig
}

// NewService creates a new email service
func NewService(cfg *config.EmailConfig) Service {
	return &service{
		config: cfg,
	}
}

// SendPasswordResetEmail sends a password reset email
func (s *service) SendPasswordResetEmail(toEmail, resetToken string) error {
	if !s.config.Enabled {
		fmt.Printf("[DEV MODE] Password reset email to %s with token %s\n", toEmail, resetToken)
		return nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.config.FrontendURL, resetToken)
	subject := "Password Reset Request"
	htmlContent, err := s.generateResetEmailHTML(resetURL)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	return s.sendEmail(toEmail, subject, htmlContent)
}

// SendInvitationEmail sends an organization invitation email
func (s *service) SendInvitationEmail(toEmail, inviterName, organizationName, invitationToken string) error {
	if !s.config.Enabled {
		fmt.Printf("[DEV MODE] Invitation email to %s from %s for %s with token %s\n",
			toEmail, inviterName, organizationName, invitationToken)
		return nil
	}

	// URL-encode the email to handle special characters like +
	encodedEmail := url.QueryEscape(toEmail)
	invitationURL := fmt.Sprintf("%s/accept-invitation?token=%s&email=%s", s.config.FrontendURL, invitationToken, encodedEmail)
	subject := fmt.Sprintf("You've been invited to join %s", organizationName)
	htmlContent, err := s.generateInvitationEmailHTML(inviterName, organizationName, invitationURL)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	return s.sendEmail(toEmail, subject, htmlContent)
}

// SendVerificationEmail sends an email verification email with 6-digit code
func (s *service) SendVerificationEmail(toEmail, verificationCode string) error {
	if !s.config.Enabled {
		fmt.Printf("[DEV MODE] 📧 Verification email to %s with code %s\n", toEmail, verificationCode)
		return nil
	}

	subject := "Verify Your Email Address"
	htmlContent, err := s.generateVerificationEmailHTML(toEmail, verificationCode)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	return s.sendEmail(toEmail, subject, htmlContent)
}

// sendEmail sends an email via SMTP or Resend API
func (s *service) sendEmail(to, subject, htmlBody string) error {
	// Check if using Resend API (RESEND_API_KEY is set)
	if s.config.ResendAPIKey != "" {
		return s.sendViaResend(to, subject, htmlBody)
	}

	// Fallback to SMTP
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Build email message
	msg := fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", s.config.FromName, s.config.FromEmail, to, subject, htmlBody)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendViaResend sends email using Resend SDK
func (s *service) sendViaResend(to, subject, htmlBody string) error {
	client := resend.NewClient(s.config.ResendAPIKey)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail),
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	fmt.Printf("✅ Email sent via Resend to %s (ID: %s)\n", to, sent.Id)
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

// generateInvitationEmailHTML generates HTML content for invitation email
func (s *service) generateInvitationEmailHTML(inviterName, organizationName, invitationURL string) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Organization Invitation</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9fafb;">
    <div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 20px; text-align: center; border-radius: 8px 8px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 28px;">You're Invited!</h1>
    </div>
    <div style="background: white; padding: 40px; border-radius: 0 0 8px 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <p style="font-size: 16px; color: #374151; line-height: 1.6;">Hi there,</p>
        <p style="font-size: 16px; color: #374151; line-height: 1.6;">
            <strong>{{.InviterName}}</strong> has invited you to join <strong>{{.OrganizationName}}</strong>.
        </p>
        <p style="font-size: 16px; color: #374151; line-height: 1.6;">
            Click the button below to accept the invitation and get started:
        </p>
        <div style="text-align: center; margin: 40px 0;">
            <a href="{{.InvitationURL}}" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 16px 32px; text-decoration: none; border-radius: 8px; display: inline-block; font-weight: 600; font-size: 16px;">Accept Invitation</a>
        </div>
        <p style="font-size: 14px; color: #6b7280; line-height: 1.6;">
            If you didn't expect this invitation, you can safely ignore this email.
        </p>
        <p style="font-size: 14px; color: #6b7280; line-height: 1.6;">
            This invitation will expire in 7 days.
        </p>
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #e5e7eb;">
        <p style="font-size: 12px; color: #9ca3af;">
            If the button doesn't work, copy and paste this URL into your browser:
        </p>
        <p style="word-break: break-all; color: #667eea; font-size: 12px;">{{.InvitationURL}}</p>
    </div>
    <div style="text-align: center; margin-top: 20px; color: #9ca3af; font-size: 12px;">
        <p>This email was sent by {{.FromName}}</p>
    </div>
</body>
</html>`

	t, err := template.New("invitationEmail").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		InviterName      string
		OrganizationName string
		InvitationURL    string
		FromName         string
	}{
		InviterName:      inviterName,
		OrganizationName: organizationName,
		InvitationURL:    invitationURL,
		FromName:         s.config.FromName,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateVerificationEmailHTML generates HTML content for email verification with 6-digit code
func (s *service) generateVerificationEmailHTML(toEmail, verificationCode string) (string, error) {
	verifyURL := fmt.Sprintf("%s/auth/verify-email?email=%s", s.config.FrontendURL, toEmail)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9fafb;">
    <div style="background: linear-gradient(135deg, #10b981 0%, #059669 100%); padding: 40px 20px; text-align: center; border-radius: 8px 8px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 28px;">Verify Your Email</h1>
    </div>
    <div style="background: white; padding: 40px; border-radius: 0 0 8px 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <p style="font-size: 16px; color: #374151; line-height: 1.6;">Thank you for registering!</p>
        <p style="font-size: 16px; color: #374151; line-height: 1.6;">
            To complete your registration, please enter the verification code below:
        </p>
        <div style="text-align: center; margin: 40px 0;">
            <div style="background: #f3f4f6; border: 2px dashed #10b981; border-radius: 8px; padding: 24px; display: inline-block;">
                <div style="font-size: 14px; color: #6b7280; margin-bottom: 8px; text-transform: uppercase; letter-spacing: 1px; font-weight: 600;">Your Verification Code</div>
                <div style="font-size: 36px; font-weight: bold; color: #059669; letter-spacing: 8px; font-family: 'Courier New', monospace;">{{.VerificationCode}}</div>
            </div>
        </div>
        <p style="text-align: center; margin: 30px 0;">
            <a href="{{.VerifyURL}}" style="background-color: #10b981; color: white; padding: 14px 32px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: 600; font-size: 16px;">Verify Email Now</a>
        </p>
        <p style="font-size: 14px; color: #6b7280; line-height: 1.6;">
            Or copy and paste this link in your browser:
        </p>
        <p style="font-size: 12px; color: #10b981; word-break: break-all; background: #f3f4f6; padding: 12px; border-radius: 4px; font-family: monospace;">
            {{.VerifyURL}}
        </p>
        <p style="font-size: 14px; color: #6b7280; line-height: 1.6; margin-top: 30px;">
            This code will expire in <strong>15 minutes</strong> for security reasons.
        </p>
        <p style="font-size: 14px; color: #6b7280; line-height: 1.6;">
            If you didn't create an account, you can safely ignore this email.
        </p>
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #e5e7eb;">
        <p style="font-size: 12px; color: #9ca3af;">
            <strong>Security Tip:</strong> Never share this code with anyone. Our team will never ask for your verification code.
        </p>
    </div>
    <div style="text-align: center; margin-top: 20px; color: #9ca3af; font-size: 12px;">
        <p>This email was sent by {{.FromName}}</p>
    </div>
</body>
</html>`

	t, err := template.New("verificationEmail").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		VerificationCode string
		VerifyURL        string
		FromName         string
	}{
		VerificationCode: verificationCode,
		VerifyURL:        verifyURL,
		FromName:         s.config.FromName,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
