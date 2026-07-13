---
outline: [2, 3]
description: 了解如何在 Olares 中使用引擎基座应用来自托管本地大语言模型，并通过克隆基座应用运行不同的推理引擎。
---

:::warning
WARNING 本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/llm-base-apps.md)为准。
:::

# 使用引擎基座应用托管本地大语言模型

Olares v1.12.6 推出了 **Model Console**，一个用于管理本地大语言模型（LLM）全生命周期的平台。该平台提供四个引擎基座应用，每个基于不同的推理引擎构建：**Ollama 引擎基座**、**vLLM 引擎基座**、**llama.cpp 引擎基座**和**SGLang 引擎基座**。

选择你所需引擎对应的基座应用，克隆它来部署模型，然后通过专属控制台运行和管理该模型。

## 开始之前

- 你的 Olares 系统已升级至 v1.12.6 或更高版本。
- 如果你想跳过手动配置、快速体验模型，可以直接从 Market 安装预构建的模型应用。这些应用打包了 Olares 验证推荐的模型与引擎组合。

  ::: details 查看可用的预构建模型应用
  - Qwen3.6-27B (llama.cpp)
  - Qwen3.6-35B-A3B (llama.cpp)
  - Ornith-1.0-35B (llama.cpp)
  - Qwen3.6-27B MTP (llama.cpp)
  - Qwopus3.6-27B MTP (llama.cpp)
  - Gemma-4-12B (vLLM)
  - Qwen3.5-9B (SGLang)
  - Gemma 4 26B (Ollama)
  - Ornith 35B (Ollama)
  :::

## 找到引擎基座应用

