---
title: Connect two Olares One devices as a cluster
outline: [2, 3]
description: Connect two Olares One devices by enabling JuiceFS on an existing master, generating a worker join command, and checking the cluster.
head:
  - - meta
    - name: keywords
      content: Olares One, Cluster, Kubernetes, Multi-node, Worker node, Master node
---

# Connect two Olares One devices <Badge type="tip" text="1.5 h" />

For advanced use cases requiring more computing resources and distributed storage, you can physically connect two Olares One units to form a single, unified cluster.

:::warning Olares 1.12.7 or later required
The master must be activated and running Olares 1.12.7 or later as a single-node system. [Back up important data](../manual/olares/settings/backup.md) before starting. Keep the master powered on during the file system migration.
:::

## Learning objectives

By the end of this guide, you will learn how to:

- Enable shared storage on the master node.
- Generate and protect the worker join command.
- Resolve hostname conflicts and join a worker node to the cluster.
- Verify that both nodes are ready.

## Before you begin

- The factory SSH username and password are both `olares`. For activated devices, use the SSH password saved in LarePass Vault. See [Access Olares One via SSH](./access-terminal-ssh.md#step-2-get-the-login-password-from-vault).
- SSH access grants powerful control over the system. Keep your credentials secure.

## Prerequisites

**Hardware**<br>
- Two Olares One units connected to the same local area network, with internet access to download the join script and installation files.
- A computer with an SSH client.

**Experience**<br>
- Familiarity with terminal commands and basic Kubernetes node management.

**Software**<br>
- LarePass installed on your mobile phone.
- SSH access with `sudo` permissions on both units. The worker must be able to sign in to the master over SSH with a password.

## Step 1: Set up the master node

### Enable JuiceFS

Your existing, activated unit serves as the master node. JuiceFS provides shared storage for the cluster. The migration command installs MinIO as the storage backend by default and migrates your existing data.

1. Access the master node via SSH using its local IP address.

   In LarePass, go to **Settings** > **System** > your device > **Nodes** > **Network** > **Intranet IP** to find the address. Use the SSH password saved in LarePass Vault.

   ```bash
   ssh olares@<master-ip-address>
   ```

2. Stop Olares.

   ```bash
   sudo olares-cli stop
   ```

3. Enable JuiceFS and migrate the existing Olares file system.

   ```bash
   sudo olares-cli node enable-juicefs
   ```

   Migration time depends on the amount of data. Large migrations can take a while without showing progress. Keep the terminal open and wait for completion.

   Olares restarts automatically when the migration finishes. Wait for confirmation that JuiceFS is enabled and the master is ready to accept worker nodes before continuing.

   Example output:

   ```plain
   [Job] Enable JuiceFS on the master node and migrate rootfs execute successfully!!! (...)
   JuiceFS is enabled on this master node (version <version>); it is now ready to accept worker nodes.
   Run 'olares-cli node join-command' on this master to print the command for a worker node.
   ```

### Generate the worker join command

1. In the master terminal, run:

   ```bash
   sudo olares-cli node join-command
   ```

2. Check the SSH address and username shown. Enter `y` to confirm, or follow the prompt to change the username. The default is `olares`.
3. Enter the SSH password. The command checks SSH sign-in and `sudo` access on the master. If authentication fails, it prompts for the username and password again. Press `Enter` to keep the suggested username, or enter a different one. If the connection still fails, check the address, SSH service, and firewall settings.
4. Copy the generated command. You will run it on the worker node in the next step.

   Example output:

   ```plain
   SSH login and sudo access verified for <username>@<master-ip-address>:22.

   Run the following command on the worker node:

   export MASTER_AUTH_INFO='<encoded-auth-info>' OLARES_SYSTEM_CDN_SERVICE='<olares-cdn-url>' && curl -fsSL '<join-script-url>' | bash

   MASTER_AUTH_INFO is Base64-encoded, not encrypted. Anyone holding this command can recover the master's SSH credentials, so share it only with the intended worker administrator.
   ```

:::danger Protect the join command
The generated `MASTER_AUTH_INFO` value is Base64-encoded, not encrypted. Anyone who has the command can recover the master's SSH credentials. Share it only with the administrator of the intended worker, and never post it in a public channel or repository.
:::

## Step 2: Set up the worker node

The second unit will serve as the worker node. Open a separate terminal and check its IP address before continuing.

:::warning Clean installation required on the worker
Uninstall any existing Olares system on the worker, including the factory-installed system. [Back up important data](../manual/olares/settings/backup.md) first. Do not uninstall Olares on the master.
:::

1. Access the worker node via SSH using its local IP address.

   ```bash
   ssh olares@<worker-ip-address>
   ```

2. On the worker, paste and run the complete command from [Generate the worker join command](#generate-the-worker-join-command). The command follows this pattern:

   ```bash
   export MASTER_AUTH_INFO='<encoded-master-connection-information>' \
     OLARES_SYSTEM_CDN_SERVICE='<olares-cdn-url>' \
     && curl -fsSL '<join-script-url>' | bash
   ```

   Use the generated command exactly as shown. Do not copy the placeholder command above.

3. If the command reports that Olares is already installed, uninstall it:

   ```plain
   Joining this machine to an Olares cluster as a worker node, using the Olares <version> installer.
   error: Olares <version> is already installed on this node; run 'sudo olares-cli uninstall' before joining it to another cluster
   ```

   ```bash
   sudo olares-cli uninstall
   ```

   When the uninstall finishes, run the generated join command again.

4. If prompted, enter a unique hostname, such as `olares-worker`. Both units use `olares` by default, but Kubernetes requires each node to have a unique name. If the current hostname is valid and unique, no change is needed.

   The hostname must:

   - Contain only lowercase letters, numbers, hyphens (`-`), or periods (`.`).
   - Start and end with a letter or number.
   - Be unique in the cluster.

The join process checks the connection to the master, updates the hostname when needed, prepares the matching Olares version, and adds the worker to the cluster.

If the worker has only been prepared for installation, rather than fully installed, the process reuses a matching version. If the prepared version differs from the master's, it automatically removes the old prepare state and prepares the matching version.

Wait for a success message similar to the following:

```plain
[Job] Add Worker Node To The Cluster execute successfully!!! (...)

This node joined the Olares cluster at <master-ip-address> as "<worker-node-name>".
Verify it on the master with: sudo /usr/local/bin/kubectl get nodes
```

## Step 3: Verify the cluster

Once the join command completes, run the following command on the master to check both nodes:

```bash
sudo /usr/local/bin/kubectl get nodes
```

Confirm that both devices appear and show `Ready`. For example:

```plain
NAME            STATUS   ROLES                         AGE   VERSION
olares          Ready    control-plane,master,worker   1h    v1.33.3+k3s1
olares-worker   Ready    worker                        1m    v1.33.3+k3s1
```

## Resources

- [Set up a multi-node Olares cluster](../manual/best-practices/install-olares-multi-node.md): Follow the general setup guide for Linux devices.
- [Olares environment variables](../developer/install/environment-variables.md): Explore advanced configuration options.
