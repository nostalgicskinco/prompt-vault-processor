package promptvaultprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	typeStr   = "promptvault"
	stability = component.StabilityLevelAlpha
)

// NewFactory creates a factory for the prompt vault processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		component.MustNewType(typeStr),
		func() component.Config { return createDefaultConfig() },
		processor.WithTraces(createTracesProcessor, stability),
	)
}

func createTracesProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	pCfg := cfg.(*Config)

	vault, err := NewFilesystemVault(pCfg.Storage.Filesystem.BasePath)
	if err != nil {
		return nil, err
	}

	return newVaultProcessor(set.Logger, pCfg, vault, nextConsumer), nil
}