---
outline: [2,3]
description: 了解 Olares One 的硬件、本地 AI 应用、订阅费用、模型升级、移动端访问、购买和售后支持。
head:
  - - meta
    - name: keywords
      content: Olares One, Olares, 个人云
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../one/faq.md)为准。
:::

# Olares One 常见问题

查找关于 Olares One 和系统功能的常见问题解答。

:::tip Olares One 用户支持
如果你需要产品功能、硬件保修或特定使用场景方面的帮助，请通过 WhatsApp 或电子邮件 hi@olares.com 直接联系技术支持。我们将在两个工作日内回复。
:::

## 产品与规格

### 什么是 Olares One？

Olares One 是一款专为本地 AI 设计的个人云设备。它将工作站级别的硬件与 Olares OS 相结合，后者是一个开源操作系统，让你能够在物理控制的硬件上运行 AI 代理并存储数据。

### 什么是个人云？

个人云是一种私有基础设施，它复制了公共云服务的实用性，例如从任何地方访问文件和计算能力。它完全在你自己的硬件上运行，以确保数据主权。

### Olares One 的技术规格是什么？

* **处理器**：Intel Ultra 9 275HX
* **GPU**：NVIDIA RTX 5090 Mobile
* **内存**：96GB RAM
* **存储**：2TB NVMe SSD

详细信息请参阅 [Olares One 规格](spec)。

### Olares One 的 RTX 5090 Mobile 与桌面版 RTX 5090 有什么区别？

两者都采用 NVIDIA Blackwell 架构，支持 CUDA。桌面版拥有更多计算资源和显存，移动版则面向功耗和散热空间更小的设备。

| 规格 | Olares One 的 RTX 5090 Mobile | 桌面版 RTX 5090 公版规格 |
| --- | --- | --- |
| CUDA 核心数 | 10,496 | 21,760 |
| 显存 | 24 GB GDDR7 | 32 GB GDDR7 |
| 显存位宽 | 256 位 | 512 位 |
| GPU 功耗 | 性能模式下最高 175 W | 整卡功耗 575 W |

对于本地 AI，桌面版多出的 8 GB 显存可以容纳更多模型权重和对话上下文，更多核心则提供更强的并行计算能力。移动版的功耗更低，散热需求也更小，让 Olares One 能在紧凑的机身中运行本地 AI。

