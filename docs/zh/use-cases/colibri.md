---
outline: [2, 3]
description: 在 Olares 上使用 Colibri 运行较大的 MoE 模型，检查一段 Python 函数，并导出对话。
head:
  - - meta
    - name: keywords
      content: Olares, Colibri, 本地大模型, GGUF, MoE, 内存, 显存, CPU 推理, Python 代码检查, Olares Router
app_version: "0.2.35"
doc_version: "3.0"
doc_updated: "2026-11-18"
---

# 使用 Colibri 在 Olares 上运行更大的 MoE 模型

Colibri 可以在本地运行受支持的 GGUF 模型，并提供网页聊天界面。混合专家模型（MoE）每次回复只用到其中一部分专家。Colibri 可以把其他专家权重留在模型文件中，需要时再从磁盘读取，让你不必将全部权重常驻内存也能尝试更大的模型。

你可以设置分给 Colibri 的系统内存和显存，也可以使用 CPU 推理。内存越少，从磁盘读取的内容可能越多，回复也可能越慢，尤其是输入较长时。

在 Olares 上，你可以复用已有的模型文件，并使用 Olares 登录直接进入聊天界面。本教程使用 Qwen3.6-35B-A3B 找出一段 Python 函数中的错误，再导出对话。换用其他受支持的 GGUF 模型时，也可以按相同的步骤操作。

## 学习目标

在本教程中，你将学习如何：

- 添加 GGUF 模型文件，根据可用资源设置内存和显存额度，并选择推理模式。
- 使用本地模型检查 Python 函数，并导出对话。
- 将聊天连接切换到 Olares Router 的默认聊天模型。

## 前提条件

- 一台运行 Olares 1.12.7 或更高版本的 amd64 设备，并支持 AVX2。需要更新时可参阅[检查和更新 Olares](/zh/manual/olares/settings/update.md)。
- 主机需有 NVIDIA 设备。采用 GPU 加速时，需要 Blackwell sm_120 GPU 和支持 CUDA 13.0 的驱动。要查看 GPU 型号，打开 Dashboard，在**集群物理资源**中点击 GPU 卡片，再打开 **Graphics management**。此构建已在 RTX 5090 Laptop GPU 上测试。
- 足够存放模型文件的磁盘空间。下文使用的 Qwen3.6-35B-A3B `UD-IQ4_XS` 文件约为 16.5 GiB。

## 添加模型文件

