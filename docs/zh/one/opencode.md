---
outline: [2, 3]
description: 了解如何在 Olares One 上安装 Qwen3.6-27B-GGUF 模型应用，将其连接到 OpenCode，并通过 Olares CLI 技能用自然语言管理设备。
head:
  - - meta
    - name: keywords
      content: OpenCode, Qwen3.6, llama.cpp, 本地 LLM, Olares CLI 技能, AI 助手, Olares One
---

:::warning
本页面由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/opencode.md)。
:::

# 使用 OpenCode 和本地模型管理 Olares <Badge type="tip" text="30 min" />

OpenCode 是一款 AI 编程助手。将它与 **Qwen3.6-27B-GGUF (llama.cpp)** 等能力出色的本地模型配对后，你可以通过自然语言管理 Olares 设备，而无需将数据发送到云端。

本指南将带你完成安装模型应用、连接 OpenCode、在 OpenCode 中登录 Olares CLI，以及使用 Olares 技能完成几个任务的全过程。

## 学习目标

- 安装 **Qwen3.6-27B-GGUF (llama.cpp)** 模型应用，并获取模型控制台端点。
- 安装 OpenCode 并连接本地模型。
- 在 OpenCode 中登录 Olares CLI 并加载 Olares 技能。
- 在 OpenCode 中通过自然语言管理 Olares，从简单问答到安装和移植应用。

## 前提条件

**硬件** <br>
- Olares One 连接到稳定的网络。
- 足够的磁盘空间用于下载模型。
- 建议至少具备 16 GB GPU VRAM 来运行 Qwen3.6 27B Q4_K_M。

**用户权限**
- 管理员权限，用于从 Market 安装共享应用和管理 GPU 资源。

## 步骤 1：安装 Qwen3.6-27B-GGUF 模型应用并获取端点

