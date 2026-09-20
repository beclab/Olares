---
outline: [2, 3]
description: 排查在显存分片或容量切分模式下，依赖 GPU 的应用因显存不足而始终处于暂停状态的问题。
head:
  - - meta
    - name: keywords
      content: Olares, GPU 应用, 显存分片, 容量切分, 显存不足, 暂停状态, 释放显存
---

# GPU 应用安装或恢复后处于暂停状态

当 GPU 模式设置为**显存分片**（Olares 1.12.5）或**容量切分**（Olares 1.12.6）时，如果某个依赖 GPU 的应用在安装后或恢复运行后仍处于**暂停**状态，可参考本指南进行排查。

## 适用情况

在 Olares 1.12.5 的**显存分片**模式或 Olares 1.12.6 的**容量切分**模式下，出现以下任一情况：

- 安装依赖 GPU 的应用后，该应用始终处于**暂停**状态。
- 点击**恢复**后，依赖 GPU 的应用无法启动，仍处于**暂停**状态。

## 原因

在 Olares 1.12.5 的**显存分片**模式和 Olares 1.12.6 的**容量切分**模式下，应用安装或恢复运行时需要有足够的可用显存配额。如果大部分显存已被其他应用占用，系统将无法为该应用分配足够的显存，导致应用无法完成初始化，始终处于**暂停**状态。

## 解决方案：释放显存

请根据 Olares 版本选择对应的解决方案。

### Olares 1.12.5

1. 前往**应用市场** > **我的 Olares**，点击目标应用的卡片。在应用详情页中，查看应用所需的显存大小。

   ![查看目标应用所需显存](/images/zh/manual/help/ts-mem-slice-vram-app-gpu.png#bordered){width=85%}

2. 前往**设置** > **GPU**。在**分配显存**区域，将各应用的显存分配量相加，再用 GPU 总显存减去该数值，得到当前可用显存。

   ![查看当前显存分配情况](/images/zh/manual/help/ts-mem-slice-vram-gpu-mode.png#bordered){width=90%}

   以上图为例，当前已分配 22 GB 显存，仅剩 2 GB 可用，而目标应用需要 4 GB，因此无法启动。

3. 根据实际情况，选择以下一种或两种方式释放显存：

   - 在**分配显存**区域，点击某个应用显存数值旁边的 <i class="material-symbols-outlined">edit_square</i>。适当降低该应用的显存分配量，但不得低于其最低显存需求，然后点击**确认**。

     ![减少显存分配](/images/zh/manual/help/ts-mem-slice-vram-reduce-vram.png#bordered){width=90%}

   - 从**应用市场** > **我的 Olares**或**设置** > **应用**暂停当前不使用的应用。返回**设置** > **GPU**，点击已暂停应用旁边的 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。

4. 从**应用市场** > **我的 Olares**或**设置** > **应用**恢复目标应用。
5. 等待应用状态变为**运行中**。

   ![释放显存后恢复目标应用](/images/zh/manual/help/ts-mem-slice-vram-resume-outcome.png#bordered){width=90%}

### Olares 1.12.6

1. 前往**设置** > **应用**或**应用商店** > **我的 Olares**。找到目标应用，点击**恢复**。
2. 在启动弹窗右上角查看**显存分配**。斜杠后的数值表示应用所需显存。例如，`0 Gi / 4 Gi` 表示应用需要 4 Gi 显存。

   ![启动窗口](/images/zh/manual/help/ts-mem-slice-insufficient-vram1.png#bordered){width=90%}

3. 在**选择 GPU**下展开兼容的 GPU，然后查看：
   - **独立显存**：GPU 的总显存。
   - **已分配的应用**：正在使用该 GPU 的应用，以及各应用分配的显存。

   ![查看显存使用情况](/images/zh/manual/help/ts-mem-slice-insufficient-vram2.png#bordered){width=90%}

4. 在**已分配的应用**中，找到当前不需要使用的应用，然后点击**移除**。重复操作，直至释放出足够的显存。

   :::info
   移除已分配的应用会释放其 GPU 资源，并停止该应用。
   :::

5. 选择该 GPU，检查目标应用的显存分配，然后点击**启动**。

   ![启动应用](/images/zh/manual/help/ts-mem-slice-insufficient-vram3.png#bordered){width=90%}

:::info 在 Olares 1.12.6 中重新分配显存
应用完成资源分配后，无法直接修改显存配额。如需调整，请先从 GPU 中移除该应用，再恢复应用并重新分配资源。
:::

有关 Olares 1.12.6 界面和资源分配的详细说明，请参阅[管理 AI 算力资源](../olares/settings/gpu-resource.md)。