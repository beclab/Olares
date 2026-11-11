---
outline: [2,3]
description: Find answers about Olares One hardware, local AI apps, subscriptions, model upgrades, mobile access, purchasing, and support.
head:
  - - meta
    - name: keywords
      content: Olares One, Olares, personal cloud
---
# Olares One FAQs

Find answers to frequently asked questions about Olares One and system capabilities.

:::tip Support for Olares One owners
If you require assistance with product features, hardware warranty, or specific use cases, contact technical support directly via WhatsApp or email at hi@olares.com. We respond to inquiries within two business days.
:::

## Product & specifications

### What is Olares One?

Olares One is a dedicated personal cloud device designed for local AI. It integrates workstation-grade hardware with Olares OS, an open-source operating system that allows you to run AI agents and store data on hardware you physically control.

### What is a personal cloud?

A personal cloud is a private infrastructure that replicates the utility of public cloud services, such as anywhere-access to files and computing power. It runs entirely on your own hardware to ensure data sovereignty.

### What are the technical specifications for Olares One?

* **Processor**: Intel Ultra 9 275HX
* **GPU**: NVIDIA RTX 5090 Mobile
* **Memory**: 96GB RAM
* **Storage**: 2TB NVMe SSD

For details, see [Olares One specifications](spec).

### How does the RTX 5090 Mobile in Olares One compare with the desktop RTX 5090?

Both GPUs use NVIDIA's Blackwell architecture and support CUDA. The desktop version has more compute resources and memory, while the mobile version operates within a smaller power and cooling budget.

| Specification | RTX 5090 Mobile in Olares One | Desktop RTX 5090 reference design |
| --- | --- | --- |
| CUDA cores | 10,496 | 21,760 |
| GPU memory | 24 GB GDDR7 | 32 GB GDDR7 |
| Memory interface | 256-bit | 512-bit |
| GPU power | Up to 175 W in Performance mode | 575 W total graphics power |

For local AI, the desktop GPU's extra 8 GB of VRAM provides more room for model weights and conversation context. Its larger core count provides more parallel computing capacity. The mobile GPU's lower power budget reduces cooling requirements, allowing Olares One to provide local AI computing in a compact enclosure.

GPU specifications are from NVIDIA's [laptop GPU comparison](https://www.nvidia.com/en-us/geforce/laptops/compare/) and [desktop RTX 5090 specifications](https://www.nvidia.com/en-us/geforce/graphics-cards/50-series/rtx-5090/). For Olares One power modes, see [What is the power consumption?](#what-is-the-power-consumption).

### How loud is the device?

Olares One operates quietly even under load. In lab testing, the device generates 19dB when idle and remains under 39dB under maximum load.

### What is the power consumption?

The device consumes 30W in standby. Under load, the GPU power depends on your selected mode:
* **Silent Mode**: GPU 50W
* **Performance Mode**: GPU 175W

When the CPU is fully utilized on its own, power consumption reaches 120W. During combined maximum load, the system balances power distribution with the CPU operating at 55W and the GPU at 175W. The thermal design sustains these performance levels.

### What ports are available?

The device includes the following physical ports:
* 1 × Thunderbolt 5 with 80 Gbps bandwidth
* 1 × RJ45 Ethernet at 2.5 Gbps
* 1 × USB-A 3.2 Gen2 with 10 Gbps speed
* 1 × HDMI 2.1 with 48 Gbps bandwidth

### How is Olares One different from DGX Spark, AI Max+ 395 and other AI PCs?

Olares One functions as a personal cloud solution rather than a traditional personal computer.

A personal cloud runs as a stable, external service accessible anytime from any device. A PC acts as client-based software designed for direct interaction via a monitor and keyboard.

Olares One utilizes the NVIDIA CUDA stack on x86 architecture. This ensures broad compatibility with standard open-source AI applications, avoiding the adaptation challenges often found in ARM-based AI PCs or devices without CUDA support.

Olares OS simplifies local AI with one-click deployment and workflow integration, distinct from the manual software management typical of standard PCs.

