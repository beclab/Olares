---
outline: [2, 3]
description: 了解如何在升级到 Olares 1.12.6 后，将 SearXNG 设置从旧架构迁移到新的共享应用架构。
head:
  - - meta
    - name: keywords
      content: Olares, SearXNG, 迁移, 共享应用, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-07"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/searxng.md)。
:::

# 将 SearXNG 迁移到新架构

SearXNG 是 Olares 上注重隐私的网页搜索共享应用。Olares 1.12.6 更新了共享应用架构，因此你无法直接更新 SearXNG。本指南介绍如何在升级到 Olares 1.12.6 后，将旧版用户偏好设置迁移到新版 SearXNG。

## 开始之前

SearXNG 不在服务器端存储用户数据。所有偏好设置（如界面语言、主题、启用的搜索引擎和插件）都保存在浏览器的 Cookie 中。因此，你只需要在卸载之前安装的 SearXNG 应用前，备份 preferences hash。

:::tip Preferences hash
Preferences hash 是一串包含你所有 SearXNG 设置的编码字符串。复制该 hash 并粘贴到新版应用中，即可恢复你的偏好设置。
:::

## 迁移你的用户偏好设置

1. 备份 preferences hash。

   a. 打开你之前安装的 SearXNG。
   
   b. 进入 **Preferences** > **Cookies**。
   
   c. 滚动到 **Copy preferences hash** 部分，复制 hash code 并保存。
   
   ![复制 SearXNG preferences hash](/images/manual/use-cases/searxng-copy-preferences-hash.png#bordered)

2. 卸载之前安装的 SearXNG。在提示时，不要勾选**同时删除所有本地数据**。
3. 安装新版 SearXNG。

   a. 打开 Market，搜索 "SearXNG"。
   
   b. 点击应用卡片，打开应用详情页。
   
   c. 查看 **Information** 面板。**Compatibility** 字段显示 `Olares >=1.12.6-0` 即为新版本。
   
   d. 点击 **Get**，然后点击 **Install**，等待安装完成。

4. 恢复偏好设置。

   a. 打开新版 SearXNG。
   
   b. 进入 **Preferences** > **Cookies**。
   
   c. 滚动到 **Insert copied preferences hash (without URL) to restore** 部分。
   
   d. 粘贴之前保存的 hash code，然后点击 **Save**。

你的 SearXNG 用户偏好设置已迁移到新版应用。

## 了解更多

- [共享应用](../manual/olares/market/shared-apps.md)：了解新的共享应用架构。
