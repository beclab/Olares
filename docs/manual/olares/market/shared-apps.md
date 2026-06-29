---
outline: [2, 3]
description: Understand Olares shared applications, why the V3 architecture replaces V2, what happened to Ollama, and how to migrate from legacy V2 shared apps.
---

# Shared applications

Shared applications are a special category of community applications in Olares. They are deployed and managed centrally by the administrator, and every member of the cluster can use them without installing their own copy.

This page explains what shared applications are, why Olares moved to a new architecture in Olares 1.12.6, and how to migrate from legacy V2 shared applications to the new V3 architecture.

## Understand

A **shared application** is a special category of community applications on Olares designed to provide unified, shared resources or services to all users within an Olares cluster.

Key characteristics of shared applications include:

- **Centralized management**: Only administrators can install, upgrade, stop, resume, and uninstall shared applications. Administrators are responsible for configuring and hosting the app's service, resources, and runtime environment within the cluster.
- **Easy identification**: In Olares Market, shared applications are typically marked with labels such as "Shared", "Shared app", or the <i class="material-symbols-outlined">group</i> icon.
- **Flexible access**: The way you access a shared application depends on its form.
    - **Headless backend service**: For backend service shared applications without a graphical UI, the service exposes a standard API through a shared entrance, which can be consumed by any compatible third-party client such as LobeChat and Open WebUI. Take LLM base applications like Ollama LLM Base as an example: in Market, the base application provides a **View** entry. After an administrator clicks **View**, they can create a specific model instance (e.g., Qwen3.5 27B). Each deployed model displays its shared entrance address in its model console, where members can get the address and configure it in their clients.
    - **Complete application with built-in UI**: For shared applications that include a complete user interface and backend service themselves (e.g., ComfyUI Shared or Dify Shared), an application entry with the same name is generated on the Launchpad after the administrator installs it, and cluster members can access it directly through this entry.
- **Unified access address with data isolation**: All shared applications follow this unified URL access rule: `https://<app-id>.<username>.<platform-domain>`. Members access the same shared application through their own usernames, and the system automatically isolates each member's data based on the username, ensuring members can only access their own data.

## What are shared applications?

A shared application provides shared resources or services to all users in an Olares cluster. Instead of every member installing and running their own instance, the administrator installs and manages one instance that everyone uses.

Key characteristics of shared applications:

- **Centralized management**: Only administrators can install, upgrade, stop, resume, and uninstall shared applications. Members simply use them.
- **Easy identification**: In Olares Market, shared applications are marked with labels such as **Shared**, **Shared app**, or the <i class="material-symbols-outlined">group</i> icon.
- **Flexible access**: Depending on the app type, members either open it from the Launchpad or connect a third-party client to its shared API entrance.
- **Data isolation**: All shared applications use a unified access address format:

  ```text
  https://<app-id>.<username>.<platform-domain>
  ```

  Members access the same shared service through their own usernames, and the system isolates each member's data automatically.

Shared applications come in two forms:

- **Headless backend service**: Provides only API services with no graphical interface. Any compatible client can call the API. Examples include LLM base apps such as **Ollama LLM Base (llm-init)**.
- **Complete application with built-in UI**: Includes both a backend service and a web UI. Members open it directly from the Launchpad. Examples include **Dify Shared** and **ComfyUI Shared**.

## Why did Olares introduce the V3 architecture?

Before Olares 1.12.6, shared applications followed a V2 architecture that split each app into a client component and a server component. This design had a serious drawback: if a member uninstalled the client component, the shared server could become inaccessible or orphaned, breaking the service for everyone.

The V3 architecture replaces the client/server split with a single, unified shared server:

| V2 architecture | V3 architecture |
|:----------------|:----------------|
| Client + server components | Single unified shared server |
| Uninstalling the client could break the server | Server lifecycle is independent of any client |
| Multiple access addresses and formats | Unified address format for all users |
| Complex management across components | Administrators manage one shared app entry |

Benefits of V3 include:

