---
outline: [2, 3]
description: 在 Olares 上把 OpenClaw 作为自托管私人 AI 助手运行：完成安装、配置，并接入 Discord 和 Slack，助手与数据都留在你自己的设备上。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, self-hosted ai agent, personal ai agent, local ai agent, openclaw on olares
app_version: "1.0.3"
doc_version: "2.0"
doc_updated: "2026-05-29"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/openclaw.md)。
:::

# 运行 OpenClaw：你的自托管私人 AI 助手

OpenClaw 是一款专为本地设备设计的个人 AI 助手，可接入 Discord、Slack 等消息平台，在这些平台中与其交互。

它相当于一位"始终在线"的助手，能够执行搜索和发送文档、管理日历、浏览网页等实际任务。

## 学习目标

在本指南中，你将学习如何：
- 安装并初始化 OpenClaw 环境。
- 将 OpenClaw 与 Discord 等聊天应用集成。
- 可选：启用网页搜索功能。
- 管理技能和插件。
- 使用 OpenClaw 助手管理 Olares 设备。
- 可选：启用沙箱以实现安全代码执行。

## 前提条件

- 本地模型：Ollama 或其他模型提供方须已安装并运行。

    :::tip 模型提供方
    本教程使用 Ollama 作为模型提供方。如使用其他提供方或本地代理，请参阅 [OpenClaw 关于使用自定义模型提供商的文档](https://docs.openclaw.ai/zh-CN/concepts/model-providers#%E9%80%9A%E8%BF%87-models.providers-%E4%BD%BF%E7%94%A8%E6%8F%90%E4%BE%9B%E5%95%86%EF%BC%88%E8%87%AA%E5%AE%9A%E4%B9%89%2F%E5%9F%BA%E7%A1%80-url%EF%BC%89)了解配置详情。
    :::
- Discord 账号：用于创建机器人应用。
- Discord 服务器：确保你在这个服务器上有添加机器人的权限。

## 升级说明

如正在升级现有的 OpenClaw，请在升级之前，查看版本特定的更改和故障排除步骤。更多信息，请参阅[升级 OpenClaw](openclaw-upgrade.md)。

## 安装 OpenClaw

1. 打开应用市场，搜索 "OpenClaw"。

    ![在应用市场中搜索 OpenClaw](/images/zh/manual/use-cases/find-openclaw1.png#bordered){width=90%}

2. 点击**获取**，然后点击**安装**。安装完成后，启动台会出现两个快捷方式：
    - **OpenClaw CLI**：命令行界面
    - **Control UI**：图形化管理面板

    ![OpenClaw 入口](/images/manual/use-cases/openclaw-entry-points1.png#bordered){width=30%}

:::tip 运行多个 OpenClaw 实例
Olares 支持克隆应用。如需同时运行多个独立的 AI 助手处理不同任务，可克隆 OpenClaw 应用。更多信息，请参阅[克隆应用](/zh/manual/olares/market/clone-apps.md)。
:::

## 初始化 OpenClaw

快速完成助手的初始化配置。

### 步骤 1：安装模型

安装一个支持工具调用的模型，例如 `glm-4.7-flash`、`qwen3.5:35b` 或 `gpt-oss:20b`。本教程使用的是 `qwen3.5:35b`。

:::tip
OpenClaw 需要较大的"上下文窗口"（即 AI 的短期记忆）来处理复杂任务而不会忘记之前的指令。如使用本地模型，建议选择原生支持至少 64K token 上下文窗口的模型。
:::

<Tabs>
<template #(推荐)-从应用市场下载>

1. 在应用市场中，搜索 "Qwen3.5 35B A3B UD-Q4 (Ollama)"。

    ![在应用市场中查找模型应用](/images/zh/manual/use-cases/qwen35b.png#bordered)    
2. 点击**获取**，然后点击**安装**。
3. 安装完成后，点击**打开**。模型将自动开始下载。
4. 模型下载完成后，准确记录 **Model Name** 处显示的模型名称。例如，`qwen3.5:35b-a3b-ud-q4_K_L`。后续配置将用到该名称。

    ![记录模型详细信息](/images/manual/use-cases/obtain-model-details2.png#bordered)

5. 打开设置，进入**应用** > **Qwen3.5 35B A3B UD-Q4 (Ollama)** > **共享入口**。

    ![在设置中获取模型的共享端点](/images/zh/manual/use-cases/obtain-model-details3.png#bordered){width=70%}

6. 点击 **Qwen3.5 35B A3B UD-Q4_K_L**，然后记录共享端点的 URL。例如，`http://026076110.shared.olares.com`。

:::tip 为什么不使用模型应用页面显示的 URL？
模型应用页面显示的 URL 与用户绑定，且依赖浏览器前端调用。当设备与 Olares 不在同一本地网络时，这些调用可能会触发 Olares 登录，并受跨域限制（CORS）的影响。为避免上述问题，建议使用共享端点 URL。
:::
</template>
<template #通过-Ollama-下载>

1. 运行以下命令，查看已安装的模型列表：

    ```bash
    ollama list
    ```
2. 复制并保存 **NAME** 列中显示的模型名称。例如，`qwen3.5:27b`。
3. 如模型未安装，请先下载并运行。详细步骤，参见 [Ollama](ollama.md)。
4. 打开设置，进入**应用** > **Ollama** > **共享入口**。

   ![设置中的 Ollama 端点](/images/zh/manual/use-cases/ollama-endpoint.png#bordered){width=80%}

5. 点击 **Ollama API**，然后记录共享端点的 URL。例如，`http://d54536a50.shared.olares.com`。
</template>
</Tabs>

### 步骤 2：验证模型的可访问性

在配置 OpenClaw 之前，先验证模型是否可通过 API 正常访问和响应。

1. 从启动台打开 OpenClaw CLI 应用。
2. 输入以下命令验证 API 地址，并获取可用模型列表。将 `{Your-Model-API}` 替换为[步骤 1](#步骤-1-安装模型)中获取的共享端点 URL。

    ```bash
    curl {Your-Model-API}/api/tags
    ```

    例如：
    ```bash
    curl http://026076110.shared.olares.com/api/tags
    ```

    终端将返回可用模型的详细信息，表明 API 可访问。例如：

    ```text
    {"models":[{"capabilities":["completion","tools","thinking"],"details":{"context_length":262144,"embedding_length":2048,"families":["qwen35moe"],"family":"qwen35moe","format":"gguf","parameter_size":"34.7B","parent_model":"qwen3.5:35b-a3b-ud-q4_K_L-base","quantization_level":"Q8_0"},"digest":"e8cb37adef5d1325d7fed17ec8124d37cb6ba5f2f357887811d75a139ddb79dc","model":"qwen3.5:35b-a3b-ud-q4_K_L","modified_at":"2026-05-27T07:07:06.19481654Z","name":"qwen3.5:35b-a3b-ud-q4_K_L","size":20205634377}]}
    ```    
 
3. 输入以下命令，强制将模型加载到内存中并测试其响应速度。将 `{Your-Model-API}` 和 `{Your-Model-Name}` 替换为[步骤 1](#步骤-1-安装模型)中获取的详细信息。

    :::info 为什么要在初始化前执行此操作？
    Ollama 默认在模型空闲 5 分钟后将其从内存中卸载。重新加载大型模型需要一定时间，可能导致后续初始化验证超时。此命令可"唤醒"模型，确保初始化顺利完成。
    ::: 

    ```bash
    curl {Your-Model-API}/api/generate -d '{
    "model": "{Your-Model-Name}",
    "prompt": "say hello world",
    "stream": false
    }'
    ```
    例如：

    ```bash
    curl http://026076110.shared.olares.com/api/generate -d '{
    "model": "qwen3.5:35b-a3b-ud-q4_K_L",
    "prompt": "say hello world",
    "stream": false
    }'
    ```

    终端将返回包含 `Hello World!` 的响应，表明模型已就绪。例如：

    ```text
    {"model":"qwen3.5:35b-a3b-ud-q4_K_L","created_at":"2026-05-27T07:17:52.542888337Z","response":"Hello, World! 🌍","done":true,"done_reason":"stop","context":[248045,846,198,35571,23066,1814,593,26003,248046,198,248045,74455,198,248068,271,248069,271,9419,11,4196,0,10838,234,235],"total_duration":22384074696,"load_duration":197437036,"prompt_eval_count":13,"prompt_eval_duration":22064969000,"eval_count":12,"eval_duration":73444000}
    ```

### 步骤 3：运行安装向导

你可以通过交互式向导分步配置 OpenClaw，或直接运行命令完成设置。

<Tabs>
<template #交互式设置>

1. 从启动台打开 OpenClaw CLI 应用。
2. 输入以下命令，启动安装向导：
    ```bash
    openclaw onboard
    ```
3. 跟随安装向导，逐步完成设置。使用方向键移动，按**回车**键确认。

    :::tip 关于配置的说明
    为便于快速上手，本教程跳过了部分高级设置，你可稍后在配置中进行调整。
    :::

    | 配置   | 选项   |
    |:-----------|:---------|
    | I understand this is<br>personal-by-default and shared/multi-user use<br>requires lock-down. <br>Continue? | 选择 **Yes**。  |
    | Setup mode   | 选择 **QuickStart**。  |
    | Config handling  | 选择 **Keep current values**。    |
    | Model/auth provider  | 选择 **More**，然后选择 **Ollama**。    |
    | Ollama mode | 选择 **Local only**。 |
    | Ollama base URL  | 删除默认占位符文本，然后输入[步骤 1](#步骤-1-安装模型)中获取<br>的共享端点 URL。<br>例如，`http://026076110.shared.olares.com`。 |
    | Default model | 选择 **Browse all models**，然后选择你安装的模型。<br>例如，**ollama/qwen3.5:35b-a3b-ud-q4_K_L**。 |
    | Select channel  | 选择 **Skip for now**。<br>（可稍后配置聊天平台）  |
    | Search provider | 选择 **Skip for now**。<br>（可稍后配置网络搜索提供方）|
    | Configure skills now   | 选择 **No**。<br>（可稍后安装技能）       |
    | Enable hooks | 选择 **Skip for now**。<br>（按**空格**选择，然后按**回车**继续） |
    | How do you want to<br>hatch your bot   | 选择 **Hatch later**。   |

4. 安装向导完成后，向上滚动至 **Control UI** 部分。
5. 找到 **Web UI (with token)**，复制 URL 末尾的 token（`#token=` 后面的文本）。此为 Gateway Token。

    在本示例中，gateway token 为 `YrzY5wk1WYWIfcTHFodyO43Ge6n1JY4T`。

    ![获取 gateway token](/images/manual/use-cases/obtain-gateway-token3.png#bordered)

6. 保持 OpenClaw CLI 窗口开启。后续步骤仍需使用。
</template>
<template #命令行设置>

1. 从启动台打开 OpenClaw CLI 应用。
2. 输入以下命令，启动安装向导。将 `{Your-Model-API}` 和 `{Your-Model-Name}` 替换为[步骤 1](#步骤-1-安装模型)中获取的详细信息。

    ```bash
    openclaw onboard --non-interactive \
    --auth-choice ollama \
    --custom-base-url "{Your-Model-API}" \
    --custom-model-id "{Your-Model-Name}" \
    --accept-risk
    ```

    例如：

    ```text
    openclaw onboard --non-interactive \
    --auth-choice ollama \
    --custom-base-url "http://026076110.shared.olares.com" \
    --custom-model-id "qwen3.5:35b-a3b-ud-q4_K_L" \
    --accept-risk
    ```

    终端将显示包含助手信息的成功消息。例如：

    ```text
    Agents: main (default)
    Heartbeat interval: 30m (main)
    Session store (main): /home/node/.openclaw/agents/main/sessions/sessions.json (0 entries)
    Tip: run `openclaw configure --section web` to store your Brave API key for web_search. Docs: https://docs.openclaw.ai/tools/web
    ```

3. 输入以下命令验证模型配置是否正确：

    ```bash
    openclaw models status --probe
    ```

    **Auth probes** 表格中 **Status** 列显示 `ok`，表示模型已成功连接并可用。
    
4. 打开文件管理器，进入**应用** > **数据** > **clawdbot** > **config**，然后打开 `openclaw.json` 文件。
5. 找到 `gateway` 部分，记录 `auth` 中的 token。例如，`YrzY5wk1WYWIfcTHFodyO43Ge6n1JY4T`。

    ![在配置文件中获取 gateway token](/images/manual/use-cases/obtain-gateway-token-in-config.png#bordered)

6. 保持 OpenClaw CLI 窗口打开。后续步骤仍需使用。
</template>
</Tabs>

### 步骤 4：配对设备

将 Control UI 与 OpenClaw CLI 配对，以使用图形化管理面板。

<Tabs>
<template #(推荐)-自动配对设备>

1. 从启动台打开 Control UI 应用。**OpenClaw Gateway Dashboard** 将打开。

    ![Gateway 面板](/images/manual/use-cases/gateway-dashboard3.png#bordered)

    如出现 `Auth did not match` 错误，属正常情况，表明尚未提供访问 token。

2. 在 **Gateway Token** 字段中，输入上一步复制的 gateway token，然后点击 **Connect**。

    如出现 `Device pairing required` 错误，属正常情况，表明设备配对请求待批准。

    ![需要设备配对](/images/manual/use-cases/gateway-device-pairing-required.png#bordered)

3. 在 `Device pairing required` 错误消息中，找到 `Approve this request:` 行，复制其后显示的命令（不含末尾句点）。

    在本示例中，需复制的命令为 `openclaw devices approve 673b3923-cb85-4a82-a8c2-f9f8327d0761`。

4. 返回 OpenClaw CLI 窗口，然后运行复制的命令以授权 Control UI。

    :::tip 关于超时错误的说明
    批准命令有时效限制。如收到 `unknown requestId` 错误，表明请求已过期。请刷新 Control UI，复制新生成的命令，在 OpenClaw CLI 中重新运行。

    ```text
    [openclaw] The CLI command failed.
    [openclaw] Reason: unknown requestId
    ```
    :::

5. 当终端显示批准的消息时，返回 Control UI。

    ```text
    Approved 005748253152b66dc0f5f6a801f35617db043f107972f259630a6bd098d5f790 (967e3732-b3df-43e3-851a-d99b43198d8e)
    ```

6. 再次点击 **Connect**，登录后将默认进入 **Chat** 页面。
7. 从左侧边栏，点击 **Overview** 查看连接状态。**Snapshot** 面板中的 **STATUS** 应显示为 **OK**。

    ![健康状态 OK](/images/manual/use-cases/openclaw-connected3.png#bordered)
</template>
<template #(可选)-手动配对设备>

:::tip 何时使用手动配对
自动配对将批准最近一次的配对请求。如需在多个待处理请求中手动选择要批准的设备，请按本节步骤操作。
:::

1. 从启动台打开 Control UI 应用。**OpenClaw Gateway Dashboard** 将打开。

    ![Gateway 面板](/images/manual/use-cases/gateway-dashboard3.png#bordered)

    如出现 `Auth did not match` 错误，属正常情况，表明尚未提供访问 token。

2. 在 **Gateway Token** 字段中，输入上一步复制的 gateway token，然后点击 **Connect**。

    如出现 `Device pairing required` 错误，属正常情况，表明设备配对请求待批准。

3. 返回 OpenClaw CLI 窗口并输入以下命令：
    ```bash
    openclaw devices list
    ```
4. 在 **Pending** 表格中，找到与你当前设备关联的 **Request** ID。

    :::info
    Request ID 有时效限制。如授权失败，请重新运行 `openclaw devices list` 以获取新的有效 ID。
    :::

    ![查看待处理设备请求](/images/manual/use-cases/pending-request.png#bordered)
    
5. 输入以下命令授权设备：

    ```bash
    openclaw devices approve {RequestID}
    ```

6. 当终端显示批准消息时，返回 Control UI。**Snapshot** 面板中的 **STATUS** 应显示为 **OK**。

    ![健康状态 OK](/images/manual/use-cases/openclaw-connected3.png#bordered)
</template>
</Tabs>

### 步骤 5：配置上下文窗口

OpenClaw 需要较大的上下文窗口（即 AI 的短期记忆）来处理复杂任务而不会忘记之前的指令。

1. 打开文件管理器，进入**数据** > **clawdbot** > **config**。
2. 双击打开 `openclaw.json` 文件。
3. 点击右上角的 <i class="material-symbols-outlined">edit_square</i> 进入编辑模式。
4. 找到 `models` 部分并定位你的模型配置块。
5. 将 `contextWindow` 值更新为至少 65536（64K）。如硬件显存允许，强烈建议将其增加到 204800（200K）。

    ![在配置文件中配置上下文窗口](/images/manual/use-cases/configure-context-win3.png#bordered)

6. 点击右上角的 <i class="material-symbols-outlined">save</i>。
7. 重启 OpenClaw 使更改生效。

### 步骤 6：个性化 OpenClaw

为使 OpenClaw 机器人更具个性化，强烈建议完成人设设置。

人设文件用于确立助手的身份、行为边界和长期记忆，确保其在所有平台和频道上的行为保持一致。

1. 在 Control UI 中，从左侧边栏选择 **Chat**。
2. 确保右上角的 <i class="material-symbols-outlined">neurology</i> 已启用。启用后可实时观察助手调用工具及编辑人设文件的过程。
3. 输入并发送以下消息开始：

    ```text
    Wake up please!
    ```
4. 助手将响应并开始与你对话，在此过程中可建立规则、个性特征和偏好。例如：

    ```text
    - Call me Bella. I like simple language without technical jargon and 
    concise bulleted answers.
    - You are John, a witty assistant who uses emojis.
    - Never access my calendar without asking first, and never execute any 
    financial operations.
    ```

5. 与助手对话过程中，可实时观察其将你的偏好写入核心人设文件，例如 `IDENTITY.md`、`USER.md` 和 `SOUL.md`。

    ![OpenClaw 正在编辑人设文件](/images/manual/use-cases/openclaw-persona-recording3.png#bordered)

    :::tip
    如未看到中间的人设文件操作，请点击右上角的 <i class="material-symbols-outlined">refresh</i> 或按 **F5** 刷新页面。
    :::

6. （可选）如助手未能更新人设文件，可在聊天中明确要求其执行。

    若问题仍未解决，可尝试以下方法：
    - **增加上下文窗口**：从左侧边栏选择**配置**，切换到**原始配置**标签页，找到 `models` 部分，然后将 `contextWindow` 值增加到至少 64K（建议 200K）。
    
        :::tip
        较大的上下文窗口会消耗更多显存，请根据硬件配置选择合适的值。
        :::

    - **更换模型**：切换到具有更好的工具调用和指令遵循能力的模型。

7. 继续对话，直到助手收集到足够的信息。
8. 验证人设文件是否成功更新：

    a. 从启动台打开文件管理器。
    
    b. 进入**应用** > **数据** > **clawdbot** > **config** > **workspace**。
    
    c. 检查 `.md` 文件的修改时间，以确定哪些文件最近更新过，例如 `USER.md`、`IDENTITY.md` 和 `SOUL.md`。

    ![OpenClaw 生成的人设文件](/images/zh/manual/use-cases/openclaw-persona-files.png#bordered){width=90%}

    d. （可选）下载文件，在支持的文本编辑器中查看文件，并验证其中是否包含新建的规则，例如你的名字、语言风格和限制。

    e. 如果存在临时的 `BOOTSTRAP.md` 文件，删除该文件。

    :::tip 修改人设设置
    如需更改这些设置，请使用以下方法之一：
    - 在聊天中要求助手更新其规则。
    - 从此文件夹中，下载 `.md` 文件，在文本编辑器中编辑文件，然后重新上传以覆盖旧文件。
    :::

## 后续步骤

1. [与 Discord 集成](openclaw-integration.md)，实现与助手的远程对话。
2. [启用网页搜索](openclaw-web-access.md)，使助手能够访问实时互联网信息。
3. [安装技能和插件](openclaw-skills.md)，进一步扩展助手的能力。

## 故障排除和常见问题

如遇常见错误或行为问题，请参阅[常见问题](openclaw-common-issues.md)获取解决方案。

## 了解更多

- [如何在 Discord 中创建服务器](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)
