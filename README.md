# prompt-vault-processor

[![CI](https://github.com/airblackbox/otel-prompt-vault/actions/workflows/ci.yml/badge.svg)](https://github.com/airblackbox/otel-prompt-vault/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/airblackbox/otel-prompt-vault/blob/main/LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)


An OpenTelemetry Collector processor that offloads LLM prompt and completion content to a secure vault, replacing sensitive data in traces with vault references.

## Why

Traces should contain **references**, not content. This processor:

1. Intercepts spans with LLM prompt/completion attributes
2. Writes the content to a storage backend (filesystem or S3)
3. Replaces the attribute value with a `vault://` reference
4. Downstream systems see references, never raw content

## Configuration

```yaml
processors:
  promptvault:
    storage:
      backend: filesystem
      filesystem:
        base_path: /data/vault
    vault:
      keys:
        - gen_ai.prompt
        - gen_ai.completion
        - gen_ai.system_instructions
      size_threshold: 0        # 0 = vault everything
      mode: replace_with_ref   # or "remove"
```

## Modes

| Mode | Behavior |
|------|----------|
| `replace_with_ref` | Replaces content with `vault://sha256hash` |
| `remove` | Removes the attribute entirely, adds `.vault_ref` attribute |

## Part of the AIR Platform

This processor is one component of the [AIR Blackbox Gateway](https://github.com/airblackbox/gateway) collector pipeline.

## License

Apache-2.0