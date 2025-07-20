# Olares Framework

## Overview

The application framework layer provides common functionality and interfaces for system and third-party applications.

## Sub-component overview

| Component | Description |
| --- | --- |
| [app-service](app-service) | Handles application lifecycle management and resource allocation. |
| [authelia](authelia) | An open-source authentication and authorization server that provides multi-factor authentication and single sign-on (SSO). |
| [backup-server](backup-server) | Supports backups for directories, applications, and clusters. |
| [bfl](bfl) | The Backend For Launcher service that aggregates backend interfaces and proxies requests for all system services. |
| [docker-nginx-headers-more](docker-nginx-headers-more) | A Docker image for Nginx with the `headers-more` module. |
| [files](files) | Provides essential file management services. |
| [headscale](headscale) | A self-hosted implementation of the Tailscale control server. |
| [infisical](infisical) | A tool for managing sensitive information and preventing secret leaks in Olares development. |
| [kube-state-metrics](kube-state-metrics) | A service that listens to the Kubernetes API server and generates metrics about the state of the objects. |
| [l4-bfl-proxy](l4-bfl-proxy) | A Layer 4 network proxy for BFL (Backend For Launcher). |
| [market](market) | A decentralized and permissionless app store for installing, uninstalling, and updating applications and recommendation algorithms. |
| [monitor](monitor) | Used for system monitoring and resource usage tracking. |
| [notifications](notifications) | Delivers system-wide notifications. |
| [osnode-init](osnode-init) | Initializes the Olares node. |
| [reverse-proxy](reverse-proxy) | Options include Cloudflare Tunnel, Olares Tunnel, and self-built FRP. |
| [seahub](seahub) | The web frontend for the Seafile file hosting platform. |
| [search3](search3) | Provides full-text search for stored content in Knowledge and Files. |
| [system-server](system-server) | Manages permissions for inter-application API calls and handles network routing between applications and database middlewares. |
| [upgrade](upgrade) | Supports automated system upgrades. |
| [vault](vault) | Protects sensitive data like accounts, passwords, and mnemonics. |
