package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigDir returns the standard Linux config directory (~/.config/agef).
func DefaultConfigDir() (string, error) {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, "agef"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "agef"), nil
}

// EnsureConfigDir guarantees that the config directory exists with 0700 permissions.
func EnsureConfigDir() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// PrivateKeyPath returns the path to the standard private key.
func PrivateKeyPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "key.txt"), nil
}

// PublicKeyPath returns the path to the standard public key.
func PublicKeyPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "key.pub"), nil
}

// HasStandardKeys checks whether both private and public keys exist in ~/.config/agef.
func HasStandardKeys() bool {
	priv, err1 := PrivateKeyPath()
	pub, err2 := PublicKeyPath()
	if err1 != nil || err2 != nil {
		return false
	}

	infoPriv, err1 := os.Stat(priv)
	infoPub, err2 := os.Stat(pub)
	return err1 == nil && !infoPriv.IsDir() && err2 == nil && !infoPub.IsDir()
}
