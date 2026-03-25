---
outline: [2,3]
description: Frequently asked questions regarding Olares One
head:
  - - meta
    - name: keywords
      content: Olares One, Olares, personal cloud
---
# Olares One FAQs
This page lists frequently asked questions from the Olares community regarding Olares One and system capabilities.

:::tip Support for Olares One owners
If you require assistance with product features, hardware warranty, or a specific use case, contact technical support directly via WhatsApp or email at hi@olares.com. We respond to inquiries within two business days.
:::

## Product & spec

### What is Olares One?
Olares One is a dedicated personal cloud device designed for local AI. It integrates workstation-grade hardware with Olares OS, an open-source operating system that allows you to run AI agents and store data on hardware you physically control.

### What is personal cloud?
A personal cloud is a private infrastructure that replicates the utility of public cloud services such as anywhere-access to files and computing power but runs entirely on your own hardware to ensure data sovereignty.

### What are the technical specifications for Olares One?
* **Processor**: Intel Ultra 9 275HX
* **GPU**: NVIDIA RTX 5090 Mobile
* **Memory**: 96GB RAM
* **Storage**: 2TB NVMe SSD

For details, see [Olares One specifications](spec).

### How loud is the device?
Olares One is designed to operate quietly even under load. In lab testing, the device generates 19dB when idle and remains under 39dB under maximum load.

### What is the power consumption?
The device consumes 30W in standby. Under load, the GPU power depends on your selected mode:
- **Silent Mode**: GPU 95W
- **Performance Mode**: GPU 175W

When the CPU is fully utilized on its own, power consumption reaches 120W. During combined maximum load, the system balances power distribution with the CPU operating at 55W and the GPU at 175W.

The thermal design is engineered to sustain these performance levels.

### What ports are available?
The device includes the following physical ports:
* 1 × Thunderbolt 5 with 80 Gbps bandwidth
* 1 × RJ45 Ethernet at 2.5 Gbps
* 1 × USB-A 3.2 Gen2 with 10 Gbps speed
* 1 × HDMI 2.1 with 48 Gbps bandwidth

### How is Olares One different from DGX Spark, AI Max+ 395 and other AI PCs?
Olares One is a personal cloud solution rather than a traditional personal computer.

A personal cloud runs as a stable, external service accessible anytime from any device, whereas a PC is client-based software designed for direct interaction via monitor and keyboard.

Specifically, Olares One utilizes the NVIDIA CUDA stack on x86 architecture. This ensures broad compatibility with standard open-source AI applications, avoiding the adaptation challenges often found in ARM-based AI PCs or devices without CUDA support.

And Olares OS simplifies local AI with one-click deployment and workflow integration, distinct from the manual software management typical of standard PCs.

### Can I upgrade or expand the hardware?
Yes. You can modify the hardware through several internal slots and expansion ports:
* **Storage**: The motherboard includes two PCIe SSD slots (one PCIe 4.0 and one PCIe 5.0). The preinstalled 2TB SSD occupies the PCIe 4.0 slot, leaving the second slot available for system storage expansion via LVM or for a dual-boot configuration.
* **Memory**: The RAM capacity can be upgraded to a maximum of 128GB.
* **External graphics & displays**: The Thunderbolt 5 port supports eGPU enclosures for connecting external graphics cards. It also supports docking stations to connect up to two external monitors simultaneously (up to 8K resolution), with video output handled by the RTX 5090 Mobile.

### Is there a way to view the Olares OS UI via HDMI?
Currently, the HDMI output does not display a full desktop UI like a traditional PC. The HDMI port currently displays an Ubuntu shell or Steam Headless output for gaming.

Future updates target the ability to run a lightweight browser or ChromeOS instance directly via HDMI, allowing users to interact with the system using a monitor, keyboard, and mouse.

### Can I update or configure the BIOS?
Yes. Users can update the BIOS themselves as new versions are released. All BIOS options are fully unlocked, allowing users to configure settings according to their own needs.

### Is Olares One designed for "always-on" operation?
Yes. Olares One is positioned as a personal AI cloud and is capable of supporting 24/7 continuous operation reliably. It is designed to support scenarios like LLM hosting.

### Are there additional fees to use Olares One?
No. Olares OS and all future system updates are completely free.

### Can I use the device without an internet connection?
Yes. After a one-time activation, the device can work entirely offline. You can access it via your local network without any internet connection required.

### Does the device have failsafes for power outages?
Olares One uses the ext4 Linux file system to prevent file system corruption during sudden power loss. Hardware protection is managed through the power supply and motherboard design, which includes voltage regulation.

Olares OS also includes a built-in automatic backup mechanism for periodic, encrypted backups to external locations.

Looking ahead, we are working on the following improvements:

1. Enhanced support for UPS devices to ensure graceful shutdowns during power interruptions.
2. Native support for Time Machine-style snapshot backups.

### Does the device support a GPU MUX switch?
Yes. Olares One supports both GPU MUX and Optimus modes. The system is set to GPU MUX mode by default to enable a direct connection to the discrete GPU, but this can be changed in the BIOS.

## AI & gaming

### Can I run Kimi K2 on Olares One?
Kimi K2 has around 1 trillion parameters and requires approximately 1024 GB of VRAM, so we haven't made it available in the Olares Market yet.

However, given that AI model sizes have been shrinking roughly tenfold each year, there is a good chance that a Kimi‑K2–level model could run on Olares One by mid‑2026.

### Can local AI models access the internet for research?
Yes. We support several applications that enable this. For example:
* **Perplexica**: An open‑source alternative to Perplexity. It retrieves the latest information using SearxNG and analyzes it with a local LLM.
* **DeerFlow**: An open‑source alternative to OpenResearch. It uses RAGFlow to configure a local knowledge base, integrates Tavily for web search, and performs analysis using a local LLM.

### Can I use a NAS for storage and use Olares for AI processing?
Yes, we highly recommend this workflow. You can mount your NAS to Olares One as an SMB share. Olares can then index, sort, and process the photos and videos stored on the NAS via apps like Immich.

### Can I use Olares AI features on my iPhone or Mac?
Yes. Olares supports secure remote access. You can use your iPhone or Mac to query the AI models running on your Olares One from anywhere.

### How does the 120B model run on Olares One and how fast is it?
We tested the `gpt-oss:120b` model. Since the model size exceeds the dedicated GPU memory, part of the model is loaded into the 96GB of system memory and processed by the CPU.

In our testing, `gpt-oss:120b` achieves approximately 36.16 tokens/s.

For better efficiency, we recommend `qwen3-30b-a3b`. It delivers superior results while being significantly smaller, reaching speeds of up to 157 tokens/s, or around 81 tokens/s with 8 concurrent requests.

For the detailed testing methodology, Refer to https://blog.olares.com/local-ai-hardware-performance-benchmarking/.

### Can I use LM Studio to manage models on Olares?
LM Studio is designed as a client-side desktop application rather than a server-side service, so it cannot be installed directly on Olares One.

However, you have two alternatives:
* Use native apps like OpenWebUI or LobeChat, which offer similar functionality.
* Run a local instance of LM Studio on your computer and configure it to access the AI models hosted on Olares One via API.

### Will performance drop if I load multiple models at once?
Under normal circumstances, loading multiple models that exceed GPU memory would cause a crash. Olares addresses this with a time-slicing mechanism.

Models not actively in use are temporarily swapped to the 96GB system memory, while the active model is loaded into the GPU. This allows you to keep multiple models "ready" with only about a 5% performance overhead during switching.

Alternatively, you can:
- Split the GPU memory so multiple apps run simultaneously without swapping.
- Assign the entire GPU to a single application for maximum raw performance.

All these modes can be switched directly from Olares Settings. No other modifications are required.

### Is the device capable of high-end gaming?
Yes. You can connect a monitor to play directly via Steam, or use it as a game server to stream titles to your laptop, TV, or phone via Moonlight.

### Can I use Olares One for VR devices such as Quest 3?
We expect to begin broad testing with VR and AR devices in early 2026.

### How does gaming performance compare to Windows?
Olares One supports gaming in two ways: streaming as a Steam server via Moonlight, or playing directly by connecting a monitor.

In direct play scenarios, our data shows Cyberpunk 2077 achieving approximately 90% of the performance of the same hardware running Windows. This is comparable to other high-end Linux gaming setups. We expect compatibility to improve further as SteamOS adoption grows.

## Operating system

### Can I run Windows on Olares One?
Yes. For users who only need to run occasional Windows applications, we recommend installing the Windows app from the Olares Market. This runs a virtual machine accessible via RDP, which provides a seamless remote desktop experience when used with the built-in VPN.

Alternatively, you can install a second internal drive for a dual-boot setup or completely replace Olares OS with a native Windows installation.

### Can the Windows VM use the GPU?
Currently, GPU passthrough is not supported. The NVIDIA RTX 5090 Mobile functions as the primary GPU for Olares OS to power local AI applications. Assigning it to a virtual machine would detach it from the host OS, disabling these features.

Windows VMs utilize the integrated Intel graphics, which handle lightweight tasks effectively. For GPU-intensive workloads like gaming, we recommend setting up a dual-boot configuration.

### Can I wipe Olares OS and install Linux or Windows natively?
Yes. You have full ownership of the hardware. You can wipe the pre-installed OS to install Windows or any Linux distribution. In this configuration, Olares One functions as a standard high-performance workstation equipped with an NVIDIA RTX 5090 Mobile GPU, allowing you to use a monitor, keyboard, and mouse like a standard desktop computer.

### How do I set up a dual-boot system?
We recommend installing a second NVMe SSD in the available slot to keep the operating systems on separate drives.

For a clean installation, install Windows first, followed by Olares OS.

Since Olares OS is based on Ubuntu 24.04, it follows standard Linux dual-boot procedures.

### Can I set Olares OS as the default boot option but switch to Windows when needed?
Yes. Both the UEFI Boot Manager and GRUB support this configuration. You can set Olares OS as the persistent default boot option in the BIOS. To boot into Windows for specific tasks like gaming, you can select the Windows Boot Manager during the startup sequence.

## Clustering

### Can I connect two Olares One units together?
Yes. Olares OS is built on Kubernetes, which allows multiple Olares devices to form a cluster. Applications are automatically scheduled within the cluster, and models can be loaded across multiple devices.

Forming a cluster currently requires command-line operations. We expect to introduce a fully UI‑based experience to manage this process in future updates.

### Does clustering make a single game or AI task run faster?
No. Clustering improves total system throughput (concurrency), but it does not double the speed of a single task. For example, by clustering two units:
- You can run Elden Ring on one unit and Cyberpunk 2077 on another simultaneously. However, you cannot combine two units to run a single instance of Cyberpunk 2077 at double the frame rate.
- You can generate two images in 6 seconds, but cannot generate a single image in 3 seconds.

### How does clustering benefit AI workloads?
Clustering allows you to run larger models or complex workflows that a single device cannot handle.

For Large Language Models (LLMs), the system uses vLLM pipeline parallelism to distribute the model across multiple units. This allows you to run much larger models than a single device could support, though inference speed will be slower than using multiple GPUs on a single motherboard due to network latency.

It also enables complex pipelines. For example, you can run a digital human application where the LLM runs on one node while the Text-to-Speech (TTS) and Automatic Speech Recognition (ASR) services run on another.

### Does clustering increase the available memory?
Yes. The total available memory capacity roughly doubles when connecting two identical units. The system distributes different applications across nodes to maximize the available resources.

