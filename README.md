<div align="center">

# Olares: An Open-Source Personal Cloud to </br>Reclaim Your Data<!-- omit in toc -->

[![Mission](https://img.shields.io/badge/Mission-Let%20people%20own%20their%20data%20again-purple)](#)<br/>
[![Last Commit](https://img.shields.io/github/last-commit/beclab/olares)](https://github.com/beclab/olares/commits/main)
![Build Status](https://github.com/beclab/olares/actions/workflows/release-daily.yaml/badge.svg)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/beclab/olares)](https://github.com/beclab/olares/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/beclab/olares?style=social)](https://github.com/beclab/olares/stargazers)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/olares)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue)](https://github.com/beclab/olares/blob/main/LICENSE)

<a href="https://trendshift.io/repositories/15376" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15376" alt="beclab%2FOlares | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

<p>
  <a href="./README.md"><img alt="Readme in English" src="https://img.shields.io/badge/English-FFFFFF"></a>
  <a href="./README_CN.md"><img alt="Readme in Chinese" src="https://img.shields.io/badge/简体中文-FFFFFF"></a>
  <a href="./README_JP.md"><img alt="Readme in Japanese" src="https://img.shields.io/badge/日本語-FFFFFF"></a>
</p>

</div>

<p align="center">
  <a href="https://olares.com">Website</a> ·
  <a href="https://docs.olares.com">Documentation</a> ·
  <a href="https://www.olares.com/larepass">Download LarePass</a> ·
  <a href="https://github.com/beclab/apps">Olares Apps</a> ·
  <a href="https://space.olares.com">Olares Space</a>
</p>

>*The modern internet built on public clouds is increasingly threatening your personal data privacy. As reliance on services like ChatGPT, Midjourney, and Facebook grows, so does the risk to your digital autonomy. Your data lives on their servers, subject to their terms, tracking, and potential censorship.*
>
>*It's time for a change.* 

![Personal Cloud](https://app.cdn.olares.com/github/olares/public-cloud-to-personal-cloud.jpg)
We believe you have a fundamental right to control your digital life. The most effective way to uphold this right is by hosting your data locally, on your own hardware.

Olares is an **open-source personal cloud operating system** designed to empower you to own and manage your digital assets locally. Instead of relying on public cloud services, you can deploy powerful open-source alternatives locally on Olares, such as Ollama for hosting LLMs, ComfyUI for image generation, and Perplexica for private, AI-driven search and reasoning. Imagine the power of the cloud, but with you in complete command.

> 🌟 *Star us to receive instant notifications about new releases and updates.* 

## Architecture 

Just as Public clouds offer IaaS, PaaS, and SaaS layers, Olares provides open-source alternatives to each of these layers.

  ![Tech Stacks](https://app.cdn.olares.com/github/olares/olares-architecture.jpg)

 For detailed description of each component, refer to [Olares architecture](https://docs.olares.com/manual/concepts/system-architecture.html).

> 🔍 **How is Olares different from traditional NAS?**
>
> Olares focuses on building an all-in-one self-hosted personal cloud experience. Its core features and target users differ significantly from traditional Network Attached Storage (NAS) systems, which primarily focus on network storage. For more details, see [Compare Olares and NAS](https://docs.olares.com/manual/olares-vs-nas.html).

## Features

Olares offers a wide array of features designed to enhance security, ease of use, and development flexibility:

- **Enterprise-grade security**: Simplified network configuration using Tailscale, Headscale, Cloudflare Tunnel, and FRP.
- **Secure and permissionless application ecosystem**: Sandboxing ensures application isolation and security.
- **Unified file system and database**: Automated scaling, backups, and high availability.
- **Single sign-on**: Log in once to access all applications within Olares with a shared authentication service.
- **AI capabilities**: Comprehensive solution for GPU management, local AI model hosting, and private knowledge bases while maintaining data privacy.
- **Built-in applications**: Includes file manager, sync drive, vault, reader, app market, settings, and dashboard.
- **Seamless anywhere access**: Access your devices from anywhere using dedicated clients for mobile, desktop, and browsers.
- **Development tools**: Comprehensive development tools for effortless application development and porting.

Here are some screenshots from the UI for a sneak peek:

| **Desktop–Streamlined and familiar portal**     |  **Files–A secure home to your data**
| :--------: | :-------: |
| ![Desktop](https://app.cdn.olares.com/github/terminus/v2/desktop.jpg) | ![Files](https://app.cdn.olares.com/github/terminus/v2/files.jpg) |
| **Vault–1Password alternative**|**Market–App ecosystem in your control** |
| ![vault](https://app.cdn.olares.com/github/terminus/v2/vault.jpg) | ![market](https://app.cdn.olares.com/github/terminus/v2/market.jpg) |
|**Wise–Your digital secret garden** | **Settings–Manage Olares efficiently** |
| ![settings](https://app.cdn.olares.com/github/terminus/v2/wise.jpg) | ![](https://app.cdn.olares.com/github/terminus/v2/settings.jpg) |
|**Dashboard–Constant system monitoring**  | **Profile–Your unique homepage** |
| ![dashboard](https://app.cdn.olares.com/github/terminus/v2/dashboard.jpg) | ![profile](https://app.cdn.olares.com/github/terminus/v2/profile.jpg) |
| **Studio–Develop, debug, and deploy**|**Control Hub–Manage Kubernetes clusters easily**  |
| ![Studio](https://app.cdn.olares.com/github/terminus/v2/devbox.jpg) | ![Controlhub](https://app.cdn.olares.com/github/terminus/v2/controlhub.jpg)|


## Key use cases

Here is why and where you can count on Olares for private, powerful, and secure sovereign cloud experience:

🤖 **Edge AI**: Run cutting-edge open AI models locally, including large language models, computer vision, and speech recognition. Create private AI services tailored to your data for enhanced functionality and privacy. <br>

📊 **Personal data repository**: Securely store, sync, and manage your important files, photos, and documents across devices and locations.<br>

🚀 **Self-hosted workspace**: Build a free collaborative workspace for your team using secure, open-source SaaS alternatives.<br>

🎥 **Private media server**: Host your own streaming services with your personal media collections. <br>

🏡 **Smart Home Hub**: Create a central control point for your IoT devices and home automation. <br>

🤝 **User-owned decentralized social media**: Easily install decentralized social media apps such as Mastodon, Ghost, and WordPress on Olares, allowing you to build a personal brand without the risk of being banned or paying platform commissions.<br>

📚 **Learning platform**: Explore self-hosting, container orchestration, and cloud technologies hands-on.

## Getting started

### System compatibility

Olares has been tested and verified on the following Linux platforms:

- Ubuntu 24.04 LTS or later
- Debian 11 or later

### Set up Olares
To get started with Olares on your own device, follow the [Getting Started Guide](https://docs.olares.com/manual/get-started/) for step-by-step instructions.

## Project navigation
This section lists the main directories in the Olares repository:

* **[`apps`](./apps)**: Contains the code for system applications, primarily for `larepass`.
* **[`cli`](./cli)**: Contains the code for `olares-cli`, the command-line interface tool for Olares.
* **[`daemon`](./daemon)**: Contains the code for `olaresd`, the system daemon process.
* **[`docs`](./docs)**: Contains documentation for the project.
* **[`framework`](./framework)**: Contains the Olares system services.
* **[`infrastructure`](./infrastructure)**: Contains code related to infrastructure components such as computing, storage, networking, and GPUs.
* **[`platform`](./platform)**: Contains code for cloud-native components like databases and message queues.
* **`vendor`**: Contains code from third-party hardware vendors.

## Contributing to Olares

We are welcoming contributions in any form:

- If you want to develop your own applications on Olares, refer to:<br>
https://docs.olares.com/developer/develop/


- If you want to help improve Olares, refer to:<br>
https://docs.olares.com/developer/contribute/olares.html

## Community & contact

* [**GitHub Discussion**](https://github.com/beclab/olares/discussions). Best for sharing feedback and asking questions.
* [**GitHub Issues**](https://github.com/beclab/olares/issues). Best for filing bugs you encounter using Olares and submitting feature proposals. 
* [**Discord**](https://discord.gg/olares). Best for sharing anything Olares.

## Special thanks

The Olares project has incorporated numerous third-party open source projects, including: [Kubernetes](https://kubernetes.io/), [Kubesphere](https://github.com/kubesphere/kubesphere), [Padloc](https://padloc.app/), [K3S](https://k3s.io/), [JuiceFS](https://github.com/juicedata/juicefs), [MinIO](https://github.com/minio/minio), [Envoy](https://github.com/envoyproxy/envoy), [Authelia](https://github.com/authelia/authelia), [Infisical](https://github.com/Infisical/infisical), [Dify](https://github.com/langgenius/dify), [Seafile](https://github.com/haiwen/seafile),[HeadScale](https://headscale.net/), [tailscale](https://tailscale.com/), [Redis Operator](https://github.com/spotahome/redis-operator), [Nitro](https://nitro.jan.ai/), [RssHub](http://rsshub.app/), [predixy](https://github.com/joyieldInc/predixy), [nvshare](https://github.com/grgalex/nvshare), [LangChain](https://www.langchain.com/), [Quasar](https://quasar.dev/), [TrustWallet](https://trustwallet.com/), [Restic](https://restic.net/), [ZincSearch](https://zincsearch-docs.zinc.dev/), [filebrowser](https://filebrowser.org/), [lego](https://go-acme.github.io/lego/), [Velero](https://velero.io/), [s3rver](https://github.com/jamhall/s3rver), [Citusdata](https://www.citusdata.com/).
