package keys

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/user/agef/internal/config"
)

// GenerateKeypair generates a new native X25519 keypair and writes key.txt and key.pub.
func GenerateKeypair(destDir string) (string, string, string, error) {
	if destDir == "" {
		var err error
		destDir, err = config.EnsureConfigDir()
		if err != nil {
			return "", "", "", fmt.Errorf("failed to prepare config directory: %w", err)
		}
	} else {
		if err := os.MkdirAll(destDir, 0700); err != nil {
			return "", "", "", fmt.Errorf("failed to create target key directory: %w", err)
		}
	}

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate X25519 key: %w", err)
	}

	privKeyContent := fmt.Sprintf("# created with agef\n# public key: %s\n%s\n",
		identity.Recipient().String(), identity.String())
	pubKeyContent := fmt.Sprintf("%s\n", identity.Recipient().String())

	privPath := filepath.Join(destDir, "key.txt")
	pubPath := filepath.Join(destDir, "key.pub")

	if err := os.WriteFile(privPath, []byte(privKeyContent), 0600); err != nil {
		return "", "", "", fmt.Errorf("failed to save private key: %w", err)
	}

	if err := os.WriteFile(pubPath, []byte(pubKeyContent), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to save public key: %w", err)
	}

	return privPath, pubPath, identity.Recipient().String(), nil
}

// LoadRecipient parses a recipient from a string (age1...) or path to a public key / identity file.
func LoadRecipient(keyInput string) (age.Recipient, error) {
	keyInput = strings.TrimSpace(keyInput)

	if keyInput == "" {
		pubPath, err := config.PublicKeyPath()
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(pubPath); err != nil {
			return nil, fmt.Errorf("no default public key found at %s. Please initialize keys first", pubPath)
		}
		keyInput = pubPath
	}

	// Case 1: Direct public key string
	if strings.HasPrefix(keyInput, "age1") {
		return age.ParseX25519Recipient(keyInput)
	}

	// Case 2: File path
	f, err := os.Open(keyInput)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file %s: %w", keyInput, err)
	}
	defer f.Close()

	recipients, err := age.ParseRecipients(f)
	if err == nil && len(recipients) > 0 {
		return recipients[0], nil
	}

	// Fallback: If it was an identity file (private key), extract recipient from it
	f.Seek(0, 0)
	identities, err := age.ParseIdentities(f)
	if err == nil && len(identities) > 0 {
		if x25519Id, ok := identities[0].(*age.X25519Identity); ok {
			return x25519Id.Recipient(), nil
		}
	}

	return nil, fmt.Errorf("could not find a valid age recipient in %s", keyInput)
}

// LoadIdentity loads the private identity key from the specified file or default ~/.config/agef/key.txt.
func LoadIdentity(keyPath string) (age.Identity, error) {
	keyPath = strings.TrimSpace(keyPath)

	if keyPath == "" {
		var err error
		keyPath, err = config.PrivateKeyPath()
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(keyPath); err != nil {
			return nil, fmt.Errorf("no default private key found at %s. Please initialize keys first", keyPath)
		}
	}

	f, err := os.Open(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file %s: %w", keyPath, err)
	}
	defer f.Close()

	identities, err := age.ParseIdentities(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private identity file %s: %w", keyPath, err)
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no valid identities found in %s", keyPath)
	}

	return identities[0], nil
}

// ReadPublicKeyString returns the recipient string from the public key file.
func ReadPublicKeyString(pubPath string) (string, error) {
	f, err := os.Open(pubPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			return line, nil
		}
	}
	return "", fmt.Errorf("no public key found in %s", pubPath)
}
