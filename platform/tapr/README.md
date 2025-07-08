# tapr

## Overview

The Olares application runtime that includes core sidecars & operators that solve the “boring but critical” plumbing required by most self-hosted apps—file uploads, secret storage, event streaming, WebSockets, and turnkey data services.

### image-uploader
An HTTP gateway that lets any authenticated component upload images directly to the requesting user’s **`$HOME`** directory (`~/Pictures` by default).  

### secret-vault
A thin wrapper around **Infisical** that exposes per-user secret management (`create`, `get`, `list`, `delete`) through a REST interface.

### sys-event
An event bus that emits structured system events—`user.created`, `app.installed`, `cpu.high_load`, and more.


### upload-sidecar
A general-purpose file-upload sidecar for in-cluster applications.

### ws-gateway
A drop-in WebSocket gateway that upgrades HTTP connections and proxies bidirectional traffic to your service.

### middleware-operator
A Kubernetes operator that provisions common middleware—PostgreSQL, Redis, MongoDB, Elasticsearch, and more—as first-class CRDs.

