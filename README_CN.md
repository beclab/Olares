<div align="center">

# Terminus——基于 Kubernetes 的自托管家庭云 <!-- omit in toc -->

[![Mission](https://img.shields.io/badge/Mission-Let%20people%20own%20their%20data%20again-purple)](#)<br />
[![Last Commit](https://img.shields.io/github/last-commit/beclab/terminus)](https://github.com/beclab/terminus/commits/main)
![Build Status](https://github.com/beclab/terminus/actions/workflows/release-daily.yaml/badge.svg)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/beclab/terminus)](https://github.com/beclab/terminus/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/beclab/terminus?style=social)](https://github.com/beclab/terminus/stargazers)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.com/invite/BzfqrgQPDK)
[![License](https://img.shields.io/badge/License-Terminus-darkblue)](https://github.com/beclab/terminus/blob/main/LICENSE.md)

<p>
  <a href="./README.md"><img alt="Readme in English" src="https://img.shields.io/badge/English-FFFFFF"></a>
  <a href="./README_CN.md"><img alt="Readme in Chinese" src="https://img.shields.io/badge/简体中文-FFFFFF"></a>
</p>

</div>


![cover](https://file.bttcdn.com/github/terminus/desktop-dark.jpeg)

*Terminus 家庭云让你体验更多可能：构建个人 AI 助理、随时随地同步数据、自托管团队协作空间、打造私人影视厅——无缝整合你的数字生活。*

<p align="center">
  <a href="https://www.jointerminus.com">网站</a> ·
  <a href="https://docs.jointerminus.com">文档</a> ·
  <a href="https://docs.jointerminus.com/how-to/termipass/overview.html#download">下载 TermiPass</a> ·
  <a href="https://github.com/beclab/apps">Terminus 应用</a> ·
  <a href="https://space.jointerminus.com">Terminus Space</a>
</p>

## 目录 <!-- omit in toc -->

- [介绍](#介绍)
- [动机与设计](#动机与设计)
- [技术栈](#技术栈)
- [功能](#功能)
- [功能对比](#功能对比)
- [快速开始](#快速开始)
- [项目目录](#项目目录)
- [社区贡献](#社区贡献)
- [社区支持](#社区支持)
- [持续关注](#持续关注)
- [特别感谢](#特别感谢)
  
## 介绍

Terminus 是一个基于 Kubernetes 的免费自托管操作系统，可将您的边缘设备转变为强大的家庭云。在保障个人隐私的同时，Terminus 让你随时随地安全访问和管理数据，完全掌控数字生活。

Terminus 支持以下应用场景：

🤖**本地 AI 助手**：在本地部署运行开源世界级 AI 模型，涵盖语言处理、图像生成和语音识别等领域。根据个人需求定制 AI 助手，确保数据隐私和控制权均处于自己手中。<br>

💻**个人数据仓库**：所有个人文件，包括照片、文档和重要资料，都可以在这个安全的统一平台上存储和同步，随时随地都能方便地访问。<br>

🛠️**自托管工作空间**：利用开源解决方案，无需成本即可为家庭或工作团队搭建一个功能强大的工作空间。<br>

🎥**私人媒体服务器**：用自己的视频和音乐库搭建一个私人流媒体服务，随时享受个性化的娱乐体验。<br>

🏡**智能家居中心**：将所有智能设备和自动化系统集中在一个易于管理的控制中心，实现家庭智能化的简便操作。<br>

🤝**独立的社交媒体平台**：在 Terminus 上部署去中心化社交媒体应用，如 Mastodon、Ghost 和 WordPress，自由建立和扩展个人品牌，无需担忧封号或支付额外费用。<br>

📚**学习探索**：深入学习自托管服务、容器技术和云计算，并上手实践。<br>

## 动机与设计

我们深知当前互联网的局限性——用户的数据被主流互联网或云服务公司掌控，并用于其商业利益。我们致力于改变这一现状，希望通过 Terminus 赋予用户真正的数据所有权和控制权。

Terminus 为此提供了一套全新的去中心化互联网框架，主要包括以下三个部分：

- **Snowinning Protocol**：一个去中心化的身份和声誉系统，融合了去中心化标识符（DIDs）、可验证凭证（VCs）以及声誉数据，帮助用户在网络世界中安全地管理自己的身份。
- **Terminus**：一个专为边缘设备设计的自托管操作系统，用户可以在此系统上自主托管自己的数据和应用，确保数据的私密性和安全性。
- **TermiPass**：一款功能全面的客户端软件，通过安全的方式将用户与其 Terminus 系统连接起来。它不仅支持远程访问、身份和设备管理，还提供数据存储和各种办公工具，让用户高效管理其日常工作和个人数据。详情请参阅[文档](https://docs.jointerminus.com/how-to/termipass/overview.html)。

## 技术栈

  ![技术栈](https://file.bttcdn.com/github/terminus/v2/tech-stack.jpeg)

## 功能

Terminus 提供了一系列功能，旨在提升安全性、使用便捷性以及开发的灵活性：

- **企业级安全**：使用 Tailscale、Headscale、Cloudflare Tunnel 和 FRP 简化网络配置，确保安全连接。
- **安全且无需许可的应用生态系统**：应用通过沙箱化技术实现隔离，保障应用运行的安全性。
- **统一文件系统和数据库**：提供自动扩展、数据备份和高可用性功能，确保数据的持久安全。
- **单点登录**：用户仅需一次登录，即可访问 Terminus 中所有应用的共享认证服务。
- **AI 功能**：包括全面的 GPU 管理、本地 AI 模型托管及私有知识库，同时严格保护数据隐私。
- **内置应用程序**：涵盖文件管理器、同步驱动器、密钥管理器、阅读器、应用市场、设置和面板等，提供全面的应用支持。
- **无缝访问**：通过移动端、桌面端和网页浏览器客户端，从全球任何地方访问设备。
- **开发工具**：提供全面的工具支持，便于开发和移植应用，加速开发进程。


| **桌面：AI驱动的个人桌面**     |  **文件：安全存储数据**
| :--------: | :-------: |
| ![桌面](https://file.bttcdn.com/github/terminus/v2/desktop.jpg) | ![文件](https://file.bttcdn.com/github/terminus/v2/files.jpg) |
| **Vault：密码无忧管理**|**市场：可控的应用生态系统** |
| ![vault](https://file.bttcdn.com/github/terminus/v2/vault.jpg) | ![市场](https://file.bttcdn.com/github/terminus/v2/market.jpg) |
|**Wise：数字后花园** | **设置：高效管理 Terminus** |
| ![设置](https://file.bttcdn.com/github/terminus/v2/wise.jpg) | ![](https://file.bttcdn.com/github/terminus/v2/settings.jpg) |
|**面板：持续监控 Terminus**  | **Profile：去中心化网络的个人主页** |
| ![面板](https://file.bttcdn.com/github/terminus/v2/dashboard.jpg) | ![profile](https://file.bttcdn.com/github/terminus/v2/profile.jpg) |
| **Devbox：一站式开发、调试和部署**|**ControlHub：轻松管理 Kubernetes 集群**  |
| ![Devbox](https://file.bttcdn.com/github/terminus/v2/devbox.jpg) | ![控制中心](https://file.bttcdn.com/github/terminus/v2/controlhub.jpg)|

</div>

## 功能对比

为了帮您快速了解 Terminus 在市场中的独特优势，我们制作了一张功能比较表，详细展示了 Terminus 的功能以及与市场上其他主流解决方案的对比。

**图例：** 

- 🚀: **自动** - 表示系统自动完成任务。
- ✅: **支持** - 表示无开发背景的用户可以通过产品的 UI 提示完成设置。
- 🛠️: **手动配置** - 表示即使是有工程背景的用户也需要参考教程来完成设置。
- ❌: **不支持** - 表示不支持该功能。



|     | Terminus | 群晖 | TrueNAS | CasaOS | Proxmox | Unraid |
| --- | --- | --- | --- | --- | --- | --- |
| 源代码许可证 | Terminus 许可证 | 闭源 | GPL 3.0 | Apache 2.0 | MIT | 闭源 |
| 开发 | Kubernetes | Linux | Kubernetes | Docker | LXC/VM | Docker |
| 多节点支持 | ✅   | ❌   | ✅   | ❌   | 🛠️ | ❌   | ❌   |
| 内置应用 | ✅（桌面应用丰富）| ✅（桌面应用丰富） | ❌ (CLI) | ✅ （桌面应用较少） | ✅（面板）| ✅（面板） |
| 免费域名 | ✅   | ✅   | ❌   | ❌   | ❌   | ❌   |
| 自动 SSL 证书 | 🚀  | ✅   | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| 反向代理 | 🚀  | ✅   | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| VPN 管理 | 🚀  | 🛠️ | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| 分级应用入口 | 🚀  | 🛠️ | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| 多用户管理 | ✅ 用户管理 <br>🚀 资源隔离 | ✅ 用户管理 <br>🛠️ 资源隔离 | ✅ 用户管理<br>🛠️ 资源隔离 | ❌   | ✅ 用户管理  <br>🛠️ 资源隔离 | ✅ 用户管理  <br>🛠️ 资源隔离 |
| 单一登录 | 🚀  | ❌   | ❌   | ❌   | ❌   |  ❌   |
| 跨节点存储 | 🚀 (Juicefs+<br>MinIO) | ❌   | ❌   | ❌   | ❌   | ❌   |
| 数据库解决方案 | 🚀 (内置云原生解决方案) | 🛠️ | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| 灾难恢复 | 🚀 (MinIO的[**纠错码**](https://min.io/docs/minio/linux/operations/concepts/erasure-coding.html)**)** | ✅ RAID | ✅ RAID | ✅ RAID | ❌   | ✅ Unraid Storage |
| 备份 | ✅ 应用数据  <br>✅ 用户数据 | ✅ 用户数据 | ✅ 用户数据 | ✅ 用户数据 | ✅ 用户数据 | ✅ 用户数据 |
| 应用沙盒 | ✅   | ❌   | ❌ （K8S的命名空间） | ❌   | ❌  | ❌   |
| 应用生态系统 | ✅ （官方 + 第三方应用） | ✅ （官方应用为主） | ✅ （官方应用 + 第三方提交）| ✅ （官方应用为主） | ❌  | ✅ （社区应用市场） |
| 开发者友好 | ✅ IDE  <br>✅ CLI  <br>✅ SDK  <br>✅ 文档| ✅ CLI  <br>✅ SDK  <br>✅ 文档 | ✅ CLI  <br>✅ 文档 | ✅ CLI  <br>✅ 文档 | ✅ SDK  <br>✅ 文档 | ✅ 文档 |
| 本地 LLM 部署 | 🚀  | 🛠️ | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| 本地 LLM 应用开发 | 🚀  | 🛠️ | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| 客户端 | ✅ Android  <br>✅ iOS  <br>✅ Windows  <br>✅ Mac  <br>✅ Chrome 插件 | ✅ Android  <br>✅ iOS | ❌   | ❌   | ❌  | ❌   |
| 客户端功能 | ✅ （一体化客户端应用） | ✅ （14个分散的客户端应用）| ❌   | ❌   | ❌   |  ❌   |

## 快速开始

> 当前文档仅有英文版本。

- [在 Linux 上开始](https://docs.jointerminus.com/overview/introduction/getting-started/linux.html)
- [在 Raspberry Pi 上开始](https://docs.jointerminus.com/overview/introduction/getting-started/raspberry.html)
- [在 macOS 上开始](https://docs.jointerminus.com/overview/introduction/getting-started/mac.html)
- [在 Windows上开始](https://docs.jointerminus.com/overview/introduction/getting-started/windows.html)

## 项目目录

Terminus 包含多个在 GitHub 上公开可用的代码仓库。当前仓库负责操作系统的最终编译、打包、安装和升级，而特定的更改主要在各自对应的仓库中进行。

以下表格列出了 Terminus 下的项目目录及其对应的仓库。

<details>
<summary><b>框架组件</b></summary>

| 路径 | 仓库 | 说明 |
| --- | --- | --- |
| [frameworks/app-service](https://github.com/beclab/terminus/tree/main/frameworks/app-service) | <https://github.com/beclab/app-service> | 系统框架组件，负责提供全系统应用的生命周期管理及多种安全控制。 |
| [frameworks/backup-server](https://github.com/beclab/terminus/tree/main/frameworks/backup-server) | <https://github.com/beclab/backup-server> | 系统框架组件，提供定时的全量或增量集群备份服务。 |
| [frameworks/bfl](https://github.com/beclab/terminus/tree/main/frameworks/bfl) | <https://github.com/beclab/bfl> | 启动器后端（Backend For Launcher, BFL），作为用户访问点的系统框架组件，整合并代理各种后端服务的接口。 |
| [frameworks/GPU](https://github.com/beclab/terminus/tree/main/frameworks/GPU) | <https://github.com/grgalex/nvshare> | GPU共享机制，允许多个进程（或运行在 Kubernetes 上的容器）安全地同时在同一物理 GPU 上运行，每个进程都可访问全部 GPU 内存。 |
| [frameworks/l4-bfl-proxy](https://github.com/beclab/terminus/tree/main/frameworks/l4-bfl-proxy) | <https://github.com/beclab/l4-bfl-proxy> | 针对 BFL 的第4层网络代理。通过预读服务器名称指示（SNI），提供一条动态路由至用户的 Ingress。 |
| [frameworks/osnode-init](https://github.com/beclab/terminus/tree/main/frameworks/osnode-init) | <https://github.com/beclab/osnode-init> | 系统框架组件，用于初始化新节点加入集群时的节点数据。 |
| [frameworks/system-server](https://github.com/beclab/terminus/tree/main/frameworks/system-server) | <https://github.com/beclab/system-server> | 作为系统运行时框架的一部分，提供应用间安全通信的机制。 |
| [frameworks/tapr](https://github.com/beclab/terminus/tree/main/frameworks/tapr) | <https://github.com/beclab/tapr> | Terminus 应用运行时组件。 |

</details>

<details>
<summary><b>系统级应用程序和服务</b></summary>

| 路径 | 仓库 | 说明 |
| --- | --- | --- |
| [apps/analytic](https://github.com/beclab/terminus/tree/main/apps/analytic) | <https://github.com/beclab/analytic> | 基于 [Umami](https://github.com/umami-software/umami) 开发的 Analytic，是一个简单、快速、注重隐私的 Google Analytics 替代品。 |
| [apps/market](https://github.com/beclab/terminus/tree/main/apps/market) | <https://github.com/beclab/market> | 此代码库部署了 Terminus 应用市场的前端部分。 |
| [apps/market-server](https://github.com/beclab/terminus/tree/main/apps/market-server) | <https://github.com/beclab/market> | 此代码库部署了 Terminus 应用市场的后端部分。 |
| [apps/argo](https://github.com/beclab/terminus/tree/main/apps/argo) | <https://github.com/argoproj/argo-workflows> | 用于协调本地推荐算法容器执行的工作流引擎。 |
| [apps/desktop](https://github.com/beclab/terminus/tree/main/apps/desktop) | <https://github.com/beclab/desktop> | 系统内置的桌面应用程序。 |
| [apps/devbox](https://github.com/beclab/terminus/tree/main/apps/devbox) | <https://github.com/beclab/devbox> | 为开发者提供的 IDE，用于移植和开发 Terminus 应用。 |
| [apps/TermiPass](https://github.com/beclab/terminus/tree/main/apps/TermiPass) | <https://github.com/beclab/TermiPass> | 基于 [Padloc](https://github.com/padloc/padloc) 开发的团队和企业的免费 1Password 和 Bitwarden 替代品，作为客户端帮助您管理 DID、Terminus 名称和 Terminus 设备。 |
| [apps/files](https://github.com/beclab/terminus/tree/main/apps/files) | <https://github.com/beclab/files> | 基于 [Filebrowser](https://github.com/filebrowser/filebrowser) 修改的内置文件管理器，管理 Drive、Sync 和各种 Terminus 物理节点上的文件。|
| [apps/notifications](https://github.com/beclab/terminus/tree/main/apps/notifications) | <https://github.com/beclab/notifications> | Terminus 的通知系统。 |
| [apps/profile](https://github.com/beclab/terminus/tree/main/apps/profile) | <https://github.com/beclab/profile> | Terminus 中的 Linktree 替代品。|
| [apps/rsshub](https://github.com/beclab/terminus/tree/main/apps/rsshub) | <https://github.com/beclab/rsshub> | 基于 [RssHub](https://github.com/DIYgod/RSSHub) 的 RSS 订阅管理器。 |
| [apps/settings](https://github.com/beclab/terminus/tree/main/apps/settings) | <https://github.com/beclab/settings> | 内置系统设置。 |
| [apps/system-apps](https://github.com/beclab/terminus/tree/main/apps/system-apps) | <https://github.com/beclab/system-apps> | 基于 *kubesphere/console* 项目构建的 system-service 提供一个自托管的云平台，通过视觉仪表板和功能丰富的 ControlHub 帮助用户了解和控制系统的运行状态和资源使用。 |
| [apps/wizard](https://github.com/beclab/terminus/tree/main/apps/wizard) | <https://github.com/beclab/wizard> | 向用户介绍系统激活过程的向导应用程序。 |
</details>

<details>
<summary><b>第三方组件和服务</b></summary>

| 路径 | 仓库 | 说明 |
| --- | --- | --- |
| [third-party/authelia](https://github.com/beclab/terminus/tree/main/third-party/authelia) | <https://github.com/beclab/authelia> | 一个开源的认证和授权服务器，通过网络门户为应用程序提供双因素认证和单点登录（SSO）。 |
| [third-party/headscale](https://github.com/beclab/terminus/tree/main/third-party/headscale) | <https://github.com/beclab/headscale> | 在 Terminus 中的 Tailscale 控制服务器的开源自托管实现，用于管理 TermiPass 中不同设备上的 Tailscale。|
| [third-party/infisical](https://github.com/beclab/terminus/tree/main/third-party/infisical) | <https://github.com/beclab/infisical> | 一个开源的密钥管理平台，可以在团队/基础设施之间同步密钥并防止泄露。 |
| [third-party/juicefs](https://github.com/beclab/terminus/tree/main/third-party/juicefs) | <https://github.com/beclab/juicefs-ext> | 基于 Redis 和 S3 之上构建的分布式 POSIX 文件系统，允许不同节点上的应用通过 POSIX 接口访问同一数据。 |
| [third-party/ks-console](https://github.com/beclab/terminus/tree/main/third-party/ks-console) | <https://github.com/kubesphere/console> | Kubesphere 控制台，允许通过 Web GUI 进行集群管理。 |
| [third-party/ks-installer](https://github.com/beclab/terminus/tree/main/third-party/ks-installer) | <https://github.com/beclab/ks-installer-ext> | Kubesphere 安装组件，根据集群资源定义自动创建 Kubesphere 集群。 |
| [third-party/kube-state-metrics](https://github.com/beclab/terminus/tree/main/third-party/kube-state-metrics) | <https://github.com/beclab/kube-state-metrics> | kube-state-metrics（KSM）是一个简单的服务，监听 Kubernetes API 服务器并生成关于对象状态的指标。 |
| [third-party/notification-mananger](https://github.com/beclab/terminus/tree/main/third-party/notification-manager) | <https://github.com/beclab/notification-manager-ext> | Kubesphere 的通知管理组件，用于统一管理多个通知渠道和自定义聚合通知内容。 |
| [third-party/predixy](https://github.com/beclab/terminus/tree/main/third-party/predixy) | <https://github.com/beclab/predixy> | Redis 集群代理服务，自动识别可用节点并添加命名空间隔离。 |
| [third-party/redis-cluster-operator](https://github.com/beclab/terminus/tree/main/third-party/redis-cluster-operator) | <https://github.com/beclab/redis-cluster-operator> | 一个基于 Kubernetes 的云原生工具，用于创建和管理 Redis 集群。 |
| [third-party/seafile-server](https://github.com/beclab/terminus/tree/main/third-party/seafile-server) | <https://github.com/beclab/seafile-server> | Seafile（同步驱动器）的后端服务，用于处理数据存储。 |
| [third-party/seahub](https://github.com/beclab/terminus/tree/main/third-party/seahub) | <https://github.com/beclab/seahub> | Seafile（同步驱动器）的前端和中间件服务，用于处理文件共享、数据同步等。 |
| [third-party/tailscale](https://github.com/beclab/terminus/tree/main/third-party/tailscale) | <https://github.com/tailscale/tailscale> | Tailscale 已在所有平台的 TermiPass 中集成。 |
</details>

<details>
<summary><b>其他库和组件</b></summary>

| 路径 | 仓库 | 说明 |
| --- | --- | --- |
| [build/installer](https://github.com/beclab/terminus/tree/main/build/installer) |     | 用于生成安装程序构建的模板。 |
| [build/manifest](https://github.com/beclab/terminus/tree/main/build/manifest) |     | 安装构建镜像列表模板。 |
| [libs/fs-lib](https://github.com/beclab/terminus/tree/main/libs) | <https://github.com/beclab/fs-lib> | 基于 JuiceFS 实现的 iNotify 兼容接口的SDK库。 |
| [scripts](https://github.com/beclab/terminus/tree/main/scripts) |     | 生成安装程序构建的辅助脚本。 |
</details>

## 社区贡献

我们欢迎任何形式的贡献！

- 如果您想在 Terminus 上开发自己的应用，请参考：<br>
https://docs.jointerminus.com/developer/develop/


- 如果您想帮助改进 Terminus，请参考：<br>
https://docs.jointerminus.com/developer/contribute/terminus-os.html

## 社区支持

* [**Github Discussion**](https://github.com/beclab/terminus/discussions) - 讨论 Terminus 使用过程中的疑问。
* [**GitHub Issues**](https://github.com/beclab/terminus/issues) - 报告 Terminus 的遇到的问题或提出功能改进建议。
* [**Discord**](https://discord.com/invite/BzfqrgQPDK) - 日常交流，分享经验，或讨论与 Terminus 相关的任何主题。
 
## 持续关注

关注 Terminus 项目，及时获取新版本和更新的通知。

 
![点亮星标](https://file.bttcdn.com/github/terminus/terminus.git.v2.gif)
 

## 特别感谢

Terminus 项目整合了许多第三方开源项目，包括：[Kubernetes](https://kubernetes.io/)、[Kubesphere](https://github.com/kubesphere/kubesphere)、[Padloc](https://padloc.app/)、[K3S](https://k3s.io/)、[JuiceFS](https://github.com/juicedata/juicefs)、[MinIO](https://github.com/minio/minio)、[Envoy](https://github.com/envoyproxy/envoy)、[Authelia](https://github.com/authelia/authelia)、[Infisical](https://github.com/Infisical/infisical)、[Dify](https://github.com/langgenius/dify)、[Seafile](https://github.com/haiwen/seafile)、[HeadScale](https://headscale.net/)、 [tailscale](https://tailscale.com/)、[Redis Operator](https://github.com/spotahome/redis-operator)、[Nitro](https://nitro.jan.ai/)、[RssHub](http://rsshub.app/)、[predixy](https://github.com/joyieldInc/predixy)、[nvshare](https://github.com/grgalex/nvshare)、[LangChain](https://www.langchain.com/)、[Quasar](https://quasar.dev/)、[TrustWallet](https://trustwallet.com/)、[Restic](https://restic.net/)、[ZincSearch](https://zincsearch-docs.zinc.dev/)、[filebrowser](https://filebrowser.org/)、[lego](https://go-acme.github.io/lego/)、[Velero](https://velero.io/)、[s3rver](https://github.com/jamhall/s3rver)、[Citusdata](https://www.citusdata.com/)。
