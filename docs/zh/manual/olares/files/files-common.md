---
outline: [2, 3]
description: 了解如何在 Olares 中使用 Common 目录管理跨应用、跨用户共享的 AI 模型。
head:
  - - meta
    - name: keywords
      content: Olares, 文件管理器, 公共目录, AI 模型, Hugging Face, ComfyUI, Ollama, 共享存储
---

# 使用公共目录管理共享 AI 模型

公共目录用于存放可供多个应用复用的 AI 模型，节省存储空间并避免重复下载。成员通过文件管理器访问这些文件时，需要单独获得授权。

## 了解公共目录

### 公共目录的工作原理

多个 AI 应用可能需要相同的模型文件。例如，vLLM 和 llama.cpp 可以使用同一份 Hugging Face 缓存。公共目录为这些模型提供统一的存放位置，让应用复用文件，无需各自保存一份副本。

应用使用共享模型，与成员通过文件管理器访问模型目录，是两种不同的访问方式。为成员授予公共目录的访问权限，只控制其在文件管理器中浏览和管理文件，不会配置应用加载模型的方式。

![公共目录界面](/images/zh/manual/olares/files-common.png#bordered)

### 公共目录的结构

公共目录默认包含以下三个子目录，分别遵循对应平台的官方存储结构：

| 子目录 | 用途 | 目录结构说明 |
|:------|:----|:-----------|
| **huggingface** | 存储通过 Hugging Face CLI 下载的缓存模型。<br>vLLM、transformers 和 llama.cpp 等应用都从<br>这个统一缓存中读取数据。 | 完全镜像官方的缓存结构。<br><br>详细了解文件组织方式，请参考 [Hugging Face 官方缓存管理指南](https://huggingface.co/docs/huggingface_hub/guides/manage-cache)。 |
| **comfyui** | 存储在 ComfyUI 及相关应用之间共享的模型<br>（如 Checkpoint、LoRA 和 VAE）。 | 遵循标准的 ComfyUI `models` 文件夹结构。<br><br>详细了解文件组织方式，请参考 [ComfyUI 官方模型文档](https://docs.comfy.org/development/core-concepts/models)。 |
| **ollama** | 存储由 Ollama 拉取和管理的模型。 | 采用 Ollama 独特的 manifests 和 blobs 存储机制。<br><br>更多信息，请参考 [Ollama 官方 FAQ](https://docs.ollama.com/faq#where-are-models-stored)。 |

## 迁移模型到公共目录

Olares v1.12.6 默认提供公共目录。如果你使用的是旧版本 Olares，请按以下步骤迁移现有模型。

1. 将 Olares 升级至 v1.12.6。
2. 卸载现有的 AI 模型服务应用：

    - 预打包的单模型应用
    - Ollama（作为基础引擎安装，用来手动拉取模型）。卸载时请勿勾选**同时删除所有本地数据**。
    - ComfyUI。卸载时请勿勾选**同时删除所有本地数据**。

3. 安装新版本并恢复模型：

    <Tabs>
    <template #单模型应用>

    使用新的[引擎基座应用](/zh/use-cases/llm-base-apps.md)（llama.cpp 引擎基座、Ollama 引擎基座、SGLang 引擎基座，或 vLLM 引擎基座）部署所需模型。

    模型文件将保存到相应的 `huggingface` 或 `ollama` 子目录中。

    </template>

    <template #Ollama-模型>

    对于通过 Ollama 拉取的模型：

    - 如需重新下载，使用新的 Ollama [引擎基座应用](/zh/use-cases/llm-base-apps.md)部署所需模型。模型文件将自动保存到 `ollama` 子目录中。
    - 如需复用现有模型，手动将其复制到 `ollama` 子目录中。

    </template>

    <template #ComfyUI>

    从应用商店安装新版本的 ComfyUI。系统会自动将现有模型文件移动到 `comfyui` 子目录中。

    :::tip 区分 v2 与新版应用
    在应用详情页，查看**信息**面板中的**兼容性**字段：
    - v2 共享应用显示 `Olares >=1.12.3-0, <1.12.6`。
    - 新版共享应用显示 `Olares >=1.12.6-0`。
    :::

    </template>
    </Tabs>

## 查找和管理共享模型文件

在 Olares 1.12.8 中，成员默认无权访问公共目录中的文件夹。请超级管理员或有管理权限的管理员，为你需要访问的文件夹授予**可查看**或**可编辑**权限。详见[管理公共目录和外部存储的访问权限](manage-common-external-access.md)。

### 访问公共目录

1. 从启动台打开文件管理器。
2. 选择左侧边栏中的 **应用** > **公共**。
3. 打开你有权访问的 `huggingface`、`ollama` 或 `comfyui` 文件夹。
4. 找到目标模型文件。

### 管理模型文件

添加或删除模型文件需要对该公共目录文件夹拥有**可编辑**或管理权限。只有**可查看**权限时，可以浏览和下载文件，但不能修改内容。

请保持各文件夹要求的目录结构，否则应用可能无法正确加载模型。

- **添加模型**：直接将模型文件拖放到对应的子目录中。
:::warning 删除模型前
删除共享模型文件会影响所有使用它的应用。删除前，请备份模型文件，或确认可以重新下载。
:::

- **删除模型**：打开你有权编辑的公共目录文件夹，右键点击其中的模型文件或文件夹，选择**删除**。拥有可编辑权限的成员可以删除授权文件夹内的文件和子文件夹，但不能删除该文件夹本身。
