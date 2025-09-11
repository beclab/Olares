---
description: Olares 是为本地 AI 打造的开源私有云操作系统，支持自托管服务、AI 应用部署、文件管理和安全协作，让你完全掌控数据。
---
# Olares 文档

## Olares 是什么？
Olares 是一款开源个人云操作系统，旨在让你能够轻松在本地拥有并管理自己的数字资产。你无需再依赖公有云服务，而可以在 Olares 上本地部署强大的开源平替服务或应用，例如可以使用 Ollama 托管大语言模型，使用 SD WebUI 用于图像生成，以及使用 Mastodon 构建不受审查的社交空间。Olares 让你坐拥云计算的强大威力，又能完全将其置于自己掌控之下。

:::info 开源与商业模式
Olares 的模式类似于 **Android**：
- **Olares OS**（软件层）完全**开源**，保证透明性、社区共建与可扩展性，详见 [GitHub](https://github.com/beclab/Olares) 页面。
- **硬件**（运行 Olares 的设备）可授权制造商生产与销售，通过硬件产品和生态合作实现可持续发展。 
:::

<div class="cta">
  <a href="./get-started/">
    <div class="content">
      <h3>初次使用 Olares？</h3>
      <p>查看入门指南，了解如何在本地设备上设置 Olares。</p>
    </div>
    <div class="arrow">→</div>
  </a>
</div>

## Olares 组成部分

Olares 由以下核心组件组成：

- [Olares ID](./concepts/olares-id.md)：一种去中心化的身份与信用系统，支持信息和价值的安全、无信任交换。Olares ID 是你在整个 Olares 生态中的数字身份。

- [**Olares OS**](https://github.com/beclab/Olares)：一个开源自主托管操作系统，可将边缘设备转化为强大的个人云。

- [**LarePass**](./larepass/)：一款安全的跨平台客户端，连接您与 Olares 系统。它提供无缝访问、统一身份管理、快速文件同步，以及强大的设备管理能力。

## 功能亮点

Olares 提供了丰富的功能，旨在提升安全性、易用性和开发灵活性：

- **企业级安全**：通过 Tailscale、Headscale、Cloudflare Tunnel 和 FRP 等工具，简化网络配置，确保私有云的安全。
- **安全且开放的应用生态**：在安全的沙箱环境中使用近百款免费应用。[查看 Olares 应用市场](https://market.olares.com/)。
- **统一文件系统和数据库**：支持自动扩容、备份和高可用。
- **统一身份认证**：一次登录，访问所有 Olares 应用。
- **AI 能力**：管理 GPU 资源，本地部署 AI 模型，构建私有知识库。
- **内置应用**：预装多款实用应用，包括文件管理器、Vault、Wise、Profile 和仪表盘，开箱即用。
- **随时随地访问**：通过移动端、桌面端和浏览器端专用客户端，随时随地访问设备资源。
- **开发工具**：配套完整开发工具，轻松构建和迁移应用。

## 使用场景

以下是 Olares 的典型应用场景。

- **本地 AI**：在你的设备上直接托管和运行最新的开源 AI 模型，包括大型语言模型、图像生成和语音识别。构建可以与你的数据和应用程序集成的自定义 AI 助手，同时保持所有内容的私密性和安全性。

- **个人数据仓库**：所有个人文件，包括照片、文档和重要资料，都可以在这个安全的统一平台上存储和同步，随时随地都能方便地访问。

- **自托管工作空间**：利用开源 SaaS 平替方案，使用开源替代方案即可为家庭或工作团队搭建一个功能强大的协同空间。

- **私人媒体服务器**：用自己的视频和音乐库搭建私人流媒体服务，随时享受个性化的娱乐体验。

- **智能家居中心**：将所有智能设备和自动化系统集中在一个易于管理的控制中心，实现家庭智能化的简便操作。

- **去中心化社交媒体**：在 Olares 上部署去中心化社交媒体应用，如 Mastodon、Ghost 和 WordPress，自由建立和扩展个人品牌，无需担忧封号或支付额外费用。

- **学习探索**：深入学习自托管服务、容器技术和云计算，并上手实践。

## 选择适合的路径

在深入了解 Olares 之前，不妨先快速浏览一番。以下是几条路径，帮助快速了解 Olares 的功能。

<div class="cta-container">
  <a href="../use-cases/" class="cta-link">
    <p class="cta-title">探索使用场景</p>
    <p class="cta-description">发现实际应用场景，看看 Olares 如何解决常见挑战。</p>
  </a>
  <a href="olares/" class="cta-link">
    <p class="cta-title">阅读操作指南</p>
    <p class="cta-description">全面了解 Olares 的各个应用。</p>
  </a>
  <a href="./concepts/" class="cta-link">
    <p class="cta-title">了解 Olares</p>
    <p class="cta-description">掌握 Olares 的基本原理和架构。</p>
  </a>
</div>

## 其他资源

- [开发 Olares 应用](../developer/develop/)
- [加入 Discord 社区](https://discord.gg/olares)
- [查看 Olares 博客](https://blog.olares.com/)