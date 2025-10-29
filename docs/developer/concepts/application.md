---
outline: [2, 3]
description: Fundamental concepts of Olares application system, explaining application identifiers and characteristics of four application types such as cluster-scoped applications. Includes service provider mechanisms and application dependencies.
---

# Applications
 
This documents covers essential concepts for managing application identifiers, types, permissions, and Market integrations within Olares. 

## Application identifier

In Olares, each application is assigned two identifiers: an application name and an application ID.

### Application name

Application names are assigned by Indexers. The official Indexer address maintained by Olares is [apps](https://github.com/beclab/apps). The directory name of an application within this repository serves as the application name.

### Application ID

The application ID is derived as the first eight characters of the MD5 hash of the application name. For instance, if the application name is "hello", the application ID becomes "b1946ac9".

Application IDs are utilized in endpoints.

## Application types

There are multiple types of applications in Olares. You can distinguish a specific application type according to the namespace shown in Control Hub.

### System applications

System applications encompass Kubernetes, Kubesphere, Olares components, and essential hardware drivers. The system-level namespaces include:

```
os-system
kubesphere-monitoring-federated
kubesphere-controls-system
kubesphere-system
kubesphere-monitoring-system
kubekey-system
default
kube-system
kube-public
kube-node-lease
gpu-system
```
`os-system` is a component developed by Olares team. Cluster-level applications and various database middleware provided by the system are installed under this namespace.

### User system applications

Olares supports multiple users and provides two distinct namespaces for system applications accessible to Admin and Member users:

- **user-space-{Local Name}**

    The `user-space` namespace is where system applications that users interact with daily are installed. These applications include:
    - Files
    - Settings 
    - Control Hub
    - Dashboard
    - Market
    - Profile 
    - Vault

   These applications interact with each other while also calling system-level interfaces (such as Kubernetes' `api-server` interface). To ensure system security, Olares deploys them in isolated user-space namespaces and uses sandbox mechanisms to prevent malicious program attacks and unauthorized access.

- **user-system-{Local Name}**

   System applications and user's built-in applications are generally restricted from direct access by third-party applications.
  
   However, if built-in applications or database clusters make specific service interfaces available through a [service provider](../develop/advanced/provider.md), community applications can request access by [declaring these permissions](../develop/package/manifest.md#sysdata).
   
   When such access is granted, the system routes these network requests through secure proxies in the `user-system` namespace, ensuring proper authentication and protection of resources.

### Community applications

Community applications are applications created and maintained by third-party developers. They encompass a wide range of purposes, from productivity tools and entertainment applications to data analysis utilities.

The namespace of community applications consists of two parts: application name and the user's [local name](olares-id.md#olares-id-structure), for example:

```
n8n-alice
gitlab-client-bob
```

### Shared applications

A **shared application** is a special category of community applications on Olares designed to provide unified, shared resources or services to all users within an Olares cluster.

Key characteristics of shared applications include:

* **Centralized management**: Only administrators can install the core service of a shared application. Administrators are responsible for installing, configuring, and hosting the app's service, resources, and runtime environment within the cluster.
* **Easy identification**: In Olares Market, shared applications are typically marked with a "Shared" label for easy identification.
* **Flexible access**: The method for accessing a shared application depends on the app's form:
    * **Headless backend service**: For shared applications that typically run as a background service without a graphical UI (e.g., Ollama), users need to install a **reference application** to call the service. For example, users within the cluster can access the Ollama service via Open WebUI or LobeChat.
    * **Complete application with built-in UI**: For shared applications that include a complete user interface and backend service themselves (e.g., ComfyUI Shared or Dify Shared), administrators and other users in the cluster can obtain the service access point by directly installing the shared application itself.

### Reference applications

Reference applications are applications that have been granted access to specific shared applications within Olares. They typically provide a user-friendly interface, allowing users to easily access the APIs or services exposed by the shared applications.

For example, Open WebUI, LobeChat, and n8n are reference applications for Ollama. Dify Shared is the reference application of itself.

### Dependencies

Dependencies are prerequisite applications that must be present for certain applications to function properly. Before installing an application with dependencies, users must ensure all required dependencies are already installed in the cluster.

### Service provider

The Service Provider mechanism enables community applications to interact with system applications and services from other community applications.

![Service Provider](/images/overview/olares/image3.jpeg)

The mechanism consists of three procedures：

1. Provider declaration: Developers must [declare their application as a provider](../../developer/develop/advanced/provider#define-provider) for specific service interfaces.
  The system includes built-in Providers.

2. Permission request: Applications seeking to use a service interface must explicitly [request provider access permissions](../../developer/develop/advanced/provider#request-permission-to-call-provider). 

3. Request handling: `system-server` services under `user-system` act as an agent that handles incoming requests and performs necessary permission validations.


## Learn more

- User

  [Manage apps in Market](../../manual/olares/market.md)<br>

- Developer

  [Learn to develop applications on Olares](../develop/index.md)<br>
