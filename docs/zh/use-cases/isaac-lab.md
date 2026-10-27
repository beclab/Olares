---
outline: [2, 3]
description: 在 Olares 上运行 NVIDIA Isaac Lab，进行 GPU 加速的机器人仿真训练，并通过 WebRTC 流式传输实时可视化训练进度。
head:
  - - meta
    - name: keywords
      content: Olares, Isaac Lab, Isaac Sim, NVIDIA, 机器人仿真, 强化学习, GPU, WebRTC 流式传输, 自托管
app_version: "1.0.18"
doc_version: "1.1"
doc_updated: "2026-10-10"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/isaac-lab.md)为准。
:::

# 使用 Isaac Lab 运行机器人仿真

Isaac Lab 是 NVIDIA 开源的机器人学习框架。在 Olares 上，它以终端应用的形式运行，你可以从命令行启动演示、训练任务或捆绑的 Isaac Sim 仿真器。

使用本指南安装 Isaac Lab、运行 Isaac Lab 工作负载、启动 Isaac Sim、连接 WebRTC 流式传输，以及管理后台进程。

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Isaac Lab。
- 准备 Isaac Lab 终端并获取 WebRTC 端点地址。
- 运行 Isaac Lab 演示和强化学习任务。
- 启动捆绑的 Isaac Sim 仿真器。
- 使用 NVIDIA WebRTC Streaming Client 连接仿真。
- 管理后台工作负载、GPU 进程和运行时日志。

## 前提条件

