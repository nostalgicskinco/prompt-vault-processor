// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// promptvaultctl is a CLI to fetch, decrypt, and display vault-stored content.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nostalgicskinco/prompt-vault-processor/processor/promptvaultprocessor/crypto"
	"github.com/nostalgicskinco/prompt-vault-processor/processor/promptvaultprocessor/storage"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: promptvaultctl get <vault-ref-json>\n")
		fmt.Fprintf(os.Stderr, "       promptvaultctl get-file <base-path> <vault-ref-json>\n")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "get", "get-file":
		if err := runGet(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func runGet(args []string) error {
	var basePath, refJSON string

	if len(args) == 1 {
		basePath = "/tmp/promptvault"
		refJSON = args[0]
	} else if len(args) == 2 {
		basePath = args[0]
		refJSON = args[1]
	} else {
		return fmt.Errorf("expected 1-2 arguments, got %d", len(args))
	}

	var ref storage.Reference
	if err := json.Unmarshal([]byte(refJSON), &ref); err != nil {
		return fmt.Errorf("failed to parse vault reference: %w", err)
	}

	be, err := storage.NewFilesystemBackend(basePath)
	if err != nil {
		return fmt.Errorf("failed to open vault: %w", err)
	}
	defer be.Close()

	data, err := be.Retrieve(context.Background(), ref)
	if err != nil {
		return fmt.Errorf("failed to retrieve content: %w", err)
	}

	// Decrypt if needed.
	if ref.Encrypted {
		hexKey := os.Getenv("PROMPTVAULT_KEY")
		hmacSecret := os.Getenv("PROMPTVAULT_HMAC_SECRET")
		if hexKey == "" {
			return fmt.Errorf("PROMPTVAULT_KEY env var required for encrypted content")
		}
		env, err := crypto.NewEnvelope(hexKey, hmacSecret)
		if err != nil {
			return fmt.Errorf("failed to init decryption: %w", err)
		}
		data, err = env.Decrypt(data)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}
	}

	fmt.Println(string(data))
	return nil
}
