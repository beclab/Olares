---
outline: [2, 3]
description: Comprehensive guide to Olares architecture covering infrastructure, platform services, and application framework. Learn about container orchestration, storage, networking, and system components.
---

# Olares architecture

This document provides a comprehensive explanation of the Olares architecture, outlining the purpose and functionality of each layer and its components.

![Olares architecture diagram](/images/manual/architecture-diagram.png)

## Infrastructure

The infrastructure layer provides essential infrastructure services such as container orchestration, storage, networking, and cluster management.

### Container orchestration

Olares supports different Kubernetes distributions depending on the underlying environment:
- Linux environments (including WSL, PVE, LXC, Raspberry Pi): Users can choose to install [Kubernetes](https://kubernetes.io/) or the lightweight [K3s](https://k3s.io/), with K3s being the default for its better performance and resource efficiency on local hardware.
- macOS: [minikube](https://minikube.sigs.k8s.io/) is used to deploy Kubernetes within a Linux virtual machine, ensuring a unified experience across platforms.

Regardless of the chosen Kubernetes distribution, users get consistent core capabilities and experience with Olares.

### Networking

The networking stack ensures seamless communication between containers, nodes, and services. Key components include:

- [CoreDNS](https://coredns.io/): Provides DNS services for the cluster, ensuring efficient name resolution.
- [Calico](https://www.tigera.io/project-calico/): A container networking interface (CNI) that facilitates communication between containers and virtual machines while offering advanced network policy controls.
- [Envoy](https://www.envoyproxy.io/): A high-performance, extensible edge and service proxy. Envoy acts as middleware for communication between services, handling load balancing, service discovery, secure communication, and observability. It can operate as a standalone reverse proxy or an API gateway and is often used as a data plane component in service meshes.

These components collectively ensure robust, scalable, and secure networking within Olares.

### Distributed storage

Olares provides flexibility in storage solutions tailored to both single-node and multi-node setups:

- Local storage (default): Ideal for single-node deployments, offering the best read/write performance.
- [S3](https://aws.amazon.com/s3/): A cloud-based storage option. Ideal for cloud deployment via S3 or any S3-compatible service.
- [MinIO](https://min.io/): A distributed storage solution for self-hosted deployment. Users can either set up a MinIO cluster through Olares or mount an existing one.

This approach ensures that applications have access to the necessary storage mechanisms, whether it's for local or distributed environments.

### Distributed key-value storage

Olares uses [etcd](https://etcd.io/) as its distributed key-value store. etcd is integral for storing and managing all cluster data for Kubernetes.

### GPU management

Olares leverages components like CUDA driver, NVIDIA device plugin, and nvShare, which work in conjunction to manage and provision GPU resources effectively: 

- CUDA: Acts as the core interface between the GPU hardware and the operating system.
- NVIDIA device plugin: Allows GPU resources to be advertised, scheduled, and allocated to containers or pods.
- [nvshare](https://github.com/grgalex/nvshare): Allows multiple containers or pods to share a single GPU, enabling both shared and exclusive GPU usage in Olares for better GPU utilization.

:::info
Currently, Olares GPU support is restricted to deployments with one GPU per node.
:::
Starting with Olares v1.11, [CUDA](https://developer.nvidia.com/cuda-toolkit) (12.4 and above) is supported. Changes in the host environment's CUDA configuration can be synchronized with the Olares cluster using `olares-cli`.

### Container management
Olares uses [containerd](../developer/install/installation-overview.md#container-runtime-containerd), a lightweight container runtime, for containerized deployments.

### Olares Controller Panel

The management of Olares is implemented through the following:

- [olares-cli](../developer/install/cli/olares-cli.md): A command-line tool for managing Olares clusters, applications, and hardware nodes.
- [olaresd](../developer/install/installation-overview.md#container-runtime-containerd): A daemon process that monitors hardware and network changes, while also managing cluster upgrades, restarts, and other maintenance operations.

These tools streamline installation, maintenance, and scaling for Olares.

## Platform

The platform layer services run in containers with middlewares such as databases, messaging system, file system, workflow orchestration, secret management, and observability.

### Relational database

Olares uses [PostgreSQL](https://www.postgresql.org/) 16 as its primary relational database. All applications share a single PostgreSQL instance, with each having dedicated accounts for isolation. PostgreSQL also serves as a full-text search engine and vector database.

For multi-nodes, [Citus](https://github.com/citusdata/citus) is used, though its production readiness is still under evaluation.

In the future, PostgreSQL is expected to migrate to the infrastructure layer for better resource management.

### Key-value cache

Olares integrates [KVRocks](https://github.com/apache/incubator-kvrocks), a Redis-compatible persistent key-value store built on RocksDB. KVRocks balances memory and disk storage, making it more resource-efficient than Redis clusters, though slightly slower in performance.

### Message queue

Olares integrates [NATS](https://nats.io/), a lightweight and high-performance message-oriented middleware, as the messaging system. NATS ensures low resource consumption while delivering reliable message queues.

### Distributed file system

Olares employs [JuiceFS](https://juicefs.com/), a cloud-native distributed file system, to provide POSIX-compatible interfaces for applications. When S3 or MinIO is used as the storage backend, JuiceFS ensures seamless file access across nodes.

### Workflow management

Olares uses [Argo Workflows](https://argoproj.github.io/) for workflow orchestration. This Kubernetes-native tool automates complex tasks, such as those required by Olares' distributed recommendation engine. Currently, this functionality is not available to third-party applications.

### Secret management

Two secret management solutions are integrated into Olares:

- [Vault](https://github.com/beclab/olares/tree/main/apps/vault): Protects sensitive data like accounts, passwords, and mnemonics. It encrypts secrets, ensuring that even if the server is compromised, the data remains secure. Vault is developed by the Olares team based on [Padloc](https://padloc.app/).
- [Infisical](https://infisical.com/): A tool for managing sensitive information and preventing secret leaks in Olares development.

### Observability

Olares provides observability through the following:

- [Prometheus](https://prometheus.io/): Used for system monitoring and resource usage tracking. It collects resource metrics for applications like Dashboard and Market.
- [OpenTelemetry](https://opentelemetry.io/)*: Enables tracing of request workflows within the Olares system using eBPF-based monitoring. *(In development)*

### Other middlewares

The Olares application store includes common middleware such as [Grafana](https://grafana.com/) for visualization, [MongoDB](https://www.mongodb.com/) for document storage, and [Chaos Mesh](https://chaos-mesh.org/) for chaos testing.

## Application framework

The application framework layer provides common functionality and interfaces for system and third-party applications.

### Authentication and authorization

Olares uses [LLDAP](https://lldap.example.com/) to manage user accounts and provide LDAP (Lightweight Directory Access Protocol) services for applications.

Additionally, [Authelia](https://www.authelia.com/) adds authentication and authorization support, including multi-factor authentication and single sign-on (SSO).

### Application governance

Components for application governance include:
- [app-service](https://github.com/beclab/app-service): Handles application lifecycle management and resource allocation.
- [system-server](https://github.com/beclab/system-server): Manages permissions for inter-application API calls and handles network routing between applications and database middlewares.
- image-server: Works with app-service to manage container images required by Olares applications.
- [bfl](https://github.com/beclab/bfl): The Backend For Launcher service that aggregates backend interfaces and proxies requests for all system services, including user-isolated system and cluster information.

### Network connectivity
Olares supports secure and flexible network connectivity through:
- Reverse proxy: Options include [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/), Olares Tunnel, and self-built FRP.
- [Tailscale](https://tailscale.com/): Enables users to securely access the system from anywhere.
- [Headscale](https://github.com/juanfont/headscale): A self-hosted implementation of the Tailscale control server.

### File service
Components for file service include:
- File server: Provides essential file management services.
- [Seafile](https://www.seafile.com/): An open-source alternative to Dropbox for file synchronization. Olares deeply integrates Seafile, enabling users to synchronize files across multiple devices into a centralized repository.
- Drive server: Provides integration with external storage services like Google Drive, Dropbox and S3.
- Media server: Streams video files using [ffmpeg](https://github.com/FFmpeg/FFmpeg). 

### Knowledge service
Components for knowledge service include:
- Knowledge: Stores content such as web pages, videos, audio files, PDFs, and EPUBs that users collect via the browser extension or share from their mobile phones using LarePass. This repository is also utilized by the decentralized recommendation engine to store its results.
- Download: Uses [aria2](https://aria2.github.io/) and [youtube-dlp](https://github.com/yt-dlp/yt-dlp) to download files, magnet links, and online videos.
- Search: Provides full-text search for stored content in Knowledge and Files.
- [RSSHub](https://github.com/DIYgod/RSSHub): Generates RSS feeds for easier content subscription.

### AI service

Olares empowers AI capabilities with:
- Model serving*: Hosts AI models for applications. *(In development)*
- RAG interface*: Provides Retrieval-Augmented Generation (RAG) services for files, articles, images, and videos. *(In development)*
- Agent & workflow orchestration*: Manages agents and tool workflows. *(In development)*

### System service

System services include:
- Notification: Delivers system-wide notifications.
- Backup*: Supports backups for directories, applications, and clusters. *(In development)*
- Upgrade*: Supports automated system upgrades. *(In development)*

## System applications

System applications offer tools for managing files, knowledge, passwords, and the system itself. These applications are pre-installed.

Users can install additional applications via the Market app.

### Files

A file management app that manages and synchronizes files across devices and sources, enabling seamless sharing and access.

### Wise

A local-first and AI-native modern reader that helps to collect, read, and manage information from various platforms. Users can run self-hosted recommendation algorithms to filter and sort online content.

### Vault

A secure password manager for storing and mangaging sensitive information across devices.

### Market

A decentralized and permissionless app store for installing, uninstalling, and updating applications and recommendation algorithms.

### Desktop

A hub for managing and interacting with installed applications. File and application searching are also supported.

### Profile

An app to customize the user's profile page.

### Settings

A system configuration application.

### Dashboard

An app for monitoring system resource usage.

### Control Hub

The console for Olares, providing precise and autonomous control over the system and its environment.

### DevBox

A development tool for building and deploying Olares applications.

## Learn more
- To get started with Olares, see the [Getting Started guide](get-started/index.md).
- To learn more about the internals of Olares, see the topics in [Concept](concepts/index.md).
- For in-depth details about how each component of Olares is orchestrated, see [Olares installation overview](../developer/install/index.md).