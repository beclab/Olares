---
outline: [2, 3]
description: Understand Olares shared applications, why the new shared architecture replaces v2, how the Engine Base architecture replaced the standalone Ollama app, and how to migrate from legacy v2 shared apps.
---

# Shared applications

Shared applications are a special category of community applications in Olares. They are deployed and managed centrally by the administrator, and every member of the cluster can use them without installing their own copy.

This page explains what shared applications are, why Olares moved to a new architecture in Olares 1.12.6, and how to migrate from legacy v2 to the new architecture.

## Understand shared applications

<!-- #region shared-apps-what-are -->
A shared application provides shared resources or services to all users in an Olares cluster. Instead of every member installing and running their own instance, the administrator installs and manages one instance that everyone uses.

Key characteristics of shared applications include:

- **Centralized management**: Only administrators can install, upgrade, stop, resume, and uninstall shared applications. Administrators are responsible for configuring and hosting the app's service, resources, and runtime environment within the cluster.
- **Easy identification**: In Olares Market, shared applications are typically marked with labels such as "Shared", "Shared app", or the <i class="material-symbols-outlined">group</i> icon.
- **Flexible access**: The way you access a shared application depends on its form.

    - **Headless backend service**: Provides API services for compatible clients, with no end-user graphical interface. For example, model instances created on Engine Base apps expose a **Base URL** in their model console. Members paste this address into clients such as Open WebUI or LobeChat.
    
    - **Applications with built-in UI**: Includes both a backend service and a web UI. Members open it directly from the Launchpad. Examples include **Dify** and **ComfyUI Shared**.
    
- **Unified HTTPS address**: All shared applications use the same URL pattern, that is `https://<app-id>.<username>.<platform-domain>`. Each member accesses the same shared application through their own username.
<!-- #endregion shared-apps-what-are -->

## Explore the architecture shift

Before v1.12.6, shared applications were split into a server component and a client-side access point. This design had an issue: the server was tightly coupled to these access points, so uninstalling one could leave the shared server inaccessible, breaking the service for everyone.

The new architecture replaces the client/server split with a single, unified shared server:

| v2 architecture | New architecture |
|:----------------|:----------------|
| Server + client-side access point | Single unified shared server |
| Uninstalling an access point might break the server | Server lifecycle is independent of any access point |
| Multiple access addresses and formats | Unified address format for all users |
| Client and server managed separately | Administrators manage one shared service |

Benefits of the new architecture include:
- **Simpler management**: Administrators manage one shared application instead of coordinating client and server components.
- **No orphaned services**: Uninstalling a user-facing app no longer affects the shared server.
- **Personalized access**: Every member accesses a shared app through their own username, using the same HTTPS URL pattern across all shared apps.

## Manage legacy v2 shared apps

Installed v2 shared apps continue to work after upgrading to Olares v1.12.6. You can start, use, stop, and resume them as before, but you cannot upgrade them directly to the new architecture.

:::warning Data is not migrated automatically
Existing data from a v2 shared app is not moved to the new architecture automatically. To move to the new architecture, uninstall the v2 app, install the new shared app, and then optionally migrate your data according to the app type. See [Migrate from v2 to the new architecture](#migrate-from-v2-to-the-new-architecture) for details.
:::

## Replace Ollama with Engine Base

The standalone **Ollama** shared app has been replaced by the **Engine Base** architecture.

Previously, Olares provided models mainly through shared model apps in Market, such as Qwen3.5 35B A3B UD-Q4 (Ollama). Each app bundled a model with its own inference engine. This design had two main problems:
- **Scheduling conflicts**: The standalone Ollama app did not work with Olares' GPU time-slicing management, so running several models at once led to conflicts over the GPU.
- **Tight coupling**: Because the engine and model were bundled in each app, updating an engine or adding a model meant shipping a new app for every model, which slowed down the release of new models.

The Engine Base architecture keeps the same basic runtime, where each model instance still runs with its own engine, but abstracts the engine into a reusable base app：
- Olares maintains a small set of Engine Base apps: **Ollama Engine Base**, **vLLM Engine Base**, **SGLang Engine Base**, and **llama.cpp Engine Base**.
- You select the base you want, and then create the model instances you need. Each instance runs as its own shared service and appears as a separate entry on the Launchpad.

This brings several benefits:
- **Works with GPU time-slicing**: Model instances follow Olares' scheduling, so you can run multiple models without the conflicts of the old standalone Ollama app.
- **Centralized management in Market**: Engines are maintained in one place through the base apps. When Olares updates a base, you can upgrade your cloned instances to the new engine version.
- **Flexible configuration**: You set each instance's model source, parameters, and capabilities yourself, and pick the best engine for each model. For example, select Ollama for quick local inference, vLLM or SGLang for high-throughput serving, or llama.cpp for edge deployments.

## Migrate from v2 to the new architecture

Different apps require different migration paths. Choose the option below that matches the app you are migrating.

### Option 1: Migrate data automatically

Use this option when the app supports automatic data migration.

- **Apps in this category**: ComfyUI
- **Steps**: For detailed steps on migrating ComfyUI, including how to preserve your data and the new data locations, see [ComfyUI migration notes](/use-cases/comfyui-common-issues.md).

### Option 2: Perform a clean reinstallation

Use this option when the app has no user-created data to migrate.

- **Apps in this category**: Apps with no significant data, such as Falco and MTranServer
- **Steps**: Uninstall the v2 shared app, and then install the new shared app from Market.

  :::tip Identify v2 and new versions
  On the app details page, check the **Compatibility** field in the **Information** panel:

  - The v2 shared app shows `Olares >=1.12.3-0, <1.12.6`.
  - The new shared app shows `Olares >=1.12.6-0`.
  :::

### Option 3: Back up and restore manually

Use this option when the app stores user-created data or settings that must be moved manually.

- **Apps in this category**: Dify, OnlyOffice, SearXNG, and Xinference
- **Steps**: Follow the migration guide for the app you are migrating: [Dify](/use-cases/dify-upgrade.md), [OnlyOffice](/use-cases/onlyoffice.md), [SearXNG](/use-cases/searxng.md), [Xinference](/use-cases/xinference.md).

### Option 4: Upgrade Ollama and models to Engine Base

Use this option when you are migrating from standalone Ollama or legacy model apps to the new Engine Base architecture.

- **Apps in this category**: Model apps, and the Ollama app installed to pull models
- **Steps**: Deploy the model on an Engine Base app, get the Base URL in the model console, and then reconfigure your clients.

## FAQs

### What about the Shared Entrance?

The **Shared entrance** is still present in the system, but you should not use it to access a shared app.

Moving forward, this non-user-specific address is reserved for internal system-level integrations, such as pre-configured Agent-to-Agent API calls that require a unified address.

To access a shared app yourself, open it from the Launchpad. To connect a client, use the Base URL shown in the model console.
