---
outline: deep
description: Learn how to enable and configure the OpenClaw sandbox for secure code execution.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, sandbox, security, code execution
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-08"
---

# Optional: Enable sandbox

By default, OpenClaw executes commands and code directly within its primary container. While this is generally safe for everyday tasks, granting your agent the ability to run arbitrary code or install external dependencies carries inherent risks. 

To maximize security and isolate potentially dangerous operations, you can enable the OpenClaw sandbox. The sandbox provides an isolated, disposable environment specifically for code execution, ensuring your core system remains protected at all times.

:::tip Prerequisites
To use this feature, your system must meet the following requirements:
- **Olares OS**: Upgraded to V1.12.5 or later.
- **OpenClaw**: Upgraded to V0.1.31 or later.
:::

## Understand sandbox modes

When configuring the sandbox, the `mode` setting specifies when the sandbox is triggered: 
- **off**: The sandbox is disabled. All commands run in the main container.
- **non-main**: The sandbox isolates commands executed via external channels such as Discord. Commands executed directly in the Control UI's Chat page bypass the sandbox and run in the main container.
- **all**: All commands run inside the sandbox, regardless of which interface or channel you use.

## Enable sandbox

The OpenClaw sandbox is disabled by default. You can enable it by modifying the configuration file or by using the Control UI.

<Tabs>
<template #Enable-via-config-file>

1. Open the Files app, and then go to **Data** > **clawdbot** > **config**.
2. Double-click the `openclaw.json` file to open it.
3. Click <i class="material-symbols-outlined">edit_square</i> in the upper-right corner to enter the edit mode.
4. Locate the `agents` > `defaults` section, and then add the following `sandbox` configuration block into it.

    :::info Note for new vs. upgraded users
    <ul><li>If you are setting up a fresh installation, the sandbox block already exists, so you only need to change the value of <code>mode</code> from <code>off</code> to <code>non-main</code>.</li><li>If you upgraded from a previous version, you must paste the entire block.</li></ul>
    :::

      ```json
          "sandbox": {
            "mode": "non-main",
            "backend": "docker",
            "scope": "agent",
            "workspaceAccess": "rw",
            "docker": {
              "image": "beclab/harveyff-openclaw-sandbox-common:2026.4.7",
              "network": "bridge",
              "user": "1000:1000"
            },
            "prune": {
              "idleHours": 24,
              "maxAgeDays": 7
            }
          }
    ```  

    ![Enable sandbox via configuration file](/images/manual/use-cases/openclaw-edit-config-file.png#bordered)

5. Click <i class="material-symbols-outlined">save</i> in the upper-right corner. The system validates the configuration and applies the change automatically.
</template>
<template #Enable-via-Control-UI>

1. Open the Control UI, and then select **AI & Agent** from the left sidebar.
2. Scroll down to locate the **Sandbox** section, and then expand it.
3. Configure the settings as follows:
    - **Backend**: Enter `docker`.
    - Expand **Docker**:
      - **Image**: Enter `beclab/harveyff-openclaw-sandbox-common:2026.4.7`.
      - **Network**: Enter `bridge`.
      - **User**: Enter `1000:1000`.    
    - **Mode**: Select **non-main**.
    - Expand **Prune**:
      - **Idle Hours**: Enter `24`.
      - **Max Age Days**: Enter `7`.    
    - **Scope**: Select **agent**.
    - **Workspace Access**: Select **rw**.

    ![Enable sandbox in Control UI](/images/manual/use-cases/openclaw-sandbox-enable-ui.png#bordered)
    
4. Click **Save** in the upper-right corner. The system validates the configuration and applies the change automatically.

</template>
</Tabs>

## Use the sandbox

Once enabled, OpenClaw automatically creates and uses the isolated sandbox environment whenever it needs to execute commands.

To verify that the sandbox is working, test it using an external channel such as Discord:

1. Ensure that your [Discord is integrated](/use-cases/openclaw-integration.md).
2. In your Discord, send the following direct message to the agent:

    ```text
    Clone the repo [https://github.com/beclab/core](https://github.com/beclab/core), read the package.json, then summarize what version it is and list its dependencies
    ```
3. The agent will spin up the isolated sandbox to safely clone the repository, read the files, and return the summary. 
4. While the agent is working, open the OpenClaw CLI and run the following command to verify the active sandbox:

    ```bash
    openclaw sandbox list
    ```
    
    The terminal will display the currently running sandbox container, confirming the isolation is active.

    ![Verify sandbox](/images/manual/use-cases/openclaw-sandbox-verify.png#bordered)

## Grant additional directory access

By default, the `workspaceAccess: "rw"` setting only gives the sandbox access to OpenClaw's own workspace, which allows the agent to update its memory files.

If you want the sandbox to interact with your Olares files, you must explicitly grant it access using custom bind mounts. This mounts a specific directory directly into the temporary sandbox container.

### Grant access

For example, to grant the sandbox read-only (`ro`) access to your Home directory:
1. Ensure OpenClaw has access to your local files in the **Home** directory by enabling the `ALLOW_HOME_DIR_ACCESS` environment variable. For more information, see [Enable file access settings](../use-cases/openclaw-local-access.md#step-1-enable-file-access-settings).
2. Open the Files app, and then go to **Data** > **clawdbot** > **config**.
3. Double-click the `openclaw.json` file to open it.
4. Click <i class="material-symbols-outlined">edit_square</i> in the upper-right corner to enter the edit mode.
5. Locate the `agents` > `defaults` > `sandbox` > `docker` section.
6. Add the `binds` and `dangerouslyAllowExternalBindSources` lines to the configuration as follows. Ensure you added a comma after the preceding `"user": "1000:1000"` line to maintain valid JSON syntax.

    ```json
    "binds": ["/home/userdata/home:/home/userdata/home:ro"],
    "dangerouslyAllowExternalBindSources": true //Allows the sandbox to access directories outside the default workspace
    ```

    ![Grant read-only access to Home directory](/images/manual/use-cases/openclaw-sandbox-readonly.png#bordered)

7. Click <i class="material-symbols-outlined">save</i> in the upper-right corner to save the changes.
8. Restart OpenClaw for the changes to take effect.

### Test access

In the previous step, the sandbox mode is `non-main`, and the bind mount is set to `ro`. To understand how these settings work together, you can test from two different interfaces. 

#### Test the main session

Open the **Chat** page in the **Control UI**, and then send the following message:

```text
Write a self-instruction file in txt format, and save it to the Documents folder in my Olares
```

**Result**: The file is successfully created in the specified directory. 

![File creation success in specified directory](/images/manual/use-cases/openclaw-sandbox-file-created.png#bordered)

**Reason**: Commands sent through the Control UI's Chat page belong to the "main" session. Because you set the sandbox mode to `non-main`, this session bypasses the sandbox entirely. The agent uses OpenClaw's default system permissions to write the file.

#### Test a non-main session

Open your Discord, and then send a similar message:

```text
Write a sci-fi story outline in txt format, and save it to the Documents folder in my Olares Files
```

**Result**: The file creation fails. 

![File creation failure in specified directory](/images/manual/use-cases/openclaw-sandbox-file-failure.png#bordered)

**Reason**: Commands sent through external channels like Discord trigger the sandbox. Because you configured the sandbox with a read-only (`ro`) bind mount for the Home directory, the agent is blocked from writing or modifying any files.



## Learn more

- [OpenClaw Sandboxing documentation](https://docs.openclaw.ai/gateway/sandboxing#sandboxing)
- [Custom bind mounts](https://docs.openclaw.ai/gateway/sandboxing#custom-bind-mounts)
