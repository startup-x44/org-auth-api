package main

import (
	"fmt"
	"log"
	"time"

	"auth-service/internal/config"
	"auth-service/pkg/email"
)

func main() {
	// Load config
	cfg := config.Load()

	fmt.Println("📧 Testing Resend Email Integration")
	fmt.Println("=====================================")
	fmt.Printf("Using API Key: %s...\n", cfg.Email.ResendAPIKey[:10])
	fmt.Printf("From: %s <%s>\n", cfg.Email.FromName, cfg.Email.FromEmail)
	fmt.Println()

	// Create email service
	emailService := email.NewService(&cfg.Email)

	// Test 1: Verification Email
	fmt.Println("Test 1: Sending verification email...")
	err := emailService.SendVerificationEmail("bagaresnilo93@gmail.com", "123456")
	if err != nil {
		log.Fatalf("❌ Failed to send verification email: %v", err)
	}
	fmt.Println("✅ Verification email sent successfully!")
	fmt.Println()

	time.Sleep(1 * time.Second) // Rate limit: 2 req/sec

	// Test 2: Password Reset Email
	fmt.Println("Test 2: Sending password reset email...")
	err = emailService.SendPasswordResetEmail("bagaresnilo93@gmail.com", "test-reset-token-123")
	if err != nil {
		log.Fatalf("❌ Failed to send password reset email: %v", err)
	}
	fmt.Println("✅ Password reset email sent successfully!")
	fmt.Println()

	time.Sleep(1 * time.Second) // Rate limit: 2 req/sec

	// Test 3: Invitation Email
	fmt.Println("Test 3: Sending invitation email...")
	err = emailService.SendInvitationEmail("bagaresnilo93@gmail.com", "Nilo", "NiloAuth", "test-invitation-token-456")
	if err != nil {
		log.Fatalf("❌ Failed to send invitation email: %v", err)
	}
	fmt.Println("✅ Invitation email sent successfully!")
	fmt.Println()

	fmt.Println("🎉 All email tests passed! Check bagaresnilo93@gmail.com inbox.")
}
