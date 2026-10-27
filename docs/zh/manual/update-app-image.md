---
outline: [2, 3]
description: 当所需版本尚未在应用市场上架时，在控制面板中手动更新应用的 Docker 镜像，并在出现问题时回滚。
head:
  - - meta
    - name: keywords
      content: Olares, 手动更新应用, 控制面板, 容器镜像, 回滚, Docker 镜像
---

# 手动更新应用镜像

我们建议始终通过应用市场更新应用，以确保稳定性和兼容性。某些情况下，应用内部可能提示有新版本可用，但该版本尚未正式在应用市场上架。如果急需使用最新功能，可以通过控制面板手动更新应用的 Docker 镜像。

在手动更新之前，应注意以下几点：

- **改动是临时的**：在控制面板中对配置所做的所有手动编辑都不会持久化。若后续通过应用市场更新该应用，应用市场的版本会覆盖你所有的手动配置，包括镜像版本。
- **可能出现异常行为**：手动更新后，应用可能会因兼容性等问题而无法启动或运行异常。

<Tabs>
<template #使用官方镜像更新>

:::warning 兼容性与权限
- 官方镜像可能未完全适配 Olares，因为配置路径或环境变量可能存在差异。
- 如果应用需要 root 权限或其他特殊权限，使用来自其他组织的镜像可能会由于权限限制而导致应用无法启动。
:::

下列步骤以 Ollama 为例演示如何手动更新。

1. 找到官方 Docker 镜像名称和最新的发布标签（tag）。
2. 记下镜像名称和标签，例如 `ollama/ollama` 和 `0.23.1`。

    ![Ollama Docker image name](/images/manual/help/faq-ollama-docker-hub.png#bordered)

    ![Ollama Docker image version tag](/images/manual/help/faq-ollama-image-tag.png#bordered)    

3. 打开控制面板，进入 **浏览** > **System** > **ollamaserver-shared** > **部署** > **ollama**，点击 <span class="material-symbols-outlined">edit_square</span>。
4. 在 YAML 编辑器中，找到 `containers` 区域，记下当前的镜像和标签，以便后续回滚。例如 `docker.io/beclab/ollama-ollama:0.20.5`。

    ![Ollama Docker image hub](/images/zh/manual/use-cases/faq-ollama-container-update.png#bordered)

5. 将该字段更新为新的官方镜像名称和标签。例如，将 `docker.io/beclab/ollama-ollama:0.20.5` 改为 `docker.io/ollama/ollama:0.23.1`。
6. 点击 **Confirm**。系统会自动拉取新镜像并重启容器。大体积镜像可能需要数分钟下载。完成后，容器状态会恢复为 **Running**。

    ![Ollama Docker image updated in Control Hub](/images/zh/manual/use-cases/faq-ollama-container-updated.png#bordered)

7. 在控制面板中打开容器的终端，执行版本命令 `ollama -v` 以确认更新成功。

    ![Ollama Docker image update verify in Control Hub](/images/zh/manual/use-cases/faq-ollama-container-update-verify.png#bordered)

</template>
<template #使用-Olares-镜像仓库中的镜像更新>

:::warning 潜在冲突
`beclab` 镜像是 Olares 为方便访问而提供的。但由于某些更新包含环境适配方面的调整，手动拉取新版本可能导致与当前环境的配置不匹配，进而使应用无法启动或运行异常。
:::

对于某些更新频率较高的 AI 应用，Olares 可能已将最新镜像同步到官方仓库，但尚未将对应的 Chart 更新推送到应用市场。

下列步骤以 OpenClaw 为例演示如何手动更新。

1. 访问 [Olares 官方 Docker 仓库](https://hub.docker.com/u/beclab)。
2. 搜索 `OpenClaw`，进入详情页，查看 **Tags** 标签页，记下最新的版本标签。例如 `2026.5.7`。

    ![Search for latest docker image in Olares Docker registry](/images/manual/help/faq-openclaw-latest-image.png#bordered)

3. 打开控制面板，进入**浏览** > **{用户名}** > **clawdbot-{用户名}** > **部署** > **clawdbot**，点击 <span class="material-symbols-outlined">edit_square</span>。
4. 在 YAML 编辑器中，找到 `containers` 区域，记下当前的镜像和标签，以便后续回滚。例如 `beclab/openclaw-openclaw:2026.3.12`。

    ![OpenClaw image tag in Control Hub](/images/zh/manual/use-cases/faq-openclaw-container-update.png#bordered)

5. 仅更新现有 `beclab` 镜像的版本标签。例如，将 `beclab/openclaw-openclaw:2026.3.12` 改为 `beclab/openclaw-openclaw:2026.5.7`。
6. 点击 **Confirm**。系统会自动拉取新镜像并重启容器。大体积镜像可能需要数分钟下载。完成后，容器状态会恢复为 **Running**。

    ![OpenClaw Docker image updated in Control Hub](/images/zh/manual/use-cases/faq-openclaw-container-updated.png#bordered)

7. 在控制面板中打开容器的终端，执行版本命令 `openclaw -v` 以确认更新成功。

    ![OpenClaw Docker image update verify in Control Hub](/images/zh/manual/use-cases/faq-openclaw-container-update-verify.png#bordered)
</template>
</Tabs>

:::tip 回滚
如果手动更新后，应用无法启动或出现兼容性问题，可以重新编辑 YAML，使用之前记下的旧镜像标签恢复。例如，将 `docker.io/ollama/ollama:0.23.1` 改回 `docker.io/beclab/ollama-ollama:0.20.5`。
:::
