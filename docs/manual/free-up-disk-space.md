---
outline: [2, 3]
description: Free Olares disk space by removing unused AI models and carefully deleting unreferenced container images on Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, disk space, storage full, image cleanup, model cleanup, crictl, Hugging Face, Ollama
---

# Free up disk space

Use this guide to reclaim storage used by AI models you no longer need and container images that Olares no longer references. It also applies when disk usage is high, an app cannot install, or sign-in reports **Authentication failed, disk space is full**.

Start with model files that you can identify. Remove container images only after reviewing the unused-image list. Stopping an app does not automatically remove its models or images because Olares may need them when the app resumes.

## Prerequisites

- Administrator access to Olares.
- Terminal access to the Olares host for removing unused images. See [Connect to Olares via SSH](/manual/access-olares-terminal).

## Check disk usage

1. Open **Dashboard** and select the **Disk** card.
2. Check the used and available space.
3. Select **Occupancy analysis** to see which file system is full.

For details, see [Monitor resource usage](olares/resources-usage.md#disk-panel).

## Remove unused AI models

:::warning Model deletion cannot be undone
Delete only models that you recognize and no longer need. Apps that use a deleted model stop working until the model is downloaded again.
:::

1. Open **Files**.
2. Check the model locations that exist on your system:
   - On Olares 1.12.6, select **Application** > **Common**, then open `huggingface` or `ollama`.
   - For models stored in the earlier per-user layout, select **Home**, then look for app-named folders such as `Huggingface` and `Ollama`.
3. Identify models that are no longer used by any app.
4. Right-click the model folder or file and select **Delete**.

Keep the required directory structure when deleting individual models. For details, see [Manage shared AI models](olares/files/files-common.md#manage-model-files).

## List unused container images

Olares 1.12.6 includes a command that lists local container images with no standard workload references. Run it before deleting anything:

```bash
olares-cli doctor images --unused
```

Review every row. The command sorts unused candidates from largest to smallest and shows the estimated reclaimable size.

On a multi-node Olares cluster, the list covers images stored on the control node where the command runs. It does not inventory images stored only on worker nodes.

:::warning Check custom workloads before continuing
The unused-image check covers standard Deployment, StatefulSet, DaemonSet, Job, and CronJob specifications. It may not detect images used only by a bare Pod, static Pod, custom controller, or a running container still pinned to an older digest. If you created custom workloads, do not run the deletion pipeline without checking them first.
:::

## Delete reviewed container images

After you confirm that the list contains only images you can re-download, run:

```bash
olares-cli doctor images --unused --no-headers \
  | awk '$1 ~ /^[0-9a-f]+$/ && length($1) == 12 {print $1}' \
  | xargs -r sudo crictl rmi
```

The filter passes only 12-character hexadecimal image IDs to `crictl`; it does not treat the `no unused images` message as an image ID.

Deleted images do not remove personal files or app data. However, resuming or reinstalling an affected app may be slower because Olares must download its image again.

## Verify available space

1. Run `olares-cli doctor images --unused` again.
2. Return to **Dashboard** > **Disk** and refresh the usage data.
3. Resume the apps you need and confirm that they start normally.

If disk usage is still high after cleanup, see [Expand Olares system storage](best-practices/expand-storage-in-olares.md).

## Avoid blanket image pruning

:::danger Not recommended
Do not use `sudo crictl rmi --prune` as routine cleanup. It removes images for all stopped apps, not just the candidates reviewed with `olares-cli doctor images --unused`. Apps can take longer to resume, and the Olares UI can temporarily show inconsistent image information.
:::

Use blanket pruning only when Olares support specifically asks you to and you understand that stopped apps may need to download their images again.
