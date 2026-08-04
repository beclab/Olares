---
outline: [2, 3]
description: 在 Olares 上使用预配置的 JupyterLab 环境和 RAPIDS-singlecell 运行 GPU 加速的单细胞 RNA 测序分析。
head:
  - - meta
    - name: keywords
      content: Olares, RNA Sequencing, scRNA-seq, single-cell, JupyterLab, RAPIDS, GPU, bioinformatics, self-hosted
app_version: "1.0.2"
doc_version: "1.0"
doc_updated: "2026-05-09"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/rna-sequencing.md)。
:::

# 分析单细胞 RNA 测序数据

RNA Sequencing 提供了一个预配置的 JupyterLab 环境，用于 GPU 加速的单细胞 RNA 测序（scRNA-seq）分析。它包含一个可立即运行的笔记本 `scRNA_analysis_preprocessing.ipynb`，演示了基于 RAPIDS-singlecell 的预处理和分析工作流。

使用本指南打开笔记本、运行完整的分析流程并查找生成的输出文件。

## 前提条件

- 一台具有 AMD64 架构和 NVIDIA GPU 的 Olares 设备，且该应用至少可以使用 10 GB VRAM。

## 安装 RNA Sequencing

1. 打开 Market 并搜索 "RNA Sequencing"。

   ![RNA Sequencing](/images/manual/use-cases/rna-sequencing.png#bordered){width=90%}

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 运行分析笔记本

1. 从 Launchpad 打开 RNA Sequencing。应用会在 JupyterLab 界面中打开。

   ![JupyterLab interface](/images/manual/use-cases/rna-sequencing-jupyter.png#bordered){width=90%}

2. 在左侧的文件浏览器中，双击 `scRNA_analysis_preprocessing.ipynb` 打开分析笔记本。

   该笔记本已预配置，可直接运行。

   ![Open notebook](/images/manual/use-cases/rna-sequencing-notebook.png#bordered){width=95%}

   :::tip 备份原始笔记本
   如需保留原始笔记本不变，请右键点击 `scRNA_analysis_preprocessing.ipynb` 并选择 **Duplicate**。这会在同一文件夹中创建一个副本。打开并运行该副本笔记本。
   :::

3. 点击 **Run** > **Run All Cells** 执行完整的分析流程。

   ![Run All Cells](/images/manual/use-cases/rna-sequencing-run-all.png#bordered){width=95%}

4. 等待所有单元格运行完毕。

   单元格运行时，其执行指示器显示 `[*]`。单元格运行完成后，已执行的单元格会显示如 `[1]` 和 `[2]` 等数字，生成的图表或表格会出现在对应单元格下方。

5. 运行完成后，在文件浏览器中打开 `h5` 文件夹以查找生成的输出文件。

以上步骤运行了预配置的流程。如需自定义工作流，你可以编辑笔记本、创建新笔记本，或使用标准 JupyterLab 功能运行选定的单元格。有关 JupyterLab 的一般用法，请参阅 [JupyterLab 文档](https://jupyterlab.readthedocs.io/en/latest/)。

## 了解更多

- [RAPIDS-singlecell 文档](https://rapids-singlecell.readthedocs.io/en/latest/)：GPU 加速 scRNA-seq 分析的 API 参考和教程。
- [NVIDIA Single-Cell Playbook](https://build.nvidia.com/spark/single-cell)：本应用所基于的原始 playbook。
