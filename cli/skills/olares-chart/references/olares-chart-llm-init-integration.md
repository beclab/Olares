# Integrate `llm-init` into a custom chart

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. Choose this route only when the chart itself embeds `llm-init` as a serving sidecar or model downloader. Official generation clones use [llm-models.md](olares-chart-llm-models.md); embeddings use `olares-cli market install embeddinggemmav3` through [`olares-market`](../../olares-market/SKILL.md).

## Chart-author recipe

1. Add a pinned `beclab/llm-init` container to the chart.
2. Set `MODEL_NAME`, `MODEL_MODE`, and `MODEL_SOURCE`. Add optional raw `supports_*` values through `MODEL_SUPPORTS`.
3. For serving, set `ENGINE_KIND` to the chosen engine and give the engine container the same `ENGINE_ARGS`. Leave `ENGINE_KIND` empty only when the chart needs download-only behavior.
4. Mount the App Common Hugging Face cache at `/cache/hf/hub` and a per-pod run-state volume at `/run/llm-init`. Ensure UID 1000 can write both.
5. Make the consumer wait until `GET /readyz` succeeds or until the supported `/run/llm-init/model_path` sentinel is present, then read the resolved path.
6. Use the upstream `llm-init` documentation for the complete environment, source, engine, and API contract.

The four official generation charts expose convenience groups for `MODEL_SUPPORTS`: `vision`, `tools`, `thinking`, and `none`. Raw `llm-init` does not expand those groups. Its optional `MODEL_SUPPORTS` value is a comma-separated list of validated `supports_*` keys, such as `supports_reasoning,supports_tool_choice`; an unknown key fails startup.

## Serving and download-only

A serving chart sets `ENGINE_KIND` to `llamacpp`, `ollama`, `vllm`, or `sglang`, shares model storage with the matching engine container, and routes consumers through the `llm-init` data plane.

A download-only chart leaves `ENGINE_KIND` empty. `llm-init` downloads the model, maintains its control plane, and exposes no `/v1/*` data plane. In this mode:

- `hf://owner/repo` downloads into the shared Hugging Face cache;
- `https://...` downloads to the explicit `MODEL_SOURCE_LOCAL` path;
- `ollama://...` is not supported because no Ollama serving engine is present.

Example download-only configuration:

```yaml
env:
  - name: MODEL_NAME
    value: Qwen/Qwen2.5-7B-Instruct
  - name: MODEL_MODE
    value: chat
  - name: MODEL_SOURCE
    value: hf://Qwen/Qwen2.5-7B-Instruct
  - name: ENGINE_KIND
    value: ""
```

## Sources and credentials

Use one source form for a simple chart:

```text
MODEL_SOURCE=hf://Qwen/Qwen2.5-7B-Instruct
MODEL_SOURCE=https://models.example/model.gguf
MODEL_SOURCE_LOCAL=/cache/manual/model.gguf
```

Raw integration also supports comma-separated sources and 1-based indexed sources. These are chart-author features: choose one form, keep exactly one main source, and mount every explicit local destination.

```text
MODEL_SOURCE=hf://org/main,hf://org/extra
```

```text
MODEL_SOURCE_NUM=2
MODEL_SOURCE_1=hf://org/main
MODEL_SOURCE_1_ROLE=main
MODEL_SOURCE_2=https://models.example/extra.bin
MODEL_SOURCE_2_LOCAL=/cache/manual/extra.bin
MODEL_SOURCE_2_ROLE=extra
```

For every `hf://` source, set deployment-level `HF_ENDPOINT` and `HF_TOKEN` as needed. Map them from the Olares Hugging Face user values described in [the chart environment reference](olares-chart-env.md), rather than placing credentials or mirror flags inside `MODEL_SOURCE`.

Mount `.Values.userspace.appCommon` with `permission.appCommon: true` so `/cache/hf/hub` is shared across applications. The cache and `/run/llm-init` run-state volume must preserve POSIX rename behavior and be writable by UID 1000. The engine or consumer should mount downloaded model data read-only when it does not need to modify it.

## Canonical `llm-init` documentation

- [README](https://github.com/beclab/llm-init)
- [`MODEL_SOURCE` contract](https://github.com/beclab/llm-init/blob/main/docs/model-source.md)
- [model-spec file and API contract](https://github.com/beclab/llm-init/blob/main/docs/model-spec-file.md)
- [`ENGINE_ARGS` contract](https://github.com/beclab/llm-init/blob/main/docs/engine-args.md)