1. 打开 Market，搜索 "引擎基座"。

    会出现四个引擎基座应用：vLLM 引擎基座、SGLang 引擎基座、Ollama 引擎基座和 llama.cpp 引擎基座。

    ![Market 中的引擎基座应用](/images/manual/olares/llm-base-apps1.png#bordered)

2. 选择适合你需求的引擎基座。每个都针对不同的推理场景做了优化：

    | 引擎基座 | 适用场景 |
    | :--- | :--- |
    | **llama.cpp 引擎基座** | 运行轻量 GGUF 模型，或在显存有限的环境中部署。<br>它是 Olares One 上推荐的引擎。 |
    | **Ollama 引擎基座** | 希望快速上手、且需要广泛的模型兼容性。它通过原生模型标签自动拉取模型，非常适合聊天和嵌入任务。 |
    | **SGLang 引擎基座** | 需要高效的结构化生成或高级推理优化。 |
    | **vLLM 引擎基座** | 在高并发负载下对 Hugging Face 模型进行高吞吐量推理服务。 |

## 创建新的模型实例

引擎基座应用只是一个模板。要运行模型，你必须先将基座克隆为独立的运行实例。

1. 选择与你所需推理引擎匹配的引擎基座，然后点击 **View**。例如 **llama.cpp 引擎基座**。
2. 点击 **Create**，初始化一个新实例。

    ![创建模型实例](/images/manual/olares/llm-base-apps-create-instance2.png#bordered)

3. 选择实例运行所用的硬件加速器，然后点击 **Confirm**。
4. 配置实例标识：

    - **New app name**：输入实例的唯一名称。该名称会作为应用名显示在 Market 和 Settings 中。例如 `Qwen3.6-35B-A3B`。
    - **Shortcut name for [client]**：输入实例的唯一快捷方式名称。该名称会显示在 Launchpad 上。例如 `qwen3.6-35b-a3b`。

5. 点击 **Create**，进入环境变量配置。

## 配置引擎环境变量

创建实例后，会弹出配置窗口。你需要定义引擎从哪里拉取模型、使用多少显存，以及向其他客户端应用暴露哪些能力。

1. 在 **Configure environment variables for [New-app-name]** 窗口中，根据目标模型和引擎填写以下信息。

    <tabs>
    <template #llama-cpp>

    | 变量 | 说明 |
    | :--- | :--- |
    | **MODEL_SOURCE** | 指定引擎从哪里拉取模型。<br><br>格式：`hf://<repo> --include <file>.gguf`。<br>如需下载多个文件，用逗号分隔每个条目：`hf://<repo> --include <file1>.gguf,hf://<repo> --include <file2>.gguf`。<br><br>示例：<ul><li>模型页面：`https://huggingface.co/unsloth/Qwen3.6-35B-A3B-GGUF`</li><li>`MODEL_SOURCE`：`hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`</li></ul> |
    | **MODEL_NAME** | 定义客户端应用调用该实例时使用的名称。<br><br>从 `MODEL_SOURCE` 推导，格式为：`<repo>:<quantization>`（每个实例只能带一种量化）。<br><br>示例：<ul><li>`MODEL_SOURCE`：`hf://unsloth/Qwen3.6-35B-A3B-GGUF --include Qwen3.6-35B-A3B-UD-Q4_K_XL.gguf`</li><li>`MODEL_NAME`：`unsloth/Qwen3.6-35B-A3B-GGUF:UD-Q4_K_XL`</li></ul> |
    | **MODEL_MODE** | 选择 **Chat** 或 **Embedding**。 |
    | **MODEL_SUPPORTS** | 选择模型支持的能力：**Vision**、**Tools** 或 **Thinking**。若为嵌入模型，请选择 **None**。 |
    | **ENGINE_ARGS** | 设置引擎启动参数。上下文长度（`-c`）为必填项。多个参数之间用空格分隔。<br><br>示例：<ul><li>`-c 65536`</li><li>`-c 65536 -ngl all`</li></ul>更多参数见[引擎调优参数](#引擎调优参数)。 |
    | **LLAMACPP<br>_REQUIRED<br>_GPU_MEMORY** | 设置实例启动所需的最小 GPU 显存，单位为 MB 或 Gi。例如 `20Gi`。<ul><li>在时间分片或独占模式下，设为小于显卡总显存。</li><li>在显存分片模式下，设为小于剩余显存。</li><li>在 CPU 模式下，设为 `0`。</li></ul> |

    </template>
    <template #Ollama>

    | 变量 | 说明 |
    | :--- | :--- |
    | **MODEL_SOURCE** | 指定引擎从哪里拉取模型。<br><br>格式：`ollama://<model>:<size-tag>`。<br><br>示例：<ul><li>模型页面：`https://ollama.com/library/qwen3.5`</li><li>`MODEL_SOURCE`：`ollama://qwen3.5:2b`</li></ul> |
    | **MODEL_NAME** | 定义客户端应用调用该实例时使用的名称。<br><br>从 `MODEL_SOURCE` 推导：取 `ollama://` 之后的字符串。<br><br>示例：<ul><li>`MODEL_SOURCE`：`ollama://qwen3.5:2b`</li><li>`MODEL_NAME`：`qwen3.5:2b`</li></ul> |
    | **MODEL_MODE** | 选择 **Chat** 或 **Embedding**。 |
    | **MODEL_SUPPORTS** | 选择模型支持的能力：**Vision**、**Tools** 或 **Thinking**。若为嵌入模型，请选择 **None**。 |
    | **ENGINE_ARGS** | 设置引擎启动参数。上下文长度为必填项。多个参数之间用空格分隔。<br><br>示例：<ul><li>`OLLAMA_CONTEXT_LENGTH=8192`</li><li>`OLLAMA_CONTEXT_LENGTH=8192 OLLAMA_KV_CACHE_TYPE=q8_0`</li></ul>更多参数见[引擎调优参数](#引擎调优参数)。 |
    | **OLLAMA<br>_REQUIRED<br>_GPU_MEMORY** | 设置实例启动所需的最小 GPU 显存，单位为 MB 或 Gi。例如 `8Gi` 或 `8192Mi`。<ul><li>在时间分片或独占模式下，设为小于显卡总显存。</li><li>在显存分片模式下，设为小于剩余显存。</li><li>在 CPU 模式下，设为 `0`。</li></ul> |

    </template>
    <template #vLLM-or-SGLang>

    | 变量 | 说明 |
    | :--- | :--- |
    | **MODEL_SOURCE** | 指定引擎从哪里拉取模型。选择包含 `.safetensors` 权重文件的仓库。<br><br>格式：`hf://<repo>`。<br><br>示例：<ul><li>模型页面：`https://huggingface.co/Qwen/Qwen3.5-2B`</li><li>`MODEL_SOURCE`：`hf://Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_NAME** | 定义客户端应用调用该实例时使用的名称。<br><br>从 `MODEL_SOURCE` 推导：取 `hf://` 之后的字符串。<br><br>示例：<ul><li>`MODEL_SOURCE`：`hf://Qwen/Qwen3.5-2B`</li><li>`MODEL_NAME`：`Qwen/Qwen3.5-2B`</li></ul> |
    | **MODEL_MODE** | 选择 **Chat** 或 **Embedding**。 |
    | **MODEL_SUPPORTS** | 选择模型支持的能力：**Vision**、**Tools** 或 **Thinking**。若为嵌入模型，请选择 **None**。 |
    | **ENGINE_ARGS** | 设置引擎启动参数。上下文长度为必填项。多个参数之间用空格分隔。<br><br>示例：<ul><li>vLLM：`--max-model-len 65536`</li><li>SGLang：`--context-length 65536`</li></ul>更多参数见[引擎调优参数](#引擎调优参数)。 |
    | **VLLM/SGLANG<br>_REQUIRED_GPU<br>_MEMORY** | 设置实例启动所需的最小 GPU 显存，单位为 MB 或 Gi。例如 `20Gi`。<ul><li>在时间分片或独占模式下，设为小于显卡总显存。</li><li>在显存分片模式下，设为小于剩余显存。</li><li>在 CPU 模式下，设为 `0`。</li></ul> |

    </template>
    </tabs>

2. 点击 **Confirm** 保存配置并开始安装实例。

    页面右侧会出现 **Instances** 面板，显示安装进度。安装完成后，实例的操作按钮会变为 **Open**，表示底层服务正在运行。同名的模型应用也会出现在 Launchpad 上。

    ![模型实例安装完成](/images/manual/olares/llm-base-model-instance-installed1.png#bordered)

    :::info
    从引擎基座应用创建的模型实例，应用名旁会显示 `From template` 标签。在 Market 或 Settings 中查看该应用时即可看到此标签。

    ![模型实例标签](/images/manual/olares/llm-base-model-instance-tag1.png#bordered){width=70%}
    :::

:::tip 后续修改环境变量
安装后如需修改这些变量，进入 Olares **Settings** > **Applications** > **[App-Name]** > **Manage environment variables**。点击变量旁的编辑图标，更新其值，保存修改，然后点击 **Apply**。
:::

## 监控部署并配置模型服务

打开内置的模型控制台，跟踪模型下载、确认模型与引擎就绪、配置客户端访问，并查看 GPU 占用与性能。

1. 在引擎基座应用详情页的 **Instances** 面板中找到该模型实例，或在 Launchpad 中找到它。
2. 打开它，启动专属的模型控制台。

    控制台默认打开 **Status** 标签页，模型文件会自动开始下载。

3. 在 **Service status** 下跟踪模型和引擎的就绪状态：

    - **Model**：文件下载并校验完成后显示 `Ready`。
    - **Engine**：推理服务上线后显示 `Running`。

    ![模型控制台就绪](/images/manual/olares/llm-base-model-console-status.png#bordered)

4. 当引擎显示 `Running` 后，配置客户端应用如何访问该服务。

    - **Connection source**：选择客户端的运行位置。
        - **Apps in Olares**：用于在 Olares 内运行的应用。
        - **Devices on your network**：用于同一局域网内的设备。
        - **Remote**：用于通过公网访问，需先在 LarePass 中开启 VPN。
    - **API format**：选择客户端所需的 API 风格：**Ollama**、**OpenAI-Compatible** 或 **Anthropic-Compatible**。
    - **Base URL**：复制客户端应用连接服务所用的 URL。
    - **Supported endpoints**：展开此列表可查看所选 API 格式暴露的每个端点，包括其 HTTP 方法、路径和用途。

5. 选择 **Configuration** 标签页查看模型详情：

    ![Configuration 标签页](/images/manual/olares/llm-base-model-console-config.png#bordered)

    - **Model**：显示模型名称、模式，以及该实例暴露的能力标签。
    - **Parameters**：查看引擎参数。展开 **Advanced parameters** 查看完整参数，并可在 **Form** 和 **Raw** 之间切换视图。

6. 在 **GPU residency** 部分，点击 **Detect**，然后：

    - **查看模式**：确认模型运行在哪种模式下。
        - **Full GPU**：整个模型运行在 GPU 上。这是速度最快的状态，也是安装时选择 GPU 加速器后的预期状态。
        - **CPU** 或 **Split**：模型的一部分或全部运行在 CPU 上，会让推理变慢。
            - 如果安装时选择了 CPU 加速器，`CPU` 是预期状态。
            - 如果安装时选择了 GPU 加速器，请检查 `[ENGINE]_REQUIRED_GPU_MEMORY` 设置和引擎参数。
    - **查看显存占用**：查看 **VRAM**、**KV cache used** 和 **GPU memory utilization**，了解模型占用了多少显存，以及还剩多少余量用于更长的上下文或更多并发请求。

7. 在 **Performance** 部分，点击 **Run test** 测量两项响应速度指标。用它们对比不同的量化级别、上下文长度或引擎参数，并在采用某项改动前先验证它确实提升了速度：

    - **TTFT**（Time To First Token，首字延迟）：你等待第一个字出现的时长。值越低，模型响应越快。
    - **Cold start**（冷启动）：引擎从零加载模型所需的时间，例如重启后。值越低，模型越快可以对外服务。

## 将客户端应用连接到模型服务

模型实例运行后，任何使用 OpenAI 兼容 API 的客户端应用都可以通过 Base URL 连接它。

下面的示例以 [OpenCode](./opencode.md) 作为客户端。

1. 在模型控制台中进入 **Status** 标签页。在 **Service status** 下：

    - **Connection source**：选择 **Apps in Olares**，因为 OpenCode 在 Olares 内运行。
    - **API format**：选择 **OpenAI-Compatible**。
    - 复制 **Base URL**，并记下 **Model name**。

2. 在 OpenCode 中，点击左下角的 <i class="material-symbols-outlined">settings</i>，选择 **Providers**，向下滚动并点击 **Custom Provider** 旁的 **Connect**。

3. 填写以下信息：

    - **Provider ID**：该 provider 的唯一标识符。例如 `olares-llm`。
    - **Display name**：在 provider 列表中显示的名称。例如 `Olares LLM`。
    - **Base URL**：你从模型控制台复制的 **Base URL**。
    - **Models**：
        - **Model ID**：你的 `MODEL_NAME`。例如 `Qwen3.6-35B-A3B`。
        - **Display Name**：该模型显示的名称。例如 `Qwen3.6 35B A3B`。

4. 点击 **Submit** 保存配置。该 provider 会出现在 provider 列表中。
5. 运行一个任务来测试连接。本示例使用 Olares skills 将一个应用部署到 Olares。

    a. 在顶部点击 **Search** 字段，选择 **Toggle terminal** 打开终端。

    b. 登录 Olares CLI 以使用内置 Olares skills。将 `alice123@olares.com` 替换为你自己的 Olares ID。

    ```bash
    olares-cli profile login --olares-id alice123@olares.com
    ```

    c. 出现提示时，输入你的 Olares 密码并按 **Enter**。输入内容不会显示。

    d. 如果你的 Olares 开启了两步验证，CLI 会提示你输入该 Olares ID 的两步验证码。在 LarePass 中获取 6 位验证码，输入后按 **Enter**。

    e. 在聊天框下方，选择 **Big Pickle** 打开模型选择器，再从列表中选择 **Qwen3.6 35B A3B**。

    f. 发送任务。下面的示例使用 `dockersamples/101-tutorial`，一个适合初学者的 Docker 教程 Web 应用。

    ```text
    Deploy this app to Olares: https://github.com/dockersamples/101-tutorial
    ```

    g. 如果出现提示，按提示操作直到部署完成。然后你可以在启动台和 **My Olares** 中找到该应用。

    ![部署到 My Olares 的应用](/images/manual/olares/llm-base-model-inst-task1.png#bordered)

## 参考资料

### 引擎调优参数

使用 `ENGINE_ARGS` 变量添加自定义设置，以调整显存占用、上下文长度和处理行为。多个参数之间用空格分隔。点击下方推理引擎查看一些常用的调优参数。

<tabs>
<template #Llama-cpp>

| 参数 | 用途 | 推荐 |
| :--- | :--- | :--- |
| `-c` | 设置最大上下文长度（以 token 为单位）。 | `65536` |
| `-ngl` | 将所有模型层加载到 GPU，避免 CPU 计算拖慢速度。 | `all` |
| `-fa` | 启用 Flash Attention 以加速注意力计算。 | `on` |
| `-ctk` / `-ctv` | 将 KV Cache 量化为 8-bit，在显存占用与精度之间取得平衡。 | `q8_0` |

更多 llama.cpp 参数见[官方文档](https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md)。
</template>
<template #Ollama>

| 参数 | 用途 | 推荐 |
| :--- | :--- | :--- |
| `OLLAMA_CONTEXT_LENGTH` | 设置默认上下文窗口大小（以 token 为单位）。<br><br>默认值随显存自动调整：<ul><li>小于 24G：4096</li><li>24G 到 48G：32768</li><li>48G 及以上：262144</li></ul> | `8192` 到 `131072` |
| `OLLAMA_KEEP_ALIVE` | 设置最后一次请求后模型在 GPU 显存中的驻留时长。<br>到期后权重会被换出到系统内存。<ul><li>`-1` 表示永久驻留在 GPU 显存中。</li><li>`3m` 表示驻留 3 分钟。</li></ul>默认值：`-1`。 | `-1` 或 `30m` |
| `OLLAMA_FLASH_ATTENTION` | 启用 Flash Attention。使用 `OLLAMA_KV_CACHE_TYPE`<br> 进行 KV 缓存量化时必须开启。<br><br>默认值：`1`。 | `1`（开启） |
| `OLLAMA_KV_CACHE_TYPE` | 设置键值（KV）缓存的量化类型以节省显存。<br><br>默认值：`f16`。 | `q8_0`（轻微精度损失）或 `q4_0` |
| `OLLAMA_NUM_PARALLEL` | 设置每个模型可同时处理的并发请求数。<br><br>默认值：`1`。 | `1` |

更多 Ollama 参数见[官方文档](https://github.com/ollama/ollama/blob/main/docs/faq.mdx)。
</template>
<template #SGLang>

| 参数 | 用途 | 推荐 |
| :--- | :--- | :--- |
| `--context-length` | 设置最大上下文长度。 | `65536` |
| `--mem-fraction-static` | 设置预分配给静态用途的显存比例，类似 vLLM 的 <br>`--gpu-memory-utilization`。 | `0.85` |
| `--chunked-prefill-size` | 将超长输入分块处理，避免长时间占用 GPU，让并发请求的流式输出 <br>保持顺畅。 | `4096` |

更多 SGLang 参数见[官方文档](https://docs.sglang.io/docs/advanced_features/server_arguments)。
</template>
<template #vLLM>

| 参数 | 用途 | 推荐 |
| :--- | :--- | :--- |
| `--max-model-len` | 设置最大上下文长度。 | `65536` |
| `--gpu-memory-utilization` | 设置 vLLM 引擎使用的显存比例。 | `0.9` |
| `--tensor-parallel-size` | 设置张量并行规模，即用多少张 GPU 切分并共同运行同一个模型。 | `1` |
| `--max-num-batched-tokens` | 限制单个批次处理的最大 token 数，使遇到超长请求时响应时间 <br>保持稳定。 | `8192` |
| `--enable-prefix-caching` | 在 KV Cache 中缓存并复用相同前缀的计算结果。 | 启用 |

更多 vLLM 参数见[官方文档](https://docs.vllm.ai/en/v0.17.0/configuration/engine_args/)。
</template>
</tabs>

### 推荐模型与参数

以下是各引擎经过验证的最佳实践推荐。可作为起点，再根据你的硬件进行调整。

<tabs>
<template #Llama-cpp>

**推荐模型 1**

- **推荐模型**：Hugging Face 上的 [`unsloth/Qwen3.6-27B-MTP-GGUF`](https://huggingface.co/unsloth/Qwen3.6-27B-MTP-GGUF)，量化参数 `UD-Q4_K_XL`
- **MODEL_SOURCE**：`hf://unsloth/Qwen3.6-27B-MTP-GGUF --include Qwen3.6-27B-UD-Q4_K_XL.gguf`
    :::tip 多模态模型
    如果模型具备多模态能力，请在 `MODEL_SOURCE` 中带上 `mmproj-F16.gguf` 文件：

    `hf://unsloth/Qwen3.6-27B-MTP-GGUF --include Qwen3.6-27B-UD-Q4_K_XL.gguf,hf://unsloth/Qwen3.6-27B-MTP-GGUF --include mmproj-F16.gguf`
    :::
- **MODEL_NAME**：`unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL`
- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`, `Tools`, `Vision`
- **ENGINE_ARGS**：`-c 100352 -ngl all -fa on -ctk q8_0 -ctv q8_0 --jinja -np 1 --spec-type draft-mtp --spec-draft-n-max 2`
- **LOG_LEVEL**：`Info`
- **LLAMACPP_REQUIRED_GPU_MEMORY**：`23Gi`

**推荐模型 2**

- **推荐模型**：Hugging Face 上的 [`Jackrong/Qwopus3.6-27B-v2-MTP-GGUF`](https://huggingface.co/Jackrong/Qwopus3.6-27B-v2-MTP-GGUF)，量化参数 `Q4_K_M`
- **MODEL_SOURCE**：`hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include Qwopus3.6-27B-v2-MTP-Q4_K_M.gguf`
    :::tip 多模态模型
    如果模型具备多模态能力，请在 `MODEL_SOURCE` 中带上 `mmproj-F32.gguf` 文件：

    `hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include Qwopus3.6-27B-v2-MTP-Q4_K_M.gguf,hf://Jackrong/Qwopus3.6-27B-v2-MTP-GGUF --include mmproj-F32.gguf`
    :::
- **MODEL_NAME**：`Jackrong/Qwopus3.6-27B-v2-MTP-GGUF:Q4_K_M`
- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`, `Tools`, `Vision`
- **ENGINE_ARGS**：`-c 100352 -ngl all -fa on -ctk q8_0 -ctv q8_0 --jinja -np 1 --spec-type draft-mtp --spec-draft-n-max 2`
- **LOG_LEVEL**：`Info`
- **LLAMACPP_REQUIRED_GPU_MEMORY**：`23Gi`

</template>
<template #Ollama>

- **推荐模型**：Ollama 库中的 [`gemma4-26b`](https://ollama.com/library/gemma4:26b)，默认使用 Q4_K_M 量化
- **MODEL_SOURCE**：`ollama://gemma4-26b`
- **MODEL_NAME**：`gemma4-26b`
- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：Thinking, Tools, Vision
- **ENGINE_ARGS**：`OLLAMA_KEEP_ALIVE=-1 OLLAMA_CONTEXT_LENGTH=131072 OLLAMA_FLASH_ATTENTION=1 OLLAMA_KV_CACHE_TYPE=q8_0 OLLAMA_NUM_PARALLEL=1`
- **OLLAMA_REQUIRED_GPU_MEMORY**：`23Gi`

</template>
<template #SGLang>

:::info
SGLang 模型可能需要较长时间加载，引擎才会进入 `RUNNING` 状态。
:::

- **推荐模型**：Hugging Face 上的 [`Qwen3.5-9B-AWQ-4bit`](https://huggingface.co/cyankiwi/Qwen3.5-9B-AWQ-4bit)
- **MODEL_SOURCE**：`hf://cyankiwi/Qwen3.5-9B-AWQ-4bit`
- **MODEL_NAME**：`cyankiwi/Qwen3.5-9B-AWQ-4bit`
- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`、`Tools`、`Vision`
- **ENGINE_ARGS**：`--context-length 65536 --mem-fraction-static 0.85 --chunked-prefill-size 4096 --reasoning-parser qwen3 --tool-call-parser qwen3_coder`
- **LOG_LEVEL**：`Info`
- **SGLANG_REQUIRED_GPU_MEMORY**：`23Gi`

</template>
<template #vLLM>

:::info
vLLM 模型可能需要较长时间加载，引擎才会进入 `RUNNING` 状态。
:::

- **推荐模型**：Hugging Face 上的 [`gemma-4-12B-it-AWQ-INT4`](https://huggingface.co/cyankiwi/gemma-4-12B-it-AWQ-INT4)
- **MODEL_SOURCE**：`hf://cyankiwi/gemma-4-12B-it-AWQ-INT4`
- **MODEL_NAME**：`cyankiwi/gemma-4-12B-it-AWQ-INT4`
- **MODEL_MODE**：`Chat`
- **MODEL_SUPPORTS**：`Thinking`、`Tools`、`Vision`
- **ENGINE_ARGS**：`--max-model-len 65536 --gpu-memory-utilization 0.9 --tensor-parallel-size 1 --max-num-batched-tokens 8192 --tool-call-parser qwen3_coder --reasoning-parser qwen3 --enable-prefix-caching --enable-auto-tool-choice`
- **LOG_LEVEL**：`Info`
- **VLLM_REQUIRED_GPU_MEMORY**：`23Gi`

</template>
</tabs>
