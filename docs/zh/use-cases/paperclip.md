---
outline: [2, 3]
description: 在 Olares 上运行 Paperclip，协调多个 AI 智能体协同完成同一组任务。添加由 Claude Code、Codex、OpenCode、Cursor 或其他提供商支持的智能体，并为它们分配任务单。
head:
  - - meta
    - name: keywords
      content: Olares, Paperclip, AI agent, multi-agent, Claude Code, Codex, OpenCode, Cursor, self-hosted
app_version: "1.0.22"
doc_version: "1.1"
doc_updated: "2026-06-12"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/paperclip.md)为准。
:::

# 使用 Paperclip 协调多个 AI 智能体

Paperclip 是一个开源平台，用于在同一个统一工作区下协调多个 AI 智能体。通过设置虚拟公司，你可以添加由 Claude Code、Codex、OpenCode、Cursor 或其他提供商驱动的 AI 智能体，并为它们分配任务单。无论是编码、研究还是内容创作，Paperclip 都能管理工作流。

将 Paperclip 作为自托管应用运行在 Olares 上，可以确保你的 API 密钥、任务历史和智能体输出完全保留在你的设备上。

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

每个智能体都需要其底层模型提供商的 API 密钥。你可以在 Settings 应用中将密钥配置为环境变量。

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

4. 命名你的公司，并按照指示完成引导流程。

## 创建你的第一个公司

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

:::warning 谨慎使用本地模型
Paperclip 是一个完全自主的多智能体协作平台。完全在本地模型上运行可能会因模型能力限制、上下文溢出或并发限制而导致工作流中断，从而可能引发级联故障。

考虑使用混合配置：将高性能云端模型分配给 CEO 或 CTO 等关键角色，同时为执行型智能体使用本地模型以降低成本。或者，将本地模型指定为 `cheap model` 用于要求较低的任务。

仔细评估每个模型的能力和你的工作流需求，以确定最佳设置。
:::

### 使用 OpenCode 调用本地模型

:::info 独立的 OpenCode 实例
此 OpenCode 运行在 Paperclip 容器内部，与从 Olares Market 安装的 OpenCode 应用是分开的。你需要独立配置它。
:::

1. 从 Market 安装一个模型应用。本示例使用 `gemma4:26b`。
2. 创建智能体时，选择 **OpenCode** 作为运行时。选择一个临时模型，如 **big-pickle**，以完成初始设置。你将在下一步将其替换为本地模型配置。
3. 初始化后，配置本地模型：

   a. 打开 Files 应用，进入 **Data** > **paperclip** > **paperclip** > **.config** > **opencode**。

   b. 将 `opencode.jsonc` 重命名为 `opencode.json`，并以编辑模式打开。

   c. 将默认配置替换为以下示例（使用 gemma4:26b 作为模型）：

   - 将配置中的 `<your-olares-id>` 替换为你的具体 Olares ID。
   - 如果你使用的不是 Gemma4 26B Q4_K_M (Ollama)，请将 `2b46296c` 替换为你的实际应用路由 ID。
   - 确保 `baseURL` 以 `/v1` 结尾，以通过 Ollama 提供 OpenAI 兼容的 API 访问。

   ```json {wrap}
   {
     "$schema": "https://opencode.ai/config.json",
     "model": "olares/gemma4:26b",
     "provider": {
       "olares": {
         "name": "Gemma4:26b (Ollama)",
         "npm": "@ai-sdk/openai-compatible",
         "models": {
           "gemma4:26b": {
             "name": "gemma4:26b"
           }
         },
         "options": {
           "baseURL": "https://2b46296c.<your-olares-id>.olares.com/v1"
         }
       }
     }
   }
   ```

4. 重启 Paperclip 容器。
5. 在 Paperclip 中，进入 **Agents** > **Configuration** > **Permissions & Configuration** 以验证新添加的本地模型。

   ![OpenCode local model in agent config](/images/manual/use-cases/paperclip-opencode-model-config.png#bordered)

   :::warning
   如果你不打算使用默认的 `openai/gpt-5.1-codex-mini` 作为 cheap model，请务必关闭此功能或切换到其他可用模型。
   :::

### 测试工作流

你可以在任务的 **Activity** > **Continuation Summary** 中监控执行过程和结果。

- 要求 CEO 雇佣一个新智能体：
   - **Task title:** Hire a CMO
   - **Task description:** Hire a content generation agent that uses opencode as the runtime and olares/gemma4:26b as the model.

   ![Agent run activity](/images/manual/use-cases/paperclip-agent-run-activity.png#bordered)

- 要求 CMO 撰写品牌故事：
   - **Task title:** Write a brand story
   - **Task description:** Output in md format and upload the final result as an attachment to the task.

   ![Brand story output](/images/manual/use-cases/paperclip-brand-story-output.png#bordered)

## 常见问题

### Paperclip 支持哪些智能体适配器？

Paperclip 目前支持以下智能体适配器。你根据选择的适配器将特定的 API 密钥配置为环境变量：

- Claude Code: 需要 `ANTHROPIC_API_KEY`。
- Codex: 需要 `OPENAI_API_KEY`。
- OpenCode: 需要 `ANTHROPIC_API_KEY` 或 `OPENAI_API_KEY`。
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
