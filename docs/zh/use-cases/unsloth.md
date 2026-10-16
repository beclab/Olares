---
outline: [2, 3]
description: 在 Olares 上使用 Unsloth Studio 运行、微调、对比和导出开源模型，并在 JupyterLab 中使用同一套环境。
head:
  - - meta
    - name: keywords
      content: Olares, Unsloth, Unsloth Studio, 模型微调, QLoRA, LoRA, GGUF, JupyterLab, 本地大模型
app_version: "0.0.5"
doc_version: "1.0"
doc_updated: "2026-10-14"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/unsloth.md)为准。
:::

# 使用 Unsloth 微调并运行开源模型

Unsloth Studio 是开源的无代码 Web 界面，可在同一个本地界面中训练、运行和导出开源模型。你可以从 Hugging Face 下载 GGUF 和 safetensors 模型并与其对话，使用 QLoRA、LoRA 或全量微调训练文本模型，再将结果导出为 GGUF 或 safetensors。

在 Olares 上，本应用同时包含 Unsloth Studio 和 JupyterLab。两者共用同一块 GPU、同一个模型缓存和持久化存储，应用重启后，模型、训练记录和导出文件仍会保留。

本文以 `Qwen/Qwen3.5-2B` 模型和 `mlabonne/FineTome-100k` 数据集为例，演示一个小规模流程。

## 学习目标

在本指南中，你将学习如何：
- 为 Unsloth 分配显存。
- 在 Unsloth Studio 中加载模型并对话。
- 运行一次简短的 QLoRA 微调，并与基础模型对比效果。
- 将微调后的模型导出为 GGUF，并保存到 Home 目录。

## 前提条件

- 一台 x86-64（amd64）架构、配备 NVIDIA GPU 的 Olares 设备。
- 能够访问 Hugging Face，用于下载模型和数据集。

## 安装 Unsloth

