---
outline: [2, 3]
description: Understand shared apps in Olares, the Engine Base architecture, differences from v2, and how to migrate legacy shared applications.
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
    
    - **Applications with built-in UI**: Includes both a backend service and a web UI. Members open it directly from the Launchpad. Examples include **Dify** and **ComfyUI**.
    
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

## Replace Ollama with Engine Base

The standalone **Ollama** shared app and legacy model apps have been replaced by the new **Engine Base** architecture.

Previously, you had two options to run local LLMs on Olares:
- **Ollama**: You installed the standalone Ollama shared app and pulled models manually via its command line. The major pain point here was scheduling conflicts: the standalone Ollama app did not work with Olares' GPU time-slicing management, so running several models at once led to conflicts over the GPU.
- **Pre-packaged model apps**: Olares also provided dedicated model apps, but each app bundled a specific model with a specific engine. The major pain point here was tight coupling: updating an underlying engine or shipping a new model required releasing a new app every time, slowing down the release of new models.

To resolve these challenges, Olares introduced the Engine Base architecture. This new design abstracts the inference engines into four reusable base applications: **Ollama Engine Base**, **vLLM Engine Base**, **SGLang Engine Base**, and **llama.cpp Engine Base**.

Instead of pulling models into a single Ollama app or installing pre-bundled apps, you now select the desired engine base and clone it into independent, tailored model instances. This architecture shift brings the following core benefits:

- **Works with GPU time-slicing**: Model instances follow Olares' scheduling rules, so you can now run multiple models smoothly without hardware or scheduling conflicts.
- **Centralized management**: Engines are maintained centrally via the base apps. When Olares updates an engine base, you can upgrade your cloned instances without waiting for a new model app release.
- **Flexible configuration**: You set each instance's model source, parameters, and capabilities yourself, and pick the best engine for each model. For example, select Ollama for quick local inference, vLLM or SGLang for high-throughput serving, or llama.cpp for edge deployments.
- **Built-in model console**: Each instance features a dedicated console to monitor GPU residency, track performance, and manage client connections.

For details on creating and configuring model instances with Engine Base apps, see [Host local large language models with Engine Base apps](/use-cases/llm-base-apps.md).

## Manage legacy v2 shared apps

Installed v2 shared apps continue to work after upgrading to Olares v1.12.6. You can start, use, stop, and resume them as before, but you cannot upgrade them directly to the new architecture.

:::tip Identify v2 shared apps
To check if an app is a legacy v2 version, open its details page in Market and check the **Compatibility** field in the **Information** panel. A v2 shared app typically shows `Olares >=1.12.3-0, <1.12.6`.
:::

:::warning Check your app data before uninstalling
Before you uninstall, you might need to back up app-specific data to avoid losing your workflows. See [Migrate from v2 to the new architecture](#migrate-from-v2-to-the-new-architecture) for app-specific guidance.
:::

## Migrate from v2 to the new architecture

Different shared apps require different migration paths. Choose the option below that matches the shared app you are migrating.

### Option 1: Migrate data automatically

Use this option when the shared app supports automatic data migration.

- **Apps in this category**: ComfyUI
- **Steps**: For detailed steps on migrating ComfyUI, including how to preserve your data and the new data locations, see [ComfyUI migration notes](/use-cases/comfyui-common-issues.md).

### Option 2: Perform a clean reinstallation

Use this option when the shared app has no user-created data to migrate.

- **Apps in this category**: Apps with no significant data, such as Falco and MTranServer
- **Steps**: Uninstall the v2 shared app, and then install the new shared app from Market.

  :::tip Identify v2 and new versions
  On the app details page, check the **Compatibility** field in the **Information** panel:

  - The v2 shared app shows `Olares >=1.12.3-0, <1.12.6`.
  - The new shared app shows `Olares >=1.12.6-0`.
  :::

### Option 3: Back up and restore manually

Use this option when the shared app stores user-created data or settings that must be moved manually.

- **Apps in this category**: Dify, OnlyOffice, SearXNG, and Xinference
- **Steps**: Follow the migration guide for the shared app you are migrating: [Dify](/use-cases/dify-upgrade.md), [OnlyOffice](/use-cases/onlyoffice.md), [SearXNG](/use-cases/searxng.md), and [Xinference](/use-cases/xinference.md).

### Option 4: Upgrade the Ollama app to Engine Base

Use this option when you are migrating from the standalone Ollama shared app to the new Engine Base architecture.

- **Apps in this category**: The Ollama app installed to pull models
- **Steps**: Deploy the model pulled via Ollama on an [Engine Base app](/use-cases/llm-base-apps.md), get the Base URL in the model console, and then reconfigure your clients.

## FAQs

### What about the Shared Entrance?

The **Shared entrance** is still present in the system, but you should not use it to access a shared app.

Moving forward, this non-user-specific address is reserved for internal system-level integrations, such as pre-configured Agent-to-Agent API calls that require a unified address.

To access a shared app yourself, open it from the Launchpad. To connect a client, use the Base URL shown in the model console.
