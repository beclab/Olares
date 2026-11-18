---
outline: [2, 3]
description: Check Olares One status and hardware details, adjust power settings, reset the SSH password, and restart or shut down your device.
head:
  - - meta
    - name: keywords
      content: Olares One, cluster overview, node settings, power mode, CPU frequency, automatic startup, SSH password
---

# Check device status and manage hardware settings

Use **Cluster overview** in **Settings** to check your Olares One status, view hardware details, and adjust settings such as power mode and automatic startup.

Your Olares One appears as a node in the cluster.

## Check cluster status

Open **Settings**, click your avatar in the upper-left corner, then click **Cluster overview**.

Check the cluster's health, connectivity, and operating phase, along with its Olares OS version and node counts.

![Cluster overview](/images/manual/olares/cluster-overview-1.12.7.png#bordered)

## View device details

Under **Nodes**, click your Olares One to open its details.

The page shows the node's status, connectivity, and operating phase, along with its CPU, GPUs, memory, and disk capacity. You can also check the Olaresd version and when the node information was last observed.

![Device details](/images/manual/olares/device-detail-1.12.7.png#bordered)


## Switch power mode

On the node details page, under **Node settings**, click **Power mode** and select a profile:

- **Silent**: Limits CPU and GPU power for quiet operation, suitable for everyday workloads.
- **Performance**: Enables maximum CPU and GPU performance for demanding tasks such as AI inference or gaming.

## Limit CPU frequency

On the node details page, find **Limit CPU frequency** under **Node settings**.

Turn on this switch to limit the CPU frequency from 5.4 GHz to 5.0 GHz. Turn it off to restore the original maximum frequency.

## Set automatic startup

On the node details page, under **Node settings**, turn on **Automatic startup** to start Olares One automatically when power is connected or restored after an outage.

:::info
Requires Olares OS 1.12.6 or later and EC firmware 1.03 or later. If either requirement is not met, the switch is disabled.

To check or update your EC firmware, see [Manage BIOS and EC](update-firmware.md).
:::

## Reset SSH password

To use your own SSH password instead of the one generated during activation, reset it in Settings:

1. On the node details page, under **Node settings**, click **Reset SSH password**.
2. Enter a new SSH password that meets all strength requirements, then click **Reset**.

   ![Reset SSH password dialog](/images/one/ssh-reset-password-in-settings-1.12.7.png#bordered)

3. Open LarePass and scan the QR code shown on the screen.
4. Tap **Confirm** in LarePass to finish. The new SSH password is saved to Vault.

## Restart or shut down the device

Save your work before restarting or shutting down. These actions interrupt running applications and tasks.

- To restart your Olares One, click **Restart node** on its details page.
- To restart all nodes, return to **Cluster overview** and click **Restart cluster**.
- To shut down all nodes, return to **Cluster overview** and click **Shut down cluster**.

If your cluster contains only one Olares One, **Restart node** and **Restart cluster** both restart that device.