1. 打开 Market，搜索 **Qwen3.6-27B-GGUF (llama.cpp)**。

   <!-- ![安装 Qwen3.6-27B-GGUF](/images/one/qwen3.6-27b-gguf-market.png#bordered){width=90%} -->

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

3. 安装完成后，点击 **Open**。模型控制台会自动打开。

   <!-- ![Qwen3.6-27B-GGUF 模型控制台](/images/one/qwen3.6-27b-gguf-console.png#bordered){width=90%} -->

4. 等待模型下载完成。在 **Service status** 下：
   - **Model** 显示 **Ready**。
   - **Engine** 显示 **Running**。

5. 配置 OpenCode 访问该服务的方式：
   - **Connection source**：选择 **Apps in Olares**。
   - **API format**：选择 **OpenAI-Compatible**。
   - 复制 **Base URL**。
   - 记下控制台中显示的 **Model name**，例如 `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`。

   <!-- ![模型控制台端点](/images/one/qwen3.6-27b-gguf-endpoint.png#bordered){width=90%} -->

   :::tip 为什么使用模型控制台的 Base URL？
   此处暴露的 Base URL 是 Olares 内其他应用可访问的服务端点，可避免使用应用页面 URL 时可能出现的用户绑定路由、登录提示和跨域限制（CORS）问题。
   :::

## 步骤 2：安装 OpenCode

1. 打开 Market，搜索 "OpenCode"。

   <!-- ![安装 OpenCode](/images/manual/use-cases/opencode.png#bordered){width=90%} -->

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

3. 从启动台打开 **OpenCode**。首次启动时，OpenCode 会下载依赖包，根据网络情况可能需要 10 到 30 分钟。

   :::tip 查看初始化进度
   如需查看下载进度，可打开 Control Hub，选择 OpenCode 项目，进入 **Deployments** > **opencode**，点击运行中的 pod，然后查看 **init-packages** 容器的日志。
   :::

## 步骤 3：将 OpenCode 连接到模型

1. 在 OpenCode 中，点击左下角的 <span class="material-symbols-outlined">settings</span>。

   <!-- ![OpenCode 设置](/images/manual/use-cases/opencode-settings.png#bordered){width=70%} -->

2. 选择 **Providers**，然后向下滚动，点击 **Custom Provider** 旁边的 **Connect**。

   <!-- ![OpenCode 自定义提供方](/images/manual/use-cases/opencode-custom-provider.png#bordered){width=70%} -->

3. 输入以下信息：
   - **Provider ID**：唯一标识符。例如，`olares-qwen3.6`。
   - **Display name**：在提供方列表中显示的名称。例如，`Olares Qwen3.6 27B`。
   - **Base URL**：步骤 1 中复制的 **Base URL**。确保它以 `/v1` 结尾。
   - **Models**：
     - **Model ID**：步骤 1 中记下的精确 **Model name**。
     - **Display Name**：模型显示名称。例如，`Qwen3.6 27B`。

   <!-- ![OpenCode 提供方配置](/images/manual/use-cases/opencode-provider-config.png#bordered){width=70%} -->

4. 点击 **Submit** 保存配置。

5. 在聊天框下方，点击 **Big Pickle** 打开模型选择器，选择刚刚添加的模型。

## 步骤 4：登录 Olares CLI 并加载技能

OpenCode 可以使用 Olares CLI Agent Skills 来管理你的设备。在助手运行命令之前，你必须先在 OpenCode 中登录 Olares CLI。

1. 在 OpenCode 中，点击右上角的 <span class="material-symbols-outlined">terminal_2</span> 打开终端面板。

   <!-- ![OpenCode 终端面板](/images/manual/use-cases/opencode-web-terminal.png#bordered){width=90%} -->

2. 登录你的 Olares 账号：

   ```bash
   olares-cli profile login --olares-id <你的-olares-id>
   ```

   例如：

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

3. 按提示输入 Olares 登录密码。输入时密码不会显示。

4. 验证登录状态：

   ```bash
   olares-cli profile list
   ```

   输出中应显示你的 profile 状态为 `logged-in`。

5. 安装 Olares CLI Agent Skills：

   ```bash
   npx skills add beclab/Olares -y -g
   ```

   这会加载 profile、files、market、settings、dashboard 和 cluster 管理的技能包。

   :::tip 技能已安装？
   部分 OpenCode 镜像或环境可能已经内置 Olares 技能。如果命令提示技能已存在，可跳过此步骤。
   :::

## 步骤 5：通过自然语言管理 Olares

模型已连接且 Olares CLI 已认证后，你就可以让 OpenCode 管理设备了。以下示例从简单到复杂依次递进。

:::info 任务列表可能会调整
以下示例仅作为起点，你可以根据实际工作流程增加、删除或调整任务。
:::

### 简单问答

先问一个简单问题，确认 Olares 技能可用：

```text
What apps are currently installed on my Olares One?
```

或询问资源使用情况：

```text
Show me the CPU and memory usage of my Olares device.
```

### 从 Market 安装应用

让 OpenCode 帮你安装一个应用：

```text
Install Firefox from the Olares Market and tell me when it's ready.
```

OpenCode 会使用 `olares-market` 技能查找、安装并确认应用状态。

### 移植应用到 Olares

进阶任务：让 OpenCode 上传并部署一个现有项目：

```text
Upload and deploy this GitHub repository as an Olares app:
https://github.com/chandruk4321/dockerize-static-web-project
```

跟随 OpenCode 的提示和确认步骤，直到应用部署到 **My Olares**。

## 故障排查

### 模型控制台显示 Model 或 Engine 未就绪

如果 **Model** 未显示 **Ready** 或 **Engine** 未显示 **Running**：

- 确保安装时选择了 GPU 加速器。
- 检查 GPU VRAM 是否足够运行 Qwen3.6 27B Q4_K_M。
- 在模型控制台中进入 **GPU residency** 部分，点击 **Detect**，确认模型正在 GPU 上运行。

### OpenCode 无法连接模型

- 确保 **Base URL** 以 `/v1` 结尾。
- 确保模型控制台中的 **API format** 设置为 **OpenAI-Compatible**。
- 确保 OpenCode 中的 **Model ID** 与模型控制台中显示的 **Model name** 完全一致。

### OpenCode 内的 Olares CLI 命令失败

- 确保已运行 `olares-cli profile login`，且 profile 状态为 `logged-in`。
- 确保已通过 `npx skills add beclab/Olares -y -g` 安装 Olares 技能。
- 如果某个技能未被触发，可在提示中显式指定。例如："Using the olares-market skill, install Firefox."

## 资源

- [Set up OpenCode as your AI coding agent](../use-cases/opencode.md)：完整的 OpenCode 设置指南。
- [Install and use Agent Skills](../developer/cli-agent-skills.md)：Olares CLI 技能包详情。
- [切换 GPU 模式](gpu.md)