Colibri 会扫描 `Common/huggingface` 中的 GGUF 文件，安装时不会自动下载模型权重。如果模型已经存放在这个目录或其下的 Hugging Face 缓存中，可以直接[安装 Colibri](#安装-colibri-并设置内存和显存)。已有缓存应保留 `snapshots` 目录结构和相对链接。

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

如果还有配套的 `mmproj` 视觉投影文件，应将它与主权重放在同一目录。Colibri 会把它列在模型下方作为视觉附件。能否输入图片取决于加载后的引擎能力。

Colibri 只读使用原有权重，将导入索引和模型选择保存在应用自己的数据目录中。建议尽量把模型放在 SSD 上：内存较少时，Colibri 会按需从磁盘读取专家权重，磁盘速度会影响回复速度。

## 安装 Colibri 并设置内存和显存

安装前，在 Dashboard 中查看可用内存，并在 **Settings** > **GPU** 中查看显存分配情况。如果设备能分出 `32Gi` 内存和 `8Gi` 显存，可以从下面这组数值开始：

1. 打开 Market，搜索“Colibri”，点击 **Get**，再点击 **Install**。
2. 在 **Configure environment variables for Colibri** 窗口中填写：

   | 设置 | 起步值 | 作用 |
   |---|---|---|
   | `MEMORY_LIMIT` | `32Gi` | Colibri 主服务可使用的系统内存上限 |
   | `GPU_MEMORY_LIMIT` | `8Gi` | 分配给 Colibri 的显存额度 |

   :::tip 按设备资源调整额度
   如果设备可用资源较少，就按能分出的额度填写，并从更小的模型开始尝试。显存填 `0` 时改用 CPU 推理。当前安装包即使使用 CPU 推理，也需要安装在有 NVIDIA 设备的 Olares 主机上。
   :::

   其他设置保持默认。

3. 点击 **Confirm**，等待安装完成。

## 扫描并选择本地模型

1. 从启动台打开 Colibri。尚未加载模型时，会进入“本地模型”页面。
2. 点击**重新扫描**，查看 `Common/huggingface` 中发现的模型。列表展示名称、主权重路径、架构和大小，视觉附件显示在所属模型下。分片 GGUF 会合并为一个选项，缺少分片时不能加载。
3. 选中 **Qwen3.6-35B-A3B**。标为不可选的项目会说明原因，例如当前 Colibri 引擎不支持其架构。
4. 点击**加载所选模型**，检查页面内的确认提示，再点击**确认继续**。
5. 等待状态显示“模型已就绪”，然后点击**进入聊天**。大型模型的加载可能需要几分钟，等待期间可以展开**加载日志**查看进度。

![在 Colibri 中选择本地模型，视觉文件作为对应模型的附件展示](/images/manual/use-cases/colibri-moe-models-en.jpg#bordered)

模型选择会保存到应用数据中，重启应用后会重新加载同一个模型，不会自动改选扫描结果中的第一个文件。保存的文件被删除或移动时，重新扫描并选择其他模型即可。

需要更换模型时，在聊天页面左侧点击本地模型图标，新标签页会打开模型管理页面。先等待当前回复结束并导出需要保存的对话，再选择其他模型。新模型就绪后，通过**进入聊天**开启新对话。

点击**停止本地模型**并确认后，推理进程退出并释放已使用的内存和显存。采用 GPU 模式时，应用仍保留 Olares 的 GPU 分配，后续可以重新选择模型。停止状态也会保存，重启后不会自动恢复加载。

## 根据模型调整资源

### 调整内存和显存

换用其他模型时，先用当前设置加载模型，再发一条短消息试用。如果 Colibri 提示内存不足，而设备还有空闲内存，就调高 `MEMORY_LIMIT`。如果模型能加载但回复太慢，可以先尝试调高内存上限。显卡还有余量时，也可以增加显存额度。如果机器已经没有余量，就换用更小的模型或量化文件。

调整内存时，打开 **Settings** > **Applications** > **Colibri** > **Manage environment variables**，修改 `MEMORY_LIMIT` 并应用更改。Colibri 随后会重启并重新加载已选择的模型，操作前请导出需要保存的对话。

调整 `GPU_MEMORY_LIMIT` 时，卸载 Colibri 并保留应用数据，再重新安装并填写新的额度。这项设置安装后不能直接编辑。`Common/huggingface` 中的模型文件仍可使用。

### 选择推理模式

Colibri 默认使用 `standard`。如果想在可用内存较少时尝试受支持的 MoE 模型，可以切换到 `low-memory-moe`。这一模式分阶段使用 GPU，回复可能更慢，尤其是输入较长时。两种模式使用的内存和显存额度都来自安装时填写的数值。

1. 打开 **Settings** > **Applications** > **Colibri** > **Manage environment variables**。
2. 将 `INFERENCE_PROFILE` 设为 `low-memory-moe`，点击 **Confirm**，再点击 **Apply**。Colibri 随后重启。
3. 返回 Colibri，等待“模型已就绪”后再开始新对话。

需要切回默认模式时，将 `INFERENCE_PROFILE` 改回 `standard` 并应用更改。切换前请导出需要保存的对话。

## 使用本地模型检查 Python 函数

### 连接本地引擎

1. 进入聊天页面后，下文按英文界面标注按钮名称。需要切换语言时，点击左下角的设置图标，在 **Settings** > **General** > **Language** 中选择 **English**。
2. 在 **Settings** > **Connection** 中，确认 **Connection service** 为 **Colibri local engine (default)**。保留自动填写的 **API endpoint**。
3. 点击 **Probe server**。出现 **Engine reachable** 后，即可开始对话。连接失败时，先回到本地模型页面确认加载状态。
4. 打开 **Model** 标签页，选择已加载的本地模型。将 **Max output tokens** 设为 `1024`，保持 **Reasoning** 关闭，用于这次简短的代码检查。

网页通过 Olares 登录完成认证，无需在 Colibri 中填写 API key。

### 找出并修复排序错误

1. 点击左侧的 **New chat** 图标，新建对话。
2. 将下面这段简短的 Python 示例粘贴到输入框，点击发送箭头。你也可以换成自己的函数。

   ```text
   这个函数应该按分数从高到低返回前几名。
   请用一句话说明错误，然后修改函数。

   def top_scores(scores, limit=3):
       ranked = sorted(scores)
       return ranked[:limit]
   ```

3. 对照原函数检查修改建议。修改后的函数应返回最高的分数，并按从高到低排序。如果解释不清楚，可以在同一对话中继续提问。

本版本默认提供 16K 上下文，输入、历史消息和输出共同占用这一空间。检查较长文件时，可以一次发送一个函数。内存或显存额度较低时，等待回复的时间可能增加。

### 导出对话

:::warning 刷新页面前先导出对话
Colibri 刷新页面后不会保留聊天记录和生成设置。关闭或刷新页面前，先导出需要保存的对话。
:::

1. 回复完成后，点击页面右上角的 **Export conversation** 图标。
2. 选择 **Markdown (.md)**，将对话下载到电脑。文件包含输入记录和模型回复，方便之后整理或分享。需要结构化保存时，也可以选择 **JSON (.json)**。

## 资源配置与性能参考

对本例的 Qwen3.6-35B-A3B `UD-IQ4_XS`，`32Gi` 内存上限和 `8Gi` 显存额度可以作为一组已测试的起点。资源允许时，增加内存或显存通常有助于缩短等待。较低额度也可能完成请求，但不代表能保持相同速度。

以下数据来自 Olares One、RTX 5090 Laptop GPU，采用 16K 上下文上限、约 1K 输入和 128 tokens 输出。表中取各档第二次请求，缓存可能已经预热。首 token 时间指开始请求至出现第一个文本片段。

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

在 `20Gi` 显存额度下，该模型的全部专家权重可放入显存，因此速度提升明显。其他模型不一定在相同额度下达到这一状态。

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

上述内存对比中的两轮请求均已完成，没有因内存不足而终止或发生引擎重启。`32Gi` 和 `48Gi` 两档可让全部专家权重驻留内存，首次响应明显快于更低的内存档位。

显存对比和内存对比分别测试，相同配置的结果会有波动。测试机与 GPU 为共享使用，每档只有两次请求，实际表现还会受到模型、输入内容和其他负载影响。

这里调整的是容器内存上限。测试主机总内存约 94GiB，共享文件缓存可能跨请求和容器保留。这些结果不能等同于在只有 8GiB 或 12GiB 物理内存的设备上运行。

## 可选：切换到 Olares Router

Colibri 也可以使用 Router 提供的模型。开始前，在 Router 中配置一个可调用的聊天模型，并在 **Default models** 页面将其设为默认聊天模型。连接原理可参阅[使用 Olares Router 作为你的 AI 网关](/zh/use-cases/olares-router.md)。

:::info 切换连接不会停止本地引擎
切换连接服务只改变聊天请求的去向，不会停止 Colibri 的后台本地引擎。采用 GPU 模式时，它仍可能占用显存。调用同一张 GPU 上的 Router 模型前，确认显存能同时容纳两项服务。资源不足时，先为 Router 选择不会与 Colibri 争用这张 GPU 的模型服务。
:::

1. 在 Colibri 中，进入 **Settings** > **Connection**。
2. 将 **Connection service** 改为 **Olares Router**，保留自动填写的地址。无需另外创建或填写 API key。
3. 点击 **Probe server**，然后在 **Model** 标签页选择 `default-chat`。
4. 点击 **New chat**，在新对话中发送同一段 `top_scores` 提示。回复由 Router 当前默认的聊天模型生成。之后在 Router 中更换默认模型时，此连接会跟随切换。

Router 的回复不会出现在 Colibri 本地引擎的性能统计中。要回到本地推理，在 **Connection service** 中选择 **Colibri local engine (default)**，重新探测并选择本地模型，再新建对话。

## 了解更多

- [AI 应用如何连接 Olares 模型](/zh/manual/best-practices/connect-ai-apps.md)：了解 Router 的模型名称、连接地址和鉴权方式。
- [将资源转存到 Olares](/zh/manual/larepass/save-resources-to-olares.md)：使用 LarePass 转存，将模型文件直接下载到 Olares。
- [使用 Olares Router 作为你的 AI 网关](/zh/use-cases/olares-router.md)：管理模型服务和默认聊天模型。
- [Colibri 官方文档（当前构建版本）](https://github.com/JustVugg/colibri/blob/0a5504a1dc64f86fcd0fe7db298b0448c730e7e6/README.md)：查看模型兼容性、运行原理和调优说明。
