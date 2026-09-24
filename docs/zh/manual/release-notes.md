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

本文档从 [Olares GitHub 仓库](https://github.com/beclab/Olares/tree/main/docs)的 `main` 分支构建并发布，因此始终反映最新的文档变更。

<details>
<summary>Olares OS 软件发行说明请参见 GitHub releases 页面。</summary>

- [Olares 1.12.6](https://github.com/beclab/Olares/releases/tag/1.12.6)
- [Olares 1.12.5](https://github.com/beclab/Olares/releases/tag/1.12.5)
</details>

关注 Olares 社交媒体，或加入 [Discord 社区](https://discord.com/invite/BzfqrgQPDK)，及时了解 Olares 新闻和文档更新。

## Olares 1.12.7

<!-- TODO: 正式发布资料上线后，填写确认的发布日期并启用以下链接。
发布日期：[确认后的发布日期]

软件变更请参阅 [Olares 1.12.7 发行说明](https://github.com/beclab/Olares/releases/tag/1.12.7)。
版本亮点与功能介绍请参阅 [Olares 1.12.7 发布博客](https://www.olares.cn/blog/olares-1-12-7/)。
-->

### 新增文档

- 新增[升级至 Olares 1.12.7 后的操作指南](/zh/manual/update-guides/1.12.7)，说明升级路径、Router 与 Lares 的配置，以及 Agent 应用更新。
- 新增[使用 Olares Router 作为 AI 网关](/zh/use-cases/olares-router)，介绍模型与工具能力、调用方鉴权、模型命名和连接信息。
- 新增[使用 Lares 管理 Olares 并开展研究](/zh/use-cases/lares)，介绍初次使用、权限设置，以及通过 Router 配置网页研究工具。
- 新增[安装 LarePass 浏览器扩展](/zh/manual/install-larepass-browser-extension)和[使用 LarePass 私密翻译网页](/zh/manual/tutorial/translate-webpages-with-larepass)，介绍如何通过 Router 调用本地模型翻译网页。
- 新增[将资源保存到 Olares](/zh/manual/larepass/save-resources-to-olares)，介绍 LarePass 桌面端的 Fetch 功能，包括网页链接、种子、Hugging Face 仓库和支持的媒体资源。
- 新增[在 Olares Space 中创建和管理工单](/zh/manual/space/tickets)，并在[获取支持](/zh/manual/help/request-technical-support)中补充 Ticket 应用、日志收集和工单跟进流程。
- 新增[恢复 LarePass、Olares 桌面或设备访问权限](/zh/manual/help/ts-access-without-mnemonic)，并更新[登录密码重置说明](/zh/manual/help/ts-forget-login-password)。
- 新增[使用 Lares 创建 Blender 场景](/zh/use-cases/blender)和[使用 Concat 和 Lares 制作视频](/zh/use-cases/concat)，涵盖 MCP 配置、创作流程，以及在文件管理器中保存和查看输出。
- 新增 [Open Design](/zh/use-cases/open-design)、[Codex CLI](/zh/use-cases/codex-cli) 和[通过 OpenCode 管理 Olares](/zh/use-cases/opencode-olares-cli) 教程，覆盖设计、编程和设备管理。
- 新增 [Minecraft](/zh/use-cases/minecraft) 和 [Palworld](/zh/use-cases/palworld) 服务器教程，介绍如何通过 Overlay Gateway 在局域网中连接。

### 更新文档

- 重写[为应用接入 AI 能力](/zh/manual/best-practices/connect-ai-apps)，介绍 Router 的 API 格式、模型与工具配置、`default-chat` 等默认系统名称，以及外部客户端的 API 密钥。
- 更新 AI 客户端教程，以 Qwen3.8-27B (llama.cpp) 为聊天示例，改用 Router 连接流程。在相关页面加入版本切换，可选择 Olares 1.12.7 或 1.12.6 的操作说明。[OpenClaw](/zh/use-cases/openclaw) 和 [Hermes Agent](/zh/use-cases/hermes) 的步骤已按交互式向导顺序调整，模型配置和上下文大小改为在 Router 中查看。
- 按用户任务重新组织 [Olares 手册](/zh/manual/overview)，按 Linux、Olares One 和 DGX Spark 区分[安装路径](/zh/manual/get-started/install-olares)。[Olares One 新手引导](/zh/one/olares-onboarding)已更新为使用 Lares。
- 更新[本地访问指南](/zh/manual/best-practices/local-access)，补充 Windows 和 macOS 上的 LarePass hosts 映射，并说明 VPN 与 `.local` 地址的使用方式。
- 更新[应用市场](/zh/manual/olares/market/market)的应用与模型发现、来源选择流程，并在[语言设置](/zh/manual/olares/settings/language-appearance)中补充新增语言选项。
- 重写 Olares Space 的[账号管理](/zh/manual/space/manage-accounts)、[资源与流量监控](/zh/manual/space/manage-olares)和[账单](/zh/manual/space/billing)说明。
- 重写 [Olares One eGPU 指南](/zh/one/egpu)，分别提供 Olares OS 和 Windows 配置步骤，并补充兼容性与故障排查说明。[Olares One 常见问题](/zh/one/faq#olares-one-支持带外管理吗)新增带外管理能力的支持情况说明。
- 更新 [Jellyfin](/zh/use-cases/jellyfin) 硬件加速、[Steam 串流](/zh/use-cases/steam-stream)网络配置，以及 [Penpot](/zh/use-cases/penpot) 内置 MCP 服务器的连接方法。

### 页面迁移与合并

- 将备份与恢复说明合并到[备份与恢复 Olares](/zh/manual/olares/settings/backup)。原恢复页面及 Olares Space 备份页面会跳转至此。
- 将原“我的 Olares”主题拆分为[密码与设备管理](/zh/manual/password-and-devices)和 [Olares One 硬件设置](/zh/one/hardware-settings)。
- 将 Olares ID 创建流程统一到[创建 Olares ID](/zh/manual/get-started/create-olares-id)，并将集成配置拆入[挂载云存储](/zh/manual/olares/files/mount-cloud-storage)等任务指南。旧链接会跳转到替代页面。

## Olares 1.12.6

发布日期：2026 年 7 月 23 日

Olares v1.12.6 的亮点和详细介绍请参见 [Olares 1.12.6 发布博客](https://www.olares.com/blog/olares-1-12-6/)。

### 新增文档

- 新增 [Olares CLI](/zh/developer/cli-overview)，介绍如何通过命令行管理 Olares。
  - 新增[安装 olares-cli](/zh/developer/cli-install)，说明如何在本地或 Agent 应用内安装 `olares-cli`。
  - 新增[登录 Olares](/zh/developer/cli-log-in)，说明如何使用 Olares ID 认证 `olares-cli`。
  - 新增[安装与使用 Agent Skills](/zh/developer/cli-agent-skills)，介绍集群管理、应用、设置等内置技能的使用方法。
- 新增[通过命令行安装并激活 Olares](/zh/manual/best-practices/activate-olares-using-cli)，介绍如何使用 Olares CLI 设置设备。
- 新增[使用公共目录管理共享 AI 模型](/zh/manual/olares/files/files-common)，说明如何使用 Common 目录在应用间共享模型。
- 新增[压缩与解压缩文件](/zh/manual/olares/files/compress-extract-files)，涵盖 Olares Files 中 ZIP、7z、TAR 和加密压缩包的操作。
- 新增[挂载 NFS 共享](/zh/manual/olares/files/mount-nfs)，说明如何从 Olares 访问 NFS 共享目录。
- 新增[关于共享应用](/zh/manual/olares/market/shared-apps)，介绍新的共享应用架构；新增[迁移旧版共享应用](/zh/manual/migrate-shared-apps)，说明 v2 应用的迁移方式。
  - [Ollama](/zh/use-cases/ollama)
  - [ComfyUI](/zh/use-cases/comfyui-common-issues)
  - [Dify](/zh/use-cases/dify-upgrade)
  - [OnlyOffice](/zh/use-cases/onlyoffice-migration)
  - [SearXNG](/zh/use-cases/searxng)
  - [Xinference](/zh/use-cases/xinference)
- 新增[配置 Overlay Gateway](/zh/manual/olares/settings/overlay-gateway)，介绍如何管理 Overlay Gateway 设置。
- 新增 [Olares One 新手引导](/zh/one/olares-onboarding)，介绍如何通过自然语言管理 Olares One。
- 新增[桌面小组件](/zh/manual/olares/desktop#widgets)，介绍 Olares 桌面上的新小组件。
- 新增[使用 Engine Base 运行本地大模型](/zh/use-cases/llm-base-apps)，介绍如何在 Olares 上部署和运行本地 AI 模型。
- 新增[通过 Overlay Gateway 使用 Home Assistant](/zh/use-cases/home-assistant#enable-the-overlay-gateway)，说明如何通过局域网访问 Home Assistant。
- 新增[通过 Overlay Gateway 使用 Jellyfin](/zh/use-cases/jellyfin#enable-overlay-gateway-for-jellyfin)，说明如何通过 Overlay Gateway 访问 Jellyfin。
- 新增 [*Arr 应用更新说明](/zh/use-cases/arrs-upgrade)，介绍 Olares v1.12.6 内部入口变化后需要调整的配置。

### 更新文档

- [连接 AI 应用](/zh/manual/best-practices/connect-ai-apps)已针对 v1.12.6+ 架构重写。
- Olares One 软件文档已整合到主手册中，ISO 下载链接也添加了版本号。
- [我的 Olares](/zh/one/hardware-settings) 已更新，在 **我的硬件** 下新增了 **限制 CPU 频率** 和 **自动开机** 两个开关。
- [基础文件操作](/zh/manual/olares/files/add-edit-download)已更新，新增排序方式、Markdown 编辑、预览以及更多支持格式。
- [管理加速器资源](/zh/manual/olares/settings/gpu-resource)已更新，涵盖 GPU 和其他加速器资源。
- [管理 BIOS 和 EC](/zh/one/update-firmware) 已更新，新增 EC 1.03 和 BIOS 1.05 的变更日志，并说明 **自动开机** 功能需要 Olares OS 1.12.6 或更高版本。
- [ComfyUI](/zh/use-cases/comfyui) 已更新，增加了迁移步骤和 v1.12.6 新目录结构说明。
- [应用提交指南](/zh/developer/develop/submit-apps)已更新，反映新的打包和提交流程。
- [OlaresManifest 规范](/zh/developer/develop/package/manifest)已更新，说明 `0.12.0` 模式，包括 `apiVersion`、`accelerator`、`workloadReplicas`、`overlayGateway`、`LLMGatewaySupported`、`appCommon` 和 `externalData` 等新字段，以及已弃用字段。
- AI 应用指南现在采用统一的本地模型和服务连接流程。各指南改为通过 Model Console 获取模型端点，不再依赖独立的 Ollama 应用或单个模型应用。

### 已下架文档

- [通过 Ollama 下载和运行本地 AI 模型](/zh/use-cases/ollama)已添加下架说明，因为独立 Ollama 应用已在 Olares 1.12.6 中从应用市场移除。
- Olares CLI 参考页面已下架。旧页面已被新的 [Olares CLI](/zh/developer/cli-overview)、[安装 olares-cli](/zh/developer/cli-install)、[登录 Olares](/zh/developer/cli-log-in) 和[安装与使用 Agent Skills](/zh/developer/cli-agent-skills) 指南取代。
- 所有 Studio 相关文档已下架。Studio 已不再上架应用市场。新的打包和移植流程请参见[安装与使用 Agent Skills](/zh/developer/cli-agent-skills) 和[应用提交指南](/zh/developer/develop/submit-apps)。
- **允许子网路由** 功能在 Olares 1.12.6 中暂时下架，[配置 VPN 访问 Olares](/zh/manual/olares/settings/remote-access#allow-subnet-routing) 中的相关内容已移除。该功能会在后续版本中恢复。
- DeerFlow 指南已下架。DeerFlow 应用已被 [DeerFlow 2.0](/zh/use-cases/deerflow2) 取代。

### 即将下架

- [通过 Ollama 下载和运行本地 AI 模型](/zh/use-cases/ollama)指南将在后续版本中移除。独立 Ollama 应用已在 Olares 1.12.6 中从应用市场移除，AI 应用指南现在通过 Model Console 获取模型端点。
