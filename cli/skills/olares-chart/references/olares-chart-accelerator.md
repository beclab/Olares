# Accelerator resources (`spec.accelerator`) — modes & sizing

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. This covers declaring accelerator modes and sizing the resource envelope. Building the CUDA image and provisioning model weights are in [olares-chart-gpu.md](olares-chart-gpu.md).

## A. Declaring accelerator modes (`spec.accelerator`)

An app that needs an accelerator declares one `spec.accelerator[]` entry **per compute mode** it supports. Without it the app is scheduled as plain `cpu` and never gets an accelerator device or GPU memory.

> **Naming gotchas (these bite):**
> - The YAML key is `spec.accelerator`, but `lint` error messages call it **`spec.resources`** — same thing.
> - The GPU-memory key inside an entry is **`requiredGPUMemory` / `limitedGPUMemory`** (a memory quantity, not a card count).
> - `lint` error messages for a missing accelerator-entry GPU field still say **`requiredGpu` / `limitedGpu`**, even though the YAML key you must write in the entry is `requiredGPUMemory` / `limitedGPUMemory`. Don't let the message rename your field.
> - Only **discrete-GPU** modes carry a GPU-memory field; unified/SoC modes (`nvidia-gb10`, `apple-m`, `intel`, `amd`, `moore-soc`) omit it and size via pod `requiredMemory` / `limitedMemory` instead.

### Accelerator modes (not just NVIDIA)

The canonical mode set is the one the platform recognizes on nodes — `gpu.bytetrade.io/<mode>` labels, defined in `framework/app-service/pkg/utils/gpu_types.go`. A node advertises a mode by carrying that label (a node can advertise several at once); `cpu` is never labeled because every node runs CPU workloads.

| `mode` | Target device | Memory model | Typical arch |
|---|---|---|---|
| `cpu` | no accelerator, CPU only (implicit on every node) | host RAM only | any |
| `nvidia` | discrete NVIDIA card (via HAMi) | discrete — own GPU-memory quota → `nvidia.com/gpumem` | `amd64` |
| `nvidia-gb10` | NVIDIA GB10 Superchip | unified system memory — uses pod memory, no separate gpumem | `arm64` |
| `apple-m` | Apple M-series SoC (Metal/MPS) | unified memory — pod memory | `arm64` |
| `intel` | Intel integrated GPU | unified memory — pod memory | `amd64` |
| `amd` | AMD integrated GPU (Ryzen AI Max) | unified memory — pod memory | `amd64` |
| `intel-gpu` | Intel discrete GPU | discrete — own GPU-memory quota | `amd64` |
| `amd-gpu` | AMD discrete GPU (ROCm) | discrete — own GPU-memory quota | `amd64` |
| `moore-soc` | Moore Threads SoC | unified — pod memory | per hardware |

Rule of thumb: **discrete** cards (`nvidia`, `intel-gpu`, `amd-gpu`) take a standalone GPU-memory quota (`requiredGPUMemory`/`limitedGPUMemory`); **unified / SoC** modes (`nvidia-gb10`, `apple-m`, `intel`, `amd`, `moore-soc`) draw from pod memory and declare no GPU-memory field. Declare modes with `spec.accelerator`.

> **Which modes `lint` accepts.** `olares-cli chart lint` accepts `cpu`, `nvidia`, `nvidia-gb10`, `apple-m`, `amd-gpu` (`validResourceModes` in `framework/oac/internal/manifest/resources.go`). So:
> - `nvidia`, `nvidia-gb10`, `apple-m`, `amd-gpu`, `cpu` pass lint — safe to use.
> - `intel`, `amd`, `intel-gpu`, `moore-soc` are real node modes but are **not accepted by `lint`** — declaring them fails validation.
> - Avoid `strix-halo` / `mthreads-m1000`.
>
> `lint` also cross-checks mode → `spec.supportArch` for the arch-bound modes (`nvidia`/`amd-gpu` → `amd64`; `nvidia-gb10` → `arm64`), and only `nvidia`/`amd-gpu` are blessed for the GPU-memory fields (`gpuMemoryModes`).

### Shape and semantics

```yaml
# OlaresManifest.yaml  (olaresManifest.version: '0.12.0', apiVersion: v3)
spec:
  supportArch:
  - amd64
  accelerator:
  - mode: cpu                  # optional CPU fallback — only if upstream runs on CPU
    requiredCpu: "1"
    limitedCpu: "4"
    requiredMemory: 4Gi
    limitedMemory: 16Gi
    requiredDisk: 2Gi
    limitedDisk: 10Gi
  - mode: nvidia
    supportMultiCard: false    # true if the app can shard across cards
    supportMultiNodes: false
    requiredCpu: "1"
    limitedCpu: "4"
    requiredMemory: 8Gi
    limitedMemory: 24Gi
    requiredDisk: 2Gi
    limitedDisk: 10Gi
    requiredGPUMemory: 16Gi    # GPU memory floor (NOT a card count) → nvidia.com/gpumem
    limitedGPUMemory: 24Gi
```

- `required*` is the **scheduling floor** (reserved); `limited*` is the **cap**. They map to Kubernetes container `requests` / `limits`.
- **GPU is allocated by memory, not whole cards.** `requiredGPUMemory` is the vGPU memory the scheduler reserves (matched against device memory); a card count is not what you request here.
- Each declared mode entry must be **complete** (all CPU/memory/disk pairs present); `lint` reports every missing field.
- The resource envelope lives entirely under `spec.accelerator[]` (mode-keyed) — there is no flat top-level `spec.requiredCpu/...`.

