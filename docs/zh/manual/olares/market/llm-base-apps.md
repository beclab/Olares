---
outline: [2, 3]
description: 了解如何在 Olares 中使用大模型底座（LLM Base App）托管本地大语言模型，并通过克隆底座运行 Ollama、vLLM、llama.cpp 或 SGLang 等推理引擎。
---

# 使用大模型底座托管本地大语言模型

Olares V1.12.6 推出了基于 `llm-init` 项目的本地大语言模型（LLM）托管与管理平台。该平台提供四个大模型底座应用，分别对应四种推理引擎：**Ollama LLM Base**、**vLLM LLM Base**、**llama.cpp LLM Base** 和 **SGLang LLM Base**。选择对应引擎的底座，部署不同模型，并通过专属面板监控模型运行状态。

## 开始之前

- 你的 Olares 系统已升级至 V1.12.6 或更高版本。

## 找到大模型底座

1. 打开 **Market**，搜索“LLM Base”。

   ![应用市场中的大模型底座](/images/manual/olares/llm-base-apps.png#bordered)

2. 每个底座应用针对不同的推理场景做了优化。根据你的模型来源、性能需求和硬件条件进行选择：

   | 底座应用 | 适用场景 |
   | :--- | :--- |
   | **Ollama LLM Base (llm-init)** | 快速上手和广泛的模型兼容。Ollama 可通过原生模型标签自动拉取模型，最适合聊天和嵌入任务。 |
   | **vLLM LLM Base (llm-init)** | 高并发场景下对 HuggingFace 模型进行高吞吐量推理服务。 |
   | **SGLang LLM Base (llm-init)** | 需要高效结构化生成和高级推理服务优化的场景。 |
   | **llama.cpp LLM Base (llm-init)** | 轻量 GGUF 模型、显存有限或资源紧张的部署环境。 |

## 创建新的模型实例

大模型底座只是一个模板。要运行模型，你需要先将底座克隆为独立的运行实例。

1. 选择与你所需推理引擎匹配的底座，然后点击 **View**。例如，**Ollama LLM Base (llm-init)**。
2. 点击 **Create**，初始化一个新的实例。
3. 配置实例标识：

   - **New app name**：输入实例的唯一名称。该名称会显示在 **Market** > **My Olares** 中。
   - **Shortcut name for {client}**：输入在启动台上显示的唯一快捷方式名称。

4. 点击 **Create**，进入环境变量配置。

## 配置引擎环境变量

创建实例后，会弹出配置窗口。你需要定义引擎从哪里拉取模型、使用多少显存，以及向其他客户端应用暴露哪些能力。

1. 在 **Configure environment variables for {New-app-name}** 窗口中，根据目标模型和引擎填写以下信息：

   | 变量 | 说明 |
   | :--- | :--- |
   | **MODEL_SOURCE** | 指定模型源地址。<br>格式取决于所选引擎。<br>示例：`ollama://qwen3.5:0.8b` 或 `hf://Qwen/Qwen3.5-2B`。 |
   | **MODEL_NAME** | 指定客户端应用连接时使用的模型名称。<br>示例：`qwen3.5-2b`。 |
   | **MODEL_MODE** | 选择 **Chat** 或 **Embedding**。<br>示例：`chat`。 |
   | **MODEL_SUPPORTS** | 输入逗号分隔的模型能力标志：<ul><li>**推理模型**：包含 `supports_reasoning`。</li><li>**工具调用模型**：包含 `supports_function_calling`。<br>如需同时处理多个任务，再添加 `supports_parallel_function_calling`。</li><li>**视觉模型**：包含 `supports_vision`。</li><li>**嵌入模型**：留空，不要填写聊天相关能力标志。</li></ul>示例：`supports_function_calling,supports_tool_choice`。<br>参考：[模型能力标志](#model-capability-flags)。 |
   | **ENGINE_ARGS** | 指定引擎启动参数。<br>多个参数之间用空格分隔。<br>示例：`OLLAMA_CONTEXT_LENGTH=4096`。<br>参考：[引擎调优参数](#engine-tuning-arguments)。 |
   | **{ENGINE}_REQUIRED<br>_GPU_MEMORY** | 设置实例所需的最低显存，单位为 MB 或 Gi。<br>示例：`8192`。 |

2. 点击 **Confirm** 保存配置并开始安装实例。

   页面右侧会出现 **Instances** 面板，显示安装进度。安装完成后，实例操作按钮会变为 **Open**，表示底层服务正在运行。

   ![大模型底座实例安装完成](/images/manual/olares/llm-base-model-instance-installed.png#bordered)

### 参考：模型能力标志 {#model-capability-flags}

`MODEL_SUPPORTS` 变量声明模型向外部客户端暴露的能力。这些标志对所有推理引擎通用。

| 类别 | 支持的标志 |
| --- | --- |
| **核心** | `supports_vision`、`supports_function_calling`、<br>`supports_reasoning`、`supports_native_streaming`、<br>`supports_response_schema`、`supports_prompt_caching`、<br>`supports_web_search`、`supports_parallel_function_calling` |
| **多模态** | `supports_audio_input`、`supports_audio_output`、<br>`supports_video_input`、`supports_pdf_input`、<br>`supports_computer_use`、`supports_url_context` |
| **推理与控制 token** | `supports_reasoning_effort`、`supports_thinking`、<br>`supports_assistant_prefill`、`supports_tool_choice`、<br>`supports_tokenizer` |
| **采样控制** | `supports_system_messages`、`supports_temperature`、<br>`supports_top_p`、`supports_top_k`、<br>`supports_stop_sequences`、`supports_frequency_penalty`、<br>`supports_presence_penalty` |
| **响应形式** | `supports_n`、`supports_logprobs`、`supports_seed`、<br>`supports_response_format`、`supports_logit_bias`、`supports_user` |

### 参考：引擎调优参数 {#engine-tuning-arguments}

使用 `ENGINE_ARGS` 变量来调整显存占用、上下文长度和处理行为。点击下方推理引擎查看可用调优参数。

<tabs>
<template #Ollama>

| 参数 | 用途 | 推荐示例 |
| :--- | :--- | :--- |
| `OLLAMA_CONTEXT_LENGTH` | 设置默认上下文窗口大小（以 token 为单位）。<br><br>默认根据显存自动调整：<ul><li>小于 24G：4096</li><li>24G 到 48G 之间：32768</li><li>48G 及以上：262144</li></ul> | `8192` 到 `131072` |
| `OLLAMA_NUM_PARALLEL` | 设置每个模型可同时处理的并发请求数。<br><br>Ollama 会根据可用显存自动决定，通常为 `1` 或 `4`。<br><br>默认值：`0` | `1` |
| `OLLAMA_KV_CACHE_TYPE` | 设置 KV 缓存量化类型以节省显存。<br><br>默认值：`f16`。 | `q8_0`（轻微精度损失）或 `q4_0` |
| `OLLAMA_FLASH_ATTENTION` | 启用 Flash Attention，以优化长上下文场景下的显存效率。<br><br>默认值：`0`（关闭）。 | `1`（开启） |
| `OLLAMA_MAX_LOADED_MODELS` | 设置可同时加载在内存中的模型数量上限。<br>默认会根据每块可用 GPU 自动扩展到约 3 个模型。<br><br>默认值：`0`。 | `1` |
| `OLLAMA_MAX_QUEUE` | 设置处理队列中允许的最大请求数。<br><br>默认值：`512`。 | `512` |
| `OLLAMA_KEEP_ALIVE` | 设置最后一次请求后模型在内存中保留的时长。使用 `-1` 表示永久保留。<br><br>默认值：`5m`。 | `30m` 或 `-1` |
| `OLLAMA_LOAD_TIMEOUT` | 设置等待模型加载完成的最大时长。<br><br>默认值：`5m`。 | `5m` |
| `OLLAMA_GPU_OVERHEAD` | 设置预留的安全显存余量（字节）。<br><br>默认值：`0`。 | `0` |
| `OLLAMA_DEBUG` | 设置系统日志级别，用于排查问题。<br><br>默认值：`0`（Info）。 | `1`（Debug） |

</template>
<template #vLLM>

占位符

| 参数 | 用途 | 推荐示例 |
| :--- | :--- | :--- |
| `--max-model-len` | 设置最大上下文窗口大小。如遇到显存不足，可适当减小。 | `8192` |
| `--gpu-memory-utilization` | 设置模型可使用的显存比例上限。 | `0.9` |
| `--tensor-parallel-size` | 设置用于张量并行的 GPU 数量。 | `1` |

</template>
<template #Llama-cpp>

占位符

| 参数 | 用途 | 推荐示例 |
| :--- | :--- | :--- |
| `-c` | 设置最大上下文窗口大小（以 token 为单位）。 | `65536` |
| `-ngl` | 将模型层 offload 到 GPU。 | `all` |
| `-fa` | 启用 Flash Attention 以降低显存占用。 | `on` |

</template>
<template #SGLang>

占位符

| 参数 | 用途 | 推荐示例 |
| :--- | :--- | :--- |
| `--context-length` | 设置最大上下文长度。 | `32768` |
| `--mem-fraction-static` | 设置用于静态用途（模型权重和 KV 缓存）的显存比例。 | `0.85` |
| `--max-running-requests` | 设置并发处理的最大请求数。 | `64` |
| `--reasoning-parser` | 配置推理模型的解析器。 | `qwen3` |

</template>
</tabs>

### 引擎配置示例

<tabs>
<template #Ollama>

Ollama 使用原生模型标签自动拉取模型。

**聊天模型示例**
```text
MODEL_SOURCE=ollama://qwen3.5:2b
MODEL_NAME=qwen3.5-2b
MODEL_MODE=chat
MODEL_SUPPORTS=supports_function_calling,supports_tool_choice
ENGINE_ARGS=OLLAMA_CONTEXT_LENGTH=8192
OLLAMA_REQUIRED_GPU_MEMORY=4096
```

**嵌入模型示例**
```text
MODEL_SOURCE=ollama://nomic-embed-text
MODEL_NAME=nomic-embed-text
MODEL_MODE=embedding
MODEL_SUPPORTS=
ENGINE_ARGS=OLLAMA_KEEP_ALIVE=-1
OLLAMA_REQUIRED_GPU_MEMORY=4096
```

</template>
<template #vLLM>

```text
MODEL_SOURCE=hf://Qwen/Qwen3.5-2B
MODEL_NAME=Qwen/Qwen3.5-2B
MODEL_MODE=chat
MODEL_SUPPORTS=supports_reasoning
ENGINE_ARGS=--max-model-len 8192 --gpu-memory-utilization 0.9 --tensor-parallel-size 1
VLLM_REQUIRED_GPU_MEMORY=10Gi
```

</template>
<template #Llama-cpp>

```text
MODEL_SOURCE=hf://unsloth/Qwen3.5-2B-GGUF --include Qwen3.5-2B-UD-Q4_K_XL.gguf,hf://unsloth/Qwen3.5-2B-GGUF --include mmproj-F16.gguf
MODEL_NAME=unsloth/Qwen3.5-2B-GGUF:UD-Q4_K_XL
MODEL_MODE=chat
MODEL_SUPPORTS=supports_vision,supports_reasoning
ENGINE_ARGS=-c 65536 -ngl all -fa on
LLAMACPP_REQUIRED_GPU_MEMORY=8192
```

</template>
<template #SGLang>

```text
MODEL_SOURCE=hf://Qwen/Qwen3.5-2B
MODEL_NAME=Qwen/Qwen3.5-2B
MODEL_MODE=chat
MODEL_SUPPORTS=supports_function_calling,supports_tool_choice,supports_reasoning,supports_thinking
ENGINE_ARGS=--context-length 32768 --mem-fraction-static 0.85 --max-running-requests 64 --reasoning-parser qwen3
SGLANG_REQUIRED_GPU_MEMORY=8192
```

</template>
</tabs>

## 监控下载与初始化状态

你可以通过实例内置面板跟踪模型下载、查看性能指标，并获取 API 连接信息。

1. 在底座详情页右侧的 **Instances** 面板中找到你的部署。
2. 当状态显示底层服务正在运行时，点击 **Open**。模型实例的 **llm-init** 页面会随之打开。
3. 在 **STATUS** 标签页中确认部署进度：
   - **DOWNLOAD**：实时显示下载百分比、速度和预计完成时间（ETA）。
   - **STATUS**：跟踪失败或重试次数。如果出现网络中断或模型源地址格式错误，修复后点击 **Retry**。
   - **ENGINE**：显示初始化状态。

     - 确认两个追踪标签均显示 **Engine alive: yes** 和 **Model exists: yes**，这表示引擎已在线并可接受请求。
     - 复制模型名称和 OpenAI 兼容 API 基础地址。

4. 进入 **CONFIG** 标签页，查看运行限制、执行性能探测、查看基准历史或更新变量。

## 将客户端应用连接到模型服务

本地模型实例成功运行后，其他客户端应用可以使用标准 OpenAI API 模式连接该服务。

1. 打开 Olares **Settings**，然后进入 **Applications** > **{Your-New-Model-Instance}** > **Shared entrance** > **{Engine} LLM API**。
2. 复制端点 URL，并与你定义的 `MODEL_NAME` 一起填入客户端应用的模型配置部分。

## 卸载模型实例

1. 从应用市场打开目标底座应用。
2. 在**实例**部分，找到目标模型实例，点击操作按钮旁的下拉箭头，然后点击**卸载**。
