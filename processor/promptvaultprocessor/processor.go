package promptvaultprocessor

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type vaultProcessor struct {
	logger       *zap.Logger
	config       *Config
	vault        VaultStorage
	nextConsumer consumer.Traces
	keysSet      map[string]bool
}

func newVaultProcessor(
	logger *zap.Logger,
	cfg *Config,
	vault VaultStorage,
	next consumer.Traces,
) *vaultProcessor {
	keysSet := make(map[string]bool, len(cfg.Vault.Keys))
	for _, k := range cfg.Vault.Keys {
		keysSet[k] = true
	}

	return &vaultProcessor{
		logger:       logger,
		config:       cfg,
		vault:        vault,
		nextConsumer: next,
		keysSet:      keysSet,
	}
}

func (p *vaultProcessor) Start(_ context.Context, _ component.Host) error {
	p.logger.Info("promptvault processor started",
		zap.Int("vault_keys", len(p.keysSet)),
		zap.String("mode", p.config.Vault.Mode),
		zap.String("backend", p.config.Storage.Backend),
	)
	return nil
}

func (p *vaultProcessor) Shutdown(_ context.Context) error {
	return nil
}func (p *vaultProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *vaultProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		ilss := rss.At(i).ScopeSpans()
		for j := 0; j < ilss.Len(); j++ {
			spans := ilss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				p.vaultSpan(spans.At(k))
			}
		}
	}
	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *vaultProcessor) vaultSpan(span ptrace.Span) {
	attrs := span.Attributes()

	// Collect keys to vault (can't modify map while iterating)
	type vaultEntry struct {
		key     string
		content string
	}
	var toVault []vaultEntry

	attrs.Range(func(key string, val ptrace.Value) bool {
		if !p.keysSet[key] {
			return true
		}

		content := val.Str()
		if len(content) < p.config.Vault.SizeThreshold {
			return true
		}

		toVault = append(toVault, vaultEntry{key: key, content: content})
		return true
	})

	for _, entry := range toVault {
		ref, err := p.vault.Store([]byte(entry.content))
		if err != nil {
			p.logger.Warn("vault store failed",
				zap.String("key", entry.key),
				zap.Error(err),
			)
			continue
		}

		switch p.config.Vault.Mode {
		case "replace_with_ref":
			attrs.PutStr(entry.key, ref)
			attrs.PutStr(entry.key+".vault_ref", ref)
		case "remove":
			attrs.Remove(entry.key)
			attrs.PutStr(entry.key+".vault_ref", ref)
		}

		p.logger.Debug("vaulted attribute",
			zap.String("key", entry.key),
			zap.String("ref", ref),
			zap.Int("content_bytes", len(entry.content)),
		)
	}
}