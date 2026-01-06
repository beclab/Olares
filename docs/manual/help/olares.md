---
outline: [2, 3]
description: Find answers to common questions about the Olares platform.
---

# Olares FAQs

This page lists most frequently asked questions about Olares.

## General information

### What is Olares?
Olares is an open-source personal cloud operating system based on Kubernetes designed to empower users to own and manage their digital assets locally. It features native resource orchestration, application sandboxing, and production-grade infrastructure for edge computing. The goal of Olares is to provide a one-stop private cloud solution that runs powerful local alternatives to public cloud services such as large language models and automation workflows. It is suitable for use cases ranging from personal media servers and AI development to decentralized identity management.

### What is "personal cloud"?

A personal cloud is a private infrastructure that replicates the utility of public cloud services such as anywhere-access to files and computing power but runs entirely on your own hardware to ensure data sovereignty.

You access Olares services exactly like a public cloud via a web browser or the LarePass client, utilizing built-in encrypted tunnels to ensure secure connectivity from anywhere without complex network configurations.

### Who is Olares for?
Olares is designed for anyone who wants to use powerful AI tools locally without dealing with complex technical setups.

* **For general users:** You can deploy complex applications like ComfyUI or Perplexica from the Market with a single click.
* **For developers:** Olares functions as an efficient local development environment. You can leverage the sandboxing and agent infrastructure to build and test applications directly on your Olares device, saving time on environment configuration.

### How is Olares different from NAS operating systems?
While traditional NAS operating systems like Synology DSM or CasaOS focus primarily on storage and lightweight container hosting, Olares is designed to power a personal AI cloud.

Olares goes beyond standard NAS capabilities by:
- **Orchestrating resources:** It natively manages multiple hardware resources such as GPUs for AI tasks.
- **Sandboxing:** It manages applications in strict isolation for security.

For detailed comparisons, refer to [Compare Olares and NAS](https://blog.olares.com/compare-olares-and-nas/).

### Why is an Olares ID required?
The Olares ID is designed to allow non-technical users to access their devices easily from anywhere without complex network configurations.

Unlike a cloud account, the Olares ID is a decentralized identifier (DID) that you alone control. We never see your credentials, and we cannot recover your data if you lose your mnemonic phrase.

### Can I use Olares offline or without internet?
Yes, we support local-first usage, though the initial activation currently requires internet access.

For users prioritizing strict local control, we offer these options:
* **VPN-Only mode:** You can restrict your Olares so it is only accessible remotely via VPN.
* **Local-Only access:** You can access Olares services via `.local` domains even if the router has no internet access.

For detailed local access options, refer to [Access Olares services locally](../get-started/local-access.md).

Note that we are also working on an option to allow full device activation in a completely offline environment.

### What is LarePass and why is it required?
LarePass is the official client for Olares. It acts as a secure bridge to enable seamless access, file synchronization across devices, etc. Currently, it is required to handle the device activation.

### Can I use Olares without the LarePass app?
We understand this is a core requirement for advanced and enterprise users. We are working on decoupling these functions:

* **CLI activation:** We plan to move activation logic into the `olares-cli`, allowing for a terminal-based setup without the app.
* **Standalone components:** We aim to provide standalone deployment options for components like the Reverse Proxy, DID service, and Market repo in future updates.

### Can I use my own domain name?
Yes. You can use your own custom domain instead of the default `olares.com` domain. Note that setting this up currently requires the LarePass app.

For details, refer to [Set up a custom domain for your Olares](../best-practices/set-custom-domain.md).

### Do I need to pay for Olares?
Olares OS itself is free and open source for self-hosting. If you purchase Olares One, it is a one-time hardware cost.

We offer two optional cloud-assisted services for convenience, but free alternatives are available so you are never locked in:
* **Cloud backup:** You can subscribe to Olares Space for integrated cloud backups. The free alternative is to back up to your own external storage or S3-compatible service.
* **Remote access (FRP):** For easy remote access, we offer a built-in FRP (Fast Reverse Proxy) service with 2 GB of free monthly traffic, with paid options for higher usage. The completely free alternative is to access Olares services via LarePass VPN, or to configure and use your own FRP server.

### How often does Olares update?
We aim for a major release approximately every 2 months. You can view specific changes in our changelog.

## License
### Is Olares open source?
Yes. The Olares OS software is open source, ensuring transparency and community collaboration. The project consists of a family of repositories licensed under appropriate models:

* **Olares and LarePass:** Licensed under AGPL-3.0. You can view our [GitHub organization](https://github.com/beclab).
* **Protocol projects:** Projects like the Smart contract system for Olares ID use Apache 2.0.
* **Third-party apps:** Developers adopt any license they choose.


### Can I build Olares from source code?
The short answer is yes, but it is currently complex.

Olares is a massive project spanning over 90 repositories. Because our architecture is evolving quickly, we currently lack a fully integrated local build system that provides a simple "what you see is what you get" experience.

We are actively working to streamline the build process and documentation. We expect to improve the local build experience and release standalone deployment guides for core services such as reverse proxy in 2026. Our goal is to refine the foundation first, then invite broader community collaboration.

## Security and privacy

### Does Olares collect my data?
No. Olares is built to reclaim your data ownership. All storage, computation, and AI processing happen locally on your hardware. Olares does not collect or transmit your private data to any centralized service.

### Does Olares support backup?
Yes. Data safety is user-controlled and private. Olares includes a [built-in backup feature](../olares/settings/backup.md) that allows you to save specific file directories and set automatic schedules.

Critically, every backup file is end-to-end encrypted. This allows you to store the backup file on any medium including external drive or third-party cloud with full confidence that the data remains inaccessible to others.

### What is app sandboxing?
Sandboxing is a security standard used to prevent a single malicious app from compromising the entire system. In Olares, every app runs in its own secure, isolated environment. If an app malfunctions, it is completely contained and cannot access or damage your other applications or personal data.

### Does the system support multi-user environments?
Yes. Olares supports sub-accounts with a built-in roles and permissions system including Super Admin, Admin, and Member.

This allows a team to access shared tools on a single server. For example, you can share files within the same Olares cluster or install a large AI model once for everyone to use.