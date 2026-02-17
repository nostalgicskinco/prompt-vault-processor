package promptvaultprocessor

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// VaultStorage handles persisting content to a backend.
type VaultStorage interface {
	Store(content []byte) (ref string, err error)
}

// FilesystemVault stores content as files on disk.
type FilesystemVault struct {
	basePath string
}

// NewFilesystemVault creates a new filesystem-based vault.
func NewFilesystemVault(basePath string) (*FilesystemVault, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("create vault dir: %w", err)
	}
	return &FilesystemVault{basePath: basePath}, nil
}

// Store writes content to a file and returns a vault reference.
// The reference format is: vault://<sha256>
func (v *FilesystemVault) Store(content []byte) (string, error) {
	hash := sha256.Sum256(content)
	hexHash := fmt.Sprintf("%x", hash)

	// Use date-partitioned directories for organization
	now := time.Now().UTC()
	dir := filepath.Join(v.basePath, now.Format("2006/01/02"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create date dir: %w", err)
	}

	path := filepath.Join(dir, hexHash+".vault")

	// Deduplicate: if same hash exists, skip write
	if _, err := os.Stat(path); err == nil {
		return fmt.Sprintf("vault://%s", hexHash), nil
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", fmt.Errorf("write vault file: %w", err)
	}

	return fmt.Sprintf("vault://%s", hexHash), nil
}// Retrieve reads content back from the vault by reference.
func (v *FilesystemVault) Retrieve(ref string) ([]byte, error) {
	// Walk the vault looking for the hash file
	hexHash := ref
	if len(ref) > 8 && ref[:8] == "vault://" {
		hexHash = ref[8:]
	}

	var found string
	err := filepath.Walk(v.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() && info.Name() == hexHash+".vault" {
			found = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil || found == "" {
		return nil, fmt.Errorf("vault ref not found: %s", ref)
	}

	return os.ReadFile(found)
}