### Is it possible to upgrade or expand the hardware?

Yes. You can modify the hardware through several internal slots and expansion ports:
* **Storage**: The motherboard includes two PCIe SSD slots, one PCIe 4.0 and one PCIe 5.0. The preinstalled 2TB SSD occupies the PCIe 5.0 slot. The PCIe 4.0 slot remains available for system storage expansion via LVM or for a dual-boot configuration.
* **Memory**: The RAM capacity can be upgraded to a maximum of 128GB.
* **External graphics**: You can connect a compatible Thunderbolt external graphics device. See [Can I use an eGPU with Olares One?](#can-i-use-an-egpu-with-olares-one)
* **External displays**: The Thunderbolt 5 port supports docking stations with up to two external monitors at up to 8K resolution. The built-in RTX 5090 Mobile handles video output.

For storage upgrade instructions, see [Expand storage with an internal SSD](expand-storage-internal-ssd.md).

### Can I use an eGPU with Olares One?

Yes. You can connect a Thunderbolt external graphics device with a GPU already installed, or install a compatible desktop NVIDIA GPU in a powered eGPU dock or enclosure. Hardware compatibility and setup differ between Olares OS and Windows 11.

See [Connect an eGPU to Olares One](egpu.md) for known compatibility, current limitations, and setup instructions.

### Is there a way to view the Olares OS UI via HDMI?

Currently, the HDMI output displays an Ubuntu shell or Steam Headless output for gaming, rather than a full desktop UI like a traditional PC. 

We plan to introduce the ability to run a lightweight browser or ChromeOS instance directly via HDMI in future updates.

### Can I update or configure the BIOS?

Yes. You can update the BIOS and Embedded Controller (EC) firmware as new versions release. While the system hides advanced BIOS options by default to maintain stability, you can also unlock them to perform deep hardware configurations according to your needs.

For detailed instructions, see [Manage BIOS](../one/update-firmware.md).

### Is Olares One designed for "always-on" operation?

Yes. Olares One functions as a personal AI cloud and reliably supports 24/7 continuous operation for scenarios like Large Language Model (LLM) hosting.

### Does Olares One support out-of-band management?

Support for common out-of-band management capabilities on Olares One is as follows:

- **Automatic startup**: Supported. Olares One can start automatically when AC power is connected or restored after a power outage.
  - On Olares OS, this feature requires Olares OS 1.12.6 or later and EC firmware 1.03 or later. See [Check device status and manage hardware settings](hardware-settings.md#set-automatic-startup).
  - On Ubuntu, configure the setting from the command line. See **Configure automatic startup** in the [Ubuntu Server](install-ubuntu-server.md#configure-automatic-startup) or [Ubuntu Desktop](install-ubuntu-desktop.md#configure-automatic-startup) installation guide.
- **Wake-on-LAN (WOL)**: Supported on Olares One running Olares OS, Ubuntu, or Windows, with EC firmware 1.01 or later and a wired network connection. For setup and wake-up instructions on the same local network, see [Set up Wake-on-LAN for Olares One](wake-on-lan.md). The Windows steps cover waking from sleep.
- **Remote KVM**: Not supported.
- **IPMI**: Not supported.

### Are there additional fees to use Olares One?

Olares OS, system updates, apps in Olares Market, and local AI inference are free to use.

Your regular running costs are electricity and internet access. If you choose to connect a cloud model API, such as OpenAI or Anthropic, its usage is billed by that provider.

### Can I use the device without an internet connection?

After a one-time activation, the device can work entirely offline. You can access it via your local network without an internet connection.

### Does the device prevent data loss during power outages?

Olares One uses the ext4 Linux file system to prevent file system corruption during sudden power loss. The power supply and motherboard design include voltage regulation for hardware protection.

Olares OS also includes a built-in automatic backup mechanism for periodic, encrypted backups to external locations.

We plan to enhance support for UPS devices to ensure graceful shutdowns during power interruptions, alongside native support for snapshot backups similar to Time Machine.

### Does the device support a GPU MUX switch?

Yes. Olares One supports both GPU MUX and Optimus modes. The system defaults to GPU MUX mode to enable a direct connection to the discrete GPU. Change this setting in the BIOS if needed.

## AI & gaming

### Can I use local AI instead of paid AI subscriptions?

Yes. You can run local models for chat, writing, coding, research, and media creation. Choose a setup based on the tasks you use a subscription for:

| Task | Local models or workflow | Example apps |
| --- | --- | --- |
| Chat and writing | A local LLM, such as Qwen, answers questions and drafts text | [LobeHub](../use-cases/lobechat.md) or [Open WebUI](../use-cases/openwebui.md) as the chat interface |
| Document Q&A | A local LLM works with documents retrieved from your knowledge base | [LobeHub](../use-cases/lobechat.md) |
| Research reports with citations | A local LLM analyzes sources retrieved through SearXNG | [Local Deep Research](../use-cases/local-deep-research.md) |
| Writing and testing code | A local model works with an agent that edits files and runs tests | [OpenCode](../use-cases/opencode.md), [Claude Code](../use-cases/claude-code.md), or [Codex CLI](../use-cases/codex-cli.md) |
| Image and video creation | A workflow calls local models to generate or edit images and videos | [FlowStudio](../use-cases/flowstudio.md) or [ComfyUI](../use-cases/comfyui.md) |
| Music creation | Local music models generate songs from lyrics and style prompts | [YuE2 Studio](../use-cases/yue2-studio.md) or [ACE-Step](../use-cases/ace-step-1.5.md) |

Browse [Olares Market](https://market.olares.com) for more apps and models.

### Can I host always-on AI agents on Olares One?

Yes. Agents hosted on Olares One can run around the clock, independently of your laptop or phone. Install an agent from Market and connect it to a local model. Available agents include:

- **[Lares](../use-cases/lares.md)**: Olares' native agent for managing files, apps, and device settings, with integrations for research and creative tasks.
- **[OpenClaw](../use-cases/openclaw.md)**: A personal agent you can interact with through messaging apps, with tools and skills for tasks such as document handling and web browsing.
- **[Hermes Agent](../use-cases/hermes.md)**: An agent with memory across sessions and reusable skills for managing files, apps, and coding tasks.

### Which local LLMs can I run?

You can install local large language models (LLMs) from Market, including:

- **Qwen3.8-27B**, served by llama.cpp.
- **Qwen3.6-35B-A3B**, served by llama.cpp.
- **Gemma 4 26B**, served by Ollama.

These model apps include a configured inference engine and download the model files during setup. Browse [Olares Market](https://market.olares.com) for the current selection, or follow [Run local LLMs](../use-cases/llm-base-apps.md) to get started.

### Can I install newer models or models that are not in Market?

Yes. Install new model apps as they become available in Market. To run a model from Hugging Face or the Ollama library, create an instance from an Engine Base app. Olares provides llama.cpp, Ollama, vLLM, and SGLang Engine Base apps.

Choose an engine that supports the model's architecture and file format, then set its memory and context limits for your device. The [local model guide](../use-cases/llm-base-apps.md) explains engine selection and configuration.

### How do GPU memory and system memory affect model choice?

Olares One has two separate memory pools: 24 GB of GPU memory (VRAM) and 96 GB of system memory (RAM). VRAM holds model weights and the context cache used during GPU inference. RAM runs the operating system and apps. With an engine that supports CPU and GPU inference, RAM can also hold the part of a larger model processed by the CPU.

Quantization reduces the memory needed for model weights by storing them at lower precision. Leave room for the context cache as well, which grows with longer conversations and more concurrent requests. See [Run local LLMs](../use-cases/llm-base-apps.md) for model and memory settings.

### Can I use coding agents with local models?

Yes. Install [OpenCode](../use-cases/opencode.md), [Claude Code](../use-cases/claude-code.md), or [Codex CLI](../use-cases/codex-cli.md) from Market and connect it to a local model through Router. The model generates responses and tool calls. The agent edits files and runs commands and tests in its development environment.

You can also run OpenCode CLI on your own computer and connect it to the model on Olares. In this setup, files and code execution stay on your computer, while model inference runs on Olares One. See [How AI apps connect to Olares models](../manual/best-practices/connect-ai-apps.md).

### Which languages does Olares One support?

Language support depends on what you are using:

- **Olares and LarePass interfaces**: English, Simplified Chinese, German, Spanish, Italian, French, and Japanese. See [Olares language settings](../manual/olares/settings/language-appearance.md) and [LarePass language settings](../manual/larepass/index.md#change-the-language).
- **Third-party apps**: Language options are provided by each app.
- **AI models**: Chat, speech recognition, and speech generation models each have their own language support. Select a model that supports the language you want to use.

### Can local AI models understand images?

Yes. A vision-capable model can analyze an uploaded image and answer questions about its contents. For example, Gemma 4 26B (Ollama) supports image input.

Connect the model to a chat app that accepts image attachments, then upload an image with your question. See [Run local LLMs](../use-cases/llm-base-apps.md) for model setup.

### Does Olares One support the Kimi K2 model?

Kimi K2 has around 1 trillion parameters and requires approximately 1024 GB of VRAM, making it unavailable in the Olares Market.

We anticipate that a Kimi-K2-level model might run on Olares One in the future as AI model sizes continue to shrink.

### Can local AI models access the internet for research?

Yes. An agent can call search and page-reading tools, then use a local model to analyze the results. Olares Market provides both types of tool:

- **[SearXNG](../use-cases/searxng.md)** searches the web and returns relevant links and snippets.
- **[Firecrawl](../use-cases/firecrawl.md)** reads web pages and turns their content into text the model can process.

Connect these tools to an agent to build a research workflow. For a setup using Lares and Router, see [Research the web with Lares](../use-cases/lares-web-research.md).

### Can I use a NAS for storage and use Olares for AI processing?

Yes, we highly recommend this workflow. You can mount your NAS to Olares One as an SMB share. Olares can then index, sort, and process the photos and videos stored on the NAS via apps like Immich.

### Can I use Olares AI features on my iPhone or Mac?

Yes. Open an AI app hosted on Olares One from your phone or computer's browser. Local model inference runs on Olares One, so your phone or computer only needs to display the app and send your requests.

For access away from home, keep Olares One powered on and connected to the internet, then enable [LarePass VPN](access-olares-via-vpn.md) on your phone or computer.

### Can I use LM Studio to manage models on Olares?

LM Studio functions as a client-side desktop application rather than a server-side service, so it cannot be installed directly on Olares One.

However, you have two alternatives:
* Use [LobeHub](../use-cases/lobechat.md) to chat with models hosted on Olares.
* Run a local instance of LM Studio on your computer and configure it to access the AI models hosted on Olares One via API.

### Does performance drop if I load multiple models simultaneously?

Loading multiple models that exceed GPU memory typically causes a crash. Olares addresses this with a time-slicing mechanism.

The system temporarily swaps models not actively in use to the 96GB system memory, and loads the active model into the GPU. This keeps multiple models ready with only about a 5% performance overhead during switching.

Alternatively, you can:
- Split the GPU memory so multiple apps run simultaneously without swapping.
- Assign the entire GPU to a single application for maximum raw performance.

Switch these modes directly from Olares Settings without additional modifications.

### Is the device capable of high-end gaming?

Yes. You can connect a monitor to play directly via Steam, or use the device as a game server to stream titles to your laptop, TV, or phone via Moonlight.

### Can I use Olares One for VR devices such as Quest 3?

We are actively testing compatibility with VR and AR devices, such as Quest 3, to include support in future updates.

### How does gaming performance compare to Windows?

Olares One supports gaming by streaming as a Steam server via Moonlight, or playing directly via a connected monitor.

In direct play scenarios, testing shows Cyberpunk 2077 achieves approximately 90% of the performance compared to the same hardware running Windows. This aligns with other high-end Linux gaming setups. We anticipate further compatibility improvements as SteamOS adoption grows.

## Operating system

### Can I run Windows on Olares One?

Yes. To run occasional Windows applications, install the Windows app from the Olares Market. This runs a virtual machine accessible via RDP, providing a seamless remote desktop experience when used with the built-in VPN.

Alternatively, you can install a second internal drive for a dual-boot setup or replace Olares OS with a native Windows installation.

See [Run Windows in a virtual machine](../use-cases/windows.md), [Set up dual boot with a second drive](dual-boot-dual-drive.md), or [Install Windows on Olares One](install-windows-primary-drive.md).

### Can the Windows VM use the GPU?

Currently, GPU passthrough is not supported. The NVIDIA RTX 5090 Mobile functions as the primary GPU for Olares OS to power local AI applications. Assigning it to a virtual machine would detach it from the host OS and disable these features.

Windows VMs utilize the integrated Intel graphics, which handle lightweight tasks effectively. For GPU-intensive workloads like gaming, we recommend setting up a dual-boot configuration.

### Can I wipe Olares OS and install Linux or Windows natively?

Yes. You have full ownership of the hardware. You can wipe the pre-installed OS to install Windows or any Linux distribution. In this configuration, Olares One functions as a standard high-performance workstation equipped with an NVIDIA RTX 5090 Mobile GPU, allowing you to use a monitor, keyboard, and mouse like a standard desktop computer.

### How to set up a dual-boot system?

We recommend installing a second NVMe SSD in the available slot to keep the operating systems on separate drives.

For a clean installation, install Windows first, followed by Olares OS.

Since Olares OS is based on Ubuntu 24.04, it follows standard Linux dual-boot procedures.

### Does the system support switching between Olares OS and Windows?

Yes. Both the UEFI Boot Manager and GRUB support this configuration. You can set Olares OS as the persistent default boot option in the BIOS. To boot into Windows for specific tasks like gaming, select the Windows Boot Manager during the startup sequence.

## Clustering

### Is it possible to connect two Olares One units together?

Yes. Olares OS is built on Kubernetes, which allows multiple Olares devices to form a cluster. The system automatically schedules applications within the cluster and loads models across multiple devices.

Forming a cluster currently requires command-line operations. We plan to introduce a fully UI‑based experience to manage this process in future updates.

### Does clustering make a single game or AI task run faster?

No. Clustering improves total system throughput and concurrency, but it does not double the speed of a single task.

For example, by clustering two units:
- You can run Elden Ring on one unit and Cyberpunk 2077 on another simultaneously. However, you cannot combine two units to run a single instance of Cyberpunk 2077 at double the frame rate.
- You can generate two images in 6 seconds, rather than generating a single image in 3 seconds.

### How does clustering benefit AI workloads?

Clustering supports larger models or complex workflows that a single device cannot handle.

For LLMs, the system uses vLLM pipeline parallelism to distribute the model across multiple units. This supports much larger models than a single device handles, though inference speed decreases compared to using multiple GPUs on a single motherboard due to network latency.

It also enables complex pipelines. For example, you can run a digital human application where the LLM runs on one node while the Text-to-Speech (TTS) and Automatic Speech Recognition (ASR) services run on another.

### Does clustering increase the available memory?

Yes. The total available memory capacity roughly doubles when connecting two identical units. The system distributes applications across nodes to maximize available resources.

### Can I cluster with third-party devices like Mac Studio or DGX Spark?

Currently, Olares supports clustering only for devices with the same architecture running the same OS. 

We plan to add clustering compatibility for third-party devices, such as Mac Studio, DGX Spark, and AI MAX 395+, in future updates.

### Can I use a NAS as part of the cluster?

No. Most NAS operating systems use closed environments and cannot run Olares nodes.

However, Olares One can mount NAS directories via the SMB protocol. This allows you to manage files on your NAS as if they were local folders on the Olares One.

### How does GPU scheduling work in a cluster?

Olares OS manages GPU allocation at the operating system level. GPU scheduling falls into four stages of complexity:
1. Single node, single GPU.
2. Single node, multiple GPUs.
3. Multiple nodes, multiple GPUs with the same architecture.
4. Multiple nodes, multiple GPUs with different architectures.

Olares currently operates at Stage 3, clustering multiple devices running the same OS and architecture.

Stage 4 remains partially automated and requires manual intervention, such as pulling the specific container images for the corresponding architecture. We plan to rewrite the scheduling algorithm to automate this support in future updates.

### Is federated learning or shared compute supported?

We are exploring advanced computing models, including federated learning, distributed AI training, and shared compute networks.

While the hardware easily handles these workloads, building a stable, production-ready solution is a long-term goal. Ultimately, these features rely on third-party applications rather than the core operating system.

### Why is the Ethernet port limited to 2.5Gbps if the device is meant for clustering?

We understand that 10Gbps is preferred for clustering to maximize data transfer between nodes. However, the inclusion of a 2.5Gbps Ethernet port is a strict limitation imposed by the hardware platform vendor.

## Purchasing & support

### Where can I check the price, shipping costs, and taxes?

See the [Olares One store page](https://shop.olares.com/products/olares-one) for current pricing, configurations, and promotions. Select your delivery destination and review the shipping and tax information at checkout. The [Shipping Policy](https://shop.olares.com/policies/shipping-policy) explains how duties and taxes are handled for different destinations.

For multi-unit purchases or help with a destination not listed at checkout, contact hi@olares.com.

### Where can I find warranty and return information?

The [Warranty Policy](https://shop.olares.com/pages/warranty-policy) explains coverage, repairs, replacements, and shipping arrangements. The [Return & Refund Policy](https://shop.olares.com/policies/refund-policy) explains return eligibility and regional terms.

For a hardware issue, contact hi@olares.com with your order number, device serial number, and a description of the problem. The support team can help diagnose the issue and provide return instructions if service is needed.

## Advanced usage and configuration

### Does Olares support multiple users with their own accounts?

Yes. Olares functions as a multi‑user system. You can create separate accounts for friends or family, and they can connect using their own LaresPass app.

The system supports three account roles:
* **Super Admin**: The user who initially activates the system and manages admins.
* **Admin**: Perform cluster‑level tasks.
* **Member**: Access shared services.

### Can I host a mail server on Olares One?

Hosting a mail server is technically possible but complex.

Olares supports open‑source solutions like Mail‑in‑a‑Box, Mailserver, and Mailcow. However, major providers often flag self-hosted email servers as spam, making reliable delivery difficult.

### Is it possible to install a VPN for outbound traffic?

Yes. Olares provides built-in support for Tailscale and Headscale. Configure and use a specific exit node for your outbound traffic.

### Do I need a static IP to host Olares One as a server?

No. Olares provides two external access options: reverse proxy and VPN. These solutions allow you to securely access your device from anywhere without a static IP.

### Can I use my own domain to access Olares One?

Yes. You can host Olares One using your own domain name.

Currently, you must point the NS records of your subdomain to Olares' name servers to set up the reverse proxy. 

We plan to introduce updates for users who wish to manage their own reverse proxy or do not require public internet access.

### Is there a way to access the device without the internet?

Yes. Olares provides a fully local access option using a `.local` domain.

If the device resides on a LAN but lacks public access, use the `.local` domain to access all features normally. Without internet connectivity, features relying on external services, such as the Olares Market, remain unavailable.

### Can Olares One function as a media server?

Yes. Olares One can function as a Plex server. It supports hardware-accelerated transcoding and utilizes both the CPU and GPU for efficient decoding.

### Can I use SMB to sync or back up files to a NAS?

Yes. Olares One supports SMB sharing, allowing you to sync files with or back up data to an external NAS.

### Is the file system encrypted?

By default, the system does not use full-disk encryption. We plan to make this a standard, user-configurable option in future updates.
