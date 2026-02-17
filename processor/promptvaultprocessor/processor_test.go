package promptvaultprocessor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestVaultReplacesContent(t *testing.T) {
	tmpDir := t.TempDir()
	vault, err := NewFilesystemVault(tmpDir)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	cfg := createDefaultConfig()
	cfg.Storage.Filesystem.BasePath = tmpDir
	sink := new(consumertest.TracesSink)
	proc := newVaultProcessor(zap.NewNop(), cfg, vault, sink)

	td := ptrace.NewTraces()
	span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetName("chat")
	span.Attributes().PutStr("gen_ai.prompt", "Tell me about quantum computing")
	span.Attributes().PutStr("gen_ai.completion", "Quantum computing uses qubits...")

	err = proc.ConsumeTraces(context.Background(), td)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	attrs := sink.AllTraces()[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes()

	prompt, _ := attrs.Get("gen_ai.prompt")
	if !strings.HasPrefix(prompt.Str(), "vault://") {
		t.Errorf("expected gen_ai.prompt to be vault ref, got: %s", prompt.Str())
	}

	completion, _ := attrs.Get("gen_ai.completion")
	if !strings.HasPrefix(completion.Str(), "vault://") {
		t.Errorf("expected gen_ai.completion to be vault ref, got: %s", completion.Str())
	}

	// Check vault ref attributes were added
	promptRef, ok := attrs.Get("gen_ai.prompt.vault_ref")
	if !ok {
		t.Error("expected gen_ai.prompt.vault_ref to exist")
	}
	if !strings.HasPrefix(promptRef.Str(), "vault://") {
		t.Errorf("expected vault ref format, got: %s", promptRef.Str())
	}
}func TestVaultWritesToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	vault, _ := NewFilesystemVault(tmpDir)

	ref, err := vault.Store([]byte("Hello, World!"))
	if err != nil {
		t.Fatalf("vault store failed: %v", err)
	}

	if !strings.HasPrefix(ref, "vault://") {
		t.Errorf("expected vault:// prefix, got: %s", ref)
	}

	// Check file exists on disk
	found := false
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".vault") {
			found = true
		}
		return nil
	})
	if !found {
		t.Error("expected vault file to exist on disk")
	}
}

func TestVaultDeduplication(t *testing.T) {
	tmpDir := t.TempDir()
	vault, _ := NewFilesystemVault(tmpDir)

	ref1, _ := vault.Store([]byte("duplicate content"))
	ref2, _ := vault.Store([]byte("duplicate content"))

	if ref1 != ref2 {
		t.Errorf("expected same ref for same content, got %s and %s", ref1, ref2)
	}
}func TestVaultSkipsSmallContent(t *testing.T) {
	tmpDir := t.TempDir()
	vault, _ := NewFilesystemVault(tmpDir)
	cfg := createDefaultConfig()
	cfg.Vault.SizeThreshold = 1000 // Only vault content > 1000 bytes
	sink := new(consumertest.TracesSink)
	proc := newVaultProcessor(zap.NewNop(), cfg, vault, sink)

	td := ptrace.NewTraces()
	span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.Attributes().PutStr("gen_ai.prompt", "short")

	proc.ConsumeTraces(context.Background(), td)

	attrs := sink.AllTraces()[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes()

	prompt, _ := attrs.Get("gen_ai.prompt")
	if prompt.Str() != "short" {
		t.Errorf("expected content under threshold to be untouched, got: %s", prompt.Str())
	}
}

func TestVaultRemoveMode(t *testing.T) {
	tmpDir := t.TempDir()
	vault, _ := NewFilesystemVault(tmpDir)
	cfg := createDefaultConfig()
	cfg.Vault.Mode = "remove"
	sink := new(consumertest.TracesSink)
	proc := newVaultProcessor(zap.NewNop(), cfg, vault, sink)

	td := ptrace.NewTraces()
	span := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.Attributes().PutStr("gen_ai.prompt", "sensitive content here")

	proc.ConsumeTraces(context.Background(), td)

	attrs := sink.AllTraces()[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes()

	if _, ok := attrs.Get("gen_ai.prompt"); ok {
		t.Error("expected gen_ai.prompt to be removed in remove mode")
	}

	if _, ok := attrs.Get("gen_ai.prompt.vault_ref"); !ok {
		t.Error("expected gen_ai.prompt.vault_ref to exist even in remove mode")
	}
}func TestVaultRetrieve(t *testing.T) {
	tmpDir := t.TempDir()
	vault, _ := NewFilesystemVault(tmpDir)

	original := "This is the content to vault and retrieve"
	ref, err := vault.Store([]byte(original))
	if err != nil {
		t.Fatalf("store failed: %v", err)
	}

	data, err := vault.Retrieve(ref)
	if err != nil {
		t.Fatalf("retrieve failed: %v", err)
	}

	if string(data) != original {
		t.Errorf("expected %q, got %q", original, string(data))
	}
}