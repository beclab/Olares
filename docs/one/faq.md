---
outline: [2,3]
description: Frequently asked questions regarding Olares One
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

### How loud is the device?

Olares One operates quietly even under load. In lab testing, the device generates 19dB when idle and remains under 39dB under maximum load.

### What is the power consumption?

The device consumes 30W in standby. Under load, the GPU power depends on your selected mode:
* **Silent Mode**: GPU 95W
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
* **Storage**: The motherboard includes two PCIe SSD slots, one PCIe 4.0 and one PCIe 5.0. The preinstalled 2TB SSD occupies the PCIe 4.0 slot. The second slot remains available for system storage expansion via LVM or for a dual-boot configuration.
* **Memory**: The RAM capacity can be upgraded to a maximum of 128GB.
* **External graphics & displays**: The Thunderbolt 5 port supports eGPU enclosures to connect external graphics cards. It also supports docking stations to connect up to two external monitors simultaneously at up to 8K resolution, with video output handled by the RTX 5090 Mobile.

### Is there a way to view the Olares OS UI via HDMI?

Currently, the HDMI output displays an Ubuntu shell or Steam Headless output for gaming, rather than a full desktop UI like a traditional PC. 

We plan to introduce the ability to run a lightweight browser or ChromeOS instance directly via HDMI in future updates.

### Can I update or configure the BIOS?

Yes. You can update the BIOS and Embedded Controller (EC) firmware as new versions release. While the system hides advanced BIOS options by default to maintain stability, you can also unlock them to perform deep hardware configurations according to your needs.

For detailed instructions, see [Manage BIOS](../one/update-firmware.md).

### Is Olares One designed for "always-on" operation?

Yes. Olares One functions as a personal AI cloud and reliably supports 24/7 continuous operation for scenarios like Large Language Model (LLM) hosting.

### Are there additional fees to use Olares One?

No. Olares OS and system updates are completely free.

### Can I use the device without an internet connection?

After a one-time activation, the device can work entirely offline. You can access it via your local network without an internet connection.

### Does the device prevent data loss during power outages?

Olares One uses the ext4 Linux file system to prevent file system corruption during sudden power loss. The power supply and motherboard design include voltage regulation for hardware protection.

Olares OS also includes a built-in automatic backup mechanism for periodic, encrypted backups to external locations.

We plan to enhance support for UPS devices to ensure graceful shutdowns during power interruptions, alongside native support for snapshot backups similar to Time Machine.

### Does the device support a GPU MUX switch?

Yes. Olares One supports both GPU MUX and Optimus modes. The system defaults to GPU MUX mode to enable a direct connection to the discrete GPU. Change this setting in the BIOS if needed.

## AI & gaming

### Does Olares One support the Kimi K2 model?

Kimi K2 has around 1 trillion parameters and requires approximately 1024 GB of VRAM, making it unavailable in the Olares Market.

We anticipate that a Kimi-K2-level model might run on Olares One in the future as AI model sizes continue to shrink.

### Can local AI models access the internet for research?

Yes. Olares supports several applications that enable internet access:
* **Vane** (formerly Perplexica): An open‑source alternative to Perplexity. It retrieves the latest information using SearxNG and analyzes it with a local LLM.
* **DeerFlow**: An open‑source alternative to OpenResearch. It uses RAGFlow to configure a local knowledge base, integrates Tavily for web search, and performs analysis using a local LLM.

### Can I use a NAS for storage and use Olares for AI processing?

Yes, we highly recommend this workflow. You can mount your NAS to Olares One as an SMB share. Olares can then index, sort, and process the photos and videos stored on the NAS via apps like Immich.

### Can I use Olares AI features on my iPhone or Mac?

Yes. Olares supports secure remote access. You can use your iPhone or Mac to query the AI models running on your Olares One from anywhere.

### How does the 120B model perform on Olares One?

We tested the `gpt-oss:120b` model. Since the model size exceeds the dedicated GPU memory, the system loads part of the model into the 96GB of system memory for CPU processing.

In our testing, `gpt-oss:120b` achieves approximately 36.16 tokens/s.

For better efficiency, we recommend `qwen3-30b-a3b`. It delivers superior results while remaining significantly smaller, reaching speeds of up to 157 tokens/s, or around 81 tokens/s with 8 concurrent requests.

For the detailed testing methodology, see [Local AI Hardware Performance Benchmarking](https://blog.olares.com/local-ai-hardware-performance-benchmarking/).

### Can I use LM Studio to manage models on Olares?

LM Studio functions as a client-side desktop application rather than a server-side service, so it cannot be installed directly on Olares One.

However, you have two alternatives:
* Use native apps, such as OpenWebUI or LobeHub (formerly LobeChat), which offer similar functionality.
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