- **No orphaned services**: Uninstalling a user-facing app no longer affects the shared server.
- **Simpler management**: Administrators manage one shared application instead of coordinating client and server components.
- **Consistent access**: Every shared app uses the same URL pattern, making it easier to connect clients and bookmarks.
- **Clearer data isolation**: Members access shared services through their own usernames, and the system keeps data separated automatically.

## What happened to Ollama?

The monolithic **Ollama** shared application has been replaced by a more flexible **LLM base** model.

In the new model:

1. The administrator installs an **LLM base app** such as **Ollama LLM Base (llm-init)** from Market.
2. From the base app, the administrator creates specific **model instances** (for example, Qwen3.5 27B).
3. Each model instance runs as a shared service and appears as its own entry on the Launchpad.
4. Members open the model console to copy the shared entrance address, then paste it into any compatible client such as **Open WebUI**, **LobeHub (LobeChat)**, or **Dify**.

This change separates the model runtime from the models themselves, so you can add, remove, or switch models without replacing the whole application.

## V2 shared app lifecycle after upgrading to 1.12.6

After you upgrade Olares to 1.12.6, previously installed V2 shared applications are **not deleted automatically**. The system cleans up legacy client components where possible and keeps the shared server running.

You can still:

- Start, stop, pause, and resume V2 shared apps.
- Use V2 shared apps through their existing access addresses.

However, V2 shared apps **cannot be directly upgraded** to V3. To move to the new architecture, you must:

1. Back up any data you want to keep.
2. Uninstall the V2 shared app, selecting **Also uninstall the shared server (affects all users)**.
3. Install the V3 version from Market.
4. Manually migrate or reconfigure your data.

:::warning Data is not migrated automatically
Existing data from a V2 shared app is **not** moved to the V3 version automatically. Each application stores data differently, so migration steps vary by app. Per-app migration guides will be provided separately.
:::

## General migration workflow

Follow this workflow when moving any shared app from V2 to V3:

1. **Back up app data**

   Open **Files** and go to **Application** > **Data** > `<app-name>`. Download or copy any files, configuration, or data you want to keep.

2. **Uninstall the V2 shared app**

   Open **Market** or the **Launchpad**, uninstall the V2 app, and make sure you select **Also uninstall the shared server (affects all users)**. This fully removes the old service and frees its resources.

3. **Install the V3 version**

   Find the V3 version of the shared app in Market and install it. The administrator must perform this step.

4. **Migrate or reconfigure your data**

   Copy your backup into the V3 app's data directory, or reconfigure the app according to its own migration instructions. See the per-app guides below for details.

5. **Update access addresses**

   V3 shared apps use the unified address format `https://<app-id>.<username>.<platform-domain>`. Update any bookmarks or client configurations that still point to the old V2 address.

## Per-app migration guides

Migration steps differ for each shared application because each app stores its data in its own way. Dedicated migration guides will be added for the following key apps:

- **ComfyUI Shared** — models, custom nodes, workflows, and generated images.
- **Dify Shared** — apps, knowledge bases, agent configurations, and settings.
- **Ollama / LLM base** — transitioning from the monolithic Ollama app to LLM base apps and model instances.

## Frequently asked questions

### Can I keep using my V2 shared apps?

Yes. Installed V2 shared apps continue to work after upgrading to Olares 1.12.6. You can start, stop, pause, and resume them as before.

### Will my V2 shared apps receive updates?

No. V2 shared apps will not be updated to new versions. To get new features, fixes, or the improved V3 architecture, you must uninstall the V2 version and install the V3 version.

### What happens if I uninstall only the V2 client?

In V2, uninstalling only the client component could leave the shared server in an inconsistent or unreachable state. If you are removing a V2 shared app, always select **Also uninstall the shared server (affects all users)** to fully clean it up.

### Do I need to change dependent client apps?

Yes. V3 shared apps use a unified access address format. Any client that connected to the old V2 address must be updated with the new V3 address. You can find the new address in the shared app's console or Launchpad entry.

### Why is Ollama no longer a standalone shared app?

Ollama has been split into LLM base apps and individual model apps. This lets administrators deploy only the models they need and connect them to any compatible client, rather than managing one monolithic application.
