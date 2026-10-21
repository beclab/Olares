---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/hermes
outline: [2,3]
description: Install Hermes Agent on Olares, connect it to a local model and Discord, manage LAN computers over SSH, and integrate it with other applications through the Gateway API.
head:
  - - meta
    - name: keywords
      content: Olares, Hermes, Hermes Agent, autonomous AI, self-improving AI, Discord bot, self-hosted, SSH, LAN automation
app_version: "1.3.49"
doc_version: "3.2"
doc_updated: "2026-10-21"
---

# Set up a self-directed AI agent with Hermes

<VersionRouteSelect />

Hermes Agent is a self-directed AI assistant that connects to your local models to execute system tasks, generate code, and manage workflows. It retains memory across sessions and creates reusable skills based on your interactions.

In Olares, you can use the built-in skills that allow your agent to manage your device's files and applications. You can also interact with your agent remotely via messaging platforms like Discord, or integrate it with other applications like Open WebUI.

## Learning objectives

In this guide, you will learn how to:
- Install Hermes Agent on Olares.
- Configure Hermes Agent to connect to a local model.
- Interact with your agent directly via the terminal.
- Integrate with Discord for remote chat.
- Use Olares skills to manage files and applications via the agent.
- Run commands and scripts over SSH and coordinate tasks across LAN computers.
- Enable the Gateway API to connect Hermes Agent with other applications like Open WebUI.

## Prerequisites

Before you begin, you need:

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.
- The following model:

    | Model type | Model | How to get it |
    | :--- | :--- | :--- |
    | Chat | Qwen3.8-27B (llama.cpp) | Install from Market |

    :::tip
    Hermes Agent requires a model with a context window of at least 64K tokens.
    :::

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install Hermes Agent

