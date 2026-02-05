---
outline: [2, 3]
description: 使用图片和图标优化你在 Olares 应用市场中的应用展示。
---
# 推广你的应用

高质量的视觉素材能帮助你的应用在 Olares 应用市场中脱颖而出。本文介绍了素材规格以及如何为应用生成并配置图片链接。

## 应用素材

在 [`OlaresManifest.yaml`](/zh/developer/develop/package/manifest.md) 中配置这些素材。

### 应用图标

**必需**。显示在启动台和市场列表中。

- **位置**：在 `OlaresManifest.yaml` 的 `metadata` 或 `entrances` 下的 `icon` 字段中配置图标 URL。
- **格式**：PNG 或 WEBP
- **尺寸**：256 × 256 像素
- **大小**：不超过 512 KB

### 宣传图

**推荐**。显示在应用详情页。建议至少上传 2 张图片。

- **位置**：在 `OlaresManifest.yaml` 的 `spec` 下的 `promoteImage` 字段中配置图片 URL。
- **格式**：JPEG、PNG 或 WEBP
- **尺寸**：1440 × 900 像素
- **大小**：每张不超过 8 MB
- **数量**：最多 8 张

### 头图

**可选**。用于 Olares 应用市场的推荐位或显示在“我的 Olares”部分。

- **位置**：在 `OlaresManifest.yaml` 的 `spec` 下的 `featuredImage` 字段中配置图片 URL。
- **格式**：JPEG、PNG 或 WEBP
- **尺寸**：1440 × 900 像素
- **大小**：不超过 8 MB
- **数量**：最多 1 张

### 图片托管服务

你可以将图片托管在自己的服务器上，或使用 Olares 图片托管服务：

1. 打开 [Olares Market 图片托管](https://imghost.olares.cn/)。
2. 选择图片类型：**应用图标**、**头图**或**宣传图**。
3. 上传并预览图片。
4. 复制生成的 URL，并粘贴到 `OlaresManifest.yaml` 对应字段中。