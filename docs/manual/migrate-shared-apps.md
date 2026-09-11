---
outline: [2, 3]
description: Choose and complete the correct migration path for a legacy v2 shared application installed before Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, migrate shared applications, v2 shared app, ComfyUI, Dify, OnlyOffice, SearXNG, Xinference, Ollama
---

# Migrate legacy shared applications

Use this guide if a shared application was installed before Olares 1.12.6 and you want to move to its current shared-app architecture. Legacy v2 apps continue to run after the Olares update, but they cannot be upgraded in place.

For an explanation of the current architecture, see [About shared applications](olares/market/shared-apps.md).

:::warning Check app data before uninstalling
Uninstalling a legacy shared app can remove data that the new app cannot restore directly. Identify the app and complete the applicable export or backup steps before uninstalling it.
:::

## Identify a legacy v2 app

Open the app details page in Market and check **Compatibility** in the **Information** panel:

- A legacy v2 shared app typically shows `Olares >=1.12.3-0, <1.12.6`.
- A current shared app shows `Olares >=1.12.6-0`.

If the app does not match these ranges, do not use this guide. Record the app and chart versions before choosing a migration path.

## Choose a migration path

| App or data state | Migration path |
|---|---|
| ComfyUI | Use the app's automatic migration workflow |
| Falco, MTranServer, or another app with no user-created data | Confirm that no data must be preserved, then reinstall |
| Dify, OnlyOffice, SearXNG, or Xinference | Export or back up app-specific data, then restore it into the new app |
| Standalone Ollama shared app | Re-create the required model on an Engine Base app, then reconnect clients |

## Migrate ComfyUI

Follow [ComfyUI migration notes](/use-cases/comfyui-common-issues.md) to run the supported migration and confirm the new data locations before removing the old app.

## Reinstall an app with no data to preserve

Use this path only after confirming that the app contains no user-created data, settings, or workflows that you need.

1. Record the legacy app name and version.
2. Uninstall the legacy v2 app.
3. Find the current shared app in Market and confirm that its compatibility range starts with `Olares >=1.12.6-0`.
4. Install the current shared app and open it from the Launchpad.

## Migrate app data manually

The data that can be preserved differs by app. Complete the app-specific guide before uninstalling the legacy app:

- [Migrate Dify](/use-cases/dify-upgrade.md)
- [Migrate OnlyOffice](/use-cases/onlyoffice-migration.md)
- [Migrate SearXNG](/use-cases/searxng.md)
- [Migrate Xinference](/use-cases/xinference.md)

Do not restore an old app data directory into a new version unless its migration guide explicitly requires it.

## Move from Ollama to Engine Base

1. Note the models and client applications that use the standalone Ollama shared app.
2. Deploy each required model with an [Engine Base app](/use-cases/llm-base-apps.md).
3. Open the model console and copy the new **Base URL**.
4. Replace the Ollama endpoint in each client application with the new Base URL.
5. Send a test request from each client. Remove the old Ollama app only after the clients use the new model successfully.
