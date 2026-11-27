---
title: 在 Olares 上部署本地大模型
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/llm-base-apps
outline: [2, 3]
description: 在 Olares 上使用 Ollama、llama.cpp、vLLM、SGLang 或 Strata 部署本地大模型，配置模型来源、显存和上下文长度。
head:
  - - meta
    - name: keywords
      content: Olares, Router, Engine Base, 本地大模型, Ollama, vLLM, llama.cpp, SGLang, Strata, 模型部署
doc_version: "2.0"
doc_updated: "2026-11-27"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/llm-base-apps.md)为准。
:::

# 使用 Ollama、llama.cpp、vLLM、SGLang 和 Strata 部署本地大模型

<VersionRouteSelect />

Olares 支持 Ollama、llama.cpp、vLLM、SGLang 和 Strata 五种推理引擎。每种引擎都有基座应用，可用于创建独立的本地模型服务。

选择适合模型的基座，填写模型来源和运行参数，然后通过 Router 将模型接入 AI 应用。

## 学习目标

- 根据模型格式选择引擎基座。
- 创建模型实例，填写配置并确认模型就绪。
- 调整 GPU 显存需求、上下文长度和引擎启动参数。

## 前提条件 {#开始之前}

Olares 版本需为 1.12.7 或更高版本。使用 1.12.6 时，请切换到页面顶部的旧版教程。

:::tip 获取更多预构建模型
[升级到最新 Olares 版本](/zh/manual/olares/settings/update.md)，可获得更多预构建模型应用。
:::

## 步骤 1：选择本地大模型推理引擎 {#找到引擎基座应用}

根据模型文件格式和使用需求选择引擎基座。

| 模型来源或需求 | 选择的基座 |
| --- | --- |
| Hugging Face 上的 GGUF 文件。Olares One 推荐使用该引擎 | llama.cpp 引擎基座 |
| Ollama 模型库中的模型标签 | Ollama 引擎基座 |
| Hugging Face 上的 `.safetensors` 权重，需要处理多个并发请求 | vLLM 引擎基座 |
| Hugging Face 上的 `.safetensors` 权重，需要结构化生成或推理优化 | SGLang 引擎基座 |
| 结合系统内存、SSD 和显存，在消费级显卡上运行 Hy4-Flash 等大型 MoE 模型 | Strata 引擎基座 |

:::info 选择 vLLM 或 SGLang
这两种引擎加载模型时，需要支持模型的架构和量化方式。可查看 Hugging Face 模型页的部署说明，选择其中推荐的引擎及版本。例如，AWQ 模型需要引擎支持该模型的 AWQ 量化实现。
:::

