---
outline: [2, 3]
description: Manage Linux computers on your LAN with Hermes Agent over SSH. Set up key-based authentication, check disk space across hosts, and troubleshoot connections.
head:
  - - meta
    - name: keywords
      content: Olares, Hermes Agent, AI agent, SSH, LAN automation, key-based authentication, disk space
app_version: "1.3.49"
doc_version: "1.0"
doc_updated: "2026-10-21"
---

# Manage LAN computers with Hermes Agent over SSH

Hermes Agent is an AI assistant that can run commands and scripts on LAN computers over SSH. Specify the destination and task in your prompt. Available operations depend on the remote account's permissions.

This tutorial uses disk-space checks on Linux as an example, starting with one computer and extending the task to a second.

## Learning objectives

In this guide, you will learn how to:

- Configure key-based SSH authentication and confirm the target computer's identity.
- Ask Hermes to check disk space on one or more computers.
- Troubleshoot SSH connection and permission errors.

## Prerequisites

- [Hermes Agent installed on Olares and connected to a model](hermes.md#configure-hermes-agent).
- A Linux target computer with a running SSH server that allows your account to log in.
- Network access from the Olares device to the target computer's SSH port, which defaults to `22`.

## Prepare the connection

Keep **Local** as the terminal backend, as configured in the [Hermes setup wizard](hermes.md#step-2-run-the-setup-wizard). Hermes runs the standard `ssh` command inside its app container to connect to the target computer. You do not need to enable the Gateway API or configure a fixed destination in Hermes.

This example uses the Linux host `ubuntu@192.168.1.50`. Replace the username, address, and port with your own settings.

:::tip Network access from Olares
Access from the computer where you open Hermes does not establish access from Olares. Routing between subnets, guest-network isolation, or firewall rules can block the connection.
:::

## Configure key-based authentication

Use key-based authentication for automated tasks. After you authorize the public key, Hermes can connect without asking for the account password. The steps below create a dedicated key. To reuse a key, follow [Use an existing private key](#existing-ssh-key).

### Step 1: Store a key in persistent app data

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

   Copy the complete public-key line. Keep the private key inside the app. Do not paste it into a chat.

Files under `/opt/data` survive app restarts and container recreation. Back up the keys before uninstalling the app or deleting its data. App clones have their own data. Set up keys for each clone that needs SSH access.

### Step 2: Authorize the key on the target computer {#authorize-ssh-key}

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

### Step 3: Check the first connection {#check-ssh-connection}

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

Use absolute paths for both the key and `known_hosts`, so the agent can find them even when its terminal has a different home directory.

## Check disk space with Hermes

Open Dashboard, or run `hermes` in Hermes CLI to start a TUI chat. Replace the example account and address in this prompt with your own:

```text
Check disk space on ubuntu@192.168.1.50 over SSH using
/opt/data/.ssh/id_ed25519 and /opt/data/.ssh/known_hosts.
Use strict host-key checking, batch mode, and a 10-second connection timeout.
Confirm the target hostname and account, then run df -h.
Summarize the size, used space, available space, and usage percentage
for each filesystem. Only inspect disk usage. Do not delete files or
change settings. Report any permission or connection errors.
```

Review the command output and summary to see which filesystems have little space left.

To check a second computer, authorize the public key on that computer and complete the first-connection check. This example uses `olares@192.168.1.60` and the same private key. You can use a separate key if you provide its path for each host.

Then send this prompt in the same conversation, replacing both accounts and addresses with your own:

```text
Check disk space on ubuntu@192.168.1.50, then olares@192.168.1.60.
Use /opt/data/.ssh/id_ed25519 and /opt/data/.ssh/known_hosts for both.
Keep strict host-key checking, batch mode, and a 10-second connection timeout.
Confirm each hostname and account, then run df -h without changing any files.
Group the results by hostname and IP address. Include each filesystem's
size, used space, available space, and usage percentage.
If a host fails, report the error and the checks completed so far,
then ask me whether to continue.
```

For other tasks, describe the commands or scripts to run and the order in which to run them. If you refer to a script, specify which computer stores it and its path.

:::info Remote sudo
Permissions on the remote computer come from its SSH account. Ask Hermes to use `sudo -n` for commands that require elevation. The account must have permission to run the command. If a password is required, `sudo -n` returns an error instead of waiting for input. Do not put a sudo password in the prompt. This is separate from sudo inside the Hermes container.
:::

## Optional: Use an existing private key {#existing-ssh-key}

You can use an existing private key, including a PEM key, instead of generating a new one. SSH does not require a `.pem` filename. For unattended tasks, the key must be usable without a passphrase prompt.

1. In Hermes CLI, prepare the persistent SSH directory:

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ```

2. In Files, open **Data** > **hermesagent** > **home**, show hidden files, and upload the private key to `.ssh` as `lan-host.pem`. Keep the private key out of chat messages.

3. In Hermes CLI, restrict access to the private key and display its public key:

   ```bash
   chmod 600 /opt/data/.ssh/lan-host.pem
   ssh-keygen -y -f /opt/data/.ssh/lan-host.pem
   ```

4. If the target account does not already authorize this key, copy the public-key output and follow [Authorize the key on the target computer](#authorize-ssh-key).

5. Follow [Check the first connection](#check-ssh-connection), replacing `/opt/data/.ssh/id_ed25519` with `/opt/data/.ssh/lan-host.pem`. Use the same replacement in your task prompts.

## Optional: Sign in manually with a password

For a manual SSH session, you can use a password if the target allows password authentication. This is an alternative to key-based login, not a prerequisite for agent tasks.

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

The standard SSH command does not automatically answer password prompts in background tasks. Use key-based authentication for tasks the agent runs independently. Do not put passwords in commands, scripts, or chat prompts.

## Troubleshoot SSH connections

| Symptom | What to check |
|:---|:---|
| Connection times out or is refused | Check the address, SSH port, server status, firewall, and routing from the Olares device to the target. Opening an Olares app entrance does not enable outgoing SSH. |
| `Permission denied (publickey)` | Check the remote username, matching public key in that account's `authorized_keys`, and private-key path. Keep the private key at mode 600 and `.ssh` at mode 700. |
| `Host key verification failed` | Complete the first connection interactively with the same `known_hosts` path. If a saved fingerprint changes, confirm the cause on the target before updating the saved entry. Keep host-key checking enabled. |
| `sudo: a password is required` | The account does not have passwordless permission for that command. Run an allowed read-only command or ask the host administrator for the required access. |

## Learn more

- [Set up Hermes Agent](hermes.md): Connect a model, use Olares skills, and integrate the agent with other applications.
