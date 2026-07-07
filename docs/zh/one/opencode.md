---
outline: [2, 3]
description: 了解如何在 Olares One 上使用 OpenCode 配合本地模型及其内置的 Olares CLI Agent Skills，通过自然语言管理设备。
head:
  - - meta
    - name: keywords
      content: OpenCode, Qwen3.6 27B, llama.cpp, 本地 LLM, Olares CLI 技能, AI 助手, Olares One
---

:::warning
本页面由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/opencode.md)。
:::

# 通过自然语言管理 Olares <Badge type="tip" text="1 h" />

`olares-cli` 是用于管理 Olares 的命令行工具。为了让 AI 智能体能够使用它，Olares 提供了 CLI Agent Skills，即以工具定义的形式将自然语言转换为正确的 `olares-cli` 命令。它们覆盖常见的任务，例如列出文件、从 Market 安装应用、查看系统指标以及部署自定义应用。

Olares 上的智能体应用已内置这些技能。本指南以 OpenCode 为例，带你完成安装 Qwen3.6-27B (llama.cpp) 模型应用、将其连接到 OpenCode、使用 Olares 账号认证 Olares CLI，并通过聊天完成几个常见任务。

## 学习目标

- 安装 Qwen3.6-27B (llama.cpp) 模型应用，并获取其连接信息。
- 安装 OpenCode 并连接该模型。
- 使用 Olares 账号认证 Olares CLI。
- 在 OpenCode 中通过自然语言管理 Olares。

## 前提条件

**系统要求**
- Olares OS 已升级至 v1.12.6。
- 如果你之前安装过 OpenCode，请在升级 Olares OS 后将其卸载，然后重新安装。

**硬件** <br>
- Olares One 连接到稳定的网络。
- 足够的磁盘空间用于下载模型及其依赖。
- Qwen3.6 27B 至少需要 23 Gi GPU 内存。

**用户权限**
- 管理员权限，用于从 Market 安装共享应用和管理 GPU 资源。

## 步骤 1：安装模型应用并获取连接信息

