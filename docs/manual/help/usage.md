---
outline: [2, 3]
description: Find answers to common questions about using Olares and community apps.
---

# Usage FAQs

Find answers to common questions about daily usage, applications, and system management.

## Applications

### What apps can I run in Olares?

The [Olares Market](https://market.olares.com/) maintains popular open-source apps like Ollama, ComfyUI, and Open WebUI.

If you have Docker experience, you can manually [deploy apps](../../developer/develop/tutorial/index.md) not listed in the Olares Market in a testing environment.

### Can I play games on my Olares device?

Yes. Install the Steam Headless app to transform your Olares device into a gaming server.

* [**Streaming**](../../use-cases/stream-game.md): Run games locally on Olares and stream them to devices like phones and tablets.
* [**Direct play**](../../use-cases/play-games-directly.md): Connect a monitor, keyboard, and mouse directly to the Olares device to play games without streaming.

### How do I access the Windows environment in Olares?

Install and run a Windows VM from the Olares Market, and access it using any standard RDP client.

For detailed instructions, refer to [Run a Windows VM on your Olares device](../../use-cases/windows.md).

### Can I develop apps on Olares?

Yes. Install [Studio](../../developer/develop/tutorial/index.md) to code directly in your browser, or connect your local VS Code to the device. This provides a development experience similar to your local machine while leveraging the greater power of your server hardware.

### Can I manually update an application version?

:::tip Important
We recommend always updating applications through the Olares Market to ensure stability and compatibility.
:::

Before publishing an update to the Market, the Olares team thoroughly tests new versions to ensure compatibility and stability. In some cases, an application might show an internal prompt for a new update before it is officially available in the Market.

If you urgently need the latest features, you can manually update the application's Docker image via the Control Hub.

Before proceeding with the manual update, review the following notes:
- **Temporary changes**: Manual edits to configurations in Control Hub are not persistent. When you apply an update through the Market later on, the Market version will overwrite all your manual configurations, including the image version.
- **Unexpected behavior**: After the manual update, the application might fail to start or run correctly due to compatibility issues.

<Tabs>
<template #Update-using-the-official-app-image>

:::warning Compatibility & privileges
- The official image might not be fully adapted for Olares because configuration paths or environment variables can vary.
- If the application requires root or other special privileges, using images from other organizations might prevent the application from starting due to permission restrictions.
:::

The following steps demonstrate how to manually update using Ollama as an example.

1. Find the official Docker image name and the latest release tag.
2. Note down the image name and tag. For example, `ollama/ollama` and `0.23.1`.

    ![Ollama Docker image name](/images/manual/help/faq-ollama-docker-hub.png#bordered)

    ![Ollama Docker image version tag](/images/manual/help/faq-ollama-image-tag.png#bordered)    

3. Open Control Hub, go to **Browse** > **System** > **ollamaserver-shared** > **Deployments** > **ollama**, and then click <span class="material-symbols-outlined">edit_square</span>.
4. In the YAML editor, find the `containers` section, and then note down the current image and tag in case you need to roll back later. For example, `docker.io/beclab/ollama-ollama:0.20.5`.

    ![Ollama Docker image hub](/images/manual/help/faq-ollama-container-update.png#bordered)

5. Update the field to the new official image name and tag. For example, change `docker.io/beclab/ollama-ollama:0.20.5` to `docker.io/ollama/ollama:0.23.1`.
6. Click **Confirm**. The system will automatically pull the new image and restart the pod. Large images might take several minutes to download. Once complete, the pod status returns to **Running**.

    ![Ollama Docker image updated in Control Hub](/images/manual/help/faq-ollama-container-updated.png#bordered)

7. Open the container's Terminal in Control Hub and run the version command `ollama -v` to confirm the update.

    ![Ollama Docker image update verify in Control Hub](/images/manual/help/faq-ollama-container-update-verify.png#bordered)

</template>
<template #Update-using-the-Olares-mirrored-image>

:::warning Potential conflicts
`beclab` images are provided by Olares for easier access. However, because some updates include environment adaptations, manually pulling a new version might cause configuration mismatches with your current setup. As a result, the application might fail to start or function correctly.
:::

For some frequently updated AI applications, Olares might have already mirrored the latest image to the official registry but hasn't manually pushed the chart update to the Market yet.

The following steps demonstrate how to manually update using OpenClaw as an example.

1. Go to the official [Olares Docker registry](https://hub.docker.com/u/beclab).
2. Search for `OpenClaw`, go to its details page, check the **Tags** tab, and then note down the latest version tag. For example, `2026.5.7`.

    ![Search for latest docker image in Olares Docker registry](/images/manual/help/faq-openclaw-latest-image.png#bordered)

3. Open Control Hub, go to **Browse** > **{Username}** > **clawdbot-{Username}** > **Deployments** > **clawdbot**, and then click <span class="material-symbols-outlined">edit_square</span>.
4. In the YAML editor, find the `containers` section, and then note down the current image and tag, in case you need to roll back later. For example, `beclab/openclaw-openclaw:2026.3.12`.

    ![OpenClaw image tag in Control Hub](/images/manual/help/faq-openclaw-container-update.png#bordered)

5. Update only the version tag of the existing `beclab` image. For example, change `beclab/openclaw-openclaw:2026.3.12` to `beclab/openclaw-openclaw:2026.5.7`.
6. Click **Confirm**. The system will automatically pull the new image and restart the pod. Large images might take several minutes to download. Once complete, the pod status returns to **Running**.

    ![OpenClaw Docker image updated in Control Hub](/images/manual/help/faq-openclaw-container-updated.png#bordered)

7. Open the container's Terminal in Control Hub and run the version command `openclaw -v` to confirm the update.

    ![OpenClaw Docker image update verify in Control Hub](/images/manual/help/faq-openclaw-container-update-verify.png#bordered)
</template>
</Tabs>

:::tip Rollback
If the application fails to start or experiences compatibility issues after the manual update, you can revert it by editing the YAML again to restore the old image tag using the one you noted down earlier. For example, change `docker.io/ollama/ollama:0.23.1` back to `docker.io/beclab/ollama-ollama:0.20.5`.
:::

## Storage

### If I add new disks to a running Olares machine, will Olares use them automatically?

It depends on the type of drive:
* **USB drives**: Yes. These are automatically mounted and will appear immediately in the Files app.
* **Internal drives**: No. Iternal HDDs or SSDs require manual configuration to join the storage pool.
* **SMB shares**: Add network storage manually. Go to **External** > **Connect to server** in the Files app.

For detailed instructions, see [Expand storage in Olares](../best-practices/expand-storage-in-olares.md).

## Multi-node clusters

### How do I add more machines to my cluster?

By default, Olares installs as a single-node cluster. To create a scalable, multi-node cluster, install Olares as a master node and then add worker nodes.

Note that this is currently an Alpha feature and works on Linux only. For detailed steps, refer to [Install a multi-node Olares cluster](../best-practices/install-olares-multi-node.md).