1. 打开 Market，搜索“引擎基座”。
2. 在搜索结果中找到所选引擎。

   ![Market 中的引擎基座应用](/images/manual/olares/llm-base-apps1.png#bordered)

## 步骤 2：创建并配置本地模型 {#创建新的模型实例}

1. 在所选引擎基座上点击 **View**，然后点击 **Create**。

   ![创建模型实例](/images/manual/olares/llm-base-apps-create-instance2.png#bordered)

2. 填写名称。以下以 Qwen3.6-27B MTP 为例：

   | 字段 | 填写内容 | 显示位置 |
   | --- | --- | --- |
   | **New app name** | 唯一的应用名，例如 `Qwen3.6-27B-MTP` | Market 和 Settings |
   | **Shortcut name for [client]** | 唯一的快捷方式名称，例如 `qwen3.6-27b-mtp` | Launchpad |

   Strata 还提供 **Shortcut name for [strataui]**，用于命名 Strata UI 快捷方式。

3. 点击 **Create**，进入 **Configure environment variables for [New-app-name]** 窗口。

:::info 硬件加速器选择
当设备上有多种可用的硬件加速方式时，Market 会提示选择。只有一种可用方式时，Olares 会自动采用该方式，继续进入实例配置。
:::

<a id="配置引擎环境变量"></a>

<a id="custom-model"></a>

<a id="推荐模型与参数"></a>

按所选引擎填写下面的字段。展开示例可查看一套完整配置。

### 使用 llama.cpp 配置 GGUF 模型 {#llama-cpp}

选择 Hugging Face 上的 GGUF 文件。`<repo>` 包含用户或组织名及仓库名，`<quantization>` 为量化类型，每个实例使用一种量化。

| 字段 | 格式与用途 |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo> --include <file>.gguf` |
| **MODEL_NAME** | `<repo>:<quantization>` |
| **ENGINE_ARGS** | 包含 `-c <tokens>`，设置上下文 token 数。多个参数用空格分隔 |
| **MODEL_MODE** | 聊天模型选 **Chat**，嵌入模型选 **Embedding** |
| **MODEL_SUPPORTS** | 按模型能力选择 **Vision**（图片理解）、**Tools**（工具调用）或 **Thinking**（推理）。嵌入模型选 **None** |
| **LLAMACPP_REQUIRED_GPU_MEMORY** | 实例启动所需的最小 GPU 显存，单位为 MB 或 Gi，例如 `23Gi` |
| **LOG_LEVEL** | 日志级别，例如 `Info` |

下载多个 GGUF 文件时，用逗号分隔完整的 `hf://... --include ...` 条目。

<details>
<summary>示例：Qwen3.6-27B MTP</summary>

模型：Hugging Face 上的 [`unsloth/Qwen3.6-27B-MTP-GGUF`](https://huggingface.co/unsloth/Qwen3.6-27B-MTP-GGUF)，量化参数 `UD-Q4_K_XL`

**MODEL_SOURCE**

```text
hf://unsloth/Qwen3.6-27B-MTP-GGUF --include Qwen3.6-27B-UD-Q4_K_XL.gguf
```

需要处理图片时，在 **MODEL_SOURCE** 中添加配套的视觉投影文件：

```text
hf://unsloth/Qwen3.6-27B-MTP-GGUF --include Qwen3.6-27B-UD-Q4_K_XL.gguf,hf://unsloth/Qwen3.6-27B-MTP-GGUF --include mmproj-F16.gguf
```

**MODEL_NAME**

```text
unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL
```

- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`, `Tools`

**ENGINE_ARGS**

```text
-c 131072 -ngl all -fa on -ctk q8_0 -ctv q8_0 --jinja -np 1 --spec-type draft-mtp --spec-draft-n-max 2
```

此模型使用 MTP 权重。`--spec-type draft-mtp` 启用 MTP 推测解码，`--spec-draft-n-max 2` 将每轮草稿长度设为最多 2 个 token。

- **LOG_LEVEL**：`Info`
- **LLAMACPP_REQUIRED_GPU_MEMORY**：`23Gi`

</details>

<details>
<summary>备选模型：Qwopus3.6-27B MTP</summary>

模型：Hugging Face 上的 [`Jackrong/Qwopus3.6-27B-v2-MTP-GGUF`](https://huggingface.co/Jackrong/Qwopus3.6-27B-v2-MTP-GGUF)，量化参数 `Q4_K_M`

**MODEL_SOURCE**

```text
hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include Qwopus3.6-27B-v2-MTP-Q4_K_M.gguf
```

需要处理图片时，在 **MODEL_SOURCE** 中添加配套的视觉投影文件：

```text
hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include Qwopus3.6-27B-v2-MTP-Q4_K_M.gguf,hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include mmproj-F32.gguf
```

**MODEL_NAME**

```text
Jackrong/Qwopus3.6-27B-v2-MTP-GGUF:Q4_K_M
```

- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`, `Tools`

**ENGINE_ARGS**

```text
-c 131072 -ngl all -fa on -ctk q8_0 -ctv q8_0 --jinja -np 1 --spec-type draft-mtp --spec-draft-n-max 2
```

此模型使用 MTP 权重。`--spec-type draft-mtp` 启用 MTP 推测解码，`--spec-draft-n-max 2` 将每轮草稿长度设为最多 2 个 token。

- **LOG_LEVEL**：`Info`
- **LLAMACPP_REQUIRED_GPU_MEMORY**：`23Gi`

</details>

### 使用 Ollama 配置模型 {#ollama}

从 Ollama 模型库选择模型及标签。`<model>` 为模型名称，`<tag>` 指定模型版本或量化类型。

| 字段 | 格式与用途 |
| --- | --- |
| **MODEL_SOURCE** | `ollama://<model>:<tag>` |
| **MODEL_NAME** | `<model>:<tag>` |
| **ENGINE_ARGS** | 包含 `OLLAMA_CONTEXT_LENGTH=<tokens>`，设置上下文 token 数。多个参数用空格分隔 |
| **MODEL_MODE** | 聊天模型选 **Chat**，嵌入模型选 **Embedding** |
| **MODEL_SUPPORTS** | 按模型能力选择 **Vision**（图片理解）、**Tools**（工具调用）或 **Thinking**（推理）。嵌入模型选 **None** |
| **OLLAMA_REQUIRED_GPU_MEMORY** | 实例启动所需的最小 GPU 显存，单位为 MB 或 Gi，例如 `23Gi` |
| **LOG_LEVEL** | 日志级别，例如 `Info` |

<details>
<summary>示例：Gemma 4 26B</summary>

模型：Ollama 库中的 [`gemma4:26b`](https://ollama.com/library/gemma4:26b)，默认使用 Q4_K_M 量化

**MODEL_SOURCE**

```text
ollama://gemma4:26b
```

**MODEL_NAME**

```text
gemma4:26b
```

- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：Thinking, Tools, Vision

**ENGINE_ARGS**

```text
OLLAMA_KEEP_ALIVE=-1 OLLAMA_CONTEXT_LENGTH=131072 OLLAMA_FLASH_ATTENTION=1 OLLAMA_KV_CACHE_TYPE=q8_0 OLLAMA_NUM_PARALLEL=1
```

- **OLLAMA_REQUIRED_GPU_MEMORY**：`23Gi`

</details>

### 使用 SGLang 配置 Hugging Face 模型 {#sglang}

选择 SGLang 支持的 Hugging Face 模型仓库，其中需包含 `.safetensors` 权重。`<repo>` 包含用户或组织名及仓库名。

| 字段 | 格式与用途 |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo>` |
| **MODEL_NAME** | `<repo>` |
| **ENGINE_ARGS** | 包含 `--context-length <tokens>`，设置上下文 token 数。多个参数用空格分隔 |
| **MODEL_MODE** | 聊天模型选 **Chat**，嵌入模型选 **Embedding** |
| **MODEL_SUPPORTS** | 按模型能力选择 **Vision**（图片理解）、**Tools**（工具调用）或 **Thinking**（推理）。嵌入模型选 **None** |
| **SGLANG_REQUIRED_GPU_MEMORY** | 实例启动所需的最小 GPU 显存，单位为 MB 或 Gi，例如 `23Gi` |
| **LOG_LEVEL** | 日志级别，例如 `Info` |

<details>
<summary>示例：Qwen3.5-9B AWQ</summary>

模型：Hugging Face 上的 [`Qwen3.5-9B-AWQ-4bit`](https://huggingface.co/cyankiwi/Qwen3.5-9B-AWQ-4bit)

**MODEL_SOURCE**

```text
hf://cyankiwi/Qwen3.5-9B-AWQ-4bit
```

**MODEL_NAME**

```text
cyankiwi/Qwen3.5-9B-AWQ-4bit
```

- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`、`Tools`、`Vision`

**ENGINE_ARGS**

```text
--context-length 65536 --mem-fraction-static 0.85 --chunked-prefill-size 4096 --reasoning-parser qwen3 --tool-call-parser qwen3_coder
```

此示例使用 `--reasoning-parser qwen3` 提取推理内容，使用 `--tool-call-parser qwen3_coder` 解析工具调用。更换模型家族时，按新模型的部署说明选择解析器。

- **LOG_LEVEL**：`Info`
- **SGLANG_REQUIRED_GPU_MEMORY**：`23Gi`

</details>

### 使用 vLLM 配置 Hugging Face 模型 {#vllm}

选择 vLLM 支持的 Hugging Face 模型仓库，其中需包含 `.safetensors` 权重。`<repo>` 包含用户或组织名及仓库名。

| 字段 | 格式与用途 |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo>` |
| **MODEL_NAME** | `<repo>` |
| **ENGINE_ARGS** | 包含 `--max-model-len <tokens>`，设置上下文 token 数。多个参数用空格分隔 |
| **MODEL_MODE** | 聊天模型选 **Chat**，嵌入模型选 **Embedding** |
| **MODEL_SUPPORTS** | 按模型能力选择 **Vision**（图片理解）、**Tools**（工具调用）或 **Thinking**（推理）。嵌入模型选 **None** |
| **VLLM_REQUIRED_GPU_MEMORY** | 实例启动所需的最小 GPU 显存，单位为 MB 或 Gi，例如 `23Gi` |
| **LOG_LEVEL** | 日志级别，例如 `Info` |

<details>
<summary>示例：Gemma-4-12B AWQ</summary>

模型：Hugging Face 上的 [`gemma-4-12B-it-AWQ-INT4`](https://huggingface.co/cyankiwi/gemma-4-12B-it-AWQ-INT4)

**MODEL_SOURCE**

```text
hf://cyankiwi/gemma-4-12B-it-AWQ-INT4
```

**MODEL_NAME**

```text
cyankiwi/gemma-4-12B-it-AWQ-INT4
```

- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`、`Tools`、`Vision`

**ENGINE_ARGS**

```text
--max-model-len 65536 --gpu-memory-utilization 0.9 --tensor-parallel-size 1 --max-num-batched-tokens 8192 --tool-call-parser gemma4 --reasoning-parser gemma4 --enable-prefix-caching --enable-auto-tool-choice
```

此示例的推理和工具调用解析器均使用 `gemma4`。`--enable-auto-tool-choice` 允许模型自行选择工具调用。更换模型家族时，按新模型的部署说明调整解析器。

- **LOG_LEVEL**：`Info`
- **VLLM_REQUIRED_GPU_MEMORY**：`23Gi`

</details>

### 使用 Strata 配置模型 {#strata}

从 Hugging Face 选择 Strata 支持的模型包及配套文件。`<repo>` 包含用户或组织名及仓库名，`<file>` 为文件名，`<model-name>` 为自定义的模型服务名称。

:::info Strata 配置入口
模型来源和能力通过 `MODEL_SOURCE`、`MODEL_NAME` 和 `MODEL_SUPPORTS` 配置。GPU 由 Olares 按应用的硬件支持情况和设备资源分配。基座负责生成模型路径和服务地址。`ENGINE_ARGS` 用于推理参数。
:::

| 字段 | 格式与用途 |
| --- | --- |
| **MODEL_SOURCE** | `hf://<repo> --include <file>`。包含主模型及其必需的附加文件 |
| **MODEL_NAME** | `<model-name>`，用于标识模型服务 |
| **ENGINE_ARGS** | 包含 `--context <tokens>`，设置上下文 token 数。多个参数用空格分隔。常用调优项见 [Strata 启动参数](#strata-tuning) |
| **MODEL_MODE** | 选择 **Chat** |
| **MODEL_SUPPORTS** | 按模型能力选择 **Vision**（图片理解）、**Tools**（工具调用）或 **Thinking**（推理）。无额外能力时选 **None** |
| **STRATA_CPU_LIMIT** | 容器最多可使用的 CPU 核数，例如 `19` |
| **STRATA_MEMORY_LIMIT** | 容器的系统内存上限，例如 `56Gi`。超过上限可能导致容器因内存不足被终止并重启 |
| **STRATA_REQUIRED_GPU_MEMORY** | 实例启动所需的最小 GPU 显存，单位为 MB 或 Gi，例如 `23Gi` |
| **LOG_LEVEL** | 日志级别，例如 `Info` |

多个下载项用逗号分隔，每项填写完整的 `hf://... --include ...`。模型包和配套文件应属于同一模型版本。

根据设备资源设置 CPU 和内存上限、GPU 显存需求，并为模型加载与推理预留空间。

<details>
<summary>示例：Hy4-Flash SQ-64</summary>

模型：[strata-engine/Hy4-Flash-Strata](https://huggingface.co/strata-engine/Hy4-Flash-Strata)，使用 SQ-64 模型包。主模型和必需的 Engram 文件合计约占 130 GB 磁盘空间。

:::info 示例资源预算
以下配置按 64 GiB 系统内存、24 GiB 显存、19 个可用 CPU 核和 128K 文本上下文估算。部署时可根据模型加载和长提示词处理期间的内存占用调整上限。
:::

**MODEL_SOURCE**

```text
hf://strata-engine/Hy4-Flash-Strata --include Hy4-Flash-SQ-64.strata,hf://strata-engine/Hy4-Flash-Strata --include Hy4-Flash-Engram-Q8.strata
```

**MODEL_NAME**

```text
hy4-flash-sq-64
```

- **MODEL_SUPPORTS**：`Tools`、`Thinking`

**ENGINE_ARGS**

```text
--context 131072
```

此配置保留默认的 DFly2 草稿模型。需要调整时，可选择以下参数：

| 调整目标 | 填写方式 |
| --- | --- |
| 关闭推测解码 | `--context 131072 --draft off` |
| 使用 Q8 KV 缓存以降低缓存占用 | `--context 131072 -- --kv q8` |
| 使用 FP16 KV 缓存 | `--context 131072 -- --kv fp16` |
| 指定专家数据的显存和系统内存预算 | 在 `ENGINE_ARGS` 中添加 `--vram-experts-gb <GiB>` 和 `--ram-experts-gb <GiB>`，并为其他模型组件和缓冲区留出空间 |

独立的 `--` 将后面的 KV 参数直接传给 Hy4-Flash 引擎。`--kv` 也接受 `auto`，沿用引擎默认设置。专家数据预算决定模型如何分配，下方资源字段决定容器的资源上限和显存需求。

| 资源字段 | 示例值 | 用途 |
| --- | --- | --- |
| **STRATA_CPU_LIMIT** | `19` | 在示例设备上最多使用 19 个 CPU 核 |
| **STRATA_MEMORY_LIMIT** | `56Gi` | 为内存增长留出余量，并为 Olares 和其他服务留下 8 GiB 系统内存 |
| **STRATA_REQUIRED_GPU_MEMORY** | `23Gi` | 调度到至少有 23 GiB 可用显存的 GPU |

预算参考了 Strata 的 [Hy4 内存规划逻辑](https://github.com/Niko1221/Strata/blob/main/terminal/strata_control/hy4_fit.py)。CPU 上限和显存需求可按设备调整。显存较少时，Strata 会将更多模型数据放在系统内存中，或从 SSD 读取。

</details>

## 步骤 3：安装模型应用 {#step-3-安装模型应用}

1. 填完所选引擎的配置后，点击 **Confirm** 保存并开始安装。
2. 在右侧 **Instances** 面板中查看安装进度，等待按钮变为 **Open**。

   ![模型实例安装完成](/images/manual/olares/llm-base-model-instance-installed1.png#bordered)

## 步骤 4：检查模型下载和引擎状态 {#监控部署并配置模型服务}

安装完成后，打开模型应用查看下载进度和引擎启动状态。

1. 在 **Instances** 面板或 Launchpad 中打开新建的模型应用。
2. 在默认打开的 **Status** 标签页查看下载进度。模型文件会自动下载。
3. 等待 **Service status** 中的两个状态都就绪：

   | 状态项 | 可以继续的状态 | 含义 |
   | --- | --- | --- |
   | **Model** | **Ready** | 模型文件已下载并校验完成 |
   | **Engine** | **Running** | 推理服务已启动 |

   ![模型控制台就绪](/images/manual/olares/llm-base-model-console-status.png#bordered)

vLLM 和 SGLang 加载模型可能需要较长时间。如果任一状态一直未就绪，参见[模型或引擎未就绪](/zh/manual/help/ts-model-engine-not-ready.md)。

## 通过 Olares Router 连接 AI 应用 {#将客户端应用连接到模型服务}

模型就绪后，请参见 [Olares Router：获取连接信息](olares-router.md#从-router-获取连接信息)，获取客户端所需的地址、模型名称和鉴权信息。

## 调整 GPU 显存与模型上下文长度 {#参考资料}

### 调整实例的 GPU 显存需求

在模型实例的环境变量设置中，调整启动所需的最小 GPU 显存。

1. 进入 **Settings** > **Applications**，选择新建的模型实例。
2. 打开 **Manage environment variables**，找到所用引擎的显存需求字段：

   | 引擎 | 字段 |
   | --- | --- |
   | llama.cpp | `LLAMACPP_REQUIRED_GPU_MEMORY` |
   | Ollama | `OLLAMA_REQUIRED_GPU_MEMORY` |
   | SGLang | `SGLANG_REQUIRED_GPU_MEMORY` |
   | vLLM | `VLLM_REQUIRED_GPU_MEMORY` |
   | Strata | `STRATA_REQUIRED_GPU_MEMORY` |

3. 点击字段旁的编辑图标，填写显存值，单位为 MB 或 Gi，例如 `23Gi`。时间分片或独占模式下，需小于总显存。显存分片模式下，需小于剩余显存。CPU 模式下填 `0`。
4. 保存修改，然后点击 **Apply**。
5. 打开模型应用，确认 **Model** 为 **Ready**、**Engine** 为 **Running**。

### 查看引擎的上下文长度 {#在-router-中查看模型详情}

客户端需要填写上下文长度时，从 Router 查看模型实例的配置。以下以 Qwen3.8-27B (llama.cpp) 为例。

<!--@include: ../reusables/ai-service-connections.md#model-context-window-->

### 修改上下文长度或引擎显存比例

1. 在 Router 的 **LLM** 页面找到模型，点击信息图标，打开 **Model card**。
2. 点击 **Edit**，修改引擎参数。例如，llama.cpp 使用 `-c` 设置上下文长度，vLLM 使用 `--gpu-memory-utilization` 设置显存比例。其他引擎的参数见下方[推理引擎启动参数参考](#引擎调优参数)。
3. 保存修改。引擎会重新启动，等待模型恢复可用后再发送请求。
4. 如果调整了上下文长度，将客户端的值设为引擎配置的长度或更小。

## 推理引擎启动参数 {#引擎调优参数}

通过 `ENGINE_ARGS` 调整上下文长度、内存占用和并发数。下表保留常用调优项，模型专用设置在上方部署示例中说明。

### llama.cpp {#llama-cpp-tuning}

参数之间用空格分隔，例如 `-c 65536 -np 1 -ngl all -fa on`。基座也接受 `LLAMA_ARG_*=value` 格式的环境变量。完整参数和适用范围见 [llama.cpp 官方文档](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md)。

| 参数 | 用途 | 示例值 |
| --- | --- | --- |
| `-c` / `--ctx-size` | 上下文 token 预算。使用多个并发槽位时，需查看服务实际报告的单槽位上下文长度。 | `65536` |
| `-np` / `--parallel` | 并发请求槽位数。设为 `1` 时，上下文预算供单个请求使用。 | `1` |
| `-ngl` / `--gpu-layers` | 放到 GPU 上的模型层数。`all` 表示全部层。减小层数可降低显存占用。 | `all` |
| `-fa` / `--flash-attn` | Flash Attention 模式。 | `on`、`off` 或 `auto` |
| `-ctk` / `-ctv` | K 和 V 缓存的数据类型。量化缓存可节省内存，支持的组合取决于模型和后端。 | `q8_0` |
| `-t` / `--threads` | 生成时使用的 CPU 线程数。 | `8` |
| `--jinja` | 启用 Jinja 对话模板，包括工具调用所需的模板。 | 无需填值 |

### Ollama {#Ollama-tuning}

使用空格分隔的环境变量赋值，例如 `OLLAMA_CONTEXT_LENGTH=8192 OLLAMA_NUM_PARALLEL=1`。这些参数配置 Ollama 服务。`temperature` 等请求参数在客户端或 API 请求中设置。完整参数和适用范围见 [Ollama 官方文档](https://docs.ollama.com/faq)。

| 参数 | 用途 | 示例值 |
| --- | --- | --- |
| `OLLAMA_CONTEXT_LENGTH` | 默认上下文长度，单位为 token。 | `8192` |
| `OLLAMA_KEEP_ALIVE` | 请求结束后模型保留在内存中的时间。基座默认值为 `-1`，表示一直保留。 | `-1` 或 `5m` |
| `OLLAMA_FLASH_ATTENTION` | 启用 Flash Attention。量化 KV 缓存时需要开启。 | `1` |
| `OLLAMA_KV_CACHE_TYPE` | KV 缓存精度。降低精度可减少内存占用，也可能影响输出质量。 | `f16`、`q8_0` 或 `q4_0` |
| `OLLAMA_NUM_PARALLEL` | 每个模型同时处理的请求数。并发越高，上下文所需内存越多。 | `1` |
| `OLLAMA_MAX_QUEUE` | 排队请求数上限，超出后拒绝新请求。 | `64` |

### SGLang {#SGLang-tuning}

参数之间用空格分隔，例如 `--context-length 65536 --mem-fraction-static 0.85`。完整参数和适用范围见 [SGLang 官方文档](https://docs.sglang.io/docs/advanced_features/server_arguments)。

:::warning SGLang 0.5.23 参数变更
该版本已移除 `--speculative-skip-dp-mlp-sync`。旧配置中保留此参数会导致启动失败。
:::

| 参数 | 用途 | 示例值 |
| --- | --- | --- |
| `--context-length` | 最大序列长度，单位为 token。 | `65536` |
| `--mem-fraction-static` | 为模型权重和 KV 缓存预留的显存比例，其余空间用于动态分配。 | `0.85` |
| `--chunked-prefill-size` | 每个预填充块的 token 数。减小可降低预填充内存峰值，但也可能降低提示词处理吞吐。 | `4096` |
| `--max-running-requests` | 同时运行的请求数上限。 | `4` |
| `--tp-size` | 张量并行使用的 GPU 数量。 | `1` |
| `--kv-cache-dtype` | KV 缓存数据类型，需模型和 GPU 支持。 | `auto` 或 `fp8_e4m3` |

### vLLM {#vLLM-tuning}

参数之间用空格分隔，例如 `--max-model-len 65536 --gpu-memory-utilization 0.9`。显存利用率以引擎可见的显存为基准。在 Olares 上，`VLLM_REQUIRED_GPU_MEMORY` 决定这部分显存预算。完整参数和适用范围见 [vLLM 官方文档](https://docs.vllm.ai/en/latest/configuration/engine_args/)。

:::tip 为 vLLM 留出显存余量
使用 vLLM 0.34.0 的基座默认启用可扩展 KV 缓存，显存利用率默认值为 `1.0`。设置较低的 `--gpu-memory-utilization`（如 `0.9`），可在已分配的显存预算内留出余量。
:::

| 参数 | 用途 | 示例值 |
| --- | --- | --- |
| `--max-model-len` | 最大序列长度，包括输入和输出 token。 | `65536` |
| `--gpu-memory-utilization` | 引擎可见显存中用于推理的比例。 | `0.9` |
| `--tensor-parallel-size` | 张量并行使用的 GPU 数量。 | `1` |
| `--max-num-seqs` | 每轮处理的序列数量上限。 | `4` |
| `--max-num-batched-tokens` | 每轮处理的 token 数上限。结合并发数和分块预填充设置，在吞吐与延迟之间取舍。 | `8192` |
| `--enable-prefix-caching` | 复用相同提示词前缀的 KV 缓存。 | 无需填值 |
| `--kv-cache-dtype` | KV 缓存数据类型，需模型和 GPU 支持。 | `auto` 或 `fp8` |

### Strata {#strata-tuning}

参数之间用空格分隔。草稿模型、KV 缓存和专家数据加载方式由所选模型家族决定。按模型家族查看可用模型、内存需求和配置选项，请参阅 [Strata 模型家族指南](https://github.com/Niko1221/Strata/blob/main/docs/MODELS.md)。启动参数的用法见 [Strata 启动参数参考](https://github.com/Niko1221/Strata/blob/main/docs/TERMINAL_LAUNCH.md)。

| 参数 | 用途和适用范围 | 示例值 |
| --- | --- | --- |
| `--context` | 上下文长度，单位为 token。基座也接受 `--max-context` 和 `--ctx` 两个别名。 | `131072` |
| `--draft` | 控制推测解码。`off` 关闭草稿模型，启用时按模型家族选择。 | `off` |

## 了解更多

- [通过 Olares Router 连接 AI 应用](olares-router.md#从-router-获取连接信息)
- [排查模型或引擎未就绪](/zh/manual/help/ts-model-engine-not-ready.md)
