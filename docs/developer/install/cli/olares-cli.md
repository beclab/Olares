---
outline: [2, 3]
description: Use common Olares CLI commands in the device terminal to check system status, collect logs, reset a password, and manage storage.
head:
  - - meta
    - name: keywords
      content: Olares, olares-cli, device terminal, system logs, reset password, disk storage, GPU status
---

# Common Olares CLI commands

Olares OS includes `olares-cli` at `/usr/local/bin/olares-cli`. This page covers the commands you are most likely to run yourself in the Olares device terminal for system maintenance and troubleshooting.

To manage files, apps, settings, and other Olares features through natural language, see [Use Olares CLI with AI agents](../../cli-install.md).

## Before you begin

- [Access the Olares terminal](../../../manual/access-olares-terminal.md) on a device you administer.
- Run commands with `sudo` when they require system privileges.
- Run `olares-cli help` or `olares-cli <command> --help` to see the commands and options available in your installed version.

## Check system status

Check whether Olares services are running and view the current installation or upgrade status:

```bash
olares-cli status
```

If the device has an NVIDIA GPU, check its driver and service status:

```bash
olares-cli gpu status
```

## Collect system logs

Collect system, Kubernetes, container, kernel, and network diagnostics into an archive:

```bash
mkdir -p "$HOME/olares-logs"
sudo olares-cli logs --max-lines 0 --output-dir "$HOME/olares-logs"
```

The last output line shows the archive path. Logs can contain private system information, so share the archive only through a private support channel.

Learn more: [Collect diagnostic information](../../../manual/collect-diagnostic-information.md).

## Reset a forgotten desktop password

Replace `<username>` with the part of the Olares ID before `@`:

```bash
olares-cli user reset-password <username> -p <new-password>
```

If the command returns a permission error, follow [Forgotten desktop login password](../../../manual/help/ts-forget-login-password.md#step-3-reset-the-password) to check the error and choose the next step.

## Reclaim disk space

In Olares 1.12.6 and later, list container images that are not referenced by standard workloads:

```bash
olares-cli doctor images --unused
```

Review the results before deleting anything. Custom workloads can still depend on an image in this list.

Learn more: [Free up disk space](../../../manual/free-up-disk-space.md).

## Expand system storage

For an LVM-based Olares installation, check which disks are available to add to the system volume:

```bash
olares-cli disk list-unmounted
```

:::warning The selected disk will be erased
Extending storage destroys all data on the selected disk and merges it into the Olares system volume. Back up the disk before continuing.
:::

Follow [Expand Olares system storage](../../../manual/best-practices/expand-storage-in-olares.md) to identify the correct disk, perform the extension, and check the result.

## Related guides

Some Olares CLI commands are part of longer procedures and should not be run in isolation:

- [Install a specific CUDA version](../../../manual/best-practices/install-specific-cuda-version.md): install or enable GPU support when the automatic setup does not meet your requirements.
- [Install and activate Olares from the command line](../../../manual/best-practices/activate-olares-using-cli.md): initialize a new device from its device terminal.
- [Set up a multi-node Olares cluster](../../../manual/best-practices/install-olares-multi-node.md): enable shared storage and add a worker node.
- [Upgrade a two-node Olares cluster](../../../one/upgrade-multi-node-cluster.md): manually coordinate an upgrade across the master and worker nodes.