1. 打开应用市场，搜索 "Unsloth"。
   <!-- ![Unsloth](/images/manual/use-cases/unsloth.png#bordered) -->

2. 点击**获取**，然后点击**安装**。按提示设置以下环境变量，然后等待安装完成：
    - **UNSLOTH_STUDIO_PASSWORD：** 输入 Unsloth Studio 密码，至少 8 个字符。
    - **JUPYTER_PASSWORD：** 输入 JupyterLab 密码，至少 8 个字符。

:::warning 妥善保存 Studio 密码
Unsloth Studio 会在应用首次启动时保存密码。之后修改 **UNSLOTH_STUDIO_PASSWORD** 不会更改已保存的密码。请将密码保存在 Vault 或其他密码管理器中。
:::

## 分配显存

Unsloth 使用一块 NVIDIA GPU。应用可用的显存取决于 GPU 模式：

| GPU 模式 | Unsloth 可用显存 |
|:---|:---|
| 分时共享或独占分配 | 应用可使用 GPU 的全部显存。 |
| 容量切分 | 应用使用启动时分配的配额。默认配额为最低要求 4 GiB，最多可分配 23 GiB。 |

在容量切分模式下，默认配额只够启动应用。训练模型或加载较大的模型前，请分配更大的配额。应用已分配到 GPU 时无法修改配额，需要先将其移除：

1. 前往**设置** > **AI 算力**，点击资源卡片上的**管理**。

2. 在**已分配的应用**中找到 Unsloth，点击**移除**并确认。应用会随之停止。

3. 前往**设置** > **应用**，找到 Unsloth，点击**恢复**。

4. 在**选择 GPU**下选择 GPU，在**当前应用配额**中输入任务所需的显存，然后点击**启动**。

应用停止期间，训练记录、导出文件和已下载的模型都会保留。详见[管理 AI 算力资源](../manual/olares/settings/gpu-resource.md#恢复应用并分配资源)。

## 登录 Unsloth Studio

1. 从启动台打开 Unsloth。

2. 使用用户名 `unsloth` 和你为 **UNSLOTH_STUDIO_PASSWORD** 设置的密码登录。
   <!-- ![登录 Unsloth Studio](/images/manual/use-cases/unsloth-login.png#bordered) -->

JupyterLab 是单独的入口。两个入口都可以在**设置** > **应用** > **Unsloth** > **入口**中找到。

## 使用 Unsloth Studio

### 与模型对话

1. 在聊天页面，打开左上角的 **Select model** 下拉菜单。如果当前在其他页面，先点击左侧边栏中的 **New Chat**。

2. 搜索模型，例如输入 `unsloth/Qwen3.5-2B-GGUF`。

3. 对于 GGUF 仓库，选择适合当前显存配额的量化版本。`Q4_K_M` 或 `UD-Q4_K_XL` 等 4-bit 版本是常用的起点。

4. 等待模型下载并加载完成，然后输入提示词开始对话。
   <!-- ![在 Unsloth Studio 中与模型对话](/images/manual/use-cases/unsloth-chat.png#bordered) -->

对于支持的模型，Studio 会自动设置 temperature、top-p、top-k 等推理参数。你也可以在聊天设置中调整这些参数并编辑系统提示词。

:::tip 复用已下载的模型
Unsloth 会将 Hugging Face 下载内容保存在 Olares 的共享 Hugging Face 缓存中。缓存中已有的模型会被直接识别，无需重新下载。不过，较大的模型仍需要时间将权重读入内存。
:::

### 微调模型

Studio 首页包含四个区域：**Model**、**Dataset**、**Parameters** 和 **Training/Config**。

1. 在 **Model** 区域配置模型：
    - **Model Type：** 选择 **Text**。
    - **Training Method：** 选择 **QLoRA**。三种方法中，它占用的显存最少。
    - **Model：** 搜索并选择 `Qwen/Qwen3.5-2B`。GGUF 模型仅用于推理，不能用于训练。

2. 在 **Dataset** 区域打开 **HuggingFace Hub** 标签页，搜索并选择 `mlabonne/FineTome-100k`。格式保持为 **auto**。

3. 如需快速试跑，将 **Dataset slice** 设置为较小的范围，例如第 `0` 到 `1000` 行。

4. 在 **Parameters** 区域，将 **Max Steps** 设置为 `30`，其他参数保持默认值。

5. 点击 **Start Training**。配置好模型和数据集后，该按钮才可用。

6. 等待模型和数据集下载并加载。随后训练视图会显示 epoch、进度、loss、学习率和 GPU 使用情况。进度条达到 100% 时训练完成。
   <!-- ![Unsloth Studio 训练进度](/images/manual/use-cases/unsloth-training.png#bordered) -->

30 步训练只用于确认整个流程能在你的设备上正常运行，不会明显改变模型行为。实际训练时，请使用针对任务准备的数据集，并增加步数或 epoch。

如需提前结束训练，点击 **Stop Training**。选择 **Stop & Save** 会保存检查点；选择 **Cancel** 则直接停止，不保存检查点。

:::tip 准备自己的数据集
在 **Dataset** 区域打开 **Local** 标签页，可以上传 PDF、DOCX、JSONL、JSON、CSV 或 Parquet 文件。如需将文档整理为训练数据集，可以使用 Data Recipes：创建配方，添加区块，点击 **Validate**，运行预览，再生成完整数据集。
:::

:::tip 受限模型
对于部分 Llama、Gemma 等受限模型，请在 **Model** 区域的 **HuggingFace Token** 字段中粘贴 Hugging Face 访问令牌。该令牌必须拥有对应模型的访问权限。
:::

### 对比微调后的模型

1. 在聊天页面点击 **Compare**，并排对比两个模型。

2. 一侧选择 `Qwen/Qwen3.5-2B`。另一侧在模型选择器中打开 **Fine-tuned** 标签页，选择刚才创建的训练记录。

3. 发送提示词。Studio 会将同一条提示词发送给两个模型。

Studio 会依次加载两个模型。请使用目标任务中的提示词，检查回复是否有所改进。

### 将模型导出为 GGUF

1. 在 Unsloth Studio 中打开 **Export**。

2. 选择训练记录及其检查点。

3. 在 **Choose a format** 下选择 **GGUF / Llama.cpp**，然后选择量化方式，例如 `Q4_K_M`。

4. 保持默认保存位置，点击 **Export**，等待导出完成。转换和量化可能需要几分钟。

5. 如需检查导出文件，在 **Select model** 下拉菜单中打开 **Fine-tuned** 标签页，加载导出的 GGUF。

导出文件保存在 Unsloth 的应用数据中。如需复制到 Home 目录，打开 JupyterLab，选择 **File** > **New** > **Terminal**，然后运行：

```bash
ls /opt/unsloth-studio/exports
cp -r /opt/unsloth-studio/exports/<export-folder> /workspace/host/
```

将 `<export-folder>` 替换为第一条命令列出的文件夹名称。之后即可在文件管理器的 Home 目录中看到这些文件。

## 使用 JupyterLab

1. 打开 JupyterLab 入口，使用你为 **JUPYTER_PASSWORD** 设置的密码登录。

2. 在默认的 `/workspace` 目录中创建 Notebook。该目录中的文件会持久保存在 Unsloth 的应用数据中。

3. 通过 `host` 文件夹读写 Olares Home 目录中的文件。

Notebook 环境已包含 Unsloth Python 包，并与 Studio 共用同一块 GPU 和 Hugging Face 缓存。Studio 和 JupyterLab 中的训练任务会共享 GPU。显存配额较小时，不要同时运行两者的训练任务。

## 常见问题

### Unsloth 的数据保存在哪里？

| 数据 | 位置 |
|:---|:---|
| Hugging Face 模型和数据集 | Olares 共享 Hugging Face 缓存 |
| 训练记录 | 应用数据中的 `/opt/unsloth-studio/outputs` |
| 导出文件 | 应用数据中的 `/opt/unsloth-studio/exports` |
| Notebook | 应用数据中的 `/workspace` |
| Home 目录 | JupyterLab 中的 `/workspace/host` |

大模型会占用大量磁盘空间。不再需要的模型，可以在 Studio 模型搜索中通过删除图标移除。

### 为什么模型无法加载，或训练时显存不足？

模型或训练任务所需的显存超过了应用当前可用的显存。在容量切分模式下，请按照[分配显存](#分配显存)中的步骤分配更大的配额。你也可以换用更小的模型、更低的量化精度，或改用 QLoRA 代替 LoRA 或全量微调。

### 可以生成图像或视频吗？

Studio 中的图像和视频模型远大于文本模型。最小的图像模型组合需要下载约 7 GB，且需要超过默认 4 GiB 的显存配额。最小的视频模型组合需要下载约 61 GB，超出了应用的资源上限，因此 Olares 上不支持视频生成。

### 忘记了 Unsloth Studio 密码，如何重置？

修改 **UNSLOTH_STUDIO_PASSWORD** 不会重置已保存的密码。打开 JupyterLab 终端并运行：

```bash
/opt/unsloth-studio/unsloth_studio/bin/unsloth studio reset-password --username unsloth
```

该命令会生成新密码并显示在终端中，无需重启应用即可生效。使用新密码登录后，请将其保存到 Vault。

## 了解更多

- [管理 AI 算力资源](../manual/olares/settings/gpu-resource.md)：切换 GPU 模式并分配显存配额。
- [Unsloth Studio 文档](https://unsloth.ai/docs/new/studio)：训练、Data Recipes 和导出选项的上游参考。
