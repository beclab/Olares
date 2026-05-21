---
outline: [2, 3]
description: Find answers to common questions about the Olares platform.
---

# Olares FAQs

This page lists most frequently asked questions about Olares.

## General information

### What is Olares?

Olares is an open-source personal cloud operating system based on Kubernetes. It empowers you to own and manage your digital assets locally.

Olares provides native resource orchestration, application sandboxing, and production-grade infrastructure for edge computing. It acts as a one-stop personal cloud solution, running powerful local alternatives to public cloud services like large language models and automation workflows.

Use Olares for personal media servers, AI development, decentralized identity management, and more.

### What is "personal cloud"?

A personal cloud is a private infrastructure that replicates the utility of public cloud services, such as anywhere-access to files and computing power. It runs entirely on your own hardware to ensure data sovereignty.

### Who is Olares for?

Olares is for anyone who wants to use powerful AI tools locally without complex technical setups.

* **General users**: Deploy complex applications like ComfyUI or Vane (formerly Perplexica) from the Market with a single click.
* **Developers**: Use Olares as an efficient local development environment. Leverage the sandboxing and agent infrastructure to build and test applications directly on your Olares device, saving time on environment configuration.

### How is Olares different from NAS operating systems?

Olares functions fundamentally as a personal AI cloud rather than a storage server. Traditional NAS systems like Synology DSM or CasaOS are optimized primarily for storing files and hosting lightweight containers.

Olares focuses on high-performance computing:
* **Orchestrating resources**: It natively manages hardware resources such as GPUs to power local AI workloads.
* **Sandboxing**: It enforces strict application isolation, providing a security model that goes beyond standard file servers.

For detailed comparisons, refer to [Compare Olares and NAS](https://blog.olares.com/compare-olares-and-nas/).
<!-- #region faq-why-olares-id -->
### Why is an Olares ID required?

The Olares ID is currently required to automate secure remote access for your device. It allows the system to configure a reverse proxy, register a subdomain, and manage HTTPS certificates on your behalf. Without it, ccessing your device from outside your home requires manual handling of complex network configurations, such as port forwarding and DNS management.

You own your Olares ID entirely. Olares does not store your credentials or recover your data if you lose your mnemonic phrase.

We plan to introduce new activation options that make the Olares ID optional for users who prefer to configure their own network access.
<!-- #endregion faq-why-olares-id -->
### Is internet access required to use Olares?

Olares supports local-first usage, though initial activation currently requires internet access.

For strict local control, choose these options:
- **VPN-only mode**: Restrict your Olares device so it remains accessible remotely only via VPN.
- **Local-only access**: Access Olares services via .local domains even if the router has no internet access.

For detailed local access options, refer to [Access Olares services securely](../get-started/local-access.md).

We are actively developing an option to support full device activation in completely offline environments.

### What is LarePass and why is it required?

LarePass is the official client for Olares. It acts as a secure bridge to enable seamless access, file synchronization across devices, and more. Currently, device activation requires LarePass.

### Can I use Olares without the LarePass app?

We are working on decoupling functions to support usage without the LarePass app:

* **CLI activation**: Use the `olares-cli` tool to activate Olares directly from the terminal without using the LarePass app for the activation step. However, note that you must still use the LarePass app to create your Olares ID before running the `olares-cli` tool. For detailed instructions, refer to [Activate an Olares device using the Olares CLI](/manual/best-practices/activate-olares-using-cli.md).
* **Standalone components**: We plan to provide standalone deployment options for components like the Reverse Proxy, DID service, and Market repo in future updates.
<!-- #region custom-domain -->
### Can I use my own domain name?

Yes. You can use your own custom domain instead of the default `olares.com` domain. Setting this up currently requires the LarePass app.
<!-- #endregion custom-domain -->
For detailed instructions, refer to [Set up a custom domain for your Olares](../best-practices/set-custom-domain.md).

### Do I need to pay for Olares?

Olares OS is free and open source for self-hosting. Olares One is a one-time hardware purchase.

Olares offers two optional cloud-assisted services for convenience, but free alternatives are available so you are never locked in:
* **Cloud backup**: Subscribe to Olares Space for integrated cloud backups. Alternatively, back up to your own external storage or an S3-compatible service for free.
* **Remote access (FRP)**: Access devices remotely using the built-in Fast Reverse Proxy (FRP) service. This includes 2 GB of free monthly traffic, with paid options for higher usage. For a completely free alternative, access Olares services via LarePass VPN, or configure and use your own FRP server.

### How often does Olares update?

Olares releases a major update approximately every 2 months. For detailed changes, refer to the [Changelog](https://www.olares.com/changelog).

## License

### Is Olares open source?

Yes. The Olares OS software is open source, ensuring transparency and community collaboration. The project includes a family of repositories licensed under different models:

* **Olares and LarePass**: Licensed under AGPL-3.0. For details, refer to the [GitHub organization](https://github.com/beclab).
* **Protocol projects**: Projects like the Smart contract system for Olares ID use Apache 2.0.
* **Third-party apps**: Developers adopt their chosen licenses.

### Is it possible to build Olares from source code?

Yes, but the process is currently complex.

Olares is a massive project spanning over 90 repositories. With a quickly evolving architecture, Olares lacks a fully integrated local build system for a simple build experience.

We are streamlining the build process and documentation to improve the local build experience. We plan to release standalone deployment guides for core services, such as the reverse proxy, in future updates. The goal is to refine the foundation first, and then invite broader community collaboration.

## Security and privacy

### Does Olares collect my data?

No. Olares reclaims your data ownership. All storage, computation, and AI processing happen locally on your hardware. Olares does not collect or transmit your private data to any centralized service.

### Does Olares support backup?

Yes. Data safety is user-controlled and private. Olares includes a [built-in backup feature](../olares/settings/backup.md) that allows you to save specific file directories and set automatic schedules.

Every backup file is end-to-end encrypted. This allows you to store the backup file on any medium, including external drive or third-party cloud, with full confidence that the data remains inaccessible to others.

### What is app sandboxing?

Sandboxing is a security standard that prevents a single malicious app from compromising the entire system. In Olares, every app runs in its own secure, isolated environment. If an app malfunctions, it remains completely contained and fails to access or damage other applications or personal data.

### Does the system support multi-user environments?

Yes. Olares supports sub-accounts with built-in roles and permissions. The available roles are Super Admin, Admin, and Member.

This allows a team to access shared tools on a single server. For example, you can share files within the same Olares cluster or install a large AI model once for everyone to use.