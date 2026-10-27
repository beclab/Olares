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

## 安装 Paperclip

1. 打开 Market 并搜索 "Paperclip"。

   ![Paperclip in Market](/images/manual/use-cases/paperclip.png#bordered)

2. 点击 **Get**，然后点击 **Install**。安装完成后，Launchpad 中会出现两个快捷方式：

   ![Paperclip entry points](/images/manual/use-cases/paperclip-install-entry.png#bordered){width=35%}

## 初始化 Paperclip

安装 Paperclip 后，创建第一个管理员账户以完成初始设置。如果你计划使用云端模型，请先配置 API 密钥。

### 配置 API 密钥

如果你计划使用 OpenAI、Anthropic（Claude）、Google Gemini 或 Cursor，可以先在 Settings 应用中通过环境变量配置对应的 API 密钥。

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

4. 输入公司名称，并按提示完成初始引导。

## 完成初始引导

首次创建组织时，Paperclip 会打开 onboarding，引导你设置组织和首个智能体。

1. 输入组织名称，并按页面提示填写组织目标。
2. 为首个智能体命名。
3. 选择模型提供商。目前默认 onboarding **仅支持 OpenAI 和 Claude**。按页面提示连接相应账户或提供 API 密钥，选择模型并验证连接。
4. 按引导完成智能体设置，进入组织工作区。

### 使用其他适配器创建首个智能体

如果要使用 **OpenCode、Hermes** 等其他适配器，先完成组织创建，然后在浏览器地址栏手动输入：

```text
https://<你的-Paperclip-域名>/agents/all
```

在智能体列表页点击创建智能体的按钮，为智能体命名并选择适配器，再完成对应的运行配置。

OpenCode 智能体创建页面目前仅支持 OpenRouter 连接。要使用本地模型，请按照[在 Paperclip 中使用本地模型](#在-paperclip-中使用本地模型)中的步骤，通过 CLI 创建未绑定 AI connection 的智能体。

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

:::warning 谨慎使用本地模型
Paperclip 是一个完全自主的多智能体协作平台。完全在本地模型上运行可能会因模型能力限制、上下文溢出或并发限制而导致工作流中断，从而可能引发级联故障。

考虑使用混合配置：将高性能云端模型分配给 CEO 或 CTO 等关键角色，同时为执行型智能体使用本地模型以降低成本。或者，将本地模型指定为 `cheap model` 用于要求较低的任务。

仔细评估每个模型的能力和你的工作流需求，以确定最佳设置。
:::

:::warning 仅修改 `opencode.json` 无法配置已绑定 OpenRouter 的智能体
最新版 Paperclip 的智能体创建页面默认将 **OpenCode 接入 OpenRouter**，而不是直接使用现有 OpenCode 配置文件中的本地模型配置。运行已绑定 OpenRouter AI connection 的智能体时，Paperclip 会使用独立的配置目录、注入 OpenRouter 凭证，并显式传入智能体选择的模型，覆盖配置文件中的默认模型。**仅修改 `opencode.json` 或 `opencode.jsonc` 已不足以配置这类智能体。**
:::

OpenCode 可以通过 Olares Router 调用兼容 OpenAI API 的本地模型。这里的 OpenCode 运行在 Paperclip 内，与 Market 中单独安装的 OpenCode 应用相互独立。

### 准备本地模型

1. 从 Market 安装并启动 Qwen3.8-27B (llama.cpp)。
2. 在 Router 中配置模型并复制连接信息。

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

3. 使用 Router 默认模型时，模型名为 `default-chat`，并须确认该别名实际路由到本地模型。直连模型服务时，改为该接口实际接受的模型 ID。

### 配置 OpenCode

Paperclip 容器内的持久化配置路径为 `/paperclip/.config/opencode/opencode.json`。

- 已有 `opencode.json`：合并下列字段，保留其他 provider、插件、MCP 服务及设置。
- 只有 `opencode.jsonc`：先备份，再将副本转换为严格 JSON，去除注释和末尾多余逗号，并保留原设置。转换后将原 JSONC 移出生效的配置目录，避免两个文件的设置冲突。
- 两个文件都不存在：创建 `opencode.json`。

以下示例使用 Router 默认模型。将示例中的 `baseURL` 替换为从 Router 复制的 API Base URL，保留 `/v1`。不要使用模型应用的网页地址。

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
        "baseURL": "https://router.your_olares_id.olares.com/v1"
      }
    }
  }
}
```

认证方式以 Router 连接说明为准。如果接口要求密钥，请配置该接口的密钥，不能用 OpenAI 或 OpenRouter 云端密钥代替。提交问题反馈时不要公开凭证文件。

### 创建不绑定 AI Connection 的 OpenCode 智能体

1. 注册并获得实例管理员或组织访问权限，然后创建组织。配置 OpenCode 本身不要求先提供云端模型密钥。
2. 使用你自己的管理员账户，通过内置 Paperclip CLI 创建不带 `runtimeConfig.aiConnection` 的智能体。
3. 从启动台打开 **Paperclip CLI**。在下方命令中，将 `https://your-paperclip-domain` 替换为 Paperclip 应用地址，不要附加 `/api` 或 Dashboard 路径：

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

