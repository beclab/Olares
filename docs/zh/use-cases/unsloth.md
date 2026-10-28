---
outline: [2, 3]
description: 在 Unsloth Studio 中使用 FineTome 数据微调 Qwen3.5-2B，对比回答，在 Olares 文件管理器中找到 GGUF，并通过 JupyterLab 运行已保存的适配器。
head:
  - - meta
    - name: keywords
      content: Olares, Unsloth, Unsloth Studio, 模型微调, QLoRA, LoRA, GGUF, JupyterLab, 本地大模型
app_version: "0.0.8"
doc_version: "1.3"
doc_updated: "2026-10-28"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/unsloth.md)为准。
:::

# 使用 Unsloth 微调并运行开源模型

Unsloth 帮助你在本地运行和微调 AI 模型。在 Olares 上，这个应用包含两个入口：Unsloth Studio 提供操作模型的 Web 界面，JupyterLab 则提供运行 Python Notebook 的编程环境。

本指南使用 `Qwen/Qwen3.5-2B`，从 `mlabonne/FineTome-100k` 中选取约 1,000 条对话，进行一轮 LoRA 微调。数据包含编程解释和通用知识问答。这次小样本训练的重点是掌握从选取数据到导出模型的完整流程。如果要让模型适应自己的任务或回答风格，请使用能体现理想回答的对话数据。

## 学习目标

在本指南中，你将学习如何：

- 为 Unsloth 分配显存。
- 在 Unsloth Studio 中加载模型并对话。
- 使用 LoRA 微调模型，并与基础模型对比效果。
- 将微调后的模型导出为 GGUF，并在 Olares 文件管理器中找到文件。
- 按需通过 Python Notebook 运行已保存的 LoRA 适配器。

## 了解 Unsloth Studio 和 JupyterLab

Olares 中的 Unsloth 应用提供两个入口，分别适合不同的使用方式：

| 入口 | 提供什么 | 什么时候使用 |
|:---|:---|:---|
| Unsloth Studio | 用于对话、训练、对比和导出模型的 Web 界面 | 通过 Web 界面完成本指南的主流程。 |
| JupyterLab | 包含 Notebook、终端和 Unsloth Python 包的编程环境 | 用 Python 加载已保存的模型、处理自己的数据，或自定义训练代码。 |

两者都运行在 Olares 设备上，共用 GPU、模型缓存和应用文件。两个入口分别使用各自的密码登录。

## 前提条件

- 一台 x86-64（amd64）架构、配备 NVIDIA GPU 的 Olares 设备。
- 能够访问 Hugging Face，用于下载模型和数据集。

## 安装 Unsloth

1. 打开应用市场，搜索 "Unsloth"。

2. 点击**获取**，然后点击**安装**。按提示设置以下环境变量：
    - **UNSLOTH_STUDIO_PASSWORD：** 输入 Unsloth Studio 密码，至少 8 个字符。
    - **JUPYTER_PASSWORD：** 输入 JupyterLab 密码，至少 8 个字符。

3. 确认设置，等待安装完成。

