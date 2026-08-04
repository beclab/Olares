---
outline: [2, 3]
description: Upgrade a two-node Olares cluster from version 1.12.5 to 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares One, multi-node, two-node, upgrade, 1.12.5, 1.12.6
---

# Upgrade a two-node Olares cluster from 1.12.5 to 1.12.6

This guide walks you through manually upgrading a two-node Olares cluster from version 1.12.5 to 1.12.6.

The procedure involves downloading upgrade packages on both nodes, upgrading the Olares CLI and daemon, and upgrading the master and worker nodes separately.

## Prerequisites

**Cluster**
- You have a two-node Olares cluster with one master node and one worker node.
- Both nodes are powered on and connected to the same local area network.
- You know the local IP addresses of both nodes.

**Access**
- You can access both nodes via SSH as a user with `sudo` privileges.

## Step 1: Connect to both nodes

Open two separate terminal windows or tabs on your computer. Use SSH to connect to each node in its own window.

1. In the first window, connect to the master node via SSH using its local IP address:

   ```bash
   ssh olares@<master-node-ip-address>
   ```

2. In the second window, connect to the worker node via SSH using its local IP address:

   ```bash
   ssh olares@<worker-node-ip-address>
   ```

3. Keep both SSH connection windows open throughout this process. Some steps must be run on both nodes, and some steps must be run while both nodes are accessible.

## Step 2: Download upgrade packages on both nodes

Download the upgrade files without installing them. Run the following commands in both your master and worker SSH windows.

1. Switch to the root user:

   ```bash
   sudo su
   ```

2. Load the Olares environment variables:

   ```bash
   source /etc/olares/release
   ```

3. Create the upgrade target file to trigger the download:

   ```bash
   echo '{"version":"1.12.6", "downloadOnly": true}' > $OLARES_BASE_DIR/upgrade.target
   ```

4. Check the download progress:

   ```bash
   curl localhost:18088/system/status | jq | grep -i download
   ```

5. Wait until the `upgradingDownloadState` value changes to `completed`.

:::warning Wait for downloads to complete
Before proceeding to Step 3, ensure that `upgradingDownloadState` shows `completed` on both nodes.
:::

## Step 3: Upgrade Olares CLI and daemon on both nodes

After the downloads finish, import the new images and update the core management tools, including the Olares CLI and the `olaresd` daemon. Run the following commands in both your master and worker SSH windows.

1. Ensure that you are still logged in as root:

   ```bash
   sudo su
   ```

2. Ensure the environment variables are loaded:

   ```bash
   source /etc/olares/release
   ```

3. Update the Olares CLI to the new version:

   ```bash
   cp -f $OLARES_BASE_DIR/pkg/components/olares-cli-v1.12.6 /usr/local/bin/olares-cli
   ```

4. Import the new container images:

   ```bash
   olares-cli prepare images
   ```

5. Upgrade the `olaresd` daemon:

   ```bash
   olares-cli prepare olaresd
   ```

## Step 4: Upgrade the master node

With both nodes prepared, upgrade the master node first.

1. In the master node SSH window, start the upgrade:

   ```bash
   sudo olares-cli upgrade
   ```

   The master node reboots when the upgrade finishes, and your SSH session will disconnect.

2. After the reboot, reconnect to the master node via SSH:

   ```bash
   ssh olares@<master-node-ip-address>
   ```

3. Verify that all system services have successfully restarted and are in the `Running` status:

   ```bash
   kubectl get pod -o wide -A
   ```

4. Temporarily set the cluster version back to `1.12.5` so the worker node can be upgraded:

   ```bash
   kubectl patch terminus terminus --type=merge -p '{"spec":{"version":"1.12.5"}}'
   ```

## Step 5: Upgrade the worker node

Now that the master node is upgraded and patched, you can upgrade the worker node.

1. In the worker node SSH window, start the upgrade:

   ```bash
   sudo olares-cli upgrade
   ```

2. The worker node reboots when the upgrade finishes. Wait for the reboot to finish.

## Step 6: Restore the cluster version on the master node

After the worker node finishes upgrading, restore the version on the master node to `1.12.6` to complete the process.

If your SSH session to the master node times out, reconnect before proceeding.

1. In the master node SSH window, run the following command:

   ```bash
   kubectl patch terminus terminus --type=merge -p '{"spec":{"version":"1.12.6"}}'
   ```

   The two-node cluster is now running Olares 1.12.6.

2. To verify the final status of both nodes in the cluster, run the following command:

   ```bash
   kubectl get nodes
   ```

## Resources

- [Connect two Olares One](./connect-two-olares-one): Learn how to set up a two-node Olares cluster.