## B. Which modes to declare — local deploy vs publish

**The bar differs by destination:**

- **Deploying to your own Olares (local):** declare **only the mode(s) your local node actually advertises**. There is no value in declaring `apple-m` or `amd-gpu` for a node that only has an NVIDIA card — it just adds unbuildable image variants. Read the node's advertised modes first:

```bash
olares-cli cluster node get <node> -o json | grep gpu.bytetrade.io   # existence-based gpu.bytetrade.io/<mode> labels
olares-cli dashboard overview gpu -o json                            # per-GPU vendor + memory
```

  Then declare that one mode (plus `cpu` only if the app genuinely runs CPU-only). For the typical single NVIDIA node, that's just `nvidia` (+ optional `cpu`).

- **Publishing to the Market:** the app may land on any user's hardware, so you must **cover every backend the upstream genuinely supports** across arches — ask the user which modes to target and declare one entry per chosen backend (each usually needs its own image / build variant). Market-readiness lives in [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).

Either way, **do not invent modes** — declare only backends the upstream project genuinely supports:

1. **Inspect the repo for its accelerator backends.** Check the Dockerfile / dependencies and build flags: CUDA / cuDNN (→ `nvidia`), ROCm / HIP (→ `amd-gpu`), Apple Metal / MPS (→ `apple-m`), Vulkan / oneAPI, or pure CPU. Read the README hardware requirements, the model card's recommended VRAM, and any device-selection logic in compose / entrypoints (e.g. `llama.cpp`'s `GGML_CUDA` / `GGML_HIP` / `GGML_METAL` / `GGML_VULKAN`, or a PyTorch backend switch).
2. **Repo supports multiple backends → ask the user** which to target, then declare one `accelerator` mode per chosen backend. Remember the arch split (`nvidia`/`amd-gpu`/`intel-gpu`/`intel`/`amd` are `amd64`; `nvidia-gb10`/`apple-m` are `arm64`), so multiple modes usually mean multiple images / build variants — extra cost.
3. **Repo supports only one backend → declare only that one** (CUDA-only → just `nvidia`; CPU-only → just `cpu`).
4. **CPU fallback only when real.** Add a `cpu` mode only if the upstream actually runs on CPU; many CUDA-only projects do not — don't add it for them.

> Decide the feasible set from the repo first, then let the user choose within that set. Never declare a device the project can't use.

## C. How much to request (sizing a ported project)

Sizing is per declared mode. Start from upstream facts, then map to `required` (floor) vs `limited` (cap):

- **Where to get the numbers:** the upstream README "requirements", a compose `deploy.resources` block, the model card's recommended VRAM/RAM, and the project's own defaults.

**GPU memory (`requiredGPUMemory`)** — rule of thumb for model-serving apps:

```
GPU memory ≈ weights + KV-cache/activations + ~1–2Gi CUDA/runtime overhead
weights ≈ params × bytes-per-param   (fp16 ≈ 2 B, int8 ≈ 1 B, 4-bit ≈ 0.5 B)
```

- e.g. a **7B** model in **fp16** ≈ 14 GB weights → `requiredGPUMemory` ≈ `16Gi` (with overhead/KV cache); 4-bit quantized ≈ `6Gi`.
- Set `requiredGPUMemory` to a realistic floor and `limitedGPUMemory` to the working peak. These are heuristics — **verify on a real GPU Olares node** and adjust.

**CPU / RAM / disk:**

- **RAM** (`requiredMemory`/`limitedMemory`): enough to load and run the model server; for AI apps RAM is often comparable to or above GPU memory; leave headroom for model load/convert.
- **CPU** (`requiredCpu`/`limitedCpu`): inference servers are usually modest (request ~1–2, limit ~4); raise it if the app does heavy CPU pre/post-processing.
- **Disk** (`requiredDisk`/`limitedDisk`): only large if weights live in per-app `appData`. With the shared `appCommon` Hugging Face cache ([olares-chart-gpu.md](olares-chart-gpu.md) §B) the app's own disk stays small.

**Align with what `lint` enforces** (CPU + memory only):

- `requiredCpu <= limitedCpu` and `requiredMemory <= limitedMemory` within the manifest.
- Each container needs `requests <= limits`, and **every container must set a memory request**.
- The **sum of all container `requests` must be `<=` the manifest `required*`**, and the **sum of `limits` `<=` the manifest `limited*`**. So size the manifest envelope to **cover** what the templates actually set.

```yaml
# manifest declared (nvidia mode)         # must be >= the rendered container totals below
requiredCpu: "1"   limitedCpu: "4"
requiredMemory: 8Gi limitedMemory: 24Gi
```
```yaml
# templates/deployment.yaml container
resources:
  requests: { cpu: "1", memory: 8Gi }     # Σ requests <= manifest required*
  limits:   { cpu: "4", memory: 24Gi }    # Σ limits   <= manifest limited*
```

> Don't over-request: an oversized `required*` reserves the user's node and can make the app unschedulable. Request the realistic floor, cap at the realistic peak. (GPU memory and disk are not part of this CPU/memory cross-check — they only drive scheduling.)
