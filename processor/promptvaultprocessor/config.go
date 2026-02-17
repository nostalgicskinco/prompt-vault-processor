package promptvaultprocessor

// Config for the prompt vault processor.
type Config struct {
	Storage StorageConfig `mapstructure:"storage"`
	Vault   VaultConfig   `mapstructure:"vault"`
}

// StorageConfig defines where vaulted content is stored.
type StorageConfig struct {
	Backend    string           `mapstructure:"backend"` // "filesystem" or "s3"
	Filesystem FilesystemConfig `mapstructure:"filesystem"`
}

// FilesystemConfig for local file-based vault storage.
type FilesystemConfig struct {
	BasePath string `mapstructure:"base_path"`
}

// VaultConfig controls which attributes get vaulted.
type VaultConfig struct {
	// Keys lists the attribute keys whose values should be vaulted.
	Keys []string `mapstructure:"keys"`
	// SizeThreshold: only vault values larger than this (bytes). 0 = vault everything.
	SizeThreshold int `mapstructure:"size_threshold"`
	// Mode: "replace_with_ref" replaces value with vault://ref, "remove" deletes the attr.
	Mode string `mapstructure:"mode"`
}

func createDefaultConfig() *Config {
	return &Config{
		Storage: StorageConfig{
			Backend: "filesystem",
			Filesystem: FilesystemConfig{
				BasePath: "/data/vault",
			},
		},
		Vault: VaultConfig{
			Keys: []string{
				"gen_ai.prompt",
				"gen_ai.completion",
				"gen_ai.system_instructions",
				"gen_ai.input.messages",
				"gen_ai.output.messages",
			},
			SizeThreshold: 0,
			Mode:          "replace_with_ref",
		},
	}
}