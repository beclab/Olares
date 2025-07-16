---
outline: [2, 3]
description: Compares Olares and general NAS, highlighting key differences in storage solutions, application ecosystem, virtual machine support, network configuration, and AI capabilities.
---


# Compare Olares and NAS 

Olares is dedicated to creating a one-stop personal cloud experience. Its core functionalities and user positioning are significantly different from traditional Network Attached System (NAS).

This document provides a detailed comparison between Olares and general NAS systems. We will highlight key differences in storage solutions, application ecosystem, virtual machine support, network configuration, and AI capabilities, using Unraid and Synology DSM as reference points.

## Overview

| Attribute         | Olares                                                                                                                                                                 | NAS                                                                                  |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Introduction         | Olares is an open-source <br/> personal cloud operating system.                                                                                                                                      | NAS systems are primarily used for network storage, allowing flexible hardware configurations to manage data and applications.                           |
| Positioning         | Focuses on helping users deploy<br/> and manage digital assets locally<br/> as an alternative to public<br/> cloud services.<br/>It supports running powerful<br/> open-source applications locally,<br/> providing users with robust<br/> cloud computing capabilities while<br/> ensuring complete data control<br/> and privacy. | Focuses on individuals and small businesses, solving low-cost data storage reliability issues, but has limitations in application hosting and network security. |
| Target users     | Regular users without<br/> a technical background                                                                                                                                                 | Geek users and small and medium business users                                                               |
| Transparency       | Open-source or source available                                                                                                                                                      | Usually closed-source                                                                             |
| Openness     | No vendor lock-in                                                                                                                                                             | Closed ecosystem                                                                             |
| Application security     | Supports application sandboxing,<br/> isolating network, storage,<br/> and computing resources. <br/>Permissions must be declared<br/> and user authorization obtained<br/> during installation.                                                                                          | No application sandboxing. Most third-party applications run with root privileges, and users bear the security risks themselves.               |
| Network security     | Provides public and private <br/> access, using reverse proxy<br/> and VPN to achieve service security.                                                                                                   | Only supports external access for limited system applications by default. Custom solutions by users may pose security risks.               |
| Developer friendliness | Provides development tools <br/>with familiar technology<br/> stacks to deploy services<br/> and develop applications.                                                                                                        | Does not support developer tools or application development.                                                         |
| AI           | 1. Advanced GPU management (v1.12)<br/>2. One-click installation support<br/> for over 30 AI applications<br/> and models<br/>3. Supports inter-application and<br/> model calls<br/>4. Supports MCP (v1.12)                                  | ❌                                                                                   |

## Project information

| Attribute | Olares                                      | Unraid                                                                     | Synology DSM                                                               |
| ---- | ------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| Year | 2022                                     | 2005                                                                    | 2000                                                                |
| Open Source | ✅                                          | ❌                                                                         | ❌                                                                     |
| Price | Free                                        | 30-day trial, then requires a one-time paid license                                        | OS sold bundled with hardware                                               |
| Positioning | Open-source personal <br/>cloud OS that focuses on <br/> helping users deploy and <br/>manage digital assets locally<br/> as an alternative to public<br/>  cloud services. | Network storage OS, allows flexible hardware configurations,<br/>simple and convenient management of data, VMs, and Docker applications | Provides secure and efficient data management systems for businesses of different scales,<br/>helping enterprises control growing data streams |

## Storage features

| Feature             | Olares                                                                     | Unraid                                            | Synology DSM                                                                |
| ---------------- | -------------------------------------------------------------------------- | ------------------------------------------------- | ----------------------------------------------------------------------- |
| Disk types         | System space, user space, <br/>application space, and<br/> application cache are on SSD                        | System runs on SSD, data stores on HDD                        |System runs on SSD, data stores on HDD                                               |
| Storage pool           | ❌                                                                     | Supports Parity-protected array similar to JBOD                  | Supports SHR, Basic, JBOD, RAID 0, RAID 1, RAID 5, RAID 6, RAID 10, and RAID F1 |
| LAN file sharing   | SMB                                                                        | SMB, NFS                                          | SMB, NFS, AFP, FTP                                                      |
| Public file sharing     | via Files application                                                         | ❌                                        | via File Station application                                                       |
| Distributed file system   | ✅                                                                       | ❌                                            | ❌                                                                  |
| Mount external cloud drives     | Supports mounting Google<br/> Drive, Dropbox, S3, etc.                             | ❌                                        | Supports mounting Google Drive, Dropbox, S3, etc.                               |
| Sync drive           | ✅ (Seafile integrated)                                                       | ❌                                        | Synology Drive                                                     |
| Mount SMB directory    | ✅                                                                       | ❌                                        | ❌                                                              |
| Mount mobile storage devices | Auto-mount                                                                   | Manual mount                                          | Auto-mount                                                                |
| Structured data support   | Supports mainstream databases <br/>and data warehouses (e.g., Redis, <br/>PostgreSQL),suitable for production <br/>environments | Can be manually installed,<br/>but not recommended for production environments                  | Can be manually installed,<br/>but not recommended for production environments                                        |
| Local data security     | No protection in single-node;<br/>disk data protection via Minio <br/>or Ceph in cluster mode    | Allows 1-2 disk failures depending on configuration,<br/>but Parity disk must not fail | Allows 1-2 disk failures depending on RAID configuration                                 |
| Remote backup         | Supports periodic incremental <br/> encrypted backups via Restic                                  | Flash drive can be manually backed up,<br/>but there is no official backup solution for data drives  | Supports multiple official backup solutions                                                    |

## Application management

| Feature             | Olares                                            | Unraid                                   | Synology DSM                       |
| ---------------- | ------------------------------------------------- | ---------------------------------------- | ------------------------------ |
| Installation format         | Olares package format based<br/> on improved Helm                    | Dockerfile or Docker Compose             | Dockerfile or Docker Compose   |
| App store         | Rich community app ecosystem<br/> with application sandboxing                            | Rich apps with no sandboxing restrictions.  | Fewer apps, extendable via third-party stores |
| Application sandbox         | ✅                                              | ❌                                   | ❌                         |
| Developer tools       | ✅ (Studio)                                    | ❌                                   | ❌                         |
| Middleware sharing       | Supports mainstream middleware like<br/>PostgreSQL, MongoDB, and Redis | ❌                                   | ❌                         |
| Cluster application support | ✅                                              | ❌                                   | ❌                         |
| LDAP integration        | ✅ (Requires third-party adaptation)                              | ❌                                   | ❌                         |
| Unified SSO login    | ✅                                              | ❌                                   | ❌                         |
| Secret management      | ✅                                              | ❌                                   | ❌                         |

## Virtual machine management

| Feature          | Olares                                                                | Unraid                         | Synology DSM      |
| ------------- | --------------------------------------------------------------------- | ------------------------------ | ------------- |
| Install via ISO | Not yet supported,<br/>will be supported via Kubevirt                                   | ✅                           | ✅          |
| Windows       | One-click install from<br/> app store, auto-integrates Tailscale,<br/>for secure external RDP access | Install via ISO                  | Install via ISO |
| Steam         | Supported in app store,<br/>auto-configures GPU mounting,<br/>streaming, and external access             | Requires manual configuration<br/>for GPU passthrough, etc. | ❌    |
| Linux         | Supported via Dev Containers<br/>in Studio                               | Install via ISO                  | Install via ISO |
| Android       | Supported via redroid                                                     | ❌                     | ❌    |
| Mac           | Coming soon                                                              | ❌                     | ❌    |
| Openwrt       | Coming soon                                                              | ❌                     | ❌    |

## Network access

| Feature                         | Olares                                                       | Unraid                                                                        | Synology DSM                                                                      |
| ---------------------------- | ------------------------------------------------------------ | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Reverse proxy                     | Integrated Cloudflare <br/>Tunnel and FRP, <br/> supports independent <br/> domain access for <br/>all apps | Only supports access to limited system services                                                        | Only supports access to limited system services                                                        |
| DDNS                         | No configuration <br/>needed                                                     | Manual configuration required.<br/>Supports domain+port access for non-system apps,<br/>but practically cannot provide external services. | Manual configuration required.<br/>Supports domain+port access for non-system apps,<br/>but practically cannot provide external services. |
| Custom domain support                 | ✅                                                         | ❌                                                                        | ❌                                                                        |
| Firewall                       | Integrated Cloudflare <br/>firewall                                       | ❌                                                                        | ❌                                                                        |
| Free HTTPS certificate              | ✅                                                         | ✅                                                                          | ✅                                                                          |
| Two-Factor Authentication Login                 | ✅                                                         | ✅                                                                          | ❌                                                                        |
| Different security policies for different directories | ✅                                                         | ✅                                                                          | ❌                                                                        |
| Private Access Endpoint                 | ✅                                                         | ✅                                                                          | ❌                                                                        |
| VPN connection                     | No configuration needed, <br/>Tailscale integrated                                     | Manual configuration required                                                                  | Manual configuration required                                                                  |
| VPN-only access mode              | ✅                                                         | ❌                                                                        | ❌                                                                        |

## AI capabilities

| Feature          | Olares                                                                                                                                                                            | Unraid                                | Synology DSM   |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- | ------ |
| GPU management | 1. Supports heterogeneous GPU clusters<br/>across multiple nodes (v1.12) <br/>2. Nvidia GPUs support memory slicing<br/>and time-slicing sharing modes (v1.12)                                                            | Supports GPU passthrough on a single node<br/>via manual configuration | ❌ |
| Models     | 1. Supports mainstream language, image,<br/>video, and voice models, such as Ollama,<br/>VLLM, ComfyUI, SD, Whisper, and ACE-STEP.<br/>2. Supports mainstream open-source AI tools,<br/> such as Dify, Ragflow, MaxKB, and LobeChat. | ❌                                | ❌ |
| Interoperability | 1. Supports mutual calls between<br/>applications and models<br/>2. System-level support for MCP calls (v1.12)                                                                                                            | ❌                                | ❌ |