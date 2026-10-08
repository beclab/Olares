---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/paperclip
outline: [2, 3]
description: 在 Olares 上运行 Paperclip，协调多个 AI 智能体协同完成同一组任务。添加由 Claude Code、Codex、OpenCode、Cursor 或其他提供商支持的智能体，并为它们分配任务单。
head:
  - - meta
    - name: keywords
      content: Olares, Paperclip, AI agent, multi-agent, Claude Code, Codex, OpenCode, Cursor, self-hosted
app_version: "1.0.34"
doc_version: "1.2"
doc_updated: "2026-10-08"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/paperclip.md)为准。
:::

# 使用 Paperclip 协调多个 AI 智能体

<VersionRouteSelect />

Paperclip 是一个开源平台，用于在同一个统一工作区下协调多个 AI 智能体。通过设置虚拟公司，你可以添加由 Claude Code、Codex、OpenCode、Cursor 或其他提供商驱动的 AI 智能体，并为它们分配任务单。无论是编码、研究还是内容创作，Paperclip 都能管理工作流。

将 Paperclip 作为自托管应用运行在 Olares 上，可以确保你的 API 密钥、任务历史和智能体输出完全保留在你的设备上。

## 前提条件

开始前，你需要：

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- 如需使用本地模型，从应用市场安装 Qwen3.8-27B (llama.cpp)。

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Paperclip。
- 为计划使用的智能体配置 API 密钥。
- 设置初始管理员账户。
- 创建你的第一个公司、智能体和任务。
- 创建任务单并跟踪智能体进度。
- 从仪表板监控操作和指标。
- 可选：在 Paperclip 中使用本地模型。

## 升级说明

从 V1.0.22 开始，Paperclip 使用官方镜像，并需要新的环境依赖。升级后你的运行环境可能会发生变化。

- 如果你尚未初始化 Paperclip，或没有重要数据，请卸载应用，删除所有相关数据，然后重新安装。
- 如果 Paperclip 已初始化，你现有的数据会被保留，但可能需要重新配置本地智能体设置和环境依赖。如果升级后遇到问题，请联系 Olares 团队。

## 安装 Paperclip

1. 打开 Market 并搜索 "Paperclip"。

   ![Paperclip in Market](/images/manual/use-cases/paperclip.png#bordered)

2. 点击 **Get**，然后点击 **Install**。安装完成后，Launchpad 中会出现两个快捷方式：

   ![Paperclip entry points](/images/manual/use-cases/paperclip-install-entry.png#bordered){width=35%}

## 初始化 Paperclip

安装 Paperclip 后，创建第一个管理员账户以完成初始设置。如果你计划使用云端模型，请先配置 API 密钥。

### 配置 API 密钥

云端模型认证取决于提供商。下列环境变量适用于使用原有认证配置的适配器；采用新版 AI Connections 的智能体通过所选连接认证。本地模型不要求必须提供云端模型 API 密钥。

:::tip
在创建管理员账户之前，先在 **Settings** 中设置 API 密钥，以便云端模型可以立即使用。
:::

1. 打开 Settings，然后进入 **Applications** > **Paperclip** > **Manage environment variables**。

   ![Manage Paperclip environment variables](/images/manual/use-cases/paperclip-manage-env-vars.png#bordered){width=70%}

2. 点击变量旁边的 <i class="material-symbols-outlined">edit_square</i>，在 **Value** 字段中输入你的 API 密钥，然后点击 **Confirm**。

   Paperclip 支持以下变量：

   | Variable | Used by |
   |:---------|:--------|
   | `ANTHROPIC_API_KEY` | Claude Code, OpenCode, Pi |
   | `OPENAI_API_KEY` | Codex, OpenCode, Pi |
   | `GEMINI_API_KEY` | Gemini CLI, Cursor |
   | `CURSOR_API_KEY` | Cursor |

3. 点击 **Apply** 保存并应用新密钥。

:::tip 稍后添加更多 API 密钥

你可以随时返回此部分添加或更新密钥。重复此步骤并重启 Paperclip 以应用新配置。

:::

### 创建管理员账户

Paperclip 默认没有用户账户。要首次访问平台，你需要通过注册流程创建管理员账户。

1. 从 Launchpad 打开 Paperclip，然后点击 **Create account**。

2. 在注册页面填写所需信息并提交。

3. 注册完成后，点击 **Claim this instance** 以关联管理员账户。

   ![Claim this instance](/images/manual/use-cases/paperclip-claim-instance.png#bordered){width=60%}

4. 命名你的公司。使用云端模型时按指示完成引导；使用本地模型时继续阅读[在 Paperclip 中使用本地模型](#在-paperclip-中使用本地模型)。

## 创建你的第一个公司

下方截图展示旧版引导流程。`2026.916.1` 和 `2026.1005.0` 的初始引导主要提供 OpenAI 和 Claude；其他提供商及本地模型请参见[本地模型说明](#在-paperclip-中使用本地模型)。

Paperclip 工作区围绕虚拟公司结构组织。该公司用于组织你的智能体和任务。

1. 在 **Company** 标签页，配置基本信息：

   a. 为公司指定名称。

   b. （可选）指定使命或目标。

   ![Set company basics](/images/manual/use-cases/paperclip-set-company.png#bordered)

   c. 点击 **Next**。

2. 在 **Agent** 标签页，创建你的第一个智能体：

   a. 指定以下设置：

      - **Agent name**: 使用默认名称 **CEO** 或自定义名称。
      - **Adapter type**: 选择底层框架。例如，选择 **Claude Code (Local Claude agent)** 或 **Codex (Local Codex agent)**。
      - **More Agent Adapter Types**: 展开以选择其他选项，如 **OpenCode** 或 **Cursor**。
      - **Model**: 从下拉列表中选择特定的 AI 模型。

   b. 点击 **Test now** 验证配置是否与你的 API 密钥正常工作。

   c. 点击 **Next**。

   ![Create an agent](/images/manual/use-cases/paperclip-create-agent.png#bordered)

   :::tip
   要查看每个适配器的 API 密钥要求，请参阅 [Paperclip 支持哪些智能体适配器](#paperclip-支持哪些智能体适配器)。
   :::

3. 在 **Task** 标签页，定义你的第一个任务：

   a. **Task title** 和 **Description**：指定标题和描述，或保留默认值。

   b. 点击 **Next**。此任务会在设置完成后自动转换为你的第一个任务单。

   ![Set up a task](/images/manual/use-cases/paperclip-set-task.png#bordered)

4. 点击 **Create & Open Issue**。**Issues** 页面将打开，显示你的工作区。

   ![Issue page after setup](/images/manual/use-cases/paperclip-issue-page.png#bordered)

## 创建并跟踪任务单

在 Paperclip 中，所有工作都通过任务单进行。创建任务单时，Paperclip 会将其分配给现有智能体。如果没有适合该请求的特定智能体，Paperclip 会自动"雇佣"一个新智能体来完成工作。

1. 在 **Issues** 页面，点击左侧边栏中的 **New issue**。
2. 指定任务单的详细信息。例如：

   - **Issue title**: Write a guide on AI servers.
   - **Description**: Hire a writing agent to research and draft a 200-word article explaining the benefits of self-hosting AI models.
   - **Assignee**: 选择 **CEO** 来评估需求。

   ![New issue dialog](/images/manual/use-cases/paperclip-new-issue.png#bordered){width=60%}

3. 点击 **Create Issue**。Paperclip 会分配任务单，智能体开始工作。
4. 点击左侧边栏中的 **Inbox** 以监控传入请求和执行进度。如果分配的智能体决定委派任务，你的收件箱中会出现雇佣请求。

   ![Inbox tracking the new issue](/images/manual/use-cases/paperclip-inbox.png#bordered)

5. 进入左侧边栏的 **Agents** 部分，查看新雇佣的写手智能体及其详细信息。

   ![Writer Agent details](/images/manual/use-cases/paperclip-agent-details.png#bordered)

6. 查看智能体的输出：

   a. 进入 **Issues** 页面，然后选择任务单以查看其详细信息。直接在聊天记录中查找输出。

   ![Agent output](/images/manual/use-cases/paperclip-agent-output.png#bordered)

   b. 如果输出不在聊天中，请在任务单中发表评论，询问智能体文件路径。打开 Olares Files 应用，然后进入指定目录以获取你的文档。

   ![Agent output in Files app](/images/manual/use-cases/paperclip-agent-output-in-files.png#bordered)

## 从仪表板监控操作

随着智能体完成任务单，你可以使用仪表板跟踪公司的整体性能、监控 API 成本，并识别潜在的瓶颈。

1. 点击左侧边栏中的 **Dashboard**。

   ![Paperclip dashboard](/images/manual/use-cases/paperclip-dashboard.png#bordered)

2. 查看顶部智能体卡片，了解智能体最近处理的任务及其执行时长。
3. 检查高级指标以监控运营状况，例如当月产生的 API 总成本。
4. 分析 14 天趋势图表以发现性能变化：

   - **Run Activity**: 检查智能体执行的总次数。
   - **Issues by Priority**: 按紧急程度（紧急、高、中、低）查看活跃任务单。
   - **Issues by Status**: 通过查看哪些任务单正在进行、已完成或被阻止来跟踪进度。
   - **Success Rate**: 监控 AI 工作力的执行成功率。

5. 滚动到活动日志以审计最近的系统事件和智能体行为。此日志提供所有最近任务的按时间顺序记录。

## 在 Paperclip 中使用本地模型

OpenCode 可以通过 Olares Router 调用兼容 OpenAI API 的本地模型。这里的 OpenCode 运行在 Paperclip 内，与 Market 中单独安装的 OpenCode 应用相互独立。

:::info 适用版本
本节区分 Paperclip `2026.609.0` 与 `2026.916.1`、`2026.1005.0`。后两个版本的新建 OpenCode 智能体页面会绑定 **OpenRouter AI connection**。手动输入 `olares/default-chat` 不会解除绑定，而会触发兼容性提示。**Environment: Local** 仅表示智能体在本机执行，不代表使用本地模型。

下文无绑定智能体的 CLI 流程依据 `2026.1005.0` 的 CLI 和服务端接口编写。本指南尚未完成该路径的本地推理端到端验证，请先完成验证步骤，再分配正式任务。
:::

### 准备模型并备份现有设置

1. 从 Market 安装并启动 Qwen3.8-27B (llama.cpp)，或其他支持 OpenAI 兼容 API 和工具调用的本地模型。对于 `gemma-4-12b-long`，请在 Olares 团队提供新移植的 v3 chart 后手动安装。本流程不会自动迁移原有模型应用。
2. 在 Router 中配置模型并复制连接信息。

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

3. 使用 Router 默认模型时，模型名为 `default-chat`，并须确认该别名实际路由到本地模型。直连模型服务时，改为该接口实际接受的模型 ID。
4. 修改前，在文件管理器中备份 **Data > paperclip > paperclip > .config > opencode**，同时记录智能体当前模型和配置。不要通过删除凭证或应用数据来重置模型配置。

### 配置 OpenCode

Paperclip 容器内的持久化配置路径为 `/paperclip/.config/opencode/opencode.json`。

- 已有 `opencode.json`：合并下列字段，保留其他 provider、插件、MCP 服务及设置。
- 只有 `opencode.jsonc`：先备份，再将副本转换为严格 JSON，去除注释和末尾多余逗号，并保留原设置。转换后将原 JSONC 移出生效的配置目录，避免两个文件的设置冲突。
- 两个文件都不存在：创建 `opencode.json`。

以下示例使用 Router 默认模型。将 `<router-base-url-including-v1>` 替换为包含 `/v1` 的实际 API Base URL，不要使用模型应用的网页地址。

```json
{
  "$schema": "https://opencode.ai/config.json",
  "autoupdate": false,
  "model": "olares/default-chat",
  "small_model": "olares/default-chat",
  "provider": {
    "olares": {
      "name": "Olares local model",
      "npm": "@ai-sdk/openai-compatible",
      "models": {
        "default-chat": {
          "name": "Local model via Router"
        }
      },
      "options": {
        "baseURL": "<router-base-url-including-v1>"
      }
    }
  }
}
```

认证方式以 Router 连接说明为准。如果接口要求密钥，请配置该接口的密钥，不能用 OpenAI 或 OpenRouter 云端密钥代替。提交问题反馈时不要公开凭证文件。

`model` 指定 OpenCode 默认模型；`small_model` 让 OpenCode 辅助请求也使用本地模型。Paperclip 独立的 **cheap model** 设置如已启用，也需改为可用本地模型或关闭。`autoupdate: false` 关闭 OpenCode 内部升级尝试；内置 CLI 应通过 Paperclip Market 镜像更新。

OpenCode 本身支持 JSONC，但 Paperclip 准备临时运行配置时会使用严格 JSON 解析器读取 `opencode.json`。使用有效 JSON 可避免智能体运行时忽略设置。

:::tip 环境变量替代方式
Paperclip `2026.1005.0` 也读取 `PAPERCLIP_OPENCODE_PROVIDERS`（内容是 `provider` 内的 JSON 对象，不是完整配置）和 `PAPERCLIP_OPENCODE_SMALL_MODEL`，可用于注入运行配置，替代编辑文件。但这不会解除 OpenRouter AI connection 绑定。只有已安装 chart 实际暴露了这些变量时，才能通过 Olares 设置配置；本指南不假设设置页面已有这些变量。
:::

### 新安装：创建不绑定 AI Connection 的 OpenCode 智能体

1. 注册并获得实例管理员或组织访问权限，然后创建组织。配置 OpenCode 本身不要求先提供云端模型密钥。
2. 组织创建后，在 Paperclip 域名下直接打开 `/agents/new`，进入完整适配器选择页。组织没有智能体时，先不要返回 Dashboard，否则可能再次打开 onboarding。直接访问路径已在有智能体的组织中检查；如果新组织仍被引导页阻挡，在组织创建成功后使用下方 CLI。
3. 在 `2026.916.1` 和 `2026.1005.0` 中，不要通过绑定 OpenRouter 的表单完成本地模型智能体创建。使用内置 Paperclip CLI，以你自己的管理员身份创建不带 `runtimeConfig.aiConnection` 的智能体。

从启动台打开 **Paperclip CLI**。将地址替换为 Paperclip 应用域名，不要附加 `/api` 或 Dashboard 路径：

```bash
export PAPERCLIP_LOCAL_API='https://your-paperclip-domain'
paperclipai auth login --api-base "$PAPERCLIP_LOCAL_API" --no-browser
```

在浏览器中打开 CLI 输出的授权链接，用你自己的 Paperclip 账户批准。这是授权 CLI 访问 Paperclip，不是登录模型提供商。然后列出组织：

```bash
paperclipai company list --api-base "$PAPERCLIP_LOCAL_API" --json
```

复制目标组织 `id` 字段中的 UUID，不要使用组织短前缀。创建一个智能体：

```bash
export PAPERCLIP_LOCAL_COMPANY='<organization-uuid>'
mkdir -p /paperclip/workspaces/local-model
paperclipai agent create \
  --api-base "$PAPERCLIP_LOCAL_API" \
  --company-id "$PAPERCLIP_LOCAL_COMPANY" \
  --payload-json '{
    "name": "Local model agent",
    "role": "general",
    "adapterType": "opencode_local",
    "adapterConfig": {
      "cwd": "/paperclip/workspaces/local-model",
      "model": "olares/default-chat"
    },
    "runtimeConfig": {
      "heartbeat": {
        "enabled": false,
        "maxConcurrentRuns": 1
      }
    },
    "permissions": {
      "canCreateAgents": false
    }
  }' --json
```

记录返回的智能体 ID。该命令会创建新智能体，重复执行会再次创建。示例关闭定时心跳和招聘权限，以便控制首次验证范围。`maxConcurrentRuns: 1` 仅限制这一个智能体，不限制所有智能体或模型的其他调用方。

检查保存的配置：

```bash
paperclipai agent get '<agent-id>' --api-base "$PAPERCLIP_LOCAL_API" --json
```

确认 `adapterType` 为 `opencode_local`、`adapterConfig.model` 为 `olares/default-chat`，且 `runtimeConfig.aiConnection` 不存在。在 Paperclip 的 **Agents** 中打开该智能体，保留原有/非托管认证方式，不要切换为 OpenRouter 连接。不要让已绑定云端连接的管理智能体代为创建，因为智能体发起的招聘可能继承托管连接。

### 从旧版本升级

升级镜像不意味着所有旧智能体已采用 Connections。修改前逐个检查配置。

| 现有智能体 | 操作方式 |
| :--- | :--- |
| OpenCode 智能体没有 `runtimeConfig.aiConnection` | 备份并合并上述 OpenCode 配置。在智能体配置中将模型设为 `olares/default-chat`，保留其他设置并验证。 |
| OpenCode 智能体已绑定 OpenRouter | 保留原智能体，通过 CLI 新建无绑定智能体，验证后再有计划地转移任务。从更新请求中删除 `aiConnection` 不会解除绑定，当前服务端会保留它。 |
| 仍在使用 `2026.609.0` | 可继续使用原有配置路径，不受新版 Connections 绑定限制。不要假设新版 CLI 命令存在，使用前运行 `paperclipai agent --help` 检查。 |

Paperclip 会通过 `--model` 参数向 OpenCode 传递智能体选择的模型。仅修改 `opencode.json` 中的默认 `model`，不能覆盖智能体仍然选定的云端模型。

共享配置修改完成后，等待活动任务停止，再通过 Market 重启 Paperclip。不要修改 `/app`、手动替换容器镜像或通过修改数据库解除连接绑定。

### 验证与恢复

1. 在 Paperclip CLI 中检查内置版本与模型发现：

   ```bash
   opencode --version
   opencode models olares
   ```

   确认出现 `olares/default-chat`。这只验证模型发现，不代表推理成功。
2. 显式向本地模型发送一个短请求：

   ```bash
   opencode run --model olares/default-chat 'Reply with OK only. Do not use tools.'
   ```

3. 在无绑定智能体的配置页运行环境/连接测试。CLI 成功不代表智能体使用相同凭证、环境和模型。
4. 手动分配一个短任务并检查运行日志。启用定时任务前，串行完成五次短任务，确认响应成功且没有上下文、KV-cache 或超时报错。验证期间保持其他模型调用方空闲；单个智能体的并发限制不是模型全局容量限制。

| 现象 | 排查方向 |
| :--- | :--- |
| “This connection does not support the current harness and model” | 检查 `runtimeConfig.aiConnection`。修改 OpenCode 文件不能解决 OpenRouter 绑定冲突。 |
| “Method not allowed” | 核对实际请求 URL、`/v1` 路径和 OpenAI 兼容协议。模型列表不能验证聊天接口。Gemma 还需确认请求指向新安装的 v3 服务。 |
| CLI 正常，但智能体报 “Internal server error” | 对比智能体模型、工作目录、连接绑定和环境，收集对应的服务端错误并隐去密钥。仅凭 500 提示无法判断根因。 |
| OpenCode 升级失败 | 使用 Market 提供的镜像更新，不要开放镜像目录写权限来修复自升级。 |
| 原 JSONC 被覆盖 | 有备份则恢复，否则需重新构建所需配置。重启或组织导出都不能恢复原文件的精确内容。 |

撤销修改时，暂停新智能体或停止向其分配任务，恢复备份的 OpenCode 配置和原智能体模型，待活动任务停止后通过 Market 重启。验证成功前保留原智能体。组织导出不是数据库、本地配置、凭证和工作区文件的完整备份；重装或降级前应另行备份这些数据。不要假设新版本迁移过的数据库兼容旧镜像。

## 常见问题

### Paperclip 支持哪些智能体适配器？

Paperclip 目前支持以下智能体适配器。你根据选择的适配器将特定的 API 密钥配置为环境变量：

- Claude Code: 需要 `ANTHROPIC_API_KEY`。
- Codex: 需要 `OPENAI_API_KEY`。
- OpenCode：认证取决于模型提供商，本地模型不要求必须提供 Anthropic 或 OpenAI 密钥。新版 OpenRouter 绑定及本地模型配置方式参见[在 Paperclip 中使用本地模型](#在-paperclip-中使用本地模型)。
- Pi: 需要 `ANTHROPIC_API_KEY` 或 `OPENAI_API_KEY`。
- Cursor: 需要 `CURSOR_API_KEY`。

### 我可以将 Hermes Agent 作为智能体运行时使用吗？

可以，但不建议用于生产环境。请注意以下事项：

1. 此 Hermes Agent 运行在 Paperclip 容器内部，与从 Olares Market 安装的 Hermes Agent 应用是分开的。你需要在 Paperclip CLI (`pip install hermes-agent`) 中单独安装和配置它。
2. Hermes Agent 和 Paperclip 适配器仍在开发中。连接稳定性问题可能需要大量调试。
3. 确保以下配置正确，否则智能体可能无法工作：

   - 关闭手动审批设置，以防止 Paperclip 在调用智能体时等待审批而超时：

   ```bash
   hermes config set approvals.mode "off"
   hermes config set approvals.cron_mode "approve"
   ```

   - 确保你的 Hermes Agent 已设置 `PAPERCLIP_API_KEY`，并安装了 Paperclip Skills，以便它可以获取和修改 Paperclip 中的任务。

### Codex 适配器认证失败

更新 `OPENAI_API_KEY` 并重启 Paperclip 后，Codex 适配器仍可能认证失败。这是因为 `codex login` 命令尝试打开本地浏览器进行 OAuth，而后台容器无法执行此操作。

要解决此问题，请运行手动设备认证登录：

1. 打开 Control Hub，然后进入 **Browse** > **paperclip-{username}** > **Deployments** > **paperclip**。
2. 在 **Pods** 下，点击 pod 名称查看其容器，然后点击 **paperclip** 容器旁边的 <i class="material-symbols-outlined">terminal</i> 打开 pod 终端。

   ![Open the Paperclip pod terminal from Control Hub](/images/manual/use-cases/paperclip-enter-container.png#bordered)

3. 输入以下命令，然后按 **Enter**：

   ```bash
   codex login --device-auth
   ```

4. 在浏览器中打开设备认证链接并登录以完成授权。然后在 Paperclip 中重试 Codex 适配器。

   ![Codex device-auth result](/images/manual/use-cases/paperclip-codex-login-result.png#bordered)

## 了解更多

- [Paperclip documentation](https://docs.paperclip.ing): 官方文档，涵盖概念、功能和配置。
- [Orchestrate multi-agent workflows with oh-my-openagent](opencode-omo.md): 在 Olares 上的单个 OpenCode 实例中运行多智能体协作。
- [Set up OpenCode as your AI coding agent](opencode.md): 安装和配置 OpenCode，一种常用的 Paperclip 适配器。