:::warning 妥善保存 Unsloth Studio 密码
Unsloth Studio 会在首次启动时使用 **UNSLOTH_STUDIO_PASSWORD** 设置登录密码。请将该密码保存在 Vault 或其他密码管理器中。后续需要更改密码时，请按照[密码重置步骤](#忘记了-unsloth-studio-密码-如何重置)操作。
:::

## 分配显存

Unsloth 使用一块 NVIDIA GPU。应用可用的显存取决于 GPU 模式：

| GPU 模式 | Unsloth 可用显存 |
|:---|:---|
| 分时共享或独占分配 | 应用可使用 GPU 的全部显存。 |
| 容量切分 | 应用使用启动时分配的配额。启动窗口会显示最低要求和可用配额。 |

Olares 显示的最低配额是应用的启动要求，训练可能需要更多显存。作为参考，Unsloth 的 [Qwen3.5 指南](https://unsloth.ai/docs/models/qwen3.5/fine-tune)列出的 Qwen3.5-2B bf16 LoRA 显存用量约为 5 GB。实际用量受上下文长度和批次大小影响，分配时应在此估算值之上留出余量。

在容量切分模式下，先将应用从已分配的 GPU 中移除，再恢复应用并设置新配额：

1. 前往**设置** > **AI 算力**，点击资源卡片上的**管理**。

2. 在**已分配的应用**中找到 Unsloth，点击**移除**并确认。应用会随之停止。

3. 前往**设置** > **应用**，找到 Unsloth，点击**恢复**。

4. 在**选择 GPU**下选择 GPU，在**当前应用配额**中输入可用配额范围内的数值，然后点击**启动**。

应用停止期间，训练记录、导出文件和已下载的模型都会保留。详见[管理 AI 算力资源](../manual/olares/settings/gpu-resource.md#恢复应用并分配资源)。

## 登录 Unsloth Studio

1. 从启动台打开 Unsloth。

2. 使用用户名 `unsloth` 和你为 **UNSLOTH_STUDIO_PASSWORD** 设置的密码登录。

## 使用 Unsloth Studio

### 与模型对话

1. 在聊天页面，打开左上角的 **Select model** 下拉菜单。如果当前在其他页面，先点击左侧边栏中的 **New chat**。

2. 搜索模型，例如输入 `unsloth/Qwen3.5-2B-GGUF`。

3. 对于 GGUF 仓库，选择适合当前显存配额的量化版本。`Q4_K_M` 或 `UD-Q4_K_XL` 等 4-bit 版本是常用的起点。

   ![在 Unsloth Studio 中选择 GGUF 量化版本](/images/manual/use-cases/unsloth-select-model.png#bordered)

4. 等待模型下载并加载完成，然后输入提示词。

   ![在 Unsloth Studio 中与模型对话](/images/manual/use-cases/unsloth-chat.png#bordered)

对于支持的模型，Studio 会自动设置 temperature、top-p、top-k 等推理参数。你也可以在聊天设置中调整这些参数并编辑系统提示词。

:::tip 复用已下载的模型
Unsloth 会将 Hugging Face 下载内容保存在 [Olares 的共享 Hugging Face 目录](../manual/olares/files/files-common.md#查找和管理共享模型文件)中。Studio 会复用缓存中的模型。加载时，请留出将权重读入内存的时间，较大的模型通常需要更久。
:::

### 微调模型

打开 **Train** > **Configure** 配置训练。LoRA 会训练一组称为适配器的额外权重，基础模型的权重保持不变。

本例按照 [Unsloth 的 Qwen3.5 训练建议](https://unsloth.ai/docs/models/qwen3.5/fine-tune)，使用 bf16 LoRA。

1. 在 **Model** 区域设置以下字段：
    - **MODEL：** 搜索并选择 `Qwen/Qwen3.5-2B`，使用其中的模型权重进行训练。对话时使用 GGUF 版本。
    - **METHOD：** 选择 **LoRA**。
    - **TRAIN AS：** 选择 **Language model**。

    :::tip 受限模型
    对于部分 Llama、Gemma 等受限模型，请在 **Model** 区域打开 **HF TOKEN**，输入 Hugging Face 访问令牌。该令牌必须拥有对应模型的访问权限。
    :::

2. 在 **Dataset** 区域，通过 **DATASET** 选择器搜索并选择 `mlabonne/FineTome-100k`。将 **Subset** 保持为 **default**，**Train Split** 保持为 **train**。点击 **View dataset** 查看对话内容。[数据集预览](https://huggingface.co/datasets/mlabonne/FineTome-100k)中包含编程、知识解释和教学等示例。为自己的任务选择数据时，需要同时检查用户问题和对应回答。
    :::tip 准备自己的数据集
    训练自己的助手时，准备包含目标指令和理想回答的对话数据，在 **Dataset** 区域点击 **Drop file or click to upload** 上传结构化数据集。单独留出一些评估样本，用于后续对比模型。

    如果需要从原始文档生成数据集，请按照 [Data Recipes 指南](https://unsloth.ai/docs/new/studio/data-recipe)操作，再选择生成的数据集进行训练。
    :::

3. 展开 **Dataset** 区域内的 **Advanced**，设置以下字段，先使用一小部分数据训练：
    - **Target Format：** 保持为 **Auto**。
    - **Train Split Start：** 输入 `0`。
    - **Train Split End：** 输入 `1000`。

4. 在 **Parameters** 区域切换到 **Advanced**。将 **Context Length** 设为 `2048`。点击 **Max Steps** 旁的 **Use Epochs**，再将训练轮数设为 `1`，将选中的数据训练一遍。
    :::tip 缩短试跑时间
    如果想用更少的步骤试跑，可以切换到 **Max Steps** 并输入 `30`。
    :::

5. 在 **Parameters** 区域内展开 **Training Hyperparameters**，选择 **Optimization**，将 **Batch Size** 设为 `1` 以减少显存占用。其他参数保持默认值。

6. 点击 **Start Training**。配置好模型和数据集后，该按钮才可用。

7. 等待依赖准备就绪，以及模型和数据集下载并加载完成。这些准备工作发生在第一个训练步骤之前。首次运行时，下载时间取决于网络连接和模型大小，可以在准备窗口中查看下载进度。

   ![训练前的模型下载和准备进度](/images/manual/use-cases/unsloth-training.png#bordered)

训练开始后，**Current Run** 页面会显示进度、损失值、学习率和 GPU 使用情况。训练完成后，页面会显示 **Completed**、**100% complete** 和模型保存路径，此时可以继续对比或导出模型。

![训练完成后显示对比和导出按钮](/images/manual/use-cases/unsloth-training-completed.png#bordered)

### 对比微调后的模型

1. 训练完成后，在训练界面点击 **Compare in Chat**。Studio 会加载训练后的模型，并打开与其基础模型的对比页面。

2. 确认两侧分别标有 **BASE MODEL** 和 **FINE-TUNED**。在 **Send to both models...** 输入框中输入以下提示词：

   ```text
   I am learning Python recursion. This function should add all the numbers in a list,
   but it gives the wrong result. Find the bug, explain why it happens, and fix it.
   Keep the recursive approach and do not use sum().

   def total(numbers):
       if not numbers:
           return 1
       return numbers[0] + total(numbers[1:])

   Show the results for [], [5], and [3, 1, 4].
   ```

3. 发送提示词，对比两侧的回答，看看模型是否修正了递归终止条件、清楚解释了错误原因，并按要求保留递归写法。

   ![对比基础模型与微调模型对递归问题的回答](/images/manual/use-cases/unsloth-compare-answers.png#bordered)

再尝试几道编程问题，对比代码正确性、指令遵循情况和解释是否清楚。保持两侧的系统提示词和生成参数一致，用未参与训练的问题评估回答，再据此改进下一轮训练的数据。

### 将模型导出为 GGUF

1. 在已完成的训练界面点击 **Export to GGUF**，打开 **Export Model** 页面并预选该训练记录。

2. 在 **Source** 中选择 **Fine-tuned**。检查 **Training Run** 和 **Checkpoint**，确认要导出的训练记录和已保存权重。

3. 在 **Export Method** 中选择 **GGUF / Llama.cpp**。将 **Export target** 设为 **Full model**，把适配器合并到基础模型后，导出完整的 GGUF 模型。

4. 在 **Quantization Levels** 中选择推荐的 **Q4_K_M**。本例保持 **Importance matrix (imatrix)** 和 **Also convert to Q4NX (AMD NPU)** 关闭。

5. 点击 **Export Configuration** 底部的 **Export Model**。在展开的面板中选择 **Save Locally**，保留默认的 **Save folder**。这里会将文件保存在 Olares 设备上。

6. 点击 **Start Export**。等待出现 **Export finished successfully.**，记下面板中显示的输出路径。转换和量化需要一些时间。导出完成后点击 **Done**。

   ![Unsloth Studio 中 GGUF 导出成功的界面](/images/manual/use-cases/unsloth-export-model.png#bordered)

### 在 Olares 文件管理器中找到 GGUF 文件

1. 打开文件管理器，在 **Application** 下选择 **Data**。

2. 依次打开 **unsloth** > **studio** > **exports**，再打开本次导出的文件夹。
    :::tip 查找自定义导出目录
    使用默认设置时，文件夹名称以 `-GGUF` 结尾。如果修改过 **Save folder**，请根据导出完成时显示的输出路径查找。应用内的 `/opt/unsloth-studio/` 对应文件管理器中的 **Data** > **unsloth** > **studio**。
    :::

3. 找到以 `.Q4_K_M.gguf` 结尾的语言模型文件，例如 `Qwen3.5-2B.Q4_K_M.gguf`。可以直接通过文件管理器下载该文件，再交给兼容 GGUF 的模型运行工具使用。

   ![Olares 文件管理器中的 Q4_K_M 模型和视觉投影文件](/images/manual/use-cases/unsloth-find-gguf-file-in-files.png#bordered)

:::info 训练文件和视觉投影组件
同级的 **outputs** 文件夹保存训练记录，包括 LoRA 适配器和检查点。上一步导出的 GGUF 位于 **exports** 中。文件名带有 `mmproj` 的文件保存模型的视觉投影组件，供兼容的运行工具处理图片输入。本例的纯文本任务使用主模型文件 `.Q4_K_M.gguf`。
:::

## 在 JupyterLab 中运行微调模型（可选）

如果想用 Python 调用模型，例如在 Notebook 中尝试多条提示词，可以按照下面的步骤操作。此示例参考 [Unsloth 官方推理指南](https://unsloth.ai/docs/basics/inference-and-deployment/unsloth-inference)，加载训练时保存的 LoRA 适配器及其基础模型。

加载模型前，先等待 Studio 中的训练或导出任务结束。Studio 和 JupyterLab 共用 GPU，而 Notebook 需要显存来单独加载一份模型。

1. 打开 JupyterLab 入口，使用安装时为 **JUPYTER_PASSWORD** 设置的密码登录。

2. 点击文件浏览器中的 **+**，在 Launcher 的 **Notebook** 下选择 **Python 3 (ipykernel)**。将 Notebook 保存为 `try-finetuned-model.ipynb`，放在默认的 `/workspace` 目录中。

3. 在文件管理器中打开 **Data** > **unsloth** > **studio** > **outputs**，找到本次训练的文件夹。将下方的 `RUN_FOLDER` 替换为该文件夹名称，再把代码粘贴到 Notebook 的第一个单元格中。例如，本指南截图中的训练文件夹名为 `Qwen_Qwen3.5-2B_1792165828`。

   ```python
   from unsloth import FastLanguageModel
   import torch

   model, tokenizer = FastLanguageModel.from_pretrained(
       model_name="/opt/unsloth-studio/outputs/RUN_FOLDER",
       max_seq_length=2048,
       dtype=torch.bfloat16,
       load_in_4bit=False,
   )
   FastLanguageModel.for_inference(model)
   ```

4. 按 **Shift + Enter** 加载模型。等待 **Loading weights** 达到 **100%**，并确认单元格执行完成、没有报错，再继续下一步。执行完成后，单元格左侧的 `[*]` 会变为数字编号。

   ![在 JupyterLab 中加载模型权重](/images/manual/use-cases/unsloth-load-model-weights-in-jupyterlab.png#bordered)

5. 将以下代码粘贴到同一个 Notebook 的下一个单元格，按 **Shift + Enter**。回答会显示在单元格下方。修改用户消息并重新运行这个单元格，即可尝试其他提示词。

   ```python
   messages = [
       {"role": "user", "content": "Explain Python recursion in one sentence."}
   ]
   prompt = tokenizer.apply_chat_template(
       messages,
       tokenize=False,
       add_generation_prompt=True,
       enable_thinking=False,
   )
   inputs = tokenizer(
       text=prompt, return_tensors="pt", add_special_tokens=False
   ).to(model.device)

   with torch.inference_mode():
       outputs = model.generate(**inputs, max_new_tokens=256, do_sample=False)

   answer = outputs[0, inputs["input_ids"].shape[1]:]
   print(tokenizer.decode(answer, skip_special_tokens=True))
   ```

   ![在 JupyterLab 中使用微调模型回答问题](/images/manual/use-cases/unsloth-try-with-prompt.png#bordered)

6. 使用完毕后保存 Notebook，选择 **Kernel** > **Shut Down Kernel**，释放 Notebook 占用的 GPU 显存。
    :::tip 在文件管理器中查找 Notebook
    保存在 `/workspace` 下的 Notebook，可以在文件管理器的 **Data** > **unsloth** > **workspace** 中找到。
    :::

## 常见问题

### Unsloth 的数据保存在哪里？

| 数据 | Olares 文件管理器中的位置 | 应用内路径 |
|:---|:---|:---|
| 训练记录和 LoRA 适配器 | **Data** > **unsloth** > **studio** > **outputs** | `/opt/unsloth-studio/outputs` |
| 默认的 GGUF 导出文件 | **Data** > **unsloth** > **studio** > **exports** | `/opt/unsloth-studio/exports` |
| Notebook | **Data** > **unsloth** > **workspace** | `/workspace` |
| Home 目录 | **Home** | `/workspace/host` |

下载的 Hugging Face 模型和数据集使用 Olares 的共享 Hugging Face 缓存。

如需释放磁盘空间，可以在 Studio 模型搜索中通过删除图标移除闲置模型。

### 为什么模型无法加载，或训练时显存不足？

如果 Studio 提示内存不足，说明任务需要更多可用内存。在容量切分模式下，可以在可用容量范围内[增加应用的显存配额](#分配显存)。同时停止使用同一块 GPU 的其他训练任务。

训练时，可以减小 **Batch Size** 或 **Context Length**，也可以换用更小的模型。对话时，可以选择更小的模型或更低位宽的 GGUF 版本。选择 LoRA 或 QLoRA 时，请遵循对应模型的训练建议。

### 可以生成图像或视频吗？

Studio 提供图像和视频生成功能。资源需求取决于模型、精度、输出分辨率和视频时长。

生成图像时，打开 **Images**，选择适合当前资源的模型。在容量切分模式下，应用的初始显存配额为 4 GiB。如果模型需要更多显存，请在加载前[增加配额](#分配显存)。同时为模型及其所需组件预留足够的磁盘空间。模型选择和参数设置可参考[官方图像生成指南](https://unsloth.ai/docs/basics/diffusion-image)。

生成视频前，分别检查下载所需的磁盘空间，以及运行模型所需的显存和系统内存。当前 Olares 应用的系统内存上限为 48 GiB。选择模型和参数时，需同时满足这一上限和可用显存的限制。

### 如何修改 JupyterLab 密码？

前往**设置** > **应用** > **Unsloth**，在**环境变量**区域点击**管理环境变量**，编辑 **JUPYTER_PASSWORD**。输入至少 8 个字符的新密码，点击**确认**，然后点击**应用**。应用会重启以使更改生效。重启完成后，使用新密码登录 JupyterLab。

### 忘记了 Unsloth Studio 密码，如何重置？

使用 Unsloth Studio 的密码重置命令生成新的登录密码。打开 JupyterLab 终端并运行：

```bash
/opt/unsloth-studio/unsloth_studio/bin/unsloth studio reset-password --username unsloth
```

新密码会显示在终端中并立即生效。使用新密码登录后，请将其保存到 Vault。

### Unsloth 会在加载模型前占用 4 GiB 显存吗？

从应用版本 0.0.8 起，Olares 安装包仅支持 NVIDIA GPU 模式。

GPU 模式显示的最低 4 GiB 是整个应用的显存配额，包含已加载模型使用的显存，并不是 Studio 在加载模型前固定占用的 4 GiB 开销。在容量切分模式下，这个配额会减少其他应用可分配的额度。实际显存占用取决于模型和任务。具体操作见[分配显存](#分配显存)。

### Unsloth 会使用系统的 Hugging Face 配置吗？

会。Unsloth Studio、训练工作进程和 JupyterLab 会继承 Olares 中配置的 Hugging Face 服务地址和访问令牌。

要修改配置，前往 Olares **设置** > **高级** > **系统环境变量**。编辑 `OLARES_SYSTEM_HUGGINGFACE_SERVICE` 设置服务地址，编辑 `OLARES_SYSTEM_HUGGINGFACE_TOKEN` 设置令牌，然后点击**确认**和**应用**。这些是共享的系统设置，修改也会影响其他使用这些配置的应用。具体操作见[设置系统级环境变量](../manual/manage-system-env.md)。

下载公开模型不需要令牌。下载受限模型时，令牌所属的 Hugging Face 账号必须拥有该模型的访问权限。

### 为什么回答结束后仍然占用显存？

Unsloth 会保留已加载的模型供下一次请求使用，即使 GPU 利用率回到 0%，模型仍可能占用显存。从应用版本 0.0.8 起，通过 llama.cpp 加载的 GGUF 对话模型默认空闲 5 分钟后卸载，不中断正在生成的回答，也不删除模型文件。此设置不影响训练任务。要释放 JupyterLab 笔记本占用的显存，请关闭其内核。

要修改应用的默认超时时间，前往 Olares **设置** > **应用** > **Unsloth** > **环境变量** > **管理环境变量**。编辑 `UNSLOTH_MODEL_IDLE_TTL`，点击**确认**，再点击**应用**。应用会重启以应用新配置。

| 值 | 默认行为 |
|:---|:---|
| `300` | 空闲 5 分钟后卸载。 |
| `0` | 关闭自动卸载。 |
| 大于或等于 `60` 的整数 | 空闲达到指定秒数后卸载。 |

尚未在 Studio 中保存空闲超时时间时，即使 **Switch model by request** 关闭，环境变量中的默认值仍会生效。Studio 已保存的设置会覆盖这个环境变量，即使环境变量设为 `0` 也不例外。如果你已经在 Studio 中保存过空闲超时时间，请在 Studio 中调整：

1. 在 Unsloth Studio 中打开 **Settings** > **API**，找到 **Model auto-switch (OpenAI API)**。
2. 开启 **Switch model by request**，在 **Idle auto-unload** 中填写所需秒数，然后点击 **Save**。设为 `0` 可关闭自动卸载。关闭 **Switch model by request** 时，已保存的超时设置不会生效。开启该开关也会允许 API 请求按名称加载已下载的模型。
3. 要让通过 Studio 手动加载的模型也能自动卸载，关闭 **Only unload models loaded by the API**。

此外，Studio **Settings** > **System** > **Model memory** 中的 **Keep model in GPU memory** 必须关闭，自动卸载才会生效。

如需手动释放显存，打开 **Chat** > **模型选择器**，选择右下角的 **Eject model** 或 **Eject all**。需要使用时，重新加载模型即可。

卸载模型不会改变 Olares 的容量切分配额。如需将配额分配给其他应用，请停止 Unsloth 或移除其 GPU 分配。具体操作见[管理 AI 算力资源](../manual/olares/settings/gpu-resource.md)。

### 升级后已有模型会怎样？

升级会保留已下载的模型文件、应用数据和 Studio 已保存的设置。服务重启后需重新加载模型。已保存的自动卸载设置仍会覆盖应用的默认超时时间，具体见[为什么回答结束后仍然占用显存？](#为什么回答结束后仍然占用显存)。

### 可以在 Unsloth 内部更新 llama.cpp 吗？

可以。使用 Unsloth Studio 中的 llama.cpp 升级按钮即可更新。更新程序会卸载当前模型，请等待更新完成并重新加载模型后，再发送请求。

Unsloth 使用自己的 llama.cpp 副本。更新它不会升级或替换 Olares 的独立 llama.cpp 引擎应用。更新后的副本保存在 Unsloth 的持久化应用数据中，停止并恢复应用后仍会保留。

## 了解更多

- [管理 AI 算力资源](../manual/olares/settings/gpu-resource.md)：切换 GPU 模式并分配显存配额。
- [Qwen3.5 微调指南](https://unsloth.ai/docs/models/qwen3.5/fine-tune)：针对该系列模型的训练建议。
- [Unsloth Studio 文档](https://unsloth.ai/docs/new/studio)：训练、Data Recipes 和导出选项的上游参考。
