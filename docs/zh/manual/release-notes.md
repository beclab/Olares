---
description: 跟踪 Olares 文档的最新变化，包括新增文档、更新内容、下架说明和页面迁移。
head:
  - - meta
    - name: keywords
      content: Olares, 文档更新, 新指南, 下架说明, 发行说明, 文档动态
---

:::warning WARNING
本页面内容由 AI 翻译生成，仅供参考。如有疑问，请以[英文原文](../../manual/release-notes.md)为准。
:::

# 文档更新动态

本页面突出显示每次 Olares 发版后对用户使用方式有影响的文档变更，包括新增指南、重大更新和下架说明。拼写修正、细节澄清等微小改动不会列出。

本文档从 [Olares GitHub 仓库](https://github.com/beclab/Olares/tree/main/docs) 的 `main` 分支构建并发布，因此始终反映最新的文档变更。

<details>
<summary>Olares OS 软件发行说明请参见 GitHub releases 页面。</summary>

- [Olares 1.12.6](https://github.com/beclab/Olares/releases/tag/1.12.6)
- [Olares 1.12.5](https://github.com/beclab/Olares/releases/tag/1.12.5)
</details>

关注 Olares 社交媒体，或加入 [Discord 社区](https://discord.com/invite/BzfqrgQPDK)，及时了解 Olares 新闻和文档更新。

## Olares 1.12.6

发布日期：2026 年 7 月 23 日

Olares v1.12.6 的亮点和详细介绍请参见 [Olares 1.12.6 发布博客](https://www.olares.com/blog/olares-1-12-6/)。

### 新增文档

- 新增 [Olares CLI](/zh/developer/cli-overview)，介绍如何通过命令行管理 Olares。
  - 新增 [安装 olares-cli](/zh/developer/cli-install)，说明如何在本地或 Agent 应用内安装 `olares-cli`。
  - 新增 [登录 Olares](/zh/developer/cli-log-in)，说明如何使用 Olares ID 认证 `olares-cli`。
  - 新增 [安装与使用 Agent Skills](/zh/developer/cli-agent-skills)，介绍集群管理、应用、设置等内置技能的使用方法。
- 新增 [使用 Olares CLI 激活 Olares 设备](/zh/manual/best-practices/activate-olares-using-cli)，介绍如何通过命令行激活设备。
- 新增 [使用公共目录管理共享 AI 模型](/zh/manual/olares/files/files-common)，说明如何使用 Common 目录在应用间共享模型。
- 新增 [压缩与解压缩文件](/zh/manual/olares/files/compress-extract-files)，涵盖 Olares Files 中 ZIP、7z、TAR 和加密压缩包的操作。
- 新增 [挂载 NFS 共享](/zh/manual/olares/files/mount-nfs)，说明如何从 Olares 访问 NFS 共享目录。
- 新增 [共享应用](/zh/manual/olares/market/shared-apps)，介绍新的共享应用架构以及旧版 v2 共享应用的迁移方式。
  - [Ollama](/zh/use-cases/ollama)
  - [ComfyUI](/zh/use-cases/comfyui-common-issues)
  - [Dify](/zh/use-cases/dify-upgrade)
  - [OnlyOffice](/zh/use-cases/onlyoffice)
  - [SearXNG](/zh/use-cases/searxng)
  - [Xinference](/zh/use-cases/xinference)
- 新增 [配置 Overlay Gateway](/zh/manual/olares/settings/overlay-gateway)，介绍如何管理 Overlay Gateway 设置。
- 新增 [Olares One 新手引导](/zh/one/olares-onboarding)，介绍如何通过自然语言管理 Olares One。
- 新增 [桌面小组件](/zh/manual/olares/desktop#widgets)，介绍 Olares 桌面上的新小组件。
- 新增 [使用 Engine Base 运行本地大模型](/zh/use-cases/llm-base-apps)，介绍如何在 Olares 上部署和运行本地 AI 模型。
- 新增 [通过 Overlay Gateway 使用 Home Assistant](/zh/use-cases/home-assistant#enable-the-overlay-gateway)，说明如何通过局域网访问 Home Assistant。
- 新增 [通过 Overlay Gateway 使用 Jellyfin](/zh/use-cases/jellyfin#enable-overlay-gateway-for-jellyfin)，说明如何通过 Overlay Gateway 访问 Jellyfin。
- 新增 [*Arr 应用升级指南](/zh/use-cases/arrs-upgrade)，在 Olares v1.12.6 内部入口变更后升级 *Arr 应用的指南。

### 更新文档

- [连接 AI 应用](/zh/manual/best-practices/connect-ai-apps) 已针对 v1.12.6+ 架构重写。
- Olares One 软件文档已整合到主手册中，ISO 下载链接也添加了版本号。
- [我的 Olares](/zh/manual/olares/settings/my-olares) 已更新，在 **我的硬件** 下新增了 **限制 CPU 频率** 和 **自动开机** 两个开关。
- [基础文件操作](/zh/manual/olares/files/add-edit-download) 已更新，新增排序方式、Markdown 编辑、预览以及更多支持格式。
- [管理加速器资源](/zh/manual/olares/settings/gpu-resource) 已更新，涵盖 GPU 和其他加速器资源。
- [管理 BIOS 和 EC](/zh/one/update-firmware) 已更新，新增 EC 1.03 和 BIOS 1.05 的变更日志，并说明 **自动开机** 功能需要 Olares OS 1.12.6 或更高版本。
- [ComfyUI](/zh/use-cases/comfyui) 已更新，增加了迁移步骤和 v1.12.6 新目录结构说明。
- [应用提交指南](/zh/developer/develop/submit-apps) 已更新，反映新的打包和提交流程。
- [OlaresManifest 规范](/zh/developer/develop/package/manifest) 已更新，说明 `0.12.0` 模式，包括 `apiVersion`、`accelerator`、`workloadReplicas`、`overlayGateway`、`LLMGatewaySupported`、`appCommon` 和 `externalData` 等新字段，以及已弃用字段。
- AI use-case 指南现在采用统一的本地模型和服务连接流程。各指南改为通过 Model Console 获取模型端点，不再依赖独立的 Ollama 应用或单个模型应用。

### 已下架文档

- [通过 Ollama 下载和运行本地 AI 模型](/zh/use-cases/ollama) 已添加下架说明，因为独立 Ollama 应用已在 Olares 1.12.6 中从应用市场移除。
- Olares CLI 参考页面已下架。旧页面已被新的 [Olares CLI](/zh/developer/cli-overview)、[安装 olares-cli](/zh/developer/cli-install)、[登录 Olares](/zh/developer/cli-log-in) 和 [安装与使用 Agent Skills](/zh/developer/cli-agent-skills) 指南取代。
- 所有 Studio 相关文档已下架。Studio 已不再上架应用市场。新的打包和移植流程请参见 [安装与使用 Agent Skills](/zh/developer/cli-agent-skills) 和 [应用提交指南](/zh/developer/develop/submit-apps)。
- **允许子网路由** 功能在 Olares 1.12.6 中暂时下架，[配置 VPN 访问 Olares](/zh/manual/olares/settings/remote-access#allow-subnet-routing) 中的相关内容已移除。该功能会在后续版本中恢复。
- DeerFlow use-case 指南已下架。DeerFlow 应用已被 [DeerFlow 2.0](/zh/use-cases/deerflow2) 取代。

### 即将下架

- [通过 Ollama 下载和运行本地 AI 模型](/zh/use-cases/ollama) use-case 指南将在后续版本中移除。独立 Ollama 应用已在 Olares 1.12.6 中从应用市场移除，AI use-case 指南现在通过 Model Console 获取模型端点。
