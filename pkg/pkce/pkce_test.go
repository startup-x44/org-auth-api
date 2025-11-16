package pkce

import (
	"strings"
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("GenerateCodeVerifier failed: %v", err)
	}

	if len(verifier) < CodeVerifierMinLength || len(verifier) > CodeVerifierMaxLength {
		t.Errorf("verifier length %d is outside valid range [%d, %d]", len(verifier), CodeVerifierMinLength, CodeVerifierMaxLength)
	}

	// Verify it's base64url encoded (no special characters except - and _)
	for _, char := range verifier {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			t.Errorf("verifier contains invalid character: %c", char)
		}
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

	challenge, err := GenerateCodeChallenge(verifier)
	if err != nil {
		t.Fatalf("GenerateCodeChallenge failed: %v", err)
	}

	// Verify it's a non-empty base64url string
	if challenge == "" {
		t.Error("challenge is empty")
	}

	// Verify no padding
	if strings.HasSuffix(challenge, "=") {
		t.Error("challenge should not have padding")
	}
}

func TestVerifyCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge, _ := GenerateCodeChallenge(verifier)

	tests := []struct {
		name      string
		verifier  string
		challenge string
		method    string
		wantErr   bool
	}{
		{
			name:      "valid S256",
			verifier:  verifier,
			challenge: challenge,
			method:    "S256",
			wantErr:   false,
		},
		{
			name:      "invalid method",
			verifier:  verifier,
			challenge: challenge,
			method:    "plain",
			wantErr:   true,
		},
		{
			name:      "mismatched verifier",
			verifier:  "different-verifier-that-is-long-enough-1234567890",
			challenge: challenge,
			method:    "S256",
			wantErr:   true,
		},
		{
			name:      "verifier too short",
			verifier:  "short",
			challenge: challenge,
			method:    "S256",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyCodeChallenge(tt.verifier, tt.challenge, tt.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyCodeChallenge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeneratePKCEPair(t *testing.T) {
	verifier, challenge, err := GeneratePKCEPair()
	if err != nil {
		t.Fatalf("GeneratePKCEPair failed: %v", err)
	}

	if verifier == "" || challenge == "" {
		t.Error("verifier or challenge is empty")
	}

	// Verify the pair is valid
	err = VerifyCodeChallenge(verifier, challenge, "S256")
	if err != nil {
		t.Errorf("generated PKCE pair is invalid: %v", err)
	}
}

func TestMultipleVerifiersDifferent(t *testing.T) {
	verifier1, _ := GenerateCodeVerifier()
	verifier2, _ := GenerateCodeVerifier()

	if verifier1 == verifier2 {
		t.Error("multiple calls to GenerateCodeVerifier should produce different values")
	}
}
