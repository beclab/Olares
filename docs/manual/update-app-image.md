---
outline: [2, 3]
description: Manually update an application's Docker image in Control Hub when the version you need is not yet available in the Market, and roll back if issues occur.
head:
  - - meta
    - name: keywords
      content: Olares, update app image, Control Hub, container image, rollback, Docker image
---

# Manually update an app image

We recommend always updating applications through the Olares Market to ensure stability and compatibility. In some cases, an application might show an internal prompt for a new update before it is officially available in the Market. If you urgently need the latest features, you can manually update the application's Docker image in Control Hub.

Before proceeding, review the following notes:

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