确认 `adapterType` 为 `opencode_local`、`adapterConfig.model` 为 `olares/default-chat`，且 `runtimeConfig.aiConnection` 不存在。在 Paperclip 的 **Agents** 中打开该智能体。

### 已有智能体的配置

对于升级后保留的智能体，先检查是否不存在 `runtimeConfig.aiConnection`。如果没有该字段，合并上述配置，将智能体模型设为 `olares/default-chat`，并保留其他设置。如果智能体已绑定 OpenRouter，请改用上面的 CLI 流程创建无绑定智能体。

在 **Paperclip CLI** 中，先按照上文使用 `paperclipai auth login` 登录，并合并 OpenCode 配置，再更新智能体。等待活动任务结束，将下方 URL 和 `<agent-id>` 替换为 Paperclip 应用地址及现有智能体的 ID。内置镜像已包含 `jq`。

```bash
(
  set -eu
  umask 077
  PAPERCLIP_LOCAL_API='https://your-paperclip-domain'
  PAPERCLIP_LOCAL_AGENT='<agent-id>'
  PAPERCLIP_CONFIG_BACKUP=$(mktemp -d /paperclip/agent-config-backup.XXXXXX)

  paperclipai agent get "$PAPERCLIP_LOCAL_AGENT" \
    --api-base "$PAPERCLIP_LOCAL_API" --json \
    > "$PAPERCLIP_CONFIG_BACKUP/agent-before.json"

  if ! jq -e '
    .adapterType == "opencode_local" and
    .runtimeConfig.aiConnection == null
  ' "$PAPERCLIP_CONFIG_BACKUP/agent-before.json" > /dev/null; then
    echo 'Stop: this is not an unbound OpenCode agent.' >&2
    exit 1
  fi

  jq '{
    adapterConfig: ((.adapterConfig // {}) + {model: "olares/default-chat"})
  }' "$PAPERCLIP_CONFIG_BACKUP/agent-before.json" \
    > "$PAPERCLIP_CONFIG_BACKUP/update.json"

  paperclipai agent update "$PAPERCLIP_LOCAL_AGENT" \
    --api-base "$PAPERCLIP_LOCAL_API" \
    --payload-json "$(cat "$PAPERCLIP_CONFIG_BACKUP/update.json")" --json \
    > "$PAPERCLIP_CONFIG_BACKUP/agent-after.json"

  jq -e '
    .adapterType == "opencode_local" and
    .adapterConfig.model == "olares/default-chat" and
    .runtimeConfig.aiConnection == null
  ' "$PAPERCLIP_CONFIG_BACKUP/agent-after.json"

  printf 'Configuration backup: %s\n' "$PAPERCLIP_CONFIG_BACKUP"
)
```

该命令会先备份智能体配置；如果目标不是无绑定的 OpenCode 智能体，则停止操作。更新时仅修改 `adapterConfig.model`，保留其他适配器设置，不修改 `runtimeConfig`。命令执行期间不要在其他页面修改该智能体配置。最后输出 `true` 表示模型设置已保存且没有 AI connection 绑定，不代表已验证模型推理。

配置可能包含凭证，请妥善保管备份目录。随后在 **Agents** 中打开该智能体，运行连接测试，再分配任务。

共享配置修改完成后，等待活动任务停止，再通过 Market 重启 Paperclip。不要修改 `/app`、手动替换容器镜像或通过修改数据库解除连接绑定。

## 常见问题

### Paperclip 支持哪些智能体适配器？

Paperclip 目前支持以下智能体适配器。你根据选择的适配器将特定的 API 密钥配置为环境变量：

- Claude Code: 需要 `ANTHROPIC_API_KEY`。
- Codex: 需要 `OPENAI_API_KEY`。
- OpenCode：认证取决于模型提供商，本地模型可能不需要 API 密钥。使用 OpenRouter 时，请在 Paperclip 界面中输入 OpenRouter API 密钥。
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
