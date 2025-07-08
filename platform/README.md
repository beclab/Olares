# Olares Platform

## Overview

The platform layer services run in containers with middlewares such as databases, messaging system, file system, workflow orchestration, secret management, and observability.

## Sub-component Overview

| Component | Description                                                                                                                                                                                                                                                          |
| --- |----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [fs-lib](fs-lib) | Provides notification for JuiceFS.                                                                                                                                                                                                                                   |
| [juicefs](juicefs) | Provides cloud-native distributed file system for multi-node clusters.                                                                                                                                                                                               |
| [mongodb](mongodb) | A general purpose, document-based, distributed database; used as the document storage in Olares.                                                                                                                                                                     |
| [nats](nats) | A lightweight and high-performance message-oriented middleware, used as the messaging system.                                                                                                                                                                        |
| [open-telemetry](open-telemetry) | Enables tracing of request workflows within Olares using eBPF-based monitoring.                                                                                                                                                                                      |
| [postgresql](postgresql) | A powerful, open source object-relational database system; Functions as the primary relational database in Olares.                                                                                                                                                   |
| [prometheus](prometheus) | An open-source monitoring system used for system monitoring and resource usage tracking.                                                                                                                                                                             |
| [redis](redis) | Contains Redis-like persistent key-value store services for KV cache in Olares.                                                                                                                                                                                      |
| [tapr](tapr) | Provides the app run time for Olares. Bundles a curated set of sidecars and Kubernetes operators to solve the “boring but critical” plumbing required by most self-hosted apps—file uploads, secret storage, event streaming, WebSockets, and turnkey data services. |
| [velero](velero) | An open-source tool to safely backup and restore, perform disaster recovery, and migrate Kubernetes cluster resources and persistent volumes.                                                                                                                        |