在开始之前，请确保：
- Olares 运行在配备 NVIDIA GPU 的机器上。
- 如果要使用 WebRTC 流式传输，Olares 主机使用 AMD64 架构。ARM64 主机（如 DGX Spark）仅支持非流式传输工作负载。
- 本地计算机上已安装 [NVIDIA WebRTC Streaming Client](https://docs.isaacsim.omniverse.nvidia.com/6.1.0/installation/download.html)。

:::info 主机与本地计算机架构
AMD64 要求适用于运行 Isaac Lab 的 Olares 主机，而非用于流式传输的本地计算机。

例如，如果 Isaac Lab 运行在配备 Intel CPU 和 NVIDIA GPU 的 Olares One 上，则 Olares 主机为 AMD64。你的本地计算机可以使用不同的架构，但必须能够运行 NVIDIA 的 WebRTC Streaming Client。
:::

:::tip 推荐的 GPU 模式
Isaac Lab 和 Isaac Sim 使用大量 GPU 资源。为了获得更好的稳定性，请将 GPU 模式设置为 [**应用独占**](/zh/manual/olares/settings/gpu-resource.md#切换资源模式)，并在运行工作负载前将 GPU 关联到 Isaac Lab。这有助于减少与其他 GPU 应用的资源争用。
:::

## 了解 Isaac Lab 和 Isaac Sim

Isaac Lab 和 Isaac Sim 服务于不同的目的：

| 组件 | 描述 | 使用场景 |
|:--|:---|:--|
| Isaac Lab | 基于 Isaac Sim 构建的机器人学习框架。 | 运行预定义演示、强化学习任务或训练脚本。 |
| Isaac Sim | 与 Isaac Lab 捆绑的底层仿真器。 | 直接启动仿真器进行场景探索、仿真测试或交互式使用。 |

- 要运行 Isaac Lab 演示或训练任务，请使用 [运行 Isaac Lab 工作负载](#运行-isaac-lab-工作负载) 中的命令。
- 要启动 Isaac Sim 仿真器本身，请使用 [运行 Isaac Sim](#运行-isaac-sim) 中的命令。

## 安装 Isaac Lab

1. 打开 Market 并搜索 "IsaacLab"。

   ![Isaac Lab](/images/manual/use-cases/isaac-lab.png#bordered){width=90%}

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 准备 Isaac Lab 终端

每当你打开 Isaac Lab 或想要切换到新任务时，请先准备终端环境。

1. 从 Launchpad 打开 Isaac Lab，访问其终端界面以运行命令行脚本。

   ![Isaac Lab 终端](/images/manual/use-cases/isaac-lab-terminal.png#bordered)

2. 获取用于 WebRTC 流式传输的公共端点地址。

   ```bash
   echo $PUBLIC_IP
   ```
   示例输出：
   ```plain
   192.168.50.169
   ```
   保存输出值。稍后连接 WebRTC Streaming Client 时需要用到它。

3. 检查 GPU 状态并查找现有工作负载。

   ```bash
   nvidia-smi
   ```
   - 如果没有工作负载在运行，`Processes` 部分将显示 `No running processes found`。

   ![无运行进程](/images/manual/use-cases/isaac-lab-no-running-process.png#bordered){width=90%}

   - 如果 `Processes` 部分列出了 Isaac Lab 或 Isaac Sim 进程，请在启动新工作负载前记下其 `PID`。在此示例中，活动进程使用 PID `2314`。

   ![活动进程](/images/manual/use-cases/isaac-lab-active-process.png#bordered){width=90%}

4. 如果现有工作负载正在运行，请在启动新工作负载前停止它：

   ```bash
   kill -9 <PID>
   ```

   将 `<PID>` 替换为终端中显示的实际进程 ID。

## 运行 Isaac Lab 工作负载

Isaac Lab 工作负载从终端启动。使用的命令取决于安装 Isaac Lab 的 Olares 主机的架构。

在启动新演示或任务前，停止任何现有工作负载。一次只运行一条命令，并在前台执行。官方 Isaac Lab 文档推荐 `uv run isaaclab`。`./isaaclab.sh` 仍可使用相同参数。

### 运行演示

<Tabs>
<template #AMD64-host>

当 Isaac Lab 运行在 AMD64 Olares 主机上且你希望启用 WebRTC 流式传输时，使用此命令。

1. 启动四足运动演示：

   ```bash
   uv run isaaclab -p scripts/demos/quadrupeds.py --livestream 1
   ```

2. 等待工作负载启动，并让本终端中的日志趋于稳定。

   启动期间可能会出现来自 Isaac Sim、Omniverse 或 PyTorch 的警告。如果没有出现致命错误且进程保持运行，请继续下一步。

3. 在本地计算机上打开 WebRTC Streaming Client，输入 `PUBLIC_IP` 值，然后点击 **连接**。

   ![连接 Streaming client](/images/manual/use-cases/isaac-lab-connect-streaming-client.png#bordered){width=90%}

4. 查看仿真进度。

   ![查看仿真进度](/images/manual/use-cases/isaac-lab-view-simulation.png#bordered){width=90%}
</template>

<template #ARM64-host>

当 Isaac Lab 运行在 ARM64 Olares 主机上（如 DGX Spark）时，使用此命令。

在 ARM64 Olares 主机上，不支持 WebRTC 流式传输。不要加 `--livestream`。

```bash
uv run isaaclab -p scripts/demos/quadrupeds.py
```

由于 ARM64 Olares 主机上不支持 WebRTC 流式传输，你无法通过 WebRTC Streaming Client 可视化演示。

</template>
</Tabs>

### 运行强化学习任务

<Tabs>
<template #AMD64-host>

当 Isaac Lab 运行在 AMD64 Olares 主机上且你希望启用 WebRTC 流式传输时，使用此命令。

1. 启动强化学习训练任务：

   ```bash
   uv run isaaclab train --rl_library rsl_rl --task Isaac-Velocity-Rough-H1 --livestream 1
   ```

2. 等待工作负载启动，并让本终端中的日志趋于稳定。

   启动期间可能会出现来自 Isaac Sim、Omniverse 或 PyTorch 的警告。如果没有出现致命错误且进程保持运行，请继续下一步。

3. 在本地计算机上打开 WebRTC Streaming Client，输入 `PUBLIC_IP` 值，然后点击 **连接**。

   ![连接 Streaming client](/images/manual/use-cases/isaac-lab-connect-streaming-client.png#bordered){width=90%}

4. 查看训练进度。

   ![查看 RL 训练进度](/images/manual/use-cases/isaac-lab-view-RL-training.png#bordered){width=90%}

</template>

<template #ARM64-host>

当 Isaac Lab 运行在 ARM64 Olares 主机上（如 DGX Spark）时，使用此命令。

在 ARM64 Olares 主机上，不支持 WebRTC 流式传输。不要加 `--livestream`。

```bash
uv run isaaclab train --rl_library rsl_rl --task Isaac-Velocity-Rough-H1
```

由于 ARM64 Olares 主机上不支持 WebRTC 流式传输，你无法通过 WebRTC Streaming Client 可视化训练过程。

</template>
</Tabs>

本页使用的训练命令见：

- [使用 RL 智能体训练](https://isaac-sim.github.io/IsaacLab/release/3.0.0/source/how-to/run_rl_training.html)

## 运行 Isaac Sim

Isaac Lab 包含 Isaac Sim 仿真器。当你想要使用仿真器本身而不是运行特定的 Isaac Lab 演示或训练脚本时，直接启动 Isaac Sim。

在启动 Isaac Sim 前，停止任何现有工作负载。

<Tabs>
<template #AMD64-host>

当 Isaac Sim 运行在 AMD64 Olares 主机上且你希望启用 WebRTC 流式传输时，使用此命令。

1. 启动 Isaac Sim：

   ```bash
   /isaac-sim/runheadless.sh
   ```

2. 等待本终端显示 Isaac Sim 流式传输应用已加载。

   例如：

   ```text
   Isaac Sim Full Streaming App is loaded.
   ```

3. 在本地计算机上打开 WebRTC Streaming Client，输入 `PUBLIC_IP` 值，然后点击 **连接**。

   ![连接 Streaming client](/images/manual/use-cases/isaac-lab-connect-streaming-client.png#bordered){width=90%}

如果流式传输视图变为灰色，请参阅 [流式传输视图变为灰色](#流式传输视图变为灰色)。

</template>

<template #ARM64-host>

当 Isaac Sim 运行在 ARM64 Olares 主机上（如 DGX Spark）时，使用此命令。

在 ARM64 Olares 主机上，不支持 WebRTC 流式传输。不要使用 `runheadless.sh` 进行流式传输。

以非图形模式启动 Isaac Sim：

```bash
./_isaac_sim/runapp.sh
```

由于 ARM64 Olares 主机上不支持图形流式传输，此模式主要用于有限的非图形用途。

</template>
</Tabs>

## 终端命令参考

命令在前台运行。关闭 WebRTC Streaming Client 不会停止工作负载。在终端按 **Ctrl + C** 即可停止。

使用以下命令进行快速终端管理和故障排除：

| 任务 | 命令 |
|:-----|:--------|
| 获取 WebRTC 端点地址 | `echo $PUBLIC_IP` |
| 检查 GPU 进程 | `nvidia-smi` |
| 停止残留进程 | `kill -9 <PID>` |

## 故障排除

### WebRTC Client 无法连接

检查以下内容：

1. 确保工作负载仍在运行：

   ```bash
   nvidia-smi
   ```

2. 确认你输入了正确的 `PUBLIC_IP` 值：

   ```bash
   echo $PUBLIC_IP
   ```

3. 确保只有一个 WebRTC Streaming Client 已连接。

4. 阅读运行工作负载的那个终端，查找以下错误：

```text
Error
Traceback
CUDA out of memory
Segmentation fault
Killed
```

如果没有出现这些错误且进程仍在运行，请等待日志稳定，然后重新连接或点击 WebRTC Streaming Client 中的 **重新加载**。

### 流式传输视图变为灰色

加载复杂场景时，流式传输视图可能会暂时变为灰色或停止响应。等待终端日志稳定。根据场景复杂度，这可能需要一到几分钟。

日志稳定后，在 WebRTC Streaming Client 中点击 **视图** > **重新加载** 以重新连接到流。

![重新加载 WebRTC 流](/images/manual/use-cases/isaac-lab-reload-WebRTC-stream.png#bordered){width=90%}

### 启动时日志显示警告

工作负载启动期间可能会出现来自 Isaac Sim、Omniverse 或 PyTorch 的警告。这些警告并不总是意味着工作负载已失败。

如果进程在 `nvidia-smi` 中保持运行且日志中没有出现致命错误，请继续执行 WebRTC 连接步骤。

## 了解更多

- [Isaac Lab 文档](https://isaac-sim.github.io/IsaacLab/release/3.0.0/index.html)：官方 Isaac Lab 文档，包括教程、环境、强化学习工作流和 API 参考。
- [Isaac Sim 文档](https://docs.isaacsim.omniverse.nvidia.com/6.1.0/index.html)：NVIDIA Isaac Sim 用户指南。
