// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Custom OTel Collector binary with the prompt vault processor.
package main

import (
	"log"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"

	"github.com/nostalgicskinco/prompt-vault-processor/processor/promptvaultprocessor"
)

func main() {
	info := component.BuildInfo{
		Command:     "otelcol-promptvault",
		Description: "OTel Collector with Prompt Vault content offload processor",
		Version:     "0.1.0",
	}

	set := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: factories,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{"file:./examples/otelcol-config.yaml"},
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
					envprovider.NewFactory(),
				},
			},
		},
		LoggingOptions: []zap.Option{},
	}

	cmd := otelcol.NewCommand(set)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func factories() (otelcol.Factories, error) {
	procs, err := processor.MakeFactoryMap(
		promptvaultprocessor.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return otelcol.Factories{
		Processors: procs,
	}, nil
}
