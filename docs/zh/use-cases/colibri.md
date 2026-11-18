---
outline: [2, 3]
description: 在 Olares 上使用 Colibri 运行较大的 MoE 模型、开始对话，并根据设备资源调整配置。
head:
  - - meta
    - name: keywords
      content: Olares, Colibri, 本地大模型, GGUF, MoE, 内存, 显存, CPU 推理, Python 代码检查, Olares Router
app_version: "0.2.35"
doc_version: "3.0"
doc_updated: "2026-11-18"
---

# 使用 Colibri 在 Olares 上运行更大的 MoE 模型

Colibri 可以在 Olares 设备上运行受支持的 GGUF 模型，让你通过网页与模型对话。运行较大的混合专家模型（MoE）时，它可以按需从磁盘读取模型数据，因此不必把整个模型都放进内存。

你可以设置分给 Colibri 的系统内存和显存，也可以使用 CPU 推理。内存越少，从磁盘读取的内容可能越多，回复也可能越慢，尤其是输入较长时。不同配置对速度的影响，可以参考后面的[性能测试结果](#资源配置与性能参考)。

本教程以 Qwen3.6-35B-A3B 为例。换用其他受支持的 GGUF 模型时，也可以按相同步骤操作，并根据设备资源调整配置。

## 学习目标

在本教程中，你将学习如何：

- 添加 GGUF 模型文件，根据可用资源设置内存和显存额度，并选择推理模式。
- 与本地模型对话并导出聊天记录。
- 换用其他模型时调整资源配置。

## 前提条件

- 一台运行 Olares 1.12.7 或更高版本的 Intel 或 AMD 64 位设备，使用 Intel Core 第 4 代或更新型号、AMD Ryzen 系列，或具备相同能力的处理器。
- Olares 设备上有 NVIDIA 显卡。使用 GPU 加速时，需要 GeForce RTX 50 系列显卡和支持 CUDA 13.0 的驱动。
- 足够存放模型文件的磁盘空间。下文使用的 Qwen3.6-35B-A3B `UD-IQ4_XS` 文件约为 16.5 GiB。

## 添加模型文件

Colibri 会扫描 `Common/huggingface` 下的 GGUF 文件。如果模型已经在这个目录下，可以直接[安装 Colibri](#安装-colibri-并设置内存和显存)。

添加本例使用的模型文件：

1. 将 [Qwen3.6-35B-A3B-UD-IQ4_XS.gguf](https://huggingface.co/unsloth/Qwen3.6-35B-A3B-GGUF/blob/main/Qwen3.6-35B-A3B-UD-IQ4_XS.gguf) 下载到电脑。
2. 从启动台打开 Files，进入 **Common** > **huggingface**。
3. 将 GGUF 文件拖入 Files 窗口，或点击右上角的上传图标。详情参阅[上传文件](/zh/manual/olares/files/add-edit-download.md#上传文件)。

:::tip 直接将模型保存到 Olares
在 LarePass 桌面端进入**传输** > **转存**，选择**保存到 Olares**。粘贴下方链接，选择 `Common/huggingface` 作为保存位置，然后开始下载。详情参阅[将资源转存到 Olares](/zh/manual/larepass/save-resources-to-olares.md)。

```text
https://huggingface.co/unsloth/Qwen3.6-35B-A3B-GGUF/resolve/main/Qwen3.6-35B-A3B-UD-IQ4_XS.gguf
```
:::

## 安装 Colibri 并设置内存和显存

如果设备能分出 `32Gi` 内存和 `8Gi` 显存，可以从这组数值开始。换用其他模型时，可以再调整内存上限。

1. 打开 Market，搜索“Colibri”，点击 **Get**，再点击 **Install**。
2. 在 **Configure environment variables for Colibri** 窗口中填写：

   | 设置 | 起步值 | 作用 |
   |---|---|---|
   | `MEMORY_LIMIT` | `32Gi` | 分配给 Colibri 的系统内存上限 |
   | `GPU_MEMORY_LIMIT` | `8Gi` | 分配给 Colibri 的显存额度 |

   :::tip 按设备资源调整额度
   在 Dashboard 查看可用内存，在 **Settings** > **GPU** 查看显存分配。如果设备可用资源较少，就按能分出的额度填写，并从更小的模型开始尝试。显存填 `0` 时使用 CPU 推理。
   :::

   其他设置保持默认。

3. 点击 **Confirm**，等待安装完成。

## 扫描并选择本地模型

1. 从启动台打开 Colibri。尚未加载模型时，会进入“本地模型”页面。
2. 点击**重新扫描**，查找 `Common/huggingface` 中的 GGUF 文件。
3. 选中 **Qwen3.6-35B-A3B**，点击**加载所选模型** > **确认继续**。
4. 等待状态显示“模型已就绪”，然后点击**进入聊天**。等待加载时，可以展开**加载日志**查看进度。

![在 Colibri 中选择本地模型](/images/manual/use-cases/colibri-moe-models-en.jpg#bordered)

Colibri 会记住已选模型，应用重启后会重新加载。

## 与本地模型对话

1. 在聊天框中输入问题并发送。例如，可以用下面的提示让模型修改一段 Python 函数：

   ```text
   这个函数应该按分数从高到低返回前几名。
   请用一句话说明错误，然后修改函数。

   def top_scores(scores, limit=3):
       ranked = sorted(scores)
       return ranked[:limit]
   ```

2. 阅读模型回复。如果修改建议不清楚，可以在同一对话中继续提问。

处理较长文件时，可以一次发送一个函数或片段。默认 16K 上下文由输入、历史消息和回复共同使用。内存较少时，长输入可能需要等待更久。

### 导出对话

Colibri 刷新页面后不会保留聊天记录和生成设置。关闭或刷新页面前，先导出需要保存的对话。

1. 回复完成后，点击页面右上角的 **Export conversation** 图标。
2. 选择 **Markdown (.md)**，将对话下载到电脑。文件包含输入记录和模型回复，方便之后整理或分享。需要结构化保存时，也可以选择 **JSON (.json)**。

## 管理本地模型

### 切换模型

1. 等待当前回复结束。如果需要保留对话，先导出聊天记录。
2. 点击聊天页面左侧的本地模型图标，打开模型管理页面。
3. 选择其他模型，点击**加载所选模型** > **确认继续**。
4. 模型就绪后，点击**进入聊天**，开始新对话。

### 停止本地模型

1. 从聊天页面左侧的本地模型图标打开模型管理页面。
2. 点击**停止本地模型**并确认。推理进程使用的内存和显存会释放。

之后可以在同一页面重新加载模型。

## 根据模型调整资源

换用其他模型时，先用当前配置加载，再发一条短消息试用。如果提示内存不足或回复太慢，而设备还有空闲内存，就调高内存上限。显卡有余量时，也可以增加显存额度。如果设备已经没有余量，就换用更小的模型或量化文件。

### 调整内存上限

1. 导出需要保存的对话。修改设置会重启 Colibri。
2. 打开 **Settings** > **Applications** > **Colibri** > **Manage environment variables**。
3. 将 `MEMORY_LIMIT` 改为设备能分出的额度，点击 **Confirm** > **Apply**。
4. 返回 Colibri，等待模型就绪。

### 调整显存额度

1. 导出需要保存的对话。
2. 卸载 Colibri，并选择保留应用数据。
3. 重新安装 Colibri，填写新的 `GPU_MEMORY_LIMIT`。`Common/huggingface` 中的模型文件仍可使用。

### 选择推理模式

Colibri 默认使用 `standard`。如果想在可用内存较少时尝试受支持的 MoE 模型，可以切换到 `low-memory-moe`。两种模式使用相同的内存和显存额度，但后一种模式的回复可能更慢，尤其是输入较长时。

1. 导出需要保存的对话。切换模式会重启 Colibri。
2. 打开 **Settings** > **Applications** > **Colibri** > **Manage environment variables**。
3. 将 `INFERENCE_PROFILE` 设为 `low-memory-moe`，点击 **Confirm** > **Apply**。
4. 返回 Colibri，等待“模型已就绪”后再开始新对话。

需要切回默认模式时，将 `INFERENCE_PROFILE` 改回 `standard` 并应用更改。

## 资源配置与性能参考

以下数据展示本例模型在 Olares One 和 RTX 5090 Laptop GPU 上使用不同资源配置时的表现。每次请求约有 1K 输入和 128 tokens 输出，上下文设为 16K。表中使用每档第二次请求的数据。

### 不同显存额度

固定 `48Gi` 系统内存上限，比较不同显存额度：

| 显存额度 | 首 token 时间 | 生成速度 |
|---|---:|---:|
| 0，纯 CPU | 8.57 秒 | 20.53 token/s |
| 4Gi | 5.89 秒 | 30.57 token/s |
| 8Gi | 1.87 秒 | 39.81 token/s |
| 12Gi | 0.98 秒 | 46.71 token/s |
| 16Gi | 0.73 秒 | 68.80 token/s |
| 20Gi | 0.25 秒 | 171.13 token/s |

对这个模型，显存达到 `20Gi` 时速度有明显提升。

### 不同内存上限

固定 `8Gi` 显存额度，比较不同系统内存上限：

| 内存上限 | 首 token 时间 | 生成速度 |
|---|---:|---:|
| 48Gi | 1.88 秒 | 42.92 token/s |
| 32Gi | 2.01 秒 | 35.30 token/s |
| 24Gi | 30.98 秒 | 22.80 token/s |
| 16Gi | 33.82 秒 | 20.45 token/s |
| 12Gi | 58.05 秒 | 16.32 token/s |
| 8Gi | 68.61 秒 | 11.18 token/s |

对这个模型，内存上限从 `24Gi` 提高到 `32Gi` 后，首次响应明显变快。可以参考表中的变化，再用自己的设备和模型测试合适的配置。

## 可选：切换到 Olares Router

如果已在 Router 中设置默认聊天模型，也可以从 Colibri 调用。设置方法参阅[使用 Olares Router 作为你的 AI 网关](/zh/use-cases/olares-router.md)。

1. 在 Colibri 中，进入 **Settings** > **Connection**。
2. 将 **Connection service** 改为 **Olares Router**。
3. 在 **Model** 标签页选择 `default-chat`。
4. 新建对话。需要切回本地模型时，在 **Connection service** 中选择 **Colibri local engine (default)**，再新建对话。

切换聊天连接不会停止已加载的本地模型。需要释放内存和显存时，可以[停止本地模型](#停止本地模型)。

## 了解更多

- [将资源转存到 Olares](/zh/manual/larepass/save-resources-to-olares.md)：使用 LarePass 转存，将模型文件直接下载到 Olares。
- [使用 Olares Router 作为你的 AI 网关](/zh/use-cases/olares-router.md)：管理模型服务和默认聊天模型。
- [Colibri 官方文档（当前构建版本）](https://github.com/JustVugg/colibri/blob/0a5504a1dc64f86fcd0fe7db298b0448c730e7e6/README.md)：查看模型兼容性、运行原理和调优说明。
