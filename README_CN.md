<div align="center">

# Olares：开源个人云操作系统，助您重获数据主权

[![Mission](https://img.shields.io/badge/Mission-Let%20people%20own%20their%20data%20again-purple)](#)<br/>
[![Last Commit](https://img.shields.io/github/last-commit/beclab/terminus)](https://github.com/beclab/olares/commits/main)
![Build Status](https://github.com/beclab/olares/actions/workflows/release-daily.yaml/badge.svg)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/beclab/terminus)](https://github.com/beclab/olares/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/beclab/terminus?style=social)](https://github.com/beclab/olares/stargazers)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.com/invite/BzfqrgQPDK)
[![License](https://img.shields.io/badge/License-Olares-darkblue)](https://github.com/beclab/olares/blob/main/LICENSE.md)

<p>
  <a href="./README.md"><img alt="Readme in English" src="https://img.shields.io/badge/English-FFFFFF"></a>
  <a href="./README_CN.md"><img alt="Readme in Chinese" src="https://img.shields.io/badge/简体中文-FFFFFF"></a>
  <a href="./README_JP.md"><img alt="Readme in Japanese" src="https://img.shields.io/badge/日本語-FFFFFF"></a>
</p>

</div>


https://github.com/user-attachments/assets/3089a524-c135-4f96-ad2b-c66bf4ee7471

*Olares 让你体验更多可能：构建个人 AI 助理、随时随地同步数据、自托管团队协作空间、打造私人影视厅——无缝整合你的数字生活。*

<p align="center">
  <a href="https://olares.com">网站</a> ·
  <a href="https://docs.olares.com">文档</a> ·
  <a href="https://olares.com/larepass">下载 LarePass</a> ·
  <a href="https://github.com/beclab/apps">Olares 应用</a> ·
  <a href="https://space.olares.com">Olares Space</a>
</p>

> *基于公有云构建的现代互联网日益威胁着您的个人数据隐私。随着您对 ChatGPT、Midjourney 和脸书等服务的依赖加深，您对数字自主权的掌控也在减弱。您的数据存储在他人服务器上，受其条款约束，被追踪并审查。*
>
> *是时候做出改变了。*

我们坚信，**您拥有掌控自己数字生活的基本权利**。维护这一权利最有效的方式，就是将您的数据托管在本地，在您自己的硬件上。

Olares 是一款开源个人云操作系统，旨在让您能够轻松在本地拥有并管理自己的数字资产。您无需再依赖公有云服务，而可以在 Olares 上本地部署强大的开源平替服务或应用，例如可以使用 Ollama 托管大语言模型，使用 SD WebUI 用于图像生成，以及使用 Mastodon 构建不受审查的社交空间。Olares 让你坐拥云计算的强大威力，又能完全将其置于自己掌控之下。

> 为 Olares 点亮 🌟 以及时获取新版本和更新的通知。

## 使用场景

在以下场景中，Olares 为您带来私密、强大且安全的私有云体验：

🤖**本地 AI 助手**：在本地部署运行顶级开源 AI 模型，涵盖语言处理、图像生成和语音识别等领域。根据个人需求定制 AI 助手，确保数据隐私和控制权均处于自己手中。<br>

💻**个人数据仓库**：所有个人文件，包括照片、文档和重要资料，都可以在这个安全的统一平台上存储和同步，随时随地都能方便地访问。<br>

🛠️**自托管工作空间**：利用开源 SaaS 平替方案，无需成本即可为家庭或工作团队搭建一个功能强大的工作空间。<br>

🎥**私人媒体服务器**：用自己的视频和音乐库搭建一个私人流媒体服务，随时享受个性化的娱乐体验。<br>

🏡**智能家居中心**：将所有智能设备和自动化系统集中在一个易于管理的控制中心，实现家庭智能化的简便操作。<br>

🤝**独立的社交媒体平台**：在 Olares 上部署去中心化社交媒体应用，如 Mastodon、Ghost 和 WordPress，自由建立和扩展个人品牌，无需担忧封号或支付额外费用。<br>

📚**学习探索**：深入学习自托管服务、容器技术和云计算，并上手实践。<br>

## 快速开始

### 系统兼容性

Olares 已在以下 Linux 平台完成测试与验证：

- Ubuntu 24.04 LTS 及以上版本
- Debian 11 及以上版本

### 安装 Olares
 
参考[快速上手指南](https://docs.joinolares.cn/zh/manual/get-started/)安装并激活 Olares。

## 系统架构
Olares 的架构设计遵循两个核心原则：
- 参考 Android 模式，控制软件权限和交互性，确保系统的流畅性和安全性。
- 借鉴云原生技术，高效管理硬件和中间件服务。

  ![架构](https://file.bttcdn.com/github/terminus/v2/olares-arch-3.png)

详细描述请参考 [Olares 架构](https://docs.joinolares.cn/zh/manual/system-architecture.html)文档。

## 功能特性

Olares 提供了一系列功能，旨在提升安全性、使用便捷性以及开发的灵活性：

- **企业级安全**：使用 Tailscale、Headscale、Cloudflare Tunnel 和 FRP 简化网络配置，确保安全连接。
- **安全且无需许可的应用生态系统**：应用通过沙箱化技术实现隔离，保障应用运行的安全性。
- **统一文件系统和数据库**：提供自动扩展、数据备份和高可用性功能，确保数据的持久安全。
- **单点登录**：用户仅需一次登录，即可访问 Olares 中所有应用的共享认证服务。
- **AI 功能**：包括全面的 GPU 管理、本地 AI 模型托管及私有知识库，同时严格保护数据隐私。
- **内置应用程序**：涵盖文件管理器、同步驱动器、密钥管理器、阅读器、应用市场、设置和面板等，提供全面的应用支持。
- **无缝访问**：通过移动端、桌面端和网页浏览器客户端，从全球任何地方访问设备。
- **开发工具**：提供全面的工具支持，便于开发和移植应用，加速开发进程。

## 项目目录

Olares 代码库中的主要目录如下：

* **`apps`**: 用于存放系统应用，主要是 `larepass` 的代码。
* **`cli`**: 用于存放 `olares-cli`（Olares 的命令行界面工具）的代码。
* **`daemon`**: 用于存放 `olaresd`（系统守护进程）的代码。
* **`docs`**: 用于存放 Olares 项目的文档。
* **`framework`**: 用来存放 Olares 系统服务代码。
* **`infrastructure`**: 用于存放计算，存储，网络，GPU 等基础设施的代码。
* **`platform`**: 用于存放数据库、消息队列等云原生组件的代码。
* **`vendor`**: 用于存放来自第三方硬件供应商的代码。

## 社区贡献

我们欢迎任何形式的贡献！

- 如果您想在 Olares 上开发自己的应用，请参考：<br>
https://docs.olares.com/developer/develop/


- 如果您想帮助改进 Olares，请参考：<br>
https://docs.olares.com/developer/contribute/olares.html

## 社区支持

* [**GitHub Discussion**](https://github.com/beclab/olares/discussions) - 讨论 Olares 使用过程中的疑问。
* [**GitHub Issues**](https://github.com/beclab/olares/issues) - 报告 Olares 的遇到的问题或提出功能改进建议。
* [**Discord**](https://discord.com/invite/BzfqrgQPDK) - 日常交流，分享经验，或讨论与 Olares 相关的任何主题。

## 特别感谢

Olares 项目整合了许多第三方开源项目，包括：[Kubernetes](https://kubernetes.io/)、[Kubesphere](https://github.com/kubesphere/kubesphere)、[Padloc](https://padloc.app/)、[K3S](https://k3s.io/)、[JuiceFS](https://github.com/juicedata/juicefs)、[MinIO](https://github.com/minio/minio)、[Envoy](https://github.com/envoyproxy/envoy)、[Authelia](https://github.com/authelia/authelia)、[Infisical](https://github.com/Infisical/infisical)、[Dify](https://github.com/langgenius/dify)、[Seafile](https://github.com/haiwen/seafile)、[HeadScale](https://headscale.net/)、 [tailscale](https://tailscale.com/)、[Redis Operator](https://github.com/spotahome/redis-operator)、[Nitro](https://nitro.jan.ai/)、[RssHub](http://rsshub.app/)、[predixy](https://github.com/joyieldInc/predixy)、[nvshare](https://github.com/grgalex/nvshare)、[LangChain](https://www.langchain.com/)、[Quasar](https://quasar.dev/)、[TrustWallet](https://trustwallet.com/)、[Restic](https://restic.net/)、[ZincSearch](https://zincsearch-docs.zinc.dev/)、[filebrowser](https://filebrowser.org/)、[lego](https://go-acme.github.io/lego/)、[Velero](https://velero.io/)、[s3rver](https://github.com/jamhall/s3rver)、[Citusdata](https://www.citusdata.com/)。
