package crypto_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/agef/internal/crypto"
	"github.com/user/agef/internal/keys"
)

func TestEncryptAndDecryptFolder(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agef_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Generate keys
	keysDir := filepath.Join(tempDir, "keys")
	privPath, pubPath, _, err := keys.GenerateKeypair(keysDir)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	recip, err := keys.LoadRecipient(pubPath)
	if err != nil {
		t.Fatalf("failed to load recipient: %v", err)
	}

	identity, err := keys.LoadIdentity(privPath)
	if err != nil {
		t.Fatalf("failed to load identity: %v", err)
	}

	// 2. Setup source directory
	srcDir := filepath.Join(tempDir, "my_source_folder")
	if err := os.MkdirAll(filepath.Join(srcDir, "nested with spaces", "sub"), 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "empty_dir"), 0755); err != nil {
		t.Fatalf("failed to create empty dir: %v", err)
	}

	file1 := filepath.Join(srcDir, "hello.txt")
	file2 := filepath.Join(srcDir, ".secret_env")
	file3 := filepath.Join(srcDir, "nested with spaces", "document.md")
	file4 := filepath.Join(srcDir, "nested with spaces", "sub", "data.bin")

	content1 := []byte("Hello, World!")
	content2 := []byte("SECRET_API_TOKEN=xyz123")
	content3 := []byte("# Classified Notes\nConfidential.")
	content4 := []byte{0x00, 0xFF, 0x12, 0x34, 0xDE, 0xAD, 0xBE, 0xEF}

	_ = os.WriteFile(file1, content1, 0644)
	_ = os.WriteFile(file2, content2, 0600)
	_ = os.WriteFile(file3, content3, 0644)
	_ = os.WriteFile(file4, content4, 0644)

	// 3. Encrypt
	encDir := filepath.Join(tempDir, "my_source_folder_encrypted")
	encCount, err := crypto.EncryptFolder(srcDir, encDir, recip, nil)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	if encCount != 4 {
		t.Errorf("expected 4 files encrypted, got %d", encCount)
	}

	// Verify plaintext is NOT in ciphertext
	encFile1, _ := os.ReadFile(filepath.Join(encDir, "hello.txt.age"))
	if bytes.Contains(encFile1, content1) {
		t.Errorf("encrypted file contains plaintext!")
	}

	// Verify empty directory exists in encrypted folder
	if _, err := os.Stat(filepath.Join(encDir, "empty_dir")); os.IsNotExist(err) {
		t.Errorf("empty directory was not replicated in destination")
	}

	// 4. Decrypt
	decDir := filepath.Join(tempDir, "my_source_folder_decrypted")
	decCount, err := crypto.DecryptFolder(encDir, decDir, identity, nil)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if decCount != 4 {
		t.Errorf("expected 4 files decrypted, got %d", decCount)
	}

	// Verify decrypted contents
	dec1, _ := os.ReadFile(filepath.Join(decDir, "hello.txt"))
	if !bytes.Equal(dec1, content1) {
		t.Errorf("decrypted file1 mismatch")
	}

	dec2, _ := os.ReadFile(filepath.Join(decDir, ".secret_env"))
	if !bytes.Equal(dec2, content2) {
		t.Errorf("decrypted file2 mismatch")
	}

	dec3, _ := os.ReadFile(filepath.Join(decDir, "nested with spaces", "document.md"))
	if !bytes.Equal(dec3, content3) {
		t.Errorf("decrypted file3 mismatch")
	}

	dec4, _ := os.ReadFile(filepath.Join(decDir, "nested with spaces", "sub", "data.bin"))
	if !bytes.Equal(dec4, content4) {
		t.Errorf("decrypted file4 mismatch")
	}

	if _, err := os.Stat(filepath.Join(decDir, "empty_dir")); os.IsNotExist(err) {
		t.Errorf("empty directory was not replicated in decrypted folder")
	}
}
