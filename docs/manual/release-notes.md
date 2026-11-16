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

- [Olares 1.12.8](https://github.com/beclab/Olares/releases/tag/1.12.8)
- [Olares 1.12.7](https://github.com/beclab/Olares/releases/tag/1.12.7)
- [Olares 1.12.6](https://github.com/beclab/Olares/releases/tag/1.12.6)
- [Olares 1.12.5](https://github.com/beclab/Olares/releases/tag/1.12.5)
</details>

To stay informed of Olares news, including documentation updates, follow Olares on X or join the [Discord community](https://discord.com/invite/BzfqrgQPDK).

## Olares 1.12.8

Released: November 16, 2026

For the software changes, see the [Olares 1.12.8 release notes](https://github.com/beclab/Olares/releases/tag/1.12.8).
For release highlights and walkthroughs, see the [Olares 1.12.8 blog post](https://www.olares.com/blog/olares-1-12-8/).

### New docs

- Added the [Olares 1.12.8 update guide](/manual/update-guides/1.12.8), covering the LarePass update requirement, device activation, GPU app restarts, Files permissions, and alternatives to the Control Hub host terminal.
- Added [Manage access to Common and External](/manual/olares/files/manage-common-external-access), explaining administrator and member permissions, folder-level access settings, and access to mounted storage.
- Added [Manage Olares apps with LarePass](/manual/larepass/manage-apps-with-larepass), covering app management from your phone.
- Added [Music Studio](/use-cases/music-studio), [YuE2 Studio](/use-cases/yue2-studio), and [YuE2 Studio with Lares](/use-cases/yue2-studio-lares) tutorials for music creation.
- Added [Hermes Workspace](/use-cases/hermes-workspace), [Hermes SSH workflows](/use-cases/hermes-ssh), [Goose](/use-cases/goose), and [OpenClaw mobile pairing](/use-cases/openclaw-mobile) guides for working with AI agents.
- Added [Cognee](/use-cases/cognee), [TrustGraph](/use-cases/trustgraph), [Hindsight](/use-cases/hindsight), and [MemPalace](/use-cases/mempalace) tutorials for document knowledge and agent memory.
- Added [Linkwarden](/use-cases/linkwarden), [Local Deep Research](/use-cases/local-deep-research), and [Research the web with Lares](/use-cases/lares-web-research), covering saved web libraries and research workflows.
- Added [Unsloth](/use-cases/unsloth), covering model fine-tuning, GGUF export, and running a saved adapter in JupyterLab.
- Added [Plex](/use-cases/plex), [Pterodactyl](/use-cases/pterodactyl), [QEMU](/use-cases/qemu), and [RustDesk](/use-cases/rustdesk) tutorials for media, game servers, virtual machines, and remote desktop access. Added [Store Immich uploads on a NAS](/use-cases/immich-store-uploads-on-nas).
- Added Olares One guides for [Wake-on-LAN](/one/wake-on-lan) and [configuring GPU mode on Ubuntu](/one/configure-gpu-mode).

### Updated docs

- Reworked [Share files](/manual/olares/files/share-files) with the sharing methods supported in each directory in 1.12.8. Updated storage guides to use **Access settings** for member access to Common and External, replacing unsupported internal shares.
- Reorganized [Wise](/manual/olares/wise/index) around [saving web pages and files](/manual/olares/wise/collect), [subscribing to feeds](/manual/olares/wise/subscribe), and [reading and organizing your library](/manual/olares/wise/basics), with updated filtering and cookie management instructions.
- Updated [accelerator resource guidance](/manual/olares/settings/gpu-resource) for the current GPU mode labels and interface. The upgrade guide explains when to restart existing GPU time-slicing apps after updating.
- Updated [password and device management](/manual/password-and-devices) to explain that changing your password signs out all devices, including the current device.
- Reworked [Olares CLI onboarding](/developer/cli-install) around the setup wizard, login, Agent Skills, and CLI updates. Consolidated the separate login guide into the onboarding guide.
- Updated [multi-node installation](/manual/best-practices/install-olares-multi-node), [connecting two Olares One devices](/one/connect-two-olares-one), and [multi-node upgrades](/one/upgrade-multi-node-cluster). Added stopping and uninstalling Olares to the DGX Spark installation guides.
- Expanded [Minecraft](/use-cases/minecraft) with version selection, mods, private servers, and app cloning. Updated [Hermes Agent](/use-cases/hermes) with Webapp login and task delegation, and [Claude Code](/use-cases/claude-code) with local model compatibility and output token configuration.

### Retired docs

- Retired the Control Hub host terminal guides after the **Terminal > Olares** entry was removed. Previous links redirect to [SSH access instructions](/manual/access-olares-terminal); container terminals remain available.
- Removed custom-domain setup instructions from [Customize application URLs](/manual/olares/settings/custom-app-domain). The guide now covers custom route IDs.
- Retired outdated Olares ID contract, DID implementation, and verifiable credential development guides. Previous URLs redirect to the relevant [Olares ID](/developer/concepts/olares-id), [DID](/developer/concepts/did), [reputation](/developer/concepts/reputation), or [verifiable credential](/developer/concepts/vc) concepts.
- Retired the Wise recommendation guide. Previous links redirect to the [Wise overview](/manual/olares/wise/index).

## Olares 1.12.7

Released: September 28, 2026

For the software changes, see the [Olares 1.12.7 release notes](https://github.com/beclab/Olares/releases/tag/1.12.7).
For release highlights and walkthroughs, see the [Olares 1.12.7 blog post](https://www.olares.com/blog/olares-1-12-7/).

### New docs

- Added [What to do after updating to Olares 1.12.7](/manual/update-guides/1.12.7), covering the upgrade path, Router and Lares setup, and agent app updates.
- Added [Use Olares Router as your AI gateway](/use-cases/olares-router), introducing model and tool capabilities, caller authentication, model naming, and connection details.
- Added [Manage Olares and run research tasks with Lares](/use-cases/lares), covering first tasks, permissions, and web research tools configured through Router.
- Added [Install the LarePass browser extension](/manual/install-larepass-browser-extension) and [Translate webpages privately with LarePass](/manual/tutorial/translate-webpages-with-larepass), for translating pages with a local model through Router.
- Added [Save resources to Olares](/manual/larepass/save-resources-to-olares), covering Fetch in LarePass Desktop for web links, torrents, Hugging Face repositories, and supported media.
- Added [Create and manage support tickets in Olares Space](/manual/space/tickets) and expanded [Get support](/manual/help/request-technical-support) with the Ticket app, log collection, and follow-up replies.
- Added [Recover access to LarePass, Olares Desktop, or your Olares device](/manual/help/ts-access-without-mnemonic), alongside updated [password reset instructions](/manual/help/ts-forget-login-password).
- Added [Create a Blender scene with Lares](/use-cases/blender-lares) and [Create videos with Concat and Lares](/use-cases/concat), covering MCP setup, creative workflows, and saving outputs in Files.
- Added [FlowStudio](/use-cases/flowstudio) and [FlowStudio through Lares](/use-cases/flowstudio-lares) guides for image and video generation.
- Added [Open Design](/use-cases/open-design), [Codex CLI](/use-cases/codex-cli), and [Manage Olares from OpenCode](/use-cases/opencode-olares-cli) tutorials for design, coding, and device management.
- Added [Minecraft](/use-cases/minecraft) and [Palworld](/use-cases/palworld) server guides with Overlay Gateway setup for local network access.

### Updated docs

- Rewrote [Connect your apps to AI capabilities](/manual/best-practices/connect-ai-apps) for Router, covering API formats, model and tool setup, default system names such as `default-chat`, and API keys for external clients.
- Updated AI client tutorials for the Router workflow, using Qwen3.8-27B (llama.cpp) as the chat example. Added an on-page version selector to affected guides so readers can choose the Olares 1.12.7 or 1.12.6 instructions.
- Reorganized the [Olares manual](/manual/overview) around user tasks and [installation guidance](/manual/get-started/install-olares) around Linux, Olares One, and DGX Spark setup paths. Updated [Olares One onboarding](/one/olares-onboarding) for Lares.
- Updated [local access guidance](/manual/best-practices/local-access) with LarePass host mappings on Windows and macOS, alongside VPN and `.local` access options.
- Updated [Market](/manual/olares/market/market) with app and model discovery and source selection, and [language settings](/manual/olares/settings/language-appearance) with the additional language options.
- Reworked [Olares Space account management](/manual/space/manage-accounts), [resource and traffic monitoring](/manual/space/manage-olares), and [billing](/manual/space/billing) instructions.
- Reworked [eGPU guidance for Olares One](/one/egpu) with separate Olares OS and Windows setup guides, compatibility information, and troubleshooting. The [Olares One FAQ](/one/faq#does-olares-one-support-out-of-band-management) now explains which out-of-band management capabilities are available.
- Updated [Jellyfin](/use-cases/jellyfin) hardware acceleration, [Steam streaming](/use-cases/steam-stream) network setup, and [Penpot](/use-cases/penpot) integration with its built-in MCP server.

### Retired docs

- Retired the standalone Ollama use-case guide, following the deprecation announced in Olares 1.12.6. Use [Ollama Engine Base](/use-cases/llm-base-apps) for new model deployments. Existing Ollama guide URLs redirect to the Engine Base guide.

### Moved and consolidated docs

- Consolidated backup and restore instructions in [Back up and restore Olares](/manual/olares/settings/backup). Previous restore and Olares Space backup links redirect to this guide.
- Split the former My Olares topic into [password and device management](/manual/password-and-devices) and [Olares One hardware settings](/one/hardware-settings).
- Consolidated Olares ID creation in [Create an Olares ID](/manual/get-started/create-olares-id), and moved integration instructions into task-specific guides such as [Mount cloud storage](/manual/olares/files/mount-cloud-storage). Previous links redirect to the replacement pages.

## Olares 1.12.6

Released: July 23, 2026

For the highlights and walk-through of Olares v1.12.6, see the [Olares 1.12.6 blog post](https://www.olares.com/blog/olares-1-12-6/).

### New docs

- Added [Olares CLI overview](/developer/cli-overview), an introduction to managing Olares from the command line.
  - Added [Use Olares CLI with AI agents](/developer/cli-install), covering installation, login, and the first AI agent task.
  - Added [Olares CLI Agent Skills](/developer/cli-agent-skills), describing the built-in skills for cluster management, apps, settings, and more.
- Added [Install and activate Olares from the command line](/manual/best-practices/activate-olares-using-cli), for setting up a device with Olares CLI.
- Added [Manage shared AI models with the Common directory](/manual/olares/files/files-common), explaining how to use the Common directory for models shared across apps.
- Added [Compress and extract files](/manual/olares/files/compress-extract-files), covering ZIP, 7z, TAR, and password-protected archives in Olares Files.
- Added [Mount NFS shares](/manual/olares/files/mount-nfs), for accessing NFS shared directories from Olares.
- Added [About shared applications](/manual/olares/market/shared-apps), explaining the new shared app architecture, and [Migrate legacy shared applications](/manual/migrate-shared-apps), covering migration options for v2 apps.
  - Ollama
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

- Added a deprecation notice to the Ollama use-case guide because the standalone Ollama app was removed from Market in Olares 1.12.6.
- Retired the Olares CLI reference pages. The deprecated pages have been superseded by the [Olares CLI overview](/developer/cli-overview), [Use Olares CLI with AI agents](/developer/cli-install), and [Olares CLI Agent Skills](/developer/cli-agent-skills).
- Retired all Studio documentation. Studio is no longer available in Market. Use [Olares CLI Agent Skills](/developer/cli-agent-skills) and the [App submission guide](/developer/develop/submit-apps) for the new packaging and porting workflow.
- The **Allow subnet routing** feature is temporarily disabled in Olares 1.12.6, and the related documentation has been removed from [Configure VPN access to Olares](/manual/olares/settings/remote-access#allow-subnet-routing). The feature will return in a future release.
- Retired the DeerFlow use-case guide. The DeerFlow app was replaced by [DeerFlow 2.0](/use-cases/deerflow2).

### Upcoming deprecations

- The Ollama use-case guide will be removed in a future release. The standalone Ollama app was already removed from Market in Olares 1.12.6, and AI use-case guides now connect to models through the Model Console.
