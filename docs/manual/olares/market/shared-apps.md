---
outline: [2, 3]
description: Understand Olares shared applications, why the V3 architecture replaces V2, what happened to Ollama, and how to migrate from legacy V2 shared apps.
---

# Shared applications

Shared applications are a special category of community applications in Olares. They are deployed and managed centrally by the administrator, and every member of the cluster can use them without installing their own copy.

This page explains what shared applications are, why Olares moved to a new architecture in Olares 1.12.6, and how to migrate from legacy V2 shared applications to the new V3 architecture.

## What are shared applications?

A shared application provides shared resources or services to all users in an Olares cluster. Instead of every member installing and running their own instance, the administrator installs and manages one instance that everyone uses.

Key characteristics of shared applications include:

- **Centralized management**: Only administrators can install, upgrade, stop, resume, and uninstall shared applications. Administrators are responsible for configuring and hosting the app's service, resources, and runtime environment within the cluster.
- **Easy identification**: In Olares Market, shared applications are typically marked with labels such as "Shared", "Shared app", or the <i class="material-symbols-outlined">group</i> icon.
- **Flexible access**: The way you access a shared application depends on its form.

    - **Headless backend service**: Provides API services for compatible clients, with no end-user graphical interface. For example, model instances created on Engine Base apps expose a shared entrance address in their model console. Members paste this address into clients such as Open WebUI or LobeChat.
    
    - **Applications with built-in UI**: Includes both a backend service and a web UI. Members open it directly from the Launchpad. Examples include **Dify Shared** and **ComfyUI Shared**.
    
- **Unified access address with data isolation**: All shared applications follow this unified URL access rule: `https://<app-id>.<username>.<platform-domain>`. Members access the same shared application through their own usernames, and the system automatically isolates each member's data based on the username, ensuring members can only access their own data.

## Why did Olares introduce the V3 architecture?

Before V1.12.6, shared applications were split into a server component and a client-side access point. This design had an issue: the server was tightly coupled to these access points, so uninstalling one could leave the shared server inaccessible, breaking the service for everyone.

The V3 architecture replaces the client/server split with a single, unified shared server:

| V2 architecture | V3 architecture |
|:----------------|:----------------|
| Server + client-side access point | Single unified shared server |
| Uninstalling an access point could break the server | Server lifecycle is independent of any access point |
| Multiple access addresses and formats | Unified address format for all users |
| Client and server managed separately | Administrators manage one shared service |

Benefits of V3 include:
- **No orphaned services**: Uninstalling a user-facing app no longer affects the shared server.
- **Simpler management**: Administrators manage one shared application instead of coordinating client and server components.
- **Consistent access**: Every shared app uses the same URL pattern, making it easier to connect clients.
- **Clearer data isolation**: Members access shared services through their own usernames, and the system keeps data separated automatically.

## V2 shared app lifecycle after upgrading to 1.12.6

Installed V2 shared apps continue to work after upgrading to Olares V1.12.6. You can start, use, stop, and resume them as before, but you cannot upgrade them directly to V3.

:::warning Data is not migrated automatically
Existing data from a V2 shared app is not moved to V3 automatically. To move to the new architecture, uninstall the V2 app, install the V3 app, and then optionally migrate your data according to the app type. See [Migrate from V2 to V3](#migrate-from-v2-to-v3) for details.
:::

## Migrate from V2 to V3

Different apps require different migration paths. Choose the option below that matches the app you are migrating.

### Option 1: Uninstall V2, install V3, system migrates data

Use this option when the app supports automatic data migration.

**Apps in this category:**
- **ComfyUI Shared**

**Steps:**
1. Uninstall the V2 shared app.
2. Install the V3 shared app.
3. The system migrates your data automatically.

### Option 2: Uninstall V2, then install V3

Use this option when the app has no user-created data to migrate.

**Apps in this category:**
- **Model apps** such as Qwen3-Coder 30B (Ollama) and Gemma3 27B (vLLM).
- **Other apps with no significant data** such as Falco and MTranServer.

**Steps:**
1. Uninstall the V2 shared app.
2. Install the V3 shared app.
3. For model apps, deploy the model again on an Engine Base app and reconfigure your clients.

### Option 3: Back up, uninstall V2, install V3, then restore data

Use this option when the app stores user-created data or settings that must be moved manually.

**Apps in this category:**
- **Dify Shared** — apps, knowledge bases, agent configurations, and settings.
- **OnlyOffice** — documents and application settings.
- **SearXNG** — search preferences and configuration.
- **Xinference** — deployed models and service settings.

**Steps:**
1. **Back up V2 app data**

   Open Files and go to **Application** > **Data** > `<app-name>`. Download or copy any files, configuration, or data you want to keep.

2. **Uninstall the V2 shared app**

   Open Market or the Launchpad, uninstall the V2 app, and make sure you select **Also uninstall the shared server (affects all users)**. This fully removes the old service and frees its resources.

3. **Install the V3 app**

   Find the V3 version of the shared app in Market and install it. The administrator must perform this step.

4. **Migrate or reconfigure your data**

   Copy your backup into the V3 app's data directory, or reconfigure the app according to its own migration instructions.

5. **Update access addresses**

   V3 shared apps use the unified address format `https://<app-id>.<username>.<platform-domain>`. Update client configurations that still point to the old V2 address.

### Option 4: Migrate from Ollama to Engine Base

Use this option when you are migrating from the standalone Ollama V2 app.

The standalone **Ollama** shared app has been replaced by the **Engine Base** architecture. Previously, each Ollama-based model was a separate shared app that bundled its own copy of the inference engine. This caused duplicated engines and heavy maintenance. In the new architecture, Olares maintains a small set of Engine Base apps, including **Ollama Engine Base**, **vLLM Engine Base**, **SGLang Engine Base**, and **llama.cpp Engine Base**, and you create the model instances you need on top of them.

**Steps:**
1. Uninstall the Ollama V2 app.
2. Select the engine base app you want.
3. Create an instance for the model you need.
4. Get the model service API address in the model console, and update the address in your clients.

## FAQs

### Will my V2 shared apps receive updates?

No. V2 shared apps will not be updated to new versions.

To get new features, fixes, or the improved V3 architecture, you must uninstall the V2 shared app and install the V3 shared app.

### What happened to Ollama?

The standalone **Ollama** shared app has been replaced by the **Engine Base** architecture.

Previously, Olares provided models mainly through individual shared model apps in Market, such as Qwen3.5 35B A3B UD-Q4 (Ollama) and Qwen3 30B (vLLM). Each of these apps bundled its own copy of the inference engine with the model. This created two problems:
- Duplicate engines: Running several models meant running several copies of the same engine, which consumed extra resources.
- Heavy maintenance: Every new model had to be packaged and published as a separate shared app by Olares. This did not scale and slowed down the release of new models.

In the new architecture, Olares only maintains a small set of Engine Base apps: **Ollama Engine Base**, **vLLM Engine Base**, **SGLang Engine Base**, and **llama.cpp Engine Base**. You select the base engine you want, and then create the model instances you need. Each instance runs as its own shared service and appears as a separate entry on the Launchpad.

Because the engine is shared, multiple models can run on top of one Engine Base without installing a new engine each time. You can also choose the best engine for each model, for example, Ollama for local CPU/GPU inference, vLLM or SGLang for high-throughput serving, or llama.cpp for edge deployments.
