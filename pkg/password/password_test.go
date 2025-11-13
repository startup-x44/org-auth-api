package password_test

import (
	"testing"

	"auth-service/pkg/password"
)

func TestPasswordService(t *testing.T) {
	svc := password.NewService()

	// Test password hashing
	plainPassword := "TestPassword123!"
	hash, err := svc.Hash(plainPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == plainPassword {
		t.Error("Hash should not equal plain password")
	}

	// Test password verification
	if valid, err := svc.Verify(plainPassword, hash); err != nil || !valid {
		t.Fatalf("Failed to verify correct password: %v", err)
	}

	if valid, err := svc.Verify("WrongPassword123!", hash); err == nil && valid {
		t.Error("Should fail to verify wrong password")
	}

	// Test password needs rehash (should be false for new hash)
	if svc.NeedsRehash(hash) {
		t.Error("Newly created hash should not need rehash")
	}
}

func TestPasswordValidation(t *testing.T) {
	svc := password.NewService()

	validPasswords := []string{
		"ValidPass123!",
		"StrongPassword456@",
		"Complex!789#Pass",
	}

	invalidPasswords := []string{
		"short",           // too short
		"nouppercase123!", // no uppercase
		"NOLOWERCASE123!", // no lowercase
		"NoNumbers!",      // no numbers
		"NoSpecial123",    // no special characters
	}

	for _, pwd := range validPasswords {
		hash, err := svc.Hash(pwd)
		if err != nil {
			t.Fatalf("Failed to hash valid password %s: %v", pwd, err)
		}
		if valid, err := svc.Verify(pwd, hash); err != nil || !valid {
			t.Fatalf("Failed to verify valid password %s: %v", pwd, err)
		}
	}

	for _, pwd := range invalidPasswords {
		_, err := svc.Hash(pwd)
		if err == nil {
			t.Fatalf("Should fail to hash invalid password: %s", pwd)
		}
	}
}

// BenchmarkPasswordVerify benchmarks password verification performance
func BenchmarkPasswordVerify(b *testing.B) {
	svc := password.NewService()
	password := "BenchmarkTestPassword123!"
	hash, err := svc.Hash(password)
	if err != nil {
		b.Fatalf("Failed to hash password: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid, err := svc.Verify(password, hash)
		if err != nil || !valid {
			b.Fatalf("Failed to verify password: %v", err)
		}
	}
}

// BenchmarkPasswordVerifyWrong benchmarks password verification with wrong password
func BenchmarkPasswordVerifyWrong(b *testing.B) {
	svc := password.NewService()
	password := "BenchmarkTestPassword123!"
	hash, err := svc.Hash(password)
	if err != nil {
		b.Fatalf("Failed to hash password: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid, err := svc.Verify("WrongPassword123!", hash)
		// We expect this to fail, but we still want to measure the performance
		_ = valid
		_ = err
	}
}

// BenchmarkPasswordHash benchmarks password hashing performance
func BenchmarkPasswordHash(b *testing.B) {
	svc := password.NewService()
	password := "BenchmarkTestPassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := svc.Hash(password)
		if err != nil {
			b.Fatalf("Failed to hash password: %v", err)
		}
	}
}