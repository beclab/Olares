

<h1 align="center">Olares</h1>

<p align="center"><strong>Your AI Agent. Your Data. Your Hardware.</strong></p>

<p align="center">
  <a href="https://olares.com">Website</a> ·
  <a href="https://www.olares.com/docs/">Docs</a> ·
  <a href="https://www.olares.com/market/">Olares Market</a> ·
  <a href="https://www.olares.com/blog/">Olares Blog</a>
</p>

<div align="center">

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/beclab/Olares)](https://github.com/beclab/olares/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/beclab/Olares?style=social)](https://github.com/beclab/Olares/stargazers)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/olares)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue)](https://github.com/beclab/olares/blob/main/LICENSE)

<a href="https://trendshift.io/repositories/15376" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15376" alt="beclab%2FOlares | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

<p>
  <a href="./README.md"><img alt="Readme in English" src="https://img.shields.io/badge/English-FFFFFF"></a>
  <a href="./README_CN.md"><img alt="Readme in Chinese" src="https://img.shields.io/badge/简体中文-FFFFFF"></a>
  <a href="./README_JP.md"><img alt="Readme in Japanese" src="https://img.shields.io/badge/日本語-FFFFFF"></a>
</p>

</div>


**Olares is an open-source personal cloud operating system you operate in plain language, built to run AI agents and LLMs on hardware you own.** 

Powered by Kubernetes, it turns your machines into a self-hosted AI platform accessible from any browser, giving everyone from individual users to small teams a unified place for compute, storage, networking, and apps.

> 🌟 *Star us to receive instant notifications about new releases and updates.*


## Why Olares

Great AI needs to know you. That requires access to your files, messages, and history. Cloud services store this sensitive data on third-party servers and charge based on usage.

Olares brings AI home, so you can run agents like OpenClaw locally, while still enjoying the access and convenience of the cloud.

![Personal Cloud](https://app.cdn.olares.com/github/olares/public-cloud-to-personal-cloud.jpg)

Features include:

- **One-click local AI:** Deploy top open-source AI apps and models from Olares Market.
- **Accelerated computing management:** Pool GPUs and other accelerators across nodes, with time-slicing, memory-slicing, and exclusive GPU modes for AI, media, and gaming workloads.
- **Unified file system and database:** A managed storage layer and built-in file manager with automated scaling, backups, and high availability, spanning local drives, sync, cloud, and external SMB/NFS mounts.
- **Enterprise-grade network:** A private VPN, reverse proxy, and public, private, and internal access controls give every app its own domain with HTTPS, no port forwarding required.
- **Anytime, anywhere access:** A single Olares ID and the LarePass client reach all your services from any phone, desktop, or browser.
- **A suite of system apps:** Files, Vault, Market, Dashboard, Control Hub, and more, ready the moment you log in.

## Key use cases

- **Run AI agents that operate your cloud.** Hand tasks to a personal agent in plain language, from coding to everyday automation, and let it act on your behalf.
- **Host local LLMs and generative AI.** Chat, generate images and video, and serve open models locally through the Model Console (llama.cpp, Ollama, vLLM, SGLang), keeping data and inference on your own hardware.
- **Build a smart home and media hub.** Make Olares the local brain for your IoT devices, home automation, and personal media collection.
- **Develop and host agentic apps.** Build, test, and run agentic apps and workflows in a secure, sandboxed environment.
- **Keep audio private.** Run meeting transcription, translation, and text-to-speech locally, without handing your recordings to a third party.
- **Collaborate on your own terms.** Stand up a free, self-hosted workspace for your team with open-source alternatives to costly SaaS.
- **Own your data repository.** Store, sync, and manage your files, photos, and documents across every device and location.

## Get started

### Requirements
Olares installs on a Linux host (bare metal or VM), with dedicated install paths for Windows, macOS, and Raspberry Pi. For a smooth experience we recommend:
- **CPU:** 4 cores or more
- **RAM:** 8 GB or more of available memory
- **Storage:** 150 GB or more of available SSD storage (an SSD is required; installation fails on an HDD)
- **OS (script install):** Ubuntu 22.04 to 25.04, or Debian 12 or 13
- **GPU (optional):** a dedicated GPU unlocks local AI acceleration

### Install and activate
1. Create your Olares ID in [LarePass](https://www.olares.com/larepass), the client app that adds secure login, a built-in VPN, and file sync.

2. On your Linux host, run:

    ```bash
    curl -fsSL https://olares.sh | bash -
    ```

    For Windows, macOS, Raspberry Pi, or a VM, choose your platform in the [installation guide](https://www.olares.com/docs/manual/get-started/install-olares).

3. Follow the guided web wizard, or do it entirely from the terminal with the [Activate using the Olares CLI](https://docs.olares.com/manual/best-practices/activate-olares-using-cli) tutorial.

Once activated, you can reach Olares from any browser at your Olares ID, in the format like `https://desktop.marvin123.olares.com`.


## System architecture

Just as public clouds offer IaaS, PaaS, and SaaS layers, Olares provides open-source alternatives to each of these layers.

  ![Tech Stacks](https://app.cdn.olares.com/github/olares/olares-architecture.jpg)

 For a detailed description of each component, refer to [Olares architecture](https://www.olares.com/docs/developer/concepts/system-architecture).

> 🔍 **How is Olares different from traditional NAS?**
>
> Olares focuses on building an all-in-one self-hosted personal cloud experience. Its core features and target users differ significantly from traditional Network Attached Storage (NAS) systems, which primarily focus on network storage. For more details, see [Compare Olares and NAS](https://www.olares.com/blog/compare-olares-and-nas/).

### What's next

- **[Explore use cases](https://docs.olares.com/use-cases/):** local LLMs, AI image generation, media servers, and more.
- **[Try the built-in apps](https://docs.olares.com/manual/olares/):** Files, Vault, Market, Dashboard, Control Hub, and others.
- **[Operate Olares in plain language](https://docs.olares.com/developer/cli-agent-skills):** drive the system with `olares-cli` and its Agent Skills.
- **[Join the community](https://discord.gg/olares):** get help and share your setup on Discord.

## Project navigation

The main directories in the Olares repository:

```
Olares/
├── apps/            # Built-in Olares system applications
├── cli/             # olares-cli: agent-native CLI with Agent Skills to install and operate Olares
├── daemon/          # olaresd, the system daemon process
├── docs/            # Project documentation
├── framework/       # Olares system services
├── infrastructure/  # Computing, storage, networking, and GPU components
├── platform/        # Cloud-native components like databases and message queues
└── vendor/          # Hardware-specific code for Olares devices
```

## Contributing

Olares is open source (AGPL-3.0) and built in the open. There are many ways to get involved:

- **Submit apps to Olares Market.** [Package and distribute](https://www.olares.com/docs/developer/develop/distribute-index) your app so anyone can install it in one click.
- **Report bugs and request features.** Open a [GitHub issue](https://github.com/beclab/olares/issues) so we can track it.


## What's Next

### Discover Olares
| Section | What's covered |
|---------|----------------|
| [Installation Guide](https://www.olares.com/docs/manual/get-started/install-olares) | Set up Olares on your device: system requirements, install paths, and activation. |
| [CLI Guide](https://www.olares.com/docs/developer/install/cli/olares-cli) | Usage and reference for `olares-cli`: install, manage, diagnose, and operate from the command line. |
| [Agent Skills](https://www.olares.com/docs/developer/cli-agent-skills) | Let AI agents operate Olares in plain language through the `olares-cli` Agent Skills. |
| [Use Cases](https://www.olares.com/docs/use-cases/) | Practical scenarios: local LLM serving, AI image generation, and media servers. |
| [Advanced Tutorials](https://www.olares.com/docs/manual/best-practices/) | Go further: CLI activation, GPU passthrough, multi-node setup, custom domains, and storage expansion. |

## Connect with Us

- [Discord](https://discord.gg/olares): Join the server for technical support, deployment discussions, and sharing agent workflows.

- [Olares Forum](https://www.olares.com/forum/): Share product design feedback, discuss user experience, and submit feature requests directly to our team.

Alternatively, follow us on social media so you don't miss a beat:

- X (Twitter)
- YouTube
- LinkedIn

## Special thanks

The Olares project has incorporated numerous third-party open source projects, including: [Kubernetes](https://kubernetes.io/), [Kubesphere](https://github.com/kubesphere/kubesphere), [Padloc](https://padloc.app/), [K3S](https://k3s.io/), [JuiceFS](https://github.com/juicedata/juicefs), [MinIO](https://github.com/minio/minio), [Envoy](https://github.com/envoyproxy/envoy), [Authelia](https://github.com/authelia/authelia), [Infisical](https://github.com/Infisical/infisical), [Dify](https://github.com/langgenius/dify), [Seafile](https://github.com/haiwen/seafile),[HeadScale](https://headscale.net/), [tailscale](https://tailscale.com/), [Redis Operator](https://github.com/spotahome/redis-operator), [Nitro](https://nitro.jan.ai/), [RssHub](http://rsshub.app/), [predixy](https://github.com/joyieldInc/predixy), [nvshare](https://github.com/grgalex/nvshare), [LangChain](https://www.langchain.com/), [Quasar](https://quasar.dev/), [TrustWallet](https://trustwallet.com/), [Restic](https://restic.net/), [ZincSearch](https://zincsearch-docs.zinc.dev/), [filebrowser](https://filebrowser.org/), [lego](https://go-acme.github.io/lego/), [Velero](https://velero.io/), [s3rver](https://github.com/jamhall/s3rver), [Citusdata](https://www.citusdata.com/).
