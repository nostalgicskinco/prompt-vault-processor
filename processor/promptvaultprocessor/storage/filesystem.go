// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// FilesystemBackend stores vault content on the local filesystem.
type FilesystemBackend struct {
	basePath string
}

// NewFilesystemBackend creates a new filesystem storage backend.
func NewFilesystemBackend(basePath string) (*FilesystemBackend, error) {
	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create vault directory: %w", err)
	}
	return &FilesystemBackend{basePath: basePath}, nil
}

// Store writes data to the filesystem keyed by trace/span/attr.
// attrKey may contain path separators (e.g. for event attributes:
// "spanID/event_0/gen_ai.prompt"), so we ensure all intermediate
// directories exist before writing.
func (f *FilesystemBackend) Store(_ context.Context, traceID, spanID, attrKey string, data []byte) (Reference, error) {
	filePath := filepath.Join(f.basePath, traceID, spanID, attrKey)
	fileDir := filepath.Dir(filePath)
	if err := os.MkdirAll(fileDir, 0750); err != nil {
		return Reference{}, fmt.Errorf("failed to create object directory: %w", err)
	}

	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	if err := os.WriteFile(filePath, data, 0640); err != nil {
		return Reference{}, fmt.Errorf("failed to write vault object: %w", err)
	}

	ref := Reference{
		URI:       fmt.Sprintf("promptvault://fs/%s/%s/%s", traceID, spanID, attrKey),
		Checksum:  checksum,
		Encrypted: false,
		SizeBytes: len(data),
	}
	return ref, nil
}

// Retrieve reads content back from the filesystem.
func (f *FilesystemBackend) Retrieve(_ context.Context, ref Reference) ([]byte, error) {
	// Parse the URI to get the path components.
	// URI format: promptvault://fs/{traceID}/{spanID}/{attrKey}
	var traceID, spanID, attrKey string
	_, err := fmt.Sscanf(ref.URI, "promptvault://fs/%s", &traceID)
	if err != nil {
		// Fallback: try to extract path parts manually.
		return nil, fmt.Errorf("invalid vault reference URI: %s", ref.URI)
	}

	// Simple path extraction from URI.
	path := ref.URI[len("promptvault://fs/"):]
	filePath := filepath.Join(f.basePath, path)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault object: %w", err)
	}

	// Verify checksum.
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])
	if checksum != ref.Checksum {
		return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", ref.Checksum, checksum)
	}

	_ = traceID
	_ = spanID
	_ = attrKey

	return data, nil
}

// Close is a no-op for filesystem backend.
func (f *FilesystemBackend) Close() error {
	return nil
}
