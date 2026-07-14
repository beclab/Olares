---
noindex: true
outline: [2, 5]
description: 在 Olares 上安装和使用 Stirling PDF 的分步指南。学习如何安全地管理 PDF 文档，涵盖合并、编辑、格式转换以及创建自动化工作流流水线等核心任务。
head:
  - - meta
    - name: keywords
      content: Olares, Stirling PDF, self-hosted PDF tools, merge PDF, edit PDF, PDF converter, Stirling PDF on Olares
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/stirling-pdf.md)。
:::

# 使用 Stirling PDF 管理 PDF 文档

Stirling PDF 是一款功能强大的本地托管工具，旨在安全高效地管理 PDF 文档。

## 学习目标

完成本教程后，你将能够：

* 安全地编辑（涂黑）敏感信息。
* 通过合并、旋转、拆分和重新排序页面来整理文档。
* 在 PDF 和其他格式之间转换文档。
* 创建自动化工作流以批量处理 PDF 文档。

## 安装 Stirling PDF

1. 打开 Olares Market，搜索 "Stirling-PDF"。

    ![从 Market 搜索 Stirling PDF](/images/manual/use-cases/install-stirling-pdf.png#bordered)

2. 点击 **Get**，然后点击 **Install**。
3. 安装完成后，从 Launchpad 打开 Stirling-PDF。Stirling PDF 主页将自动在你的默认浏览器中打开。

    ![Stirling PDF 主页](/images/manual/use-cases/stirlingpdf-landing.png#bordered)

## 编辑敏感信息

**Manual Redaction** 工具允许你隐藏或涂黑文档中的敏感信息。你可以清除特定文本、视觉区域或整页内容。

1. 在主页上，点击 **Manual Redaction**。

    ![Stirling PDF 中的 Manual Redaction](/images/manual/use-cases/manual-redaction.png#bordered)

2. 上传你的 PDF 文件。
3. 要编辑特定文本：

    a. 点击工具栏上的 **Text-based Redaction** 图标。

    ![基于文本的编辑图标](/images/manual/use-cases/text-based-redaction-icon.png#bordered){width=65%}

    b. 选择你要隐藏的文本。

    c. （可选）点击 **Colour Picker** 图标以更改编辑颜色。默认为黑色。

    ![颜色选择器图标](/images/manual/use-cases/color-picker-icon.png#bordered){width=65%}

    d. 点击 **Apply changes** 图标以确认你的选择。

    ![对勾确认图标](/images/manual/use-cases/green-check-icon.png#bordered){width=65%}

4. 要编辑图像或区域：

    a. 点击工具栏上的 **Box draw redaction** 图标。

    ![框选编辑图标](/images/manual/use-cases/box-draw-redaction.png#bordered){width=65%}

    b. 拖动鼠标在敏感区域周围绘制一个框。释放鼠标时，框从红色变为绿色，表示该区域已成功选中。

    c. （可选）点击 **Colour Picker** 图标以更改编辑颜色。默认为黑色。

    ![颜色选择器图标](/images/manual/use-cases/color-picker-icon.png#bordered){width=65%}

5. 要编辑整页：

    a. 点击工具栏上的 **Page-based Redaction** 图标。

    ![基于页面的编辑](/images/manual/use-cases/page-based-redaction.png#bordered){width=25%}

    b. 输入特定的页码或范围。

    c. 选择编辑颜色。

    d. 点击 **Apply**。

6. 导出编辑后的文件。

    a. 点击 **Export** 图标下载处理后的 PDF。

    ![导出 PDF](/images/manual/use-cases/export-pdf.png#bordered){width=25%}

    b. 检查你的文件，确保编辑已正确实施。

    下图展示了一个以橙色编辑段落的示例：

    ![文本编辑示例（橙色）](/images/manual/use-cases/text-redact-example.png#bordered){width=55%}

## 整理页面

**Multi Tool** 允许你调整 PDF 文档的结构，包括合并、拆分、旋转和重新排序页面。

1. 在主页上，点击 **PDF Multi Tool**。

    ![主页上的 Multi Tool 菜单选项](/images/manual/use-cases/pdf-multi-tool.png#bordered)

2. 点击 **Add File** 图标添加一个或多个 PDF 文件。

    ![Multi Tool 添加文件](/images/manual/use-cases/add-pdf.png#bordered){width=65%}

3. 要旋转页面：

    - 要旋转单个页面，将鼠标悬停在特定页面缩略图上，然后点击 **Rotate Left** 图标或 **Rotate Right** 图标。

    ![旋转单个页面](/images/manual/use-cases/rotate-single-page.png#bordered){width=65%}

    - 要一次性旋转所有页面，点击 **File Name** 区域中的 **Rotate Left** 图标或 **Rotate Right** 图标。

    ![批量旋转](/images/manual/use-cases/batch-rotate.png#bordered){width=65%}

4. 要重新排序页面，将页面拖动到网格中的新位置。
5. 要删除页面：

    - 要删除单个页面，将鼠标悬停在页面上，然后点击缩略图上的 **Delete** 图标。

    ![删除单个页面](/images/manual/use-cases/delete-single-page.png#bordered){width=65%}

    - 要删除多个页面，点击 **File Name** 区域中的 **Page Select** 图标，选中所有目标页面的复选框，然后点击 **Delete Selected**。

    ![批量删除](/images/manual/use-cases/page-select.png#bordered){width=65%}

6. 要拆分页面：

    - 要将特定页面拆分为新文件，点击该页面缩略图左侧的 **Scissors** 图标。

    ![拆分单个页面](/images/manual/use-cases/split-single-page.png#bordered){width=65%}

    - 要将每个页面拆分为单独的文件，点击 **File Name** 区域中的 **Scissors** 图标。

    ![拆分所有页面](/images/manual/use-cases/split-all-page.png#bordered){width=65%}

7. 要导出页面：

    - 要将重新整理后的文件下载为单个 PDF，点击 **Export**。

    ![导出所有页面](/images/manual/use-cases/export-all-page.png#bordered){width=65%}

    - 要导出特定页面，选中这些页面，然后点击 **Export Selected**。

    ![导出特定页面](/images/manual/use-cases/export-selected-page.png#bordered){width=65%}

## 转换文件格式

Stirling PDF 支持超过 50 种文件类型的转换，允许你将图像转换为 PDF，或将 PDF 转换为其他格式。

1. 在主页上找到 Convert 部分，点击你特定的转换任务。

    ![主页上的 Convert 菜单选项](/images/manual/use-cases/select-conversion.png#bordered)

2. 上传你的源文件。
3. 配置可用设置，例如颜色模式或布局偏好。设置可能因文件类型而异。
4. 点击 **Convert**。等待处理完成。转换后的文件将自动下载。

## 阅读和批注

使用此功能来审阅文档并添加注释或高亮。

1. 在主页上，点击 **View/Edit PDF**。

    ![查看和编辑 PDF](/images/manual/use-cases/view-edit-pdf.png#bordered)

2. 点击右上角的 **Open File** 图标以加载你的文档。

    ![打开 PDF 图标](/images/manual/use-cases/open-file.png#bordered){width=38%}

3. 要添加文本批注：

    a. 点击 **Text** 图标，然后选择你喜欢的字体颜色和大小。

     ![添加文本批注](/images/manual/use-cases/text-annotation.png#bordered){width=38%}

    b. 点击文档中你希望注释出现的位置，然后输入你的评论。

    c. 要删除注释，点击文本框并点击 <i class="material-symbols-outlined">delete</i>。

4. 要高亮或自由绘制：

    a. 点击 Draw 并调整画笔颜色、粗细和不透明度。

     ![绘制线条以高亮](/images/manual/use-cases/draw-line.png#bordered){width=38%}

    b. 在文档上点击并拖动以应用你的标记。

5. 要完成你的编辑，点击 **Save** 图标。包含你的批注的文档将自动下载。

     ![保存批注](/images/manual/use-cases/save-edits.png#bordered){width=38%}

## 自动化多步骤工作流

Pipeline 功能类似于 PDF 的批量处理工作流。
你无需逐一执行任务，而是可以创建自定义流水线，在单个序列中执行多个操作。本场景通过合并文档并添加密码来演示该功能。

:::info
Pipeline 功能目前处于 Beta 阶段。你可能会遇到偶尔的不稳定性或界面变化。建议在处理关键文档时验证最终输出。
:::

### 创建流水线

1. 在主页的 **Advanced** 部分，点击 **Pipeline**。

    ![主页上的 Pipeline 菜单选项](/images/manual/use-cases/pipeline.png#bordered)

2. 点击 **Configure**。

    ![Pipeline](/images/manual/use-cases/pipeline-menu.png#bordered){width=65%}

3. 在 **Pipeline Name** 框中，为你的工作流输入一个唯一的名称。例如，`Monthly invoice combination`。
4. 添加并配置合并操作：

    a. 从 **Select Operation** 列表中，选择 **merge-pdfs**，然后点击 **Add operation**。该操作将添加到 **Pipeline** 区域中，相关设置将显示在其下方。

    b. 定义 **merge-pdfs** 操作的具体参数，然后点击 **Save Operation Settings**。

    ![为流水线添加和配置操作](/images/manual/use-cases/merge-pdf.png#bordered){width=45%}

5. 添加并配置添加密码操作：

    a. 从 **Select Operation** 列表中，选择 **add-password**，然后点击 **Add operation**。

    b. 定义 **add-password** 操作的具体参数，然后点击 **Save Operation Settings**。

6. 点击 **Save to Browser** 以保存你的流水线。
7. 关闭 **Pipeline Configuration** 窗口，然后刷新 **Pipeline Menu** 页面。

### 运行流水线

1. 在 **Pipeline Menu** 页面上，选择新创建的流水线 "Monthly invoice combination"。

    ![选择要使用的流水线](/images/manual/use-cases/select-pipeline.png#bordered){width=65%}

2. 上传你的 PDF 文件，然后点击 **Submit**。

    Stirling PDF 自动合并文件并一步保护最终文档。批处理完成后，文件将自动下载。
3. 双击下载的文件，并输入你在流水线配置期间指定的密码以打开它。
4. 验证文件是否按预期合并和保护。