显卡规格来自 NVIDIA 的[移动版 GPU 对比](https://www.nvidia.com/en-us/geforce/laptops/compare/)和[桌面版 RTX 5090 规格](https://www.nvidia.com/en-us/geforce/graphics-cards/50-series/rtx-5090/)。Olares One 的功耗模式请参阅[功耗是多少](#功耗是多少)。

### 设备噪音有多大？

Olares One 即使在负载下也能安静运行。在实验室测试中，设备空闲时产生 19dB 噪音，最大负载下仍保持在 39dB 以下。

### 功耗是多少？

设备待机时消耗 30W。负载下，GPU 功耗取决于你选择的模式：
* **静音模式**：GPU 50W
* **性能模式**：GPU 175W

当 CPU 单独满负荷运行时，功耗达到 120W。在组合最大负载下，系统平衡功耗分配，CPU 运行在 55W，GPU 运行在 175W。散热设计可维持这些性能水平。

### 有哪些可用端口？

设备包含以下物理端口：
* 1 × Thunderbolt 5，带宽 80 Gbps
* 1 × RJ45 以太网，2.5 Gbps
* 1 × USB-A 3.2 Gen2，速度 10 Gbps
* 1 × HDMI 2.1，带宽 48 Gbps

### Olares One 与 DGX Spark、AI Max+ 395 和其他 AI PC 有何不同？

Olares One 作为个人云解决方案运行，而非传统的个人电脑。

个人云作为稳定的外部服务运行，可随时从任何设备访问。PC 是作为客户端软件设计的，通过显示器和键盘直接交互。

Olares One 在 x86 架构上利用 NVIDIA CUDA 技术栈。这确保了对标准开源 AI 应用的广泛兼容性，避免了基于 ARM 的 AI PC 或没有 CUDA 支持的设备中常见的适配挑战。

Olares OS 通过一键部署和工作流集成简化了本地 AI，这与标准 PC 典型的手动软件管理不同。

### 是否可以升级或扩展硬件？

可以。你可以通过多个内部插槽和扩展端口修改硬件：
* **存储**：主板包含两个 PCIe SSD 插槽，一个 PCIe 4.0 和一个 PCIe 5.0。预装的 2TB SSD 占用 PCIe 5.0 插槽。空闲的 PCIe 4.0 插槽可用于通过 LVM 扩展系统存储，或用于双启动配置。操作步骤请参阅[使用内置 SSD 扩展存储](expand-storage-internal-ssd.md)。
* **内存**：RAM 容量可升级至最大 128GB。
* **外接显卡**：可以连接兼容的雷电外置显卡设备。详情请参阅[可以给 Olares One 连接 eGPU 吗？](#可以给-olares-one-连接-egpu-吗)
* **外接显示器**：Thunderbolt 5 端口支持通过扩展坞同时连接最多两台外接显示器，最高支持 8K 分辨率。视频输出由内置 RTX 5090 Mobile 处理。

### 可以给 Olares One 连接 eGPU 吗？

可以。你可以连接已安装 GPU 的雷电外置显卡设备，也可以通过独立供电的 eGPU dock 或 eGPU enclosure 使用兼容的桌面版 NVIDIA 显卡。不同硬件组合在 Olares OS 和 Windows 11 上的兼容状态及设置方法并不相同。

请参阅[将 eGPU 连接到 Olares One](egpu.md)，查看已知兼容状态、当前限制和设置方法。

### 有没有办法通过 HDMI 查看 Olares OS 界面？

目前，HDMI 输出显示的是 Ubuntu shell 或用于游戏的 Steam Headless 输出，而不是像传统 PC 那样的完整桌面 UI。

我们计划在未来的更新中引入通过 HDMI 直接运行轻量级浏览器或 ChromeOS 实例的功能。

### 我可以更新或配置 BIOS 吗？

可以。随着新版本的发布，你可以更新 BIOS 和嵌入式控制器（EC）固件。虽然系统默认隐藏高级 BIOS 选项以保持稳定，但你也可以解锁它们，根据自己的需求进行深度硬件配置。

详细说明请参阅[管理 BIOS](../one/update-firmware.md)。

### Olares One 是否设计为"始终在线"运行？

是的。Olares One 作为个人 AI 云运行，可可靠支持 24/7 连续运行，适用于大语言模型（LLM）托管等场景。

### Olares One 支持带外管理吗？

Olares One 对常见带外管理功能的支持情况如下：

- **自动开机**：支持。接通电源或停电后恢复供电时，Olares One 可以自动开机。
  - 使用 Olares OS 时，需要 Olares OS 1.12.6 或更高版本以及 EC 固件 1.03 或更高版本。操作方法请参阅[查看设备状态与硬件设置](hardware-settings.md#设置自动开机)。
  - 使用 Ubuntu 时，需要通过命令行完成配置。操作方法请参阅 [Ubuntu Server](install-ubuntu-server.md#配置自动开机) 或 [Ubuntu Desktop](install-ubuntu-desktop.md#配置自动开机) 安装指南中的**配置自动开机**章节。
- **网络唤醒（WOL）**：运行 Olares OS、Ubuntu 或 Windows 的 Olares One 支持此功能，需要 EC 固件不低于 1.01，并使用有线网络连接。同一局域网内的配置和唤醒方法，请参阅[为 Olares One 设置网络唤醒](wake-on-lan.md)。其中 Windows 步骤适用于从睡眠状态唤醒。
- **远程 KVM**：不支持。
- **IPMI**：不支持。

### 使用 Olares One 是否有额外费用？

Olares OS、系统更新、Olares Market 中的应用和本地 AI 推理均可免费使用。

日常运行成本是电费和网络费。如果你选择接入 OpenAI、Anthropic 等云端模型 API，调用费用由该提供方收取。

### 我可以在没有互联网连接的情况下使用该设备吗？

一次性激活后，设备可以完全离线工作。你可以通过本地网络访问它，无需互联网连接。

### 设备能否在断电期间防止数据丢失？

Olares One 使用 ext4 Linux 文件系统，防止突然断电期间的文件系统损坏。电源和主板设计包含电压调节以保护硬件。

Olares OS 还包含内置的自动备份机制，用于定期将加密备份到外部位置。

我们计划增强对 UPS 设备的支持，以确保断电期间优雅关机，同时支持类似 Time Machine 的快照备份。

### 设备是否支持 GPU MUX 切换？

是的。Olares One 支持 GPU MUX 和 Optimus 两种模式。系统默认使用 GPU MUX 模式以启用与独立显卡的直接连接。如有需要，可以在 BIOS 中更改此设置。

## AI 与游戏

### 可以用本地 AI 替代付费 AI 订阅吗？

可以。你可以用本地模型完成对话、写作、编程、研究和媒体创作，按订阅的实际用途选择方案：

| 用途 | 本地模型或工作流 | 应用示例 |
| --- | --- | --- |
| 对话和写作 | 由 Qwen 等本地大语言模型回答问题、起草文字 | 使用 [LobeHub](../use-cases/lobechat.md) 或 [Open WebUI](../use-cases/openwebui.md) 作为聊天界面 |
| 文档问答 | 从知识库检索文档，再由本地大语言模型回答 | [LobeHub](../use-cases/lobechat.md) |
| 带引用的研究报告 | 通过 SearXNG 检索资料，再由本地大语言模型分析 | [Local Deep Research](../use-cases/local-deep-research.md) |
| 编写和测试代码 | 本地模型配合 Agent 修改文件、运行测试 | [OpenCode](../use-cases/opencode.md)、[Claude Code](../use-cases/claude-code.md) 或 [Codex CLI](../use-cases/codex-cli.md) |
| 图片和视频创作 | 工作流调用本地模型，生成或编辑图片和视频 | [FlowStudio](../use-cases/flowstudio.md) 或 [ComfyUI](../use-cases/comfyui.md) |
| 音乐创作 | 本地音乐模型根据歌词和风格提示生成歌曲 | [YuE2 Studio](../use-cases/yue2-studio.md) 或 [ACE-Step](../use-cases/ace-step-1.5.md) |

更多应用和模型请浏览 [Olares Market](https://market.olares.com)。

### 可以在 Olares One 上托管全天运行的 AI Agent 吗？

可以。Agent 托管在 Olares One 上后，可以全天运行，无需保持笔记本或手机在线。从 Market 安装 Agent 并连接本地模型后便可使用，例如：

- **[Lares](../use-cases/lares.md)**：Olares 原生 Agent，可以管理文件、应用和设备设置，也可以连接工具完成研究和创作任务。
- **[OpenClaw](../use-cases/openclaw.md)**：可以通过聊天应用交互的个人 Agent，结合工具和技能处理文档、浏览网页。
- **[Hermes Agent](../use-cases/hermes.md)**：支持跨会话记忆和可复用技能，可以管理文件、应用和处理编程任务。

### 可以运行哪些本地大语言模型？

你可以从 Market 安装本地大语言模型，例如：

- **Qwen3.8-27B**：使用 llama.cpp 推理引擎。
- **Qwen3.6-35B-A3B**：使用 llama.cpp 推理引擎。
- **Gemma 4 26B**：使用 Ollama 推理引擎。

这些模型应用已配置好推理引擎，并在安装设置时下载模型文件。最新列表请查看 [Olares Market](https://market.olares.com)，使用方法请参阅[运行本地大语言模型](../use-cases/llm-base-apps.md)。

### 可以安装新模型或 Market 中没有的模型吗？

可以。Market 上架新模型应用后，你可以直接安装。如果想运行 Hugging Face 或 Ollama 模型库中的模型，可以通过 Engine Base 应用创建实例。Olares 提供 llama.cpp、Ollama、vLLM 和 SGLang 四种 Engine Base 应用。

选择支持该模型架构和文件格式的引擎，再按设备资源设置内存和上下文上限。[本地模型指南](../use-cases/llm-base-apps.md)介绍了引擎选择和配置方法。

### 显存和系统内存会怎样影响模型选择？

Olares One 配备 24 GB 显存（VRAM）和 96 GB 系统内存（RAM），两者独立。显存存放 GPU 推理所需的模型权重和上下文缓存。系统内存用于运行操作系统和应用。使用支持 CPU 与 GPU 协同推理的引擎时，也可以把较大模型的一部分放入系统内存，由 CPU 处理。

量化通过降低模型权重的存储精度来减少内存占用。除了模型权重，还需要为上下文缓存预留空间，更长的对话和更多并发请求都会增加这部分开销。模型和内存设置请参阅[运行本地大语言模型](../use-cases/llm-base-apps.md)。

### 可以让编程 Agent 使用本地模型吗？

可以。从 Market 安装 [OpenCode](../use-cases/opencode.md)、[Claude Code](../use-cases/claude-code.md) 或 [Codex CLI](../use-cases/codex-cli.md)，通过 Router 连接本地模型。模型生成回复和工具调用，Agent 在自己的开发环境中修改文件、运行命令和测试。

你也可以在自己的电脑上运行 OpenCode CLI，连接 Olares 上的模型。此时，文件和代码执行都留在自己的电脑上，模型推理由 Olares One 完成。请参阅 [AI 应用如何连接 Olares 模型](../manual/best-practices/connect-ai-apps.md)。

### Olares One 支持哪些语言？

语言支持取决于你使用的功能：

- **Olares 和 LarePass 界面**：支持英语、简体中文、德语、西班牙语、意大利语、法语和日语。请参阅 [Olares 语言设置](../manual/olares/settings/language-appearance.md)和 [LarePass 语言设置](../manual/larepass/index.md#更改语言)。
- **第三方应用**：由各应用提供自己的语言选项。
- **AI 模型**：对话、语音识别和语音生成模型各有自己的语言支持范围，可以按目标语言选择模型。

### 本地 AI 模型可以理解图片吗？

可以。支持视觉的模型能够分析上传的图片，并回答有关图片内容的问题。例如，Gemma 4 26B (Ollama) 支持图片输入。

将模型连接到支持图片附件的聊天应用，再上传图片并提问。模型设置请参阅[运行本地大语言模型](../use-cases/llm-base-apps.md)。

### Olares One 是否支持 Kimi K2 模型？

Kimi K2 约有 1 万亿参数，需要大约 1024 GB 显存，因此无法在 Olares Market 中使用。

我们预计，随着 AI 模型尺寸持续缩小，未来可能会在 Olares One 上运行 Kimi-K2 级别的模型。

### 本地 AI 模型能否访问互联网进行研究？

可以。Agent 可以调用搜索和网页读取工具，再用本地模型分析结果。Olares Market 提供这两类工具：

- **[SearXNG](../use-cases/searxng.md)**：搜索网页，返回链接和摘要。
- **[Firecrawl](../use-cases/firecrawl.md)**：读取网页，将内容转成模型可以处理的文本。

将这些工具连接到 Agent 后，就可以组成研究工作流。使用 Lares 和 Router 的设置方法请参阅[使用 Lares 研究网络资料](../use-cases/lares-web-research.md)。

### 我可以将 NAS 用于存储，将 Olares 用于 AI 处理吗？

是的，我们强烈推荐这种工作流。你可以将 NAS 作为 SMB 共享挂载到 Olares One 上。然后 Olares 可以通过 Immich 等应用对 NAS 上存储的照片和视频进行索引、排序和处理。

### 我可以在 iPhone 或 Mac 上使用 Olares AI 功能吗？

可以。在手机或电脑的浏览器中打开托管在 Olares One 上的 AI 应用。本地模型推理由 Olares One 完成，手机或电脑负责显示界面和发送请求。

外出访问时，保持 Olares One 开机并连接互联网，然后在手机或电脑上开启 [LarePass VPN](access-olares-via-vpn.md)。

### 我可以使用 LM Studio 管理 Olares 上的模型吗？

LM Studio 是一个客户端桌面应用，而非服务器端服务，因此无法直接安装在 Olares One 上。

但是，你有两个替代方案：
* 使用 [LobeHub](../use-cases/lobechat.md) 与 Olares 托管的模型对话。
* 在你的电脑上运行 LM Studio 的本地实例，并将其配置为通过 API 访问托管在 Olares One 上的 AI 模型。

### 如果同时加载多个模型，性能会下降吗？

加载多个超过 GPU 显存的模型通常会导致崩溃。Olares 通过时间切片机制解决了这个问题。

系统会暂时将不活跃的模型交换到 96GB 系统内存中，并将活跃模型加载到 GPU 中。这样可以在切换期间保持多个模型就绪，仅产生约 5% 的性能开销。

或者，你可以：
- 分割 GPU 显存，让多个应用同时运行而无需交换。
- 将整个 GPU 分配给单个应用以获得最大原始性能。

直接在 Olares 设置中切换这些模式，无需额外修改。

### 该设备是否支持高端游戏？

是的。你可以连接显示器直接通过 Steam 游玩，或将设备用作游戏服务器，通过 Moonlight 将游戏串流到笔记本电脑、电视或手机上。

### 我可以将 Olares One 用于 Quest 3 等 VR 设备吗？

我们正在积极测试与 Quest 3 等 VR 和 AR 设备的兼容性，以纳入未来的更新中。

### 游戏性能与 Windows 相比如何？

Olares One 支持通过 Moonlight 作为 Steam 服务器进行串流游戏，或通过连接的显示器直接游玩。

在直接游玩场景中，测试显示《赛博朋克 2077》达到了与相同硬件运行 Windows 相比约 90% 的性能。这与其他高端 Linux 游戏设置一致。随着 SteamOS 采用率的增长，我们预计兼容性将进一步改善。

## 操作系统

### 我可以在 Olares One 上运行 Windows 吗？

可以。要运行偶尔的 Windows 应用，可以从 Olares Market 安装 Windows 应用。这会运行一个可通过 RDP 访问的虚拟机，与内置 VPN 配合使用时提供无缝的远程桌面体验。

或者，你可以安装第二个内部驱动器用于双启动设置，或将 Olares OS 替换为原生 Windows 安装。

操作步骤请参阅[运行 Windows 虚拟机](../use-cases/windows.md)、[使用第二块硬盘配置双系统](dual-boot-dual-drive.md)或[在 Olares One 上安装 Windows](install-windows-primary-drive.md)。

### Windows 虚拟机可以使用 GPU 吗？

目前不支持 GPU 直通。NVIDIA RTX 5090 Mobile 作为 Olares OS 的主 GPU，为本地 AI 应用提供支持。将其分配给虚拟机会使其脱离主机操作系统，并禁用这些功能。

Windows 虚拟机使用集成的 Intel 显卡，可有效处理轻量级任务。对于游戏等 GPU 密集型工作负载，我们建议设置双启动配置。

### 我可以抹掉 Olares OS 并原生安装 Linux 或 Windows 吗？

可以。你完全拥有硬件的所有权。你可以抹掉预装的操作系统来安装 Windows 或任何 Linux 发行版。在这种配置下，Olares One 作为配备 NVIDIA RTX 5090 Mobile GPU 的标准高性能工作站运行，让你可以像标准台式电脑一样使用显示器、键盘和鼠标。

### 如何设置双启动系统？

我们建议在可用的插槽中安装第二个 NVMe SSD，将操作系统放在不同的驱动器上。

对于干净的安装，先安装 Windows，然后安装 Olares OS。

由于 Olares OS 基于 Ubuntu 24.04，它遵循标准的 Linux 双启动程序。

### 系统支持在 Olares OS 和 Windows 之间切换吗？

是的。UEFI 启动管理器和 GRUB 都支持这种配置。你可以在 BIOS 中将 Olares OS 设置为持久的默认启动选项。要在游戏等特定任务中启动到 Windows，请在启动序列中选择 Windows 启动管理器。

## 集群

### 可以将两台 Olares One 设备连接在一起吗？

可以。Olares OS 基于 Kubernetes 构建，允许多台 Olares 设备组成集群。系统会自动在集群内调度应用，并在多个设备间加载模型。

目前组建集群需要命令行操作。我们计划在未来的更新中引入完全基于 UI 的体验来管理此过程。

### 集群化能让单个游戏或 AI 任务运行得更快吗？

不能。集群化提高了总系统吞吐量和并发能力，但不会让单个任务的速度翻倍。

例如，通过将两台设备集群化：
- 你可以在一台设备上运行《艾尔登法环》，同时在另一台上运行《赛博朋克 2077》。但是，你不能将两台设备组合起来以双倍的帧率运行单个《赛博朋克 2077》实例。
- 你可以在 6 秒内生成两张图像，而不是在 3 秒内生成一张图像。

### 集群化如何使 AI 工作负载受益？

集群化支持单个设备无法处理的大型模型或复杂工作流。

对于 LLM，系统使用 vLLM 流水线并行性将模型分布在多个设备上。这支持比单个设备处理的模型大得多的模型，尽管由于网络延迟，推理速度相比在单个主板上使用多个 GPU 有所下降。

它还支持复杂的流水线。例如，你可以运行一个数字人应用，其中 LLM 在一个节点上运行，而文本转语音（TTS）和自动语音识别（ASR）服务在另一个节点上运行。

### 集群化会增加可用内存吗？

是的。连接两台相同设备时，总可用内存容量大约翻倍。系统会将应用分布在节点上以最大化可用资源。

### 我可以与 Mac Studio 或 DGX Spark 等第三方设备组建集群吗？

目前，Olares 仅支持具有相同架构并运行相同操作系统的设备之间的集群。

我们计划在未来的更新中增加对 Mac Studio、DGX Spark 和 AI MAX 395+ 等第三方设备的集群兼容性。

### 我可以将 NAS 作为集群的一部分吗？

不能。大多数 NAS 操作系统使用封闭环境，无法运行 Olares 节点。

但是，Olares One 可以通过 SMB 协议挂载 NAS 目录。这允许你将 NAS 上的文件管理得就像 Olares One 上的本地文件夹一样。

### 集群中的 GPU 调度如何工作？

Olares OS 在操作系统级别管理 GPU 分配。GPU 调度分为四个复杂度阶段：
1. 单节点，单 GPU。
2. 单节点，多 GPU。
3. 多节点，多 GPU，相同架构。
4. 多节点，多 GPU，不同架构。

Olares 目前处于第 3 阶段，即集群化多台运行相同操作系统和架构的设备。

第 4 阶段仍部分自动化，需要手动干预，例如拉取对应架构的特定容器镜像。我们计划在未来的更新中重写调度算法以自动化此支持。

### 是否支持联邦学习或共享计算？

我们正在探索高级计算模型，包括联邦学习、分布式 AI 训练和共享计算网络。

虽然硬件可以轻松处理这些工作负载，但构建稳定、生产就绪的解决方案是一个长期目标。最终，这些功能依赖于第三方应用而非核心操作系统。

### 如果设备用于集群，为什么以太网端口限制为 2.5Gbps？

我们理解 10Gbps 更适合集群以最大化节点间的数据传输。但是，包含 2.5Gbps 以太网端口是硬件平台供应商施加的严格限制。

## 购买与售后

### 在哪里查看价格、运费和税费？

当前价格、配置和优惠活动请查看 [Olares One 商店页面](https://shop.olares.com/products/olares-one)。选择收货地区后，在结账时查看运费和税费。是否包含关税和税费取决于目的地，详情请参阅[配送政策](https://shop.olares.com/policies/shipping-policy)。

如果需要购买多台设备，或结账时找不到收货地区，可以联系 hi@olares.com。

### 在哪里查看质保和退货信息？

[质保政策](https://shop.olares.com/pages/warranty-policy)说明保修范围、维修、换货和运输安排。[退货与退款政策](https://shop.olares.com/policies/refund-policy)说明退货条件和地区条款。

遇到硬件问题时，请联系 hi@olares.com，提供订单号、设备序列号和问题描述。支持团队会协助排查，如需返修，会提供寄回说明。

## 高级使用与配置

### Olares 是否支持多个用户拥有各自独立的账户？

是的。Olares 作为多用户系统运行。你可以为朋友或家人创建独立的账户，他们可以使用自己的 LarePass 应用连接。

系统支持三种账户角色：
* **超级管理员**：最初激活系统的用户，管理管理员。
* **管理员**：执行集群级任务。
* **成员**：访问共享服务。

### 我可以在 Olares One 上托管邮件服务器吗？

托管邮件服务器在技术上是可行的，但很复杂。

Olares 支持 Mail-in-a-Box、Mailserver 和 Mailcow 等开源解决方案。但是，主要提供商通常将自托管邮件服务器标记为垃圾邮件，这使得可靠投递变得困难。

### 是否可以为出站流量安装 VPN？

可以。Olares 内置支持 Tailscale 和 Headscale。为你的出站流量配置并使用特定的出口节点。

### 我需要静态 IP 才能将 Olares One 作为服务器托管吗？

不需要。Olares 提供两种外部访问选项：反向代理和 VPN。这些解决方案允许你在没有静态 IP 的情况下从任何地方安全访问你的设备。

### 我可以使用自己的域名访问 Olares One 吗？

可以。你可以使用自己的域名托管 Olares One。

目前，你必须将子域的 NS 记录指向 Olares 的名称服务器以设置反向代理。

我们计划为希望管理自己的反向代理或不需要公共互联网访问的用户引入更新。

### 有没有办法在没有互联网的情况下访问设备？

有。Olares 提供使用 `.local` 域名的完全本地访问选项。

如果设备位于局域网但缺乏公共访问，可以使用 `.local` 域名正常访问所有功能。在没有互联网连接的情况下，依赖外部服务的功能（如 Olares Market）将不可用。

### Olares One 可以作为媒体服务器使用吗？

可以。Olares One 可以作为 Plex 服务器运行。它支持硬件加速转码，并利用 CPU 和 GPU 进行高效解码。

### 我可以使用 SMB 将文件同步或备份到 NAS 吗？

可以。Olares One 支持 SMB 共享，允许你与外部 NAS 同步文件或备份数据。

### 文件系统是否加密？

默认情况下，系统不使用全盘加密。我们计划在未来的更新中将其作为标准的用户可配置选项。
