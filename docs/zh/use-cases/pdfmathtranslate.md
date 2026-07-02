---
outline: [2, 4]
description: 学习如何在 Olares 上安装和配置 PDFMathTranslate。本教程将指导你使用本地 AI 模型 Ollama 翻译学术 PDF，同时保留原始排版和数学公式。
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/pdfmathtranslate.md)。
:::

# 翻译学术 PDF 并保留排版

PDFMathTranslate 是一款用于翻译学术 PDF 文档的应用，能够在保留原始排版和数学公式的同时完成翻译。

本教程提供在 Olares 上安装和使用 PDFMathTranslate 的说明，使用本地 AI 模型 Ollama 进行翻译。

## 学习目标

完成本教程后，你将能够：
- 配置本地 AI 模型用于 PDF 翻译。
- 翻译学术 PDF 并管理输出文件。

## 前提条件

开始前，请确保：
- Ollama 已安装并在你的 Olares 环境中运行。
- 已使用 Ollama 安装至少一个模型。更多信息，请参阅 [Ollama](ollama.md)。

## 安装 PDFMathTranslate

1. 打开 Olares Market，搜索 "PDFMathTranslate"。
2. 点击 **Get**，然后点击 **Install**。

    ![Install PDFMathTranslate](/images/manual/use-cases/install-pdfmathtranslate.png#bordered)

3. 安装完成后，点击 **Open**。PDFMathTranslate 工作区将显示出来。

    ![Open PDFMathTranslate](/images/manual/use-cases/open-pdfmathtranslate.png#bordered)

## 翻译

### 上传你的 PDF 文档

:::warning PDF 格式要求
请确保 PDF 文件是标准 PDF 文档，未受密码保护且未损坏。该应用无法处理无效的 PDF 文件。
:::

在 **File** 区域，选择你的输入 **Type**：
- 如果选择 **File**，将 PDF 文档拖放到上传区域，或点击该区域浏览本地存储。

    文档上传后，右侧 **Preview** 窗格中会显示其预览。

- 如果选择 **Link**，输入 PDF 文档的链接地址。

    此步骤中 **Preview** 窗格保持空白。翻译完成后，文档内容才会显示在该区域。

### 配置翻译服务

要使用本地 AI 服务 Ollama 进行翻译，请配置以下设置：

1. 从 **Service** 列表中，选择 **Ollama**。
2. （可选）获取 Ollama 主机地址：

    a. 前往 Olares **Settings** > **Application** > **Ollama**。

    b. 在 **Entrances** 部分，点击 **Ollama API**。

    c. 点击 **Set up endpoint**，然后点击 <i class="material-symbols-outlined">content_copy</i> 复制端点地址。

    ![Obtain Ollama host address](/images/manual/use-cases/onetest02-endpoint-entrances-ollama-api.png#bordered){width=80%}


3. 在 **OLLAMA_HOST** 字段中，输入 Ollama 主机地址。
4. 在 **OLLAMA_MODEL** 字段中，输入你已下载的 Ollama 模型的精确标识符。例如，`gemma3:4b`。

    ![Open PDFMathTranslate](/images/manual/use-cases/local-model-setup.png#bordered)

### 选择语言和范围

1. 选择源语言和目标语言：

    a. **Translate from** 表示原始文档的语言。

    b. **Translate to** 表示你想要阅读的语言。

    :::info
    PDFMathTranslate 不会自动检测语言。你必须手动选择。支持的语言包括英语、简体中文、繁体中文、法语、德语、日语、韩语、俄语、西班牙语和意大利语。
    :::

2. 指定要翻译的页面：
    * **All**：翻译整个文档。
    * **First**：仅翻译第一页。
    * **First 5 pages**：翻译前五页。
    * **Others**：翻译自定义页面范围。

    ![Set translation scope in PDFMathTranslate](/images/manual/use-cases/set-translation-scope.png#bordered)

3. 点击 **Translate**。翻译将立即开始。

    :::warning
    翻译过程中请勿点击 **Cancel** 按钮。这可能会导致进程报错。
    :::

### 下载你的文件

翻译完成后，翻译后的文件将显示在 **Preview** 窗格中，同时应用会生成三个文件：

- 原始源文件
- 翻译后的文件
- 双语版本

你可以在 Files 应用中找到这些文件。要将它们下载到本地计算机，可以直接从 PDFMathTranslate 工作区下载，或使用 Files 应用。

#### 在 PDFMathTranslate 工作区中

在 pdfmathtranslate 工作区左侧的 **Translated** 部分，点击文件旁边的下载按钮。

![Download files translated in PDFMathTranslate](/images/manual/use-cases/download-translated-files.png#bordered)

#### 从 Olares Files 应用

1. 打开 Files 应用，然后前往 **Data** > **pdfmathtranslate** > **pdfmathtranslate**。

    ![Access files translated by PDFMathTranslate](/images/manual/use-cases/access-translated-files.png#bordered)

2. 双击文件，然后点击右上角的下载图标。

    ![Download files translated from Olares Files](/images/manual/use-cases/download-in-files.png#bordered)

## 常见问题

### 为什么翻译进度条消失了？

如果翻译过程中刷新页面，进度条可能会从屏幕上消失。这只是显示问题，翻译仍在后台处理中。

等待其完成或检查输出文件夹。

### 应用会保留同一文件的多个版本吗？

如果你多次翻译同一个文件名，系统会用新版本替换旧版本。它不会创建带编号的副本，例如 `file_1.pdf`。要保留多个版本，请在再次翻译前重命名你的源文件。

### 如果应用无响应或耗时异常长，该怎么办？

如果翻译耗时明显比平时长，或应用停止响应，后台进程可能已卡住。要解决此问题，请卸载然后重新安装 pdfmathtranslate。

### 如何彻底卸载 PDFMathTranslate？

要完全移除应用及其数据：
1. 从 Market 或 Desktop 卸载应用。更多信息，请参阅 [卸载应用](../manual/olares/market/market.md#uninstall-applications)。
2. 打开 Files 应用，前往 **Application** > **Data**，然后删除 `pdfmathtranslate` 文件夹。
