---
outline: [2, 3]
description: Troubleshoot common issues and find answers to frequently asked questions about OpenClaw on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, troubleshoot, FAQ, common issues, errors
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-10"
---

# Common issues

This page provides solutions to common issues and answers to frequently asked questions when running OpenClaw on Olares.

If you encounter a problem not listed here, check the [Upgrade OpenClaw](openclaw-upgrade.md) page for version-specific changes or refer to the [official OpenClaw documentation](https://docs.openclaw.ai).

### Cannot restart OpenClaw in CLI

If you attempt to manually start, stop, or restart OpenClaw using commands like `openclaw gateway` or `openclaw gateway stop` in the OpenClaw CLI, you receive the following error messages:
- `Gateway failed to start: gateway already running (pid 1); lock timeout after 5000ms`
- `Gateway service check failed: Error: systemctl --user unavailable: spawn systemctl ENOENT`

#### Cause

OpenClaw is deployed as a containerized app in Olares, where the gateway runs as the primary container process `pid 1` and is always active. This environment does not use standard Linux system and service management tools such as `systemd` and `systemctl`, so these commands do not work. 

#### Solution

Do not use the OpenClaw CLI to manage the gateway service. Instead, restart OpenClaw using one of the following methods.

**Method 1: Restart OpenClaw from Settings or Market**
    
- Open **Settings**, go to **Applications** > **OpenClaw**, click **Stop**, and then click **Resume**.
- Open **Market**, go to **My Olares**, find **OpenClaw**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the operation button, select **Stop**, and then select **Resume**.

**Method 2: Restart the container**

Open **Control Hub**, click `clawdbot` under **Deployments**, and then click **Restart**.

### OpenClaw automatically stops during long tasks

When you ask the OpenClaw agent to perform tasks that take a long time to process like massive web scrapes or deep analysis, the task is abruptly terminated before returning the result.

#### Cause

By default, OpenClaw sets a maximum runtime limit of 10 minutes per task. If a task exceeds this limit, the system forcefully terminates it to save resources.

#### Solution

Extend this timeout limit by modifying the configuration file as follows:
1. Open the Control UI, go to **Config** > **Raw**, and then find the `agents` section.
2. In the `defaults` block, add the `timeoutSeconds` field or modify the existing one in it. 

    To set it to 1 hour, specify `3600` for the value:

    ```json
    "agents": {
        "defaults": {
            "timeoutSeconds": 3600
        }
    }
    ```
3. Click **Save** to restart the gateway and apply the changes.

### "Rate limit exceeded" error when installing skills

Installing a skill fails with a `429` error:

```text
Downloading xurl@1.0.0 from ClawHub...
ClawHub /api/v1/download failed (429): Rate limit exceeded
```

#### Cause

The ClawHub registry temporarily limits downloads due to high traffic to maintain server stability.

#### Solution

Wait a few hours and run the installation command again.

## Model responds slowly

There is a noticeable delay before the agent begins typing its first response.

#### Cause

This usually happens due to the way Ollama manages system resources and application settings:
- **Automatic offloading**: To save resources, Ollama unloads models from memory by default when they are idle. The next time you interact with the model, it takes time to reload it, causing a noticeable delay in the first response.
- **Context setting clashes**: If you have multiple applications using the same model but with different context settings, Ollama is forced to constantly unload and reload the model to switch between those different configurations.

#### Solution

To fix the issue, try one of the following methods.

#### Method 1: Prevent automatic offloading for model apps

Keep the model permanently in memory by enabling the `KEEP_ALIVE` environment variable for your model app.

1. Open **Settings**, and then go to **Applications** > **{Your Model App}** > **Manage environment variables**.
2. Find **KEEP_ALIVE**, click <i class="material-symbols-outlined">edit_square</i>, set the value to **true**, and then click **Confirm**.

    ![Enable Keep Alive for model app in Settings](/images/manual/use-cases/keep-alive-enable.png#bordered){width=70%}

3. Click **Apply**.

#### Method 2: Unify context sizes across apps

Use the same context size for all apps that share the same model to reduce reload times.

1. Check the current context size of your running models:

    - In the Ollama terminal, run `ollama ps`. The `CONTEXT` column shows the context size in use.

        ![View model details in Ollama terminal](/images/manual/use-cases/ollama-ps.png#bordered)

    - For a standalone model app, check the context size using Control Hub:
    
        a. Under the **System** namespace, find the model app's project (typically named `{model-name}server-shared`), and then open its pod terminal.

        ![Open pod terminal in Control Hub](/images/manual/use-cases/pod-terminal-ctrl-hub.png#bordered)        
        
        b. Run `ollama run {model-name}`, press **Ctrl**+**D**, and then run `ollama ps`.

        ![View model details in Control Hub](/images/manual/use-cases/ollama-ps-ctrl-hub.png#bordered)

2. Set all apps to use the same context size.

## Clean reinstall OpenClaw

If you want to uninstall OpenClaw and start fresh, simply uninstalling the app is not enough. By default, Olares preserves your application data such as configurations and persona files, so you do not lose your work.

To completely remove OpenClaw and all of its data before reinstalling, follow the steps based on your Olares OS version.

<Tabs>
<template #V1.12.5-and-later>

1. Open **Market**, go to **My Olares**, and then find **OpenClaw**.
2. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the app's operation button, and select **Uninstall**.
3. In the **Uninstall** window, select **Also remove all local data**. Then app data (in the Data directory) and cache data (in the Cache directory) will be permanently deleted and cannot be recovered.
    ![Remove local app data option during uninstallation](/images/manual/use-cases/uninstall-remove-local-data.png#bordered){width=65%}
4. Click **Confirm**.
5. Return to Market and reinstall OpenClaw. It will now install from a completely clean state.
</template>
<template #V1.12.4-and-earlier>

1. Open **Market**, go to **My Olares**, and then find **OpenClaw**.
2. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the app's operation button, and select **Uninstall**.
3. When the installation finishes, open the **Files** app, and then go to **Application** > **Data**.
4. Find the `clawdbot` folder, right-click it, select **Delete**, and then click **Confirm**. This permanently removes all the  previous configurations and workspaces.
    ![Remove OpenClaw app data](/images/manual/use-cases/remove-app-data.png#bordered){width=80%}
5. Return to Market and reinstall OpenClaw. It will now install from a completely clean state.
</template>
</Tabs>