### Can I cluster with third-party devices like Mac Studio or DGX Spark?
Currently, Olares only supports clustering devices with the same architecture running the same OS.

Support for third-party devices is in development. We target initial clustering compatibility for Mac Studio, DGX Spark, and AI MAX 395+ in future updates.

### Can I use a NAS as part of the cluster?
No. Most NAS operating systems are closed environments and cannot run Olares nodes.

However, Olares One can mount NAS directories via the SMB protocol. This allows you to manage files on your NAS as if they were local folders on the Olares One.

### How does GPU scheduling work in a cluster?
Olares OS manages GPU allocation at the operating system level. Ideally, we categorize GPU scheduling into four stages of complexity:
1. Single node, single GPU.
2. Single node, multiple GPUs.
3. Multiple nodes, multiple GPUs (same architecture).
4. Multiple nodes, multiple GPUs (different architectures).

Olares currently operates at Stage 3. We can cluster multiple devices running the same OS and architecture (e.g., multiple Olares One units).

Stage 4, such as clustering Olares One and Mac devices together, is not yet fully automated. Currently, this setup requires manual intervention, such as pulling the specific container images for the corresponding architecture yourself. We are expect to rewrite our scheduling algorithm to automate this support in the near future.

### Is federated learning or shared compute supported?

We are exploring directions such as federated learning, distributed training of smaller LLMs, and providing services similar to Vast.ai.

While feasible given the hardware's VRAM and power profile, delivering a production-ready solution is a long-term goal that may take over a year. We view this primarily as the responsibility of application-layer protocols (like EXO) rather than the core OS kernel.

### Why is the Ethernet port limited to 2.5Gbps if the device is meant for clustering?
We understand that 10Gbps is preferred for clustering to maximize data transfer between nodes. However, the inclusion of a 2.5Gbps Ethernet port is a strict limitation imposed by the hardware platform vendor.

## Advanced usage and configuration
### Does Olares support multiple users with their own accounts?
Yes. Olares is a multi‑user system. You can create separate accounts for friends or family, and they can connect using their own LaresPass.

The system currently supports three account roles:
* **Super Admin**: The user who initially activates the system and manages admins.
* **Admin**: Can perform cluster‑level tasks.
* **Member**: Standard access to shared services.

### Can I host a mail server on Olares One?
Hosting a mail server is technically possible but complex.

We support open‑source solutions like Mail‑in‑a‑Box, Mailserver, and Mailcow. However, self‑hosted email servers are often flagged as spam by major providers, which makes reliable delivery difficult.

### Is it possible to install a VPN for outbound traffic?
We have built‑in support for Tailscale and Headscale. This allows you to configure and use a specific exit node for your outbound traffic.

### Do I need a static IP to host Olares One as a server?
No. A static IP address is not required. We provide two external access options: reverse proxy and VPN. These solutions allow you to securely access your device from anywhere without needing a static IP.

### Can I use my own domain to access Olares One?
Yes. You can host Olares One using your own domain name.

Currently, you must point the NS records of your subdomain to Olares' name servers so we can set up the reverse proxy for you. We are working on an update for users who wish to manage their own reverse proxy or do not require public internet access.

### Is there a way to access the device without the internet?
Yes. We provide a fully local access option using a `.local` domain.

If the device is on a LAN but blocked from public access, you can use the `.local` domain to access all features normally. If the device has zero internet connectivity, features relying on external services (like the App Market) will be unavailable.

### Can Olares One function as a media server?
Yes. Olares One can function as a Plex server. It supports hardware-accelerated transcoding, utilizing both the CPU and GPU for efficient decoding.

### Can I use SMB to sync or back up files to a NAS?
Yes. Olares One supports SMB sharing, allowing you to sync files with or back up data to an external NAS.

### Is the filesystem encrypted?
By default, the system does not use full-disk encryption. We are working to make this a standard, user-configurable option in the near future.