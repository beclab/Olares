---
outline: [2, 3]
description: 了解如何将 Xinference 从 v2 架构迁移到 Olares 1.12.6 的新共享应用架构。
head:
  - - meta
    - name: keywords
      content: Olares, Xinference, 迁移, 共享应用, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-13"
---

# 将 Xinference 迁移到新架构

Xinference 是 Olares 上用于部署和提供模型服务的共享应用。Olares 1.12.6 更新了共享应用架构，因此无法直接原地升级 Xinference。本文介绍如何在升级到 Olares 1.12.6 后重新安装 Xinference 并重新下载模型。

## 迁移前须知

在 v2 应用中，Xinference 将所有模型以 local files 的形式存放在自己的 Cache 目录中。

Olares 1.12.6 引入了[公共目录](/zh/manual/olares/files/files-common.md)，用于跨应用管理共享的 AI 模型。在新架构中，Xinference 会根据模型的来源将其存放到不同位置：

- **从 Hugging Face 下载的模型**存放在 **应用** > **公共** > **huggingface** 中，遵循 Hugging Face 官方缓存结构。
- **从其他渠道下载的模型**存放在 **应用** > **数据** > **xinferencesv3** 中。

因此，已有模型无法自动迁移。你必须重新安装应用并重新下载模型。

## 重新安装 Xinference 并重新下载模型

1. 卸载之前安装的 Xinference 应用。提示时，请勿勾选**同时删除所有本地数据**。
2. 安装新版 Xinference 应用。

   a. 打开应用市场，搜索 "Xinference"。

   b. 点击应用卡片进入详情页。

   c. 查看**信息**面板中的**兼容性**字段，新版应用显示为 `Olares >=1.12.6-0`。

   d. 点击**获取**，然后点击**安装**，等待安装完成。

3. 打开新版 Xinference 应用，然后重新下载你需要的所有模型。系统会自动将它们存放到正确的位置。

## 了解更多

- [共享应用](../manual/olares/market/shared-apps.md)：了解新的共享应用架构。
