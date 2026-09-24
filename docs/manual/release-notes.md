---
description: Track the latest Olares documentation changes, including new guides, updates, deprecations, and moved pages.
head:
  - - meta
    - name: keywords
      content: Olares, documentation updates, what's new, new guides, deprecations, release notes
---

# What's new in docs

This page highlights documentation changes after each Olares release that affect how you use Olares. It focuses on new guides, major updates, and deprecations. Minor corrections and clarifications are not listed here.

The docs are built and published from the `main` branch of the [Olares GitHub repository](https://github.com/beclab/Olares/tree/main/docs), so they always reflect the latest documentation changes.

<details>
<summary>For Olares OS software release notes, see the GitHub releases page.</summary>

- [Olares 1.12.6](https://github.com/beclab/Olares/releases/tag/1.12.6)
- [Olares 1.12.5](https://github.com/beclab/Olares/releases/tag/1.12.5)
</details>

To stay informed of Olares news, including documentation updates, follow Olares on X or join the [Discord community](https://discord.com/invite/BzfqrgQPDK).

## Olares 1.12.7

<!-- TODO: Add the confirmed release date and enable these links after the official release materials are published.
Released: [confirmed release date]

For the software changes, see the [Olares 1.12.7 release notes](https://github.com/beclab/Olares/releases/tag/1.12.7).
For release highlights and walkthroughs, see the [Olares 1.12.7 blog post](https://www.olares.com/blog/olares-1-12-7/).
-->

### New docs

- Added [What to do after updating to Olares 1.12.7](/manual/update-guides/1.12.7), covering the upgrade path, Router and Lares setup, and agent app updates.
- Added [Use Olares Router as your AI gateway](/use-cases/olares-router), introducing model and tool capabilities, caller authentication, model naming, and connection details.
- Added [Manage Olares and run research tasks with Lares](/use-cases/lares), covering first tasks, permissions, and web research tools configured through Router.
- Added [Install the LarePass browser extension](/manual/install-larepass-browser-extension) and [Translate webpages privately with LarePass](/manual/tutorial/translate-webpages-with-larepass), for translating pages with a local model through Router.
- Added [Save resources to Olares](/manual/larepass/save-resources-to-olares), covering Fetch in LarePass Desktop for web links, torrents, Hugging Face repositories, and supported media.
- Added [Create and manage support tickets in Olares Space](/manual/space/tickets) and expanded [Get support](/manual/help/request-technical-support) with the Ticket app, log collection, and follow-up replies.
- Added [Recover access to LarePass, Olares Desktop, or your Olares device](/manual/help/ts-access-without-mnemonic), alongside updated [password reset instructions](/manual/help/ts-forget-login-password).
- Added [Create a Blender scene with Lares](/use-cases/blender) and [Create videos with Concat and Lares](/use-cases/concat), covering MCP setup, creative workflows, and saving outputs in Files.
- Added [Open Design](/use-cases/open-design), [Codex CLI](/use-cases/codex-cli), and [Manage Olares from OpenCode](/use-cases/opencode-olares-cli) tutorials for design, coding, and device management.
- Added [Minecraft](/use-cases/minecraft) and [Palworld](/use-cases/palworld) server guides with Overlay Gateway setup for local network access.

### Updated docs

- Rewrote [Connect your apps to AI capabilities](/manual/best-practices/connect-ai-apps) for Router, covering API formats, model and tool setup, default system names such as `default-chat`, and API keys for external clients.
- Updated AI client tutorials for the Router workflow, using Qwen3.8-27B (llama.cpp) as the chat example. Added an on-page version selector to affected guides so readers can choose the Olares 1.12.7 or 1.12.6 instructions. [OpenClaw](/use-cases/openclaw) and [Hermes Agent](/use-cases/hermes) now follow the interactive wizard order, with model configuration and context size checked in Router.
- Reorganized the [Olares manual](/manual/overview) around user tasks and [installation guidance](/manual/get-started/install-olares) around Linux, Olares One, and DGX Spark setup paths. Updated [Olares One onboarding](/one/olares-onboarding) for Lares.
- Updated [local access guidance](/manual/best-practices/local-access) with LarePass host mappings on Windows and macOS, alongside VPN and `.local` access options.
- Updated [Market](/manual/olares/market/market) with app and model discovery and source selection, and [language settings](/manual/olares/settings/language-appearance) with the additional language options.
- Reworked [Olares Space account management](/manual/space/manage-accounts), [resource and traffic monitoring](/manual/space/manage-olares), and [billing](/manual/space/billing) instructions.
- Reworked [eGPU guidance for Olares One](/one/egpu) with separate Olares OS and Windows setup guides, compatibility information, and troubleshooting. The [Olares One FAQ](/one/faq#does-olares-one-support-out-of-band-management) now explains which out-of-band management capabilities are available.
- Updated [Jellyfin](/use-cases/jellyfin) hardware acceleration, [Steam streaming](/use-cases/steam-stream) network setup, and [Penpot](/use-cases/penpot) integration with its built-in MCP server.

### Moved and consolidated docs

- Consolidated backup and restore instructions in [Back up and restore Olares](/manual/olares/settings/backup). Previous restore and Olares Space backup links redirect to this guide.
- Split the former My Olares topic into [password and device management](/manual/password-and-devices) and [Olares One hardware settings](/one/hardware-settings).
- Consolidated Olares ID creation in [Create an Olares ID](/manual/get-started/create-olares-id), and moved integration instructions into task-specific guides such as [Mount cloud storage](/manual/olares/files/mount-cloud-storage). Previous links redirect to the replacement pages.

## Olares 1.12.6

Released: July 23, 2026

For the highlights and walk-through of Olares v1.12.6, see the [Olares 1.12.6 blog post](https://www.olares.com/blog/olares-1-12-6/).

### New docs

- Added [Olares CLI overview](/developer/cli-overview), an introduction to managing Olares from the command line.
  - Added [Install Olares CLI](/developer/cli-install), covering how to install `olares-cli` locally or inside an agent app.
  - Added [Log in with Olares CLI](/developer/cli-log-in), explaining how to authenticate `olares-cli` with your Olares ID.
  - Added [Olares CLI agent skills](/developer/cli-agent-skills), describing how to use built-in skills for cluster management, apps, settings, and more.
- Added [Install and activate Olares from the command line](/manual/best-practices/activate-olares-using-cli), for setting up a device with Olares CLI.
- Added [Manage shared AI models with the Common directory](/manual/olares/files/files-common), explaining how to use the Common directory for models shared across apps.
- Added [Compress and extract files](/manual/olares/files/compress-extract-files), covering ZIP, 7z, TAR, and password-protected archives in Olares Files.
- Added [Mount NFS shares](/manual/olares/files/mount-nfs), for accessing NFS shared directories from Olares.
- Added [About shared applications](/manual/olares/market/shared-apps), explaining the new shared app architecture, and [Migrate legacy shared applications](/manual/migrate-shared-apps), covering migration options for v2 apps.
  - [Ollama](/use-cases/ollama)
  - [ComfyUI](/use-cases/comfyui-common-issues)
  - [Dify](/use-cases/dify-upgrade)
  - [OnlyOffice](/use-cases/onlyoffice-migration)
  - [SearXNG](/use-cases/searxng)
  - [Xinference](/use-cases/xinference)
- Added [Configure Overlay Gateway](/manual/olares/settings/overlay-gateway), for managing Overlay Gateway settings.
- Added [Olares One onboarding](/one/olares-onboarding), a guided walkthrough for managing Olares One through natural language.
- Added [Desktop widgets](/manual/olares/desktop#widgets), documenting the new widgets on the Olares desktop.
- Added [Run local AI models with Engine Base apps](/use-cases/llm-base-apps), for deploying and running local AI models on Olares.
- Added [Home Assistant with Overlay Gateway](/use-cases/home-assistant#enable-the-overlay-gateway), for accessing Home Assistant over the local network.
- Added [Jellyfin with Overlay Gateway](/use-cases/jellyfin#enable-overlay-gateway-for-jellyfin), for accessing Jellyfin over the Overlay Gateway.
- Added [Update notes for *Arr apps](/use-cases/arrs-upgrade), covering the configuration changes required after the Olares v1.12.6 internal entrance update.

### Updated docs

- Rewrote [Connect AI apps](/manual/best-practices/connect-ai-apps) to align with the v1.12.6+ architecture.
- Consolidated Olares One software documentation into the main manual and versioned the ISO download links.
- Updated [My Olares](/one/hardware-settings), adding the **Limit CPU frequency** and **Automatic startup** toggles under **My Hardware**.
- Updated [Basic file operations](/manual/olares/files/add-edit-download) with new sorting options, Markdown editing, preview, and additional supported formats.
- Updated [Managing accelerator resources](/manual/olares/settings/gpu-resource) to cover GPU and other accelerator resources.
- Updated [Manage BIOS and EC](/one/update-firmware) with EC 1.03 and BIOS 1.05 changelogs, including the note that **Automatic startup** requires Olares OS 1.12.6 or later.
- Updated [ComfyUI](/use-cases/comfyui) with migration steps and the new v1.12.6 directory structure.
- Updated [App submission guide](/developer/develop/submit-apps) to reflect the new packaging and submission workflow.
- Updated [OlaresManifest Specification](/developer/develop/package/manifest) to document the `0.12.0` schema, including new fields such as `apiVersion`, `accelerator`, `workloadReplicas`, `overlayGateway`, `LLMGatewaySupported`, `appCommon`, and `externalData`, plus deprecated fields.
- Standardized the local model and service connection workflow across AI use-case guides. AI use-case guides now connect to models through the Model Console instead of relying on the standalone Ollama app or individual model apps.

### Deprecated docs

- Added a deprecation notice to the [Ollama](/use-cases/ollama) use-case guide because the standalone Ollama app was removed from Market in Olares 1.12.6.
- Retired the Olares CLI reference pages. The deprecated pages have been superseded by the new [Olares CLI overview](/developer/cli-overview), [Install Olares CLI](/developer/cli-install), [Log in with Olares CLI](/developer/cli-log-in), and [Olares CLI agent skills](/developer/cli-agent-skills) guides.
- Retired all Studio documentation. Studio is no longer available in Market. Use the [Olares CLI agent skills](/developer/cli-agent-skills) guide and the [App submission guide](/developer/develop/submit-apps) for the new packaging and porting workflow.
- The **Allow subnet routing** feature is temporarily disabled in Olares 1.12.6, and the related documentation has been removed from [Configure VPN access to Olares](/manual/olares/settings/remote-access#allow-subnet-routing). The feature will return in a future release.
- Retired the DeerFlow use-case guide. The DeerFlow app was replaced by [DeerFlow 2.0](/use-cases/deerflow2).

### Upcoming deprecations

- The [Ollama](/use-cases/ollama) use-case guide will be removed in a future release. The standalone Ollama app was already removed from Market in Olares 1.12.6, and AI use-case guides now connect to models through the Model Console.
