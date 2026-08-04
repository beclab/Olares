# Accelerator resources (`spec.accelerator`) — modes & sizing

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. This covers declaring accelerator modes and sizing the resource envelope. Building the CUDA image and provisioning model weights are in the GPU / models capability.

## A. Declaring accelerator modes (`spec.accelerator`)

`spec.accelerator[]` is **only** for apps that need an accelerator / GPU device. A **non-accelerator (CPU-only) app does not declare any `mode`** — it sets the flat resource envelope `spec.requiredCpu` / `limitedCpu` / `requiredMemory` / `limitedMemory` / `requiredDisk` directly under `spec` instead (see §A.1 below). Only read on if your app targets an accelerator.

An app that needs an accelerator declares one `spec.accelerator[]` entry **per compute mode** it supports. Without it the app is scheduled as plain `cpu` and never gets an accelerator device or GPU memory.

> **Naming gotchas (these bite):**
> - The YAML key is `spec.accelerator`, but `lint` error messages call it **`spec.resources`** — same thing.
> - The GPU-memory key inside an entry is **`requiredGPUMemory` / `limitedGPUMemory`** (a memory quantity, not a card count).
> - `lint` error messages for a missing accelerator-entry GPU field still say **`requiredGpu` / `limitedGpu`**, even though the YAML key you must write in the entry is `requiredGPUMemory` / `limitedGPUMemory`. Don't let the message rename your field.
> - Only **discrete-GPU** modes carry a GPU-memory field; unified/SoC modes (`nvidia-gb10`, `apple-m`, `intel`, `amd`, `moore-soc`) omit it and size via pod `requiredMemory` / `limitedMemory` instead.

### Accelerator modes (not just NVIDIA)

The canonical mode set is the one the platform recognizes on nodes via `gpu.bytetrade.io/<mode>` labels. A node advertises a mode by carrying that label (a node can advertise several at once); `cpu` is never labeled because every node runs CPU workloads.

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

> **Which modes `lint` accepts.** `olares-cli chart lint` accepts **all nine** modes in the table above — `cpu`, `nvidia`, `nvidia-gb10`, `apple-m`, `intel`, `amd`, `intel-gpu`, `amd-gpu`, `moore-soc`. All nine pass validation; only off-list names like `strix-halo` / `mthreads-m1000` are rejected.
>
> `lint` also cross-checks two things:
> - **mode → `spec.supportArch`** for every arch-bound mode: `amd64` modes are `nvidia` / `intel` / `amd` / `intel-gpu` / `amd-gpu`; `arm64` modes are `nvidia-gb10` / `apple-m` / `moore-soc`. The manifest's `spec.supportArch` must contain the mode's required arch.
> - **GPU-memory fields are allowed only for the discrete-GPU modes** `nvidia` / `amd-gpu` / `intel-gpu`; declaring `requiredGPUMemory` / `limitedGPUMemory` on any other mode fails lint.

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
- **Two mutually-exclusive ways to express the envelope** — a chart uses one, never both (`lint` rejects mixing them):
  - **Accelerator / GPU app:** the mode-keyed `spec.accelerator[]` shown here.
  - **Non-accelerator app:** the flat top-level `spec.requiredCpu` / `limitedCpu` / `requiredMemory` / `limitedMemory` / `requiredDisk` (optional `limitedDisk`) — no `mode`. See §A.1.

### A.1 Non-accelerator (CPU-only) envelope — flat fields, no mode

An app that needs no accelerator declares its resources flat under `spec`, with no `spec.accelerator[]` and no `mode`:

```yaml
spec:
  requiredCpu: 100m
  limitedCpu: "1"
  requiredMemory: 128Mi
  limitedMemory: 512Mi
  requiredDisk: 1Gi
```

app-service derives an implicit `cpu` mode from these fields at install. This and `spec.accelerator[]` are **mutually exclusive** — declaring both fails `lint`. (`from-compose` instead scaffolds the equivalent `spec.accelerator[mode=cpu]` shape; both are valid, so leave whichever the chart already uses rather than mixing them.)

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
2. **Repo supports multiple backends → ask the user** which to target, then declare one `accelerator` mode per chosen backend. Remember the arch split (`nvidia`/`amd-gpu`/`intel-gpu`/`intel`/`amd` are `amd64`; `nvidia-gb10`/`apple-m`/`moore-soc` are `arm64`), so multiple modes usually mean multiple images / build variants — extra cost.
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
- **Disk** (`requiredDisk`/`limitedDisk`): only large if weights live in per-app `appData`. With the shared `appCommon` Hugging Face cache (the GPU / models capability §B) the app's own disk stays small.

**Align with what `lint` enforces** (CPU + memory only):

- `requiredCpu <= limitedCpu` and `requiredMemory <= limitedMemory` within the manifest.
- Each container needs `requests <= limits`, and **every container must set a memory request**.
- The **sum of all container `requests` must be `<=` the manifest `required*`**, and the **sum of `limits` `<=` the manifest `limited*`**. So size the manifest envelope to **cover** what the templates actually set — e.g. an `nvidia` mode declaring `requiredMemory: 8Gi` / `limitedMemory: 24Gi` must be `>=` the rendered container `requests`/`limits` (`8Gi`/`24Gi`).

> Don't over-request: an oversized `required*` reserves the user's node and can make the app unschedulable. Request the realistic floor, cap at the realistic peak. (GPU memory and disk are not part of this CPU/memory cross-check — they only drive scheduling.)
