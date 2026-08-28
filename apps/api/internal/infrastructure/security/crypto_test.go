package security

import (
	"strings"
	"testing"
)

// ============================================================================
// Happy Path
// ============================================================================

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := "01234567890123456789012345678901" // exactly 32 bytes
	plaintext := "Hello, Dboke!"

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted == plaintext {
		t.Error("Encrypted text should not equal plaintext")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	key := "01234567890123456789012345678901"
	plaintext := ""

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt of empty string failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt of empty string failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_LargePayload(t *testing.T) {
	key := "01234567890123456789012345678901"
	plaintext := strings.Repeat("A", 100_000) // 100KB

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt of large payload failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt of large payload failed: %v", err)
	}

	if decrypted != plaintext {
		t.Error("Large payload round-trip failed")
	}
}

func TestEncryptDecrypt_SpecialCharacters(t *testing.T) {
	key := "01234567890123456789012345678901"
	cases := []string{
		`{"host":"localhost","port":5432,"password":"p@$$w0rd!"}`,
		"こんにちは世界",              // Japanese unicode
		"SELECT * FROM users WHERE id = 1; DROP TABLE users;--", // SQL injection-like
		"\x00\x01\x02\x03",       // binary data
		"\n\t\r",                  // whitespace characters
		"🔑🗝️🔐",                // emoji
	}

	for _, tc := range cases {
		encrypted, err := Encrypt(tc, key)
		if err != nil {
			t.Fatalf("Encrypt(%q) failed: %v", tc, err)
		}

		decrypted, err := Decrypt(encrypted, key)
		if err != nil {
			t.Fatalf("Decrypt failed for input %q: %v", tc, err)
		}

		if decrypted != tc {
			t.Errorf("Round-trip failed: got %q, want %q", decrypted, tc)
		}
	}
}

func TestEncrypt_ProducesUniqueOutput(t *testing.T) {
	key := "01234567890123456789012345678901"
	plaintext := "same-input"

	enc1, err1 := Encrypt(plaintext, key)
	enc2, err2 := Encrypt(plaintext, key)

	if err1 != nil || err2 != nil {
		t.Fatalf("Encrypt errors: %v, %v", err1, err2)
	}

	// GCM uses random nonce, so two encryptions of the same plaintext should differ
	if enc1 == enc2 {
		t.Error("Two encryptions of the same plaintext produced identical ciphertext — nonce reuse vulnerability")
	}
}

func TestEncrypt_OutputIsHex(t *testing.T) {
	key := "01234567890123456789012345678901"
	encrypted, err := Encrypt("test", key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	for _, c := range encrypted {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Encrypted output contains non-hex character: %c", c)
			break
		}
	}
}

// ============================================================================
// Error Cases & Edge Cases
// ============================================================================

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	cases := []struct {
		name string
		key  string
	}{
		{"empty key", ""},
		{"short key (16 bytes)", "0123456789012345"},
		{"long key (64 bytes)", strings.Repeat("A", 64)},
		{"1 byte key", "A"},
		{"31 bytes", strings.Repeat("A", 31)},
		{"33 bytes", strings.Repeat("A", 33)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Encrypt("test", tc.key)
			if err == nil {
				t.Error("Expected error for invalid key length, got nil")
			}
		})
	}
}

func TestDecrypt_InvalidKeyLength(t *testing.T) {
	// First encrypt with valid key
	validKey := "01234567890123456789012345678901"
	encrypted, _ := Encrypt("test", validKey)

	cases := []struct {
		name string
		key  string
	}{
		{"empty key", ""},
		{"short key", "short"},
		{"31 byte key", strings.Repeat("B", 31)},
		{"33 byte key", strings.Repeat("B", 33)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(encrypted, tc.key)
			if err == nil {
				t.Error("Expected error for invalid key length, got nil")
			}
		})
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := "01234567890123456789012345678901"
	key2 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"

	encrypted, err := Encrypt("secret data", key1)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key, got nil")
	}
}

func TestDecrypt_InvalidHex(t *testing.T) {
	key := "01234567890123456789012345678901"
	cases := []struct {
		name  string
		input string
	}{
		{"not hex", "this-is-not-hex-data!!"},
		{"odd length hex", "abc"},
		{"invalid hex chars", "ZZZZZZ"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(tc.input, key)
			if err == nil {
				t.Errorf("Expected error for invalid hex input %q, got nil", tc.input)
			}
		})
	}
}

func TestDecrypt_TruncatedCiphertext(t *testing.T) {
	key := "01234567890123456789012345678901"

	encrypted, err := Encrypt("test data", key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Truncate to just a few hex chars (less than nonce size)
	if len(encrypted) > 4 {
		truncated := encrypted[:4]
		_, err = Decrypt(truncated, key)
		if err == nil {
			t.Error("Expected error for truncated ciphertext, got nil")
		}
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := "01234567890123456789012345678901"

	encrypted, err := Encrypt("test data", key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Flip last hex character
	runes := []rune(encrypted)
	if runes[len(runes)-1] == 'a' {
		runes[len(runes)-1] = 'b'
	} else {
		runes[len(runes)-1] = 'a'
	}
	tampered := string(runes)

	_, err = Decrypt(tampered, key)
	if err == nil {
		t.Error("Expected error for tampered ciphertext (GCM integrity check), got nil")
	}
}

func TestDecrypt_EmptyCiphertext(t *testing.T) {
	key := "01234567890123456789012345678901"
	_, err := Decrypt("", key)
	if err == nil {
		t.Error("Expected error for empty ciphertext, got nil")
	}
}