1. Open Market and search for "Hermes".

   ![Install Hermes Agent](/images/manual/use-cases/hermes-agent.png#bordered)

2. Click **Get**, and then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:

    - Hermes CLI: The command line interface
    - Dashboard: The graphical dashboard

    ![Hermes entry points](/images/manual/use-cases/hermes-entry-points.png#bordered){width=50%}

:::tip Run multiple Hermes agents
Olares supports app cloning. If you want to run multiple independent AI agents for different tasks, you can clone the Hermes Agent app. For more information, see [Clone applications](../manual/olares/market/clone-apps.md).
:::

## Configure Hermes Agent

Run a quick setup to connect Hermes Agent to your local model.

### Step 1: Get model connection details

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

5. In Router, open **LLM** and click the information icon on the Qwen3.8-27B (llama.cpp) row to open **Model card**. Under **Engine args**, read the number after `-c`. For example, `-c 104448` means you should enter `104448` in the next step. See [Check the configured context size](/manual/best-practices/connect-ai-apps.md#check-context-window) for screenshots.

### Step 2: Run the setup wizard

1. Open the Hermes CLI app from the Launchpad.
2. Enter the following command to start the configuration wizard:

   ```bash
   hermes setup
   ```

3. The wizard guides you through a series of steps. Use the arrow keys to navigate and press **Enter** to confirm.

    The model-selection prompts below are alternative branches. Follow the row that matches the number of models detected, then continue to **Context length in tokens**.

    | Settings   | Option   |
    |:-----------|:---------|
    | How would you like to set up Hermes | Select **Full setup — configure every provider, tool & option yourself (bring your own keys)**. |
    | Select provider | Select **Custom endpoint (enter URL manually)**.  |
    | API base URL  | Enter the **Base URL** you copied from Router.<br>For example, `https://router.<your-olares-domain>/v1`.  |
    | API key  | Enter any text as a placeholder value, such as `local`.<br>The input remains hidden for security. |
    | Select API compatibility mode | Enter `1` to select **Auto-detect**. |
    | Use this model? [Y/n] — only when one model is detected | Enter `n`. At the following **Model name** prompt, enter `default-chat`. This confirmation accepts yes/no, not a model name. |
    | Select model [1-N] or type name — only when multiple models are detected | Type `default-chat` instead of selecting a numbered model. |
    | Model name — when no model is detected | Enter `default-chat`. |
    | Context length in tokens | Enter the exact context size from Router's **Model card > Engine args**, such as `104448` for `-c 104448`. Hermes requires at least `65536` tokens. The value must not exceed the engine configuration. |
    | Display name |  Enter a name to identify this model, such as `router-chat`.|
    | Select terminal backend | Select **Local - run directly on this machine**. |
    | Select platforms to configure | Press **Enter** to skip for now. |
    | Tools for CLI | Press **Enter** to skip for now. |
    | Choose a provider | Select **Skip - keep defaults / configure later**. |
    | Select Search Provider | Select **Skip - keep defaults / configure later**. |

4. Keep the Hermes CLI window open for the next step.

## Interact with Hermes Agent

### Option 1: Terminal chat

The Terminal User Interface (TUI) runs directly in the Hermes CLI with no extra setup. It is ideal for quick tests.

1. When the setup wizard completes, it shows `Ready to go` with the available commands:

    - `hermes`: Start chatting.
    - `hermes gateway`: Start the messaging gateway.
    - `hermes doctor`: Check for issues.

2. Type `hermes` and press **Enter** to launch the TUI.

   :::tip
   If you exited the wizard, manually start the chat by entering `hermes` in the Hermes CLI.
   :::

3. Send a message, such as `what can you do`, to verify that the agent responds correctly.

    ![Hermes TUI chat](/images/manual/use-cases/hermes-tui1.png#bordered)

4. To close the TUI and return to the standard CLI, type `/exit`, and then press **Enter**.

### Option 2: Remote chat via Discord

To chat with your agent remotely, connect it to a Discord bot.

#### Step 1: Create a Discord bot

1. Log in to the [Discord Developer Portal](https://discord.com/developers/home) with your Discord account.
2. Select **Applications** from the left sidebar, and then click **New Application**.

    ![New application in Discord developer portal](/images/manual/use-cases/hermes-new-app.png#bordered)

3. Enter a name for the new app, agree to the terms, and then click **Create**.

    ![Create an application window](/images/manual/use-cases/hermes-create-app.png#bordered){width=40%}

4. From the left sidebar, select **Bot**.
5. Scroll down to the **Privileged Gateway Intents** section and enable the following settings:

    - Presence Intent
    - Server Members Intent
    - Message Content Intent

    ![Privileged gateway intents](/images/manual/use-cases/hermes-privileged-gateway-intents.png#bordered)

6. Click **Save Changes**.
7. Scroll up to the **Token** section, click **Reset Token**, and then copy the generated token for your Discord bot. You need the token for messaging platform configuration later in the Hermes CLI.

    ![Reset token](/images/manual/use-cases/hermes-reset-token.png#bordered)

#### Step 2: Invite the bot to your server

1. From the left sidebar, select **OAuth2**, and then find the **OAuth2 URL Generator** section:

    a. In **Scopes**, select **Bot** and **applications.commands**.

    ![OAuth2 URL Generator](/images/manual/use-cases/hermes-oauth21.png#bordered)

    b. Scroll down to **Bot Permissions**, set the permissions as shown in the following image. You can modify the settings later.

    ![Bot permissions](/images/manual/use-cases/hermes-bot-permissions.png#bordered)

2. Copy the **Generated URL** at the bottom.
3. Paste the URL into a new browser tab, select your Discord server from **Add to server**, click **Continue**, and then click **Authorize**. 

    ![Add Discord bot to server](/images/manual/use-cases/hermes-add-bot.png#bordered){width=50%}

    The bot is authorized and added to your server.

    ![Bot added to server](/images/manual/use-cases/hermes-bot-added.png#bordered)

#### Step 3: Configure the messaging platform

Connect Hermes Agent to your Discord bot by configuring and running the Hermes gateway.

1. Open the Hermes CLI, and then enter the following command to start the configuration wizard:

   ```bash
   hermes gateway setup
   ```

2. Select **Discord** as your messaging platform.
3. Follow the prompts to configure the bot integration:

   - **Discord bot token**: Enter the bot token generated from your Discord Developer Portal. The input remains hidden.
   - **Allowed user IDs or usernames**: Enter your Discord user ID to restrict access to yourself.
   - (Optional) **Home channel ID**: Enter the ID of the Discord channel where the bot operates. 

4. Select **Done**.
5. When prompted to **Restart the gateway to pick up changes**, enter `y`.
6. Check your bot status in Discord. A green status icon confirms the bot is online, which means the gateway restarted successfully and the configuration is correct.

#### Step 4: Authorize your account

For security, the bot does not talk to unauthorized users. You must pair your Discord account with the bot.

1. Send a Direct Message to your new bot. The bot will reply with an error message containing a pairing code.
2. Open the Hermes CLI and enter the following command:

    ```bash
    hermes pairing approve discord {Your-Pairing-Code}
    ```

3. Once approved, you can start chatting with your agent in Discord. If you are talking to it in a channel, mention it first.

## Manage Olares with your Hermes Agent

Use the pre-installed Olares CLI skills directly in Hermes Agent to manage files and applications on your Olares device. For example, ask it to list files, read logs, or install apps from Olares Market.

1. Open the Hermes CLI from the Launchpad.
2. Run the following commands one by one to confirm that both `olares-cli` and its skills are properly installed and enabled:

   ```bash
   olares-cli -v
   hermes skills list
   ```
   
3. Before using the skills, run the following command to log in to your Olares account. Replace {your-olares-id} with your Olares ID (e.g. olaresdemo@olares.com):

   ```bash
   olares-cli profile login --olares-id {your-olares-id}
   ```

4. Follow the prompt to enter your Olares login password. For security, the password is hidden.
5. You can now enter the TUI, and ask the agent to interact with your Olares environment. For example, ask it to install an app from Olares Market.

## Manage LAN computers over SSH

Hermes can connect to other LAN computers over SSH to run commands and scripts, manage services, and coordinate operations across multiple hosts. Specify the destination and task in each prompt; no fixed destination IP is needed in the configuration. Available operations depend on the remote account's permissions.

Keep **Local** as the terminal backend in Hermes. Here, local means the Hermes app container: the agent runs `ssh` inside the container, and SSH runs commands on the target computer. This workflow uses the standard SSH command and does not require an enabled Gateway API or a custom SSH tool.

### Prepare the connection

Check the following:

- The Olares device can reach the target computer's SSH port, which defaults to `22`.
- The target computer has a running SSH server and allows your account to log in.
- Hermes is connected to an available model and uses the **Local** terminal backend.

This example uses the Linux host `ubuntu@192.168.1.50`. Replace the username, address, and port with your own settings.

:::tip Where does the connection start?
The SSH connection starts from the Hermes container on Olares. Opening Hermes from your Mac does not automatically give the container access to your Mac's network. On the tested setup, outgoing SSH worked without host networking or privileged containers. Routing between subnets, guest-network isolation, or firewall rules can still block a connection.
:::

### Sign in with a password

If the target allows password authentication, you can sign in manually from Hermes CLI without setting up a key first.

1. Open Hermes CLI from Launchpad. If the chat interface is open, enter `/exit` to return to the shell.

2. Prepare the persistent SSH directory and connect:

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ssh -o PubkeyAuthentication=no \
     -o PreferredAuthentications=password,keyboard-interactive \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     ubuntu@192.168.1.50
   ```

   For a different SSH port, add `-p PORT`. On the first connection, check the host-key fingerprint with the target computer's administrator before entering `yes`.

3. At the `ubuntu@192.168.1.50's password:` prompt, enter the remote account's password and press **Enter**. The input stays hidden. Once connected, you can run commands allowed for that account. Enter `exit` to close the remote session when finished.

This example is an interactive sign-in from Hermes CLI. The standard SSH command does not automatically answer password prompts in background tasks. For tasks the agent runs independently, configure key-based authentication below. Do not put passwords in commands, scripts, or chat prompts.

### Configure key-based authentication

Key-based authentication suits automated tasks. After you authorize the public key once, Hermes can connect with the private key without asking for the account password each time.

#### Step 1: Store a key in persistent app data

1. Open Hermes CLI from Launchpad. If the chat interface is open, enter `/exit` to return to the shell.

2. Create a key in the persistent app-data directory:

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ssh-keygen -t ed25519 -C "hermes-lan" -f /opt/data/.ssh/id_ed25519
   ```

   Keep an existing key instead of overwriting it. For unattended tasks, the key must be usable without a passphrase prompt. Press **Enter** twice to leave the passphrase empty when generating a dedicated automation key.

3. Restrict access to the private key and display the public key:

   ```bash
   chmod 600 /opt/data/.ssh/id_ed25519
   cat /opt/data/.ssh/id_ed25519.pub
   ```

   Copy the complete public-key line. Keep the private key inside the app; do not paste it into a chat.

:::tip Existing PEM key
You can also use an existing private key. In Files, open **Data** > **hermesagent** > **home**, show hidden files, and upload it to `.ssh` as `lan-host.pem`. In Hermes CLI, run `chmod 600 /opt/data/.ssh/lan-host.pem` and use that path in place of `id_ed25519` below. A `.pem` filename is not required for SSH authentication. To generate an RSA key in PEM format, run `ssh-keygen -t rsa -b 3072 -m PEM -f /opt/data/.ssh/lan-host.pem`, then install its matching `.pub` key on the target computer.
:::

Files under `/opt/data` survive app restarts and container recreation. Back up the keys before uninstalling the app or deleting its data. App clones have their own data; set up keys for each clone that needs SSH access.

#### Step 2: Authorize the key on the target computer

1. On the target computer, open a terminal as the account Hermes will use, such as `ubuntu`. Prepare its SSH directory:

   ```bash
   mkdir -p "$HOME/.ssh"
   chmod 700 "$HOME/.ssh"
   touch "$HOME/.ssh/authorized_keys"
   chmod 600 "$HOME/.ssh/authorized_keys"
   ```

2. Open `$HOME/.ssh/authorized_keys` in a text editor. Append the public-key line from Hermes on a new line, preserving existing keys, and save the file. The directory and file must belong to this target account.

3. Read the target's SSH host-key fingerprint for the first connection:

   ```bash
   ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub
   ```

#### Step 3: Check the first connection

1. In Hermes CLI, connect once interactively and store the host key in persistent app data:

   ```bash
   ssh -i /opt/data/.ssh/id_ed25519 \
     -o IdentitiesOnly=yes \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     ubuntu@192.168.1.50 'hostname; id'
   ```

   Compare the fingerprint with the one shown on the target computer before accepting it. For a different SSH port, add `-p PORT`.

2. Check that the same command works without a password prompt:

   ```bash
   ssh -i /opt/data/.ssh/id_ed25519 \
     -o IdentitiesOnly=yes \
     -o BatchMode=yes \
     -o StrictHostKeyChecking=yes \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     -o ConnectTimeout=10 \
     ubuntu@192.168.1.50 'hostname; id'
   ```

   The command should return the target hostname and account information without a password prompt. `BatchMode=yes` returns an error when automatic authentication fails, so a task does not wait at a password prompt.

Use absolute paths for both the key and `known_hosts`, so the agent can find them even when its terminal has a different home directory. The host key is also stored in `/opt/data/.ssh/known_hosts` and remains available after an app restart.

### Assign remote tasks in natural language

Start a chat in the TUI or Dashboard and provide the destination, account, and key path. For example:

```text
Connect to ubuntu@192.168.1.50 over SSH using /opt/data/.ssh/id_ed25519.
Use /opt/data/.ssh/known_hosts, strict host-key checking, batch mode,
and a 10-second connection timeout. Confirm the target host and account,
then run commands or scripts according to the instructions I provide next.
Explain the steps and their impact before executing them, and summarize
the results afterward. Report permission or connection errors.
```

For another computer, repeat key authorization and the first-connection check, then change the destination in your prompt. You can use a separate private key for each computer. No fixed destination IP is needed in the Hermes terminal-backend configuration.

Once the computers are configured, you can assign a task across multiple hosts in the same conversation. Replace the placeholder with your actual task:

```text
Complete this task on ubuntu@192.168.1.50 and olares@192.168.1.60:
[Describe the goal, commands or scripts, and execution order across hosts].
Use /opt/data/.ssh/id_ed25519 for both and
/opt/data/.ssh/known_hosts for host-key verification.
Confirm each target host and account before following the specified order.
Organize each step's results by hostname and IP address. If a host fails,
report the cause and completed actions, then let me decide whether to
continue with the remaining steps.
```

:::info Remote sudo
Permissions on the remote computer come from its SSH account. Ask Hermes to use `sudo -n` for commands that require elevation. It succeeds only when that account is allowed to run the command without a password; otherwise, it returns an error instead of waiting for input. Do not put a sudo password in the prompt. This is separate from sudo inside the Hermes container.
:::

### Troubleshoot SSH connections

| Symptom | What to check |
|:---|:---|
| Connection times out or is refused | Check the address, SSH port, server status, firewall, and routing from the Olares device to the target. Opening an Olares app entrance does not enable outgoing SSH. |
| `Permission denied (publickey)` | Check the remote username, matching public key in that account's `authorized_keys`, and private-key path. Keep the private key at mode 600 and `.ssh` at mode 700. |
| `Host key verification failed` | Complete the first connection interactively with the same `known_hosts` path. If a saved fingerprint changes, confirm the cause on the target before updating the saved entry. Keep host-key checking enabled. |
| `sudo: a password is required` | The account does not have passwordless permission for that command. Run an allowed read-only command or ask the host administrator for the required access. |

## Integrate Hermes Agent with other applications

Hermes Agent provides an OpenAI-compatible API, allowing you to integrate it with other applications like Open WebUI or Hermes Workspace.

### Step 1: Enable the Gateway API

1. Open Olares Settings, and then go to **Applications** > **Hermes Agent** > **Manage environment variables**.

    ![Hermes Agent environment variables](/images/manual/use-cases/hermes-env-var.png#bordered)

2. Find **API_SERVER_ENABLED**, click <i class="material-symbols-outlined">edit_square</i>, and then set its value to `true`.
3. Click **Confirm**.
4. Find **HERMES_API_SERVER_KEY**, click <i class="material-symbols-outlined">edit_square</i>, and then enter the value. Ensure the key meets the following requirements:

    - It must be at least 16 characters long.
    - It can only contain letters, numbers, and common symbols.
    - It cannot be a predictable placeholder value, such as "your_API_key".
  
5. Click **Confirm**, and then click **Apply**.

### Step 2: Get the Gateway API URL

1. Go to **Applications** > **Hermes Agent** > **Hermes Gateway API**.
2. Copy the endpoint URL. For example:
   ```
   https://baf3d7172.olaresdemo.olares.com
   ```

### Step 3: Verify the API is running

Open your browser and go to your gateway URL appended with `/health`:

```
https://{your-gateway-url}/health
```

If the API is enabled successfully, the page displays a response confirming the service is active. For example, `{"status": "ok", "platform": "hermes-agent"}`.

### Step 4: Connect your application

The following steps demonstrate how to connect with Open WebUI.

1. In Open WebUI, click your profile icon and select **Admin Panel**.
2. Go to **Settings** > **Connections**.
3. On the right of **Manage OpenAI API Connections**, click <i class="material-symbols-outlined">add</i> to add a new connection.

    ![Hermes Agent config in Open WebUI](/images/manual/use-cases/hermes-integrate-openwebui.png#bordered)
4. Configure the settings as follows:
    - **API Base URL**: Enter your Hermes Gateway API URL and append `/v1`. For example, `https://baf3d7172.olaresdemo.olares.com/v1`.
    - **Auth**: Select **Bearer**, and then enter the `HERMES_API_SERVER_KEY` you set previously.
5. Click <i class="material-symbols-outlined">refresh</i> to check the connection, and then click **Save**.
6. Go to the **New chat** page and verify that **hermes-agent** appears in the model drop-down list. You can start chatting with it now.

## Advanced configuration

To manually adjust parameters, edit the configuration files directly, and then restart the Hermes CLI to apply the changes. 

1. Open the Files app from the Launchpad.
2. Go to **Data** > **hermesagent** > **home**, and locate the configuration files, such as `config.yaml` and `.env`. 

The file structure and configuration options match the official defaults. For detailed parameter descriptions, see the [Hermes configuration guide](https://hermes-agent.nousresearch.com/docs/user-guide/configuration).

## FAQs

### How to manually restart the Hermes gateway?

You can restart the gateway manually using one of the following methods:

- **Use the Hermes CLI**

    This is the fastest method. Open the Hermes CLI from the Launchpad, and then run the following command:

    ```bash
    restart-gateway
    ```
- **Use the Hermes Dashboard**

    This is the most intuitive method. Open the Hermes Dashboard from the Launchpad, and then select **Restart Gateway** in the left sidebar.

- **Use Control Hub**

    Open Control Hub, and then go to **Browse** > **{username}** > **hermesagent-{username}** > **Deployments** > **hermesagent**, and then click **Restart** in the upper-right corner. This method completely restarts the entire service, which takes slightly longer.

### Why does the gateway fail to start after configuring the API?

Your gateway will fail to start if the configured API key does not meet Hermes security requirements. When this happens:
- The Gateway API displays an `upstream connect error`.
- The system logs show the message `[Api_Server] Refusing to start: API_SERVER_KEY is set to a placeholder value`.

To resolve this issue, reset your `HERMES_API_SERVER_KEY`:
1. Open Olares Settings, and then go to **Applications** > **Hermes Agent** > **Manage environment variables**.
2. Set a new key for `HERMES_API_SERVER_KEY` that meets the following requirements:

    - It must be at least 16 characters long.
    - It can only contain letters, numbers, and common symbols.
    - It cannot be a predictable placeholder value, such as "your_API_key".
    
 3. Click **Save**, and then click **Apply**.

## Next steps

After you complete the basic setup, explore the official resources to expand your agent's capabilities:
- For advanced terminal commands, see [CLI Interface](https://hermes-agent.nousresearch.com/docs/user-guide/cli).
- For integration with Slack, Telegram, and other platforms, see [Messaging Platforms](https://hermes-agent.nousresearch.com/docs/user-guide/messaging/).
- For installation of skills, see [Skills System](https://hermes-agent.nousresearch.com/docs/user-guide/features/skills).
- For installation of tools, see [Tools & Toolsets](https://hermes-agent.nousresearch.com/docs/user-guide/features/tools).
- For information about best practices and optimization, see [Tips & Best Practices](https://hermes-agent.nousresearch.com/docs/guides/tips).

:::info Note on sudo commands
The Hermes container runs as an unprivileged user. Do not use instructions that require local `sudo` inside the container. Commands executed over SSH use the remote account's permissions; see [Manage LAN computers over SSH](#manage-lan-computers-over-ssh).
:::

## Learn more

- [How do I create a server in Discord](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)
- [How to find the Channel ID number](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID)
- [OpenClaw](openclaw.md)