1. 打开 Market，搜索 **Qwen3.6-27B (llama.cpp)**。

   ![安装 Qwen3.6-27B](/images/one/qwen3.6-27b-llamacpp-market.png#bordered)

2. 点击 **Get**，然后点击 **Install**。
3. 选择 **GPU** 作为硬件加速器，然后点击 **Confirm**。安装开始。

   :::tip 首次部署耗时较长
   首次部署可能需要较长时间，因为需要下载模型权重。具体时间取决于模型大小和网络速度。
   :::

4. 安装完成后，点击 **Open**。模型控制台会自动打开，模型也会自动开始下载。

   :::tip 首次下载耗时较长
   首次打开模型控制台时，下载模型文件可能需要一些时间，具体取决于文件大小和网络速度。
   :::

5. 等待下载完成。当看到以下状态时即表示就绪：
   - **Model**：**READY**
   - **Engine**：**RUNNING**

   ![Qwen3.6-27B 模型控制台](/images/one/qwen3.6-27b-model-console.png#bordered)

6. 配置 OpenCode 访问该服务的方式：

   - **Connection source**：选择 **Apps in Olares**。
   - **API format**：选择 **OpenAI-Compatible**。
   - 记下 **Base URL**。例如：`https://b11a5b8a.laresprime.olares.com/v1`。
   - 记下 **Model name**。例如：`unsloth/Qwen3.6-27B-GGUF:Q4_K_M`。

## 步骤 2：安装 OpenCode

1. 打开 Market，搜索 "OpenCode"。

   ![安装 OpenCode](/images/manual/use-cases/opencode.png#bordered)

2. 点击 **Get**，然后点击 **Install**。安装完成后，启动台上会出现两个快捷方式：

   - **OpenCode**：用于与智能体对话和管理项目的图形化 Web 界面。
   - **OpenCode Terminal**：用于运行 CLI 命令或启动 TUI（Terminal User Interface）的终端。

3. 打开 **OpenCode**。

   首次启动时，OpenCode 会下载依赖包，根据网络情况可能需要 10 到 30 分钟。

   :::tip 查看初始化进度
   如需查看下载进度，可打开 Control Hub，选择 OpenCode 项目，进入 **Deployments** > **opencode**，点击运行中的 pod，然后查看 **init-packages** 容器的日志。
   :::

## 步骤 3：将 OpenCode 连接到模型

1. 在 OpenCode 中，点击左下角的 <span class="material-symbols-outlined">settings</span>。

   ![OpenCode 设置](/images/manual/use-cases/opencode-settings.png#bordered)

2. 选择 **Providers**，然后向下滚动，点击 **Custom provider** 旁边的 **Connect**。

   ![OpenCode 自定义提供方](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. 配置以下设置：
   - **Provider ID**：唯一标识符。例如，`olares-engine-base`。
   - **Display name**：在提供方列表中显示的名称。例如，`local-llamacpp`。
   - **Base URL**：步骤 1 中记下的 **Base URL**。确保它以 `/v1` 结尾。
   - **Models**：
     - **Model ID**：步骤 1 中记下的精确 **Model name**。
     - **Display Name**：模型显示名称。例如，`Qwen3.6 27B`。

   ![OpenCode 提供方配置](/images/one/opencode-provider-config.png#bordered){width=70%}

4. 点击 **Submit** 保存配置。
5. 开始一个新聊天。
6. 在聊天框下方，点击 **Big Pickle** 打开模型选择器，然后选择刚刚添加的模型。

## 步骤 4：使用 Olares 账号认证 Olares CLI

在 OpenCode 能够代你运行 Olares CLI Agent Skills 之前，你必须使用 Olares 账号认证 Olares CLI。

1. 在 OpenCode 中，点击页面顶部的 **Search project**，然后选择 **Toggle terminal**。

   ![OpenCode 终端面板](/images/one/opencode-terminal.png#bordered)

2. 运行以下命令登录你的 Olares 账号。将 `<你的-olares-id>` 替换为你的真实 Olares ID。

   ```bash
   olares-cli profile login --olares-id <你的-olares-id>
   ```

   例如：

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

3. 按提示输入 Olares 登录密码。输入时密码不会显示。

4. 运行以下命令验证 profile 已创建并处于登录状态：

   ```bash
   olares-cli profile list
   ```

   示例输出：

   ```text
      NAME                   OLARES-ID              STATUS
   *  laresprime@olares.com  laresprime@olares.com  logged-in
   ```

## 步骤 5：通过自然语言管理 Olares

模型已连接且 Olares CLI 已认证后，你就可以在 OpenCode 中通过自然语言对话来管理 Olares 设备了。以下示例覆盖了一些常见场景。

### 提问

先问一个简单问题，确认 Olares 技能可用：

```text
Show me the CPU and memory usage of my Olares device.
```

![在 OpenCode 中提问](/images/one/onboard-scenario-question.png#bordered)

### 从 Market 安装应用

让 OpenCode 帮你安装一个应用：

```text
Install Chromium from the Olares Market and tell me when it's ready.
```

![在 OpenCode 中使用 Olares skill 安装应用](/images/one/onboard-scenario-install.png#bordered)

### 部署模型

让 OpenCode 使用新的引擎基座应用部署指定模型：

```text
Deploy qwen3.5:0.8b on this Olares device using the Ollama Engine Base.
```

![在 OpenCode 中使用 Olares skill 部署模型](/images/one/onboard-scenario-model-deployment.png#bordered)

模型部署完成后，模型应用会出现在启动台上。打开它即可直接进入模型控制台。

### 移植应用到 Olares

进阶任务：让 OpenCode 上传并部署一个 GitHub 仓库中的项目：

```text
Upload and deploy this GitHub repository as an Olares app:
https://github.com/chandruk4321/dockerize-static-web-project
```

跟随 OpenCode 的提示和确认步骤，直到部署完成。然后你可以在启动台和 **My Olares** 中找到该应用。

![在 OpenCode 中使用 Olares skill 部署应用](/images/one/onboard-scenario-port.png#bordered)

## 资源

- [将 OpenCode 设置为你的 AI 编码代理](../use-cases/opencode.md)：完整的 OpenCode 设置使用指南。
- [安装与使用 Agent Skills](../developer/cli-agent-skills.md)：Olares CLI 技能包详情。
- [切换 GPU 模式](gpu.md)
