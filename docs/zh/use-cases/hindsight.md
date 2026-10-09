---
outline: [2, 3]
description: 在 Olares 上使用 Hindsight 为 Hermes Agent 添加长期记忆。配置模型、连接记忆库，并在不同对话中保存和检索信息。
head:
  - - meta
    - name: keywords
      content: Olares, Hindsight, Hermes Agent, 长期记忆, 自托管, 记忆库, 保存记忆, 检索记忆
app_version: "0.0.17"
doc_version: "1.0"
doc_updated: "2026-10-09"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/hindsight.md)为准。
:::

# 使用 Hindsight 为智能体添加长期记忆

Hindsight 是开源的智能体记忆服务，让 AI 在不同对话中记住并使用之前的信息。记忆保存在你的 Olares 设备上，Hindsight 通过 Router 调用模型来整理和检索记忆。

本文以 Hermes Agent 为例，介绍如何保存信息、在新对话中检索记忆，以及控制何时保存记忆。

## 前提条件

- 一台运行 Olares 1.12.7 或更高版本、具有足够磁盘空间和内存的 Olares 设备。
- 能够访问 [Router](olares-router.md)，用于连接本地模型。
- 已安装并配置 [Hermes Agent](hermes.md)，能够正常对话。
- 以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.8-27B (llama.cpp) | 从市场安装 |
  | 嵌入 | EmbeddingGemma | 从市场安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 检查模型是否可用

Hindsight 需要聊天模型来整理记忆，嵌入模型来检索记忆。安装前，这两个模型都应能通过 Router 正常调用。

:::warning 安装前配置嵌入模型
如果未配置默认嵌入模型，或模型无法调用，Hindsight 将无法启动。请先完成以下检查，再安装 Hindsight。
:::

1. 从启动台打开 Router，进入 **Default models** 页面。
2. 将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型，将 EmbeddingGemma 设为默认嵌入模型。如果已选中这两个模型，确认选择无误，并检查两者是否都显示 **Callable**。

   ![Router 默认模型页面中的 Qwen3.8-27B 和 EmbeddingGemma 均显示 Callable](/images/manual/use-cases/hindsight-router-default-models.png#bordered)

## 安装 Hindsight

1. 打开市场，搜索“Hindsight”。

2. 点击**获取**，然后点击**安装**，等待安装完成。

## 获取 Hindsight API 地址

Hermes 通过 Hindsight API 保存和检索记忆。这个 API 有独立入口，与网页控制台、Router 的模型 API 使用不同地址。

Hindsight 的默认设置可直接连接 Router，无需填写模型地址或 Router API 密钥。如需使用其他聊天或嵌入模型，参见[更换 Hindsight 使用的模型](#更换-hindsight-使用的模型)。

1. 打开**设置** > **应用** > **Hindsight**，在**入口**下选择 **Hindsight API**，复制端点地址。

   地址格式类似以下示例：

   ```text
   https://2b85162e1.laresprime.olares.com
   ```

2. 从启动台打开 Hermes CLI，在 Hermes 中检查连接。将下面的示例地址替换为刚复制的端点地址，再运行：

   ```bash
   export HINDSIGHT_API_URL="https://your-hindsight-api-host"
   curl -fsS "${HINDSIGHT_API_URL%/}/health"
   ```

   输出示例：

   ```json
   {"status":"healthy","database":"connected","db_acquire_ms":1.4,"db_pool_waiting":0,"db_pool_in_use":0,"db_pool_max":100,"db_pool_idle":1}
   ```

   `"status":"healthy"` 和 `"database":"connected"` 表示 Hermes 能够访问 API，且 Hindsight 已连接数据库。如果返回登录页面或重定向，参见[为什么 API 健康检查会显示登录提示？](#为什么-api-健康检查会显示登录提示)。

## 连接 Hermes Agent

Hindsight 插件让 Hermes 能够保存对话中的信息，并在后续对话中使用。

每个记忆库（bank）都有自己的 ID，Hermes 通过它确定在哪里保存和检索记忆。插件默认使用名为 `hermes` 的记忆库。可以先使用这个库，需要区分项目或在智能体之间共享记忆时，再[创建或选择其他记忆库](#管理记忆库)。

1. 在 Hermes CLI 中安装 Hindsight 插件：

   ```bash
   hermes plugins install hindsight
   ```

2. 出现是否启用插件的提示时，输入 `y` 并按 **Enter** 确认。
3. 启动记忆配置向导：

   ```bash
   hermes memory setup
   ```

4. 使用方向键浏览向导，按**空格键**或 **Enter** 选择选项。按下表完成配置：

   | 设置项 | 操作 |
   | :--- | :--- |
   | Memory provider setup | 选择 **hindsight**。 |
   | Select mode | 选择 **Local External**。 |
   | Hindsight API URL | 粘贴 [Hindsight API 地址](#获取-hindsight-api-地址)。 |
   | API key | 按 **Enter** 跳过。 |
   | Starter memory template | 选择适合的模板，或选择 **Blank**，从空白记忆库开始。 |

5. 从启动台打开 Hermes Dashboard，进入 **Plugins** 页面。在 **Runtime Provider Plugins** 下，确认 **Memory Provider** 为 `hindsight`，状态为 ready 或 active。

   ![Hermes Plugins 页面中已启用的 Hindsight 记忆服务](/images/manual/use-cases/hindsight-hermes-provider-active.png#bordered)

## 确认 Hermes 已保存记忆

1. 在 Hermes CLI 或 Hermes Dashboard 中开始新对话，告诉它一条信息。例如：

   ```text
   我在 Olares 上开发和部署应用。
   ```

   Hermes 会自动通过 Hindsight 保存这条信息。

2. 从启动台打开 Hindsight，选择 `hermes` 记忆库，等待这条信息出现。这表示 Hermes 已通过 Hindsight 保存了信息。

   ![hermes 记忆库中由 Hermes 自动保存的信息](/images/manual/use-cases/hindsight-hermes-auto-saved-memory.png#bordered)

## 控制 Hermes 如何保存和检索记忆

### 关闭自动记忆

Hermes 可以在每轮对话前自动检索记忆，并保存新的对话内容。

若希望只在主动要求时读写记忆，关闭自动检索和自动保存：

1. 在 Hermes Dashboard 的 **Plugins** 页面，确认选中的是刚才配置的 profile（配置档案）。
2. 在 Hindsight 设置中，关闭 **Auto Recall** 和 **Auto Retain**。保持 **Mode**、**API URL** 和 **Bank ID** 不变。
3. 点击 **Save**，刷新页面，确认两个开关仍处于关闭状态。

:::details 找不到开关时，通过配置文件修改
这些设置属于 Hermes Agent 中的 Hindsight 插件。

1. 打开文件管理器，进入 **Data** > **hermesagent** > **home** > **hindsight**。

2. 打开 `config.json`，添加或修改以下字段。保留其他字段，包括连接和记忆库设置：

   ```json
   {
     "auto_retain": false,
     "auto_recall": false
   }
   ```

3. 保存文件。

4. 开始新的 Hermes 对话，让修改生效。
:::

### 手动保存和检索记忆

在 Hermes CLI 或 Hermes Dashboard 中，可以让 Hermes 保存信息、检索信息，或结合已有记忆分析问题。

1. 要保存一条信息或偏好，让 Hermes 使用 `hindsight_retain`：

   ```text
   使用 hindsight_retain 记住：我在 Olares 上部署应用，并且偏好自托管服务。
   ```

2. 要在后续对话中检索信息，让 Hermes 使用 `hindsight_recall`：

   ```text
   使用 hindsight_recall 查询我在哪个平台部署应用。
   ```

3. 要根据已保存的偏好获取建议，让 Hermes 使用 `hindsight_reflect`：

   ```text
   使用 hindsight_reflect，根据已保存的偏好，建议我在选择部署平台时应优先考虑哪些因素。
   ```

4. 打开 Hindsight，选择 Hermes 使用的记忆库，默认为 `hermes`。确认部署偏好已出现在记忆库中。

要让 Hermes 自动管理记忆，重新开启 **Auto Recall** 和 **Auto Retain**。

## 管理记忆库

不同项目或用户可以使用独立的记忆库。将智能体连接到同一个库，可以共享其中的记忆。

### 创建记忆库

1. 从启动台打开 Hindsight。
2. 为 Hermes Agent 创建名为 `hermes-personal` 的记忆库。

   ![在 Hindsight 中创建 hermes-personal 记忆库](/images/manual/use-cases/hindsight-create-bank.png#bordered)

3. 复制记忆库 ID `hermes-personal`，用于[配置 Hermes](#选择-hermes-使用的记忆库)。

### 选择 Hermes 使用的记忆库

1. 打开 Hermes Dashboard，进入 **Plugins** 页面，在 **Runtime Provider Plugins** 下找到 Hindsight 设置。
2. 将 **Bank ID** 设为 `hermes-personal`，保持 **Bank ID Template** 为空，以使用这个固定的记忆库。
3. 点击 **Save**，刷新页面，确认记忆库 ID 已保存。
4. 开始新的 Hermes 对话，[保存一条信息](#手动保存和检索记忆)，确认它出现在 `hermes-personal` 中。

:::details 通过配置文件更换记忆库
也可以在 Hindsight 插件的配置文件中更换记忆库。

1. 打开文件管理器，进入 **Data** > **hermesagent** > **home** > **hindsight**。

2. 打开 `config.json`，修改以下字段。将 `hermes-personal` 替换为你的记忆库 ID，保留其他字段，包括连接设置：

   ```json
   {
     "bank_id": "hermes-personal",
     "bank_id_template": ""
   }
   ```

3. 保存文件。

4. 开始新的 Hermes 对话，让修改生效。[保存一条信息](#手动保存和检索记忆)，确认它出现在所选记忆库中。
:::

## 更换 Hindsight 使用的模型

要更改所有使用 `default-chat` 或 `default-embedding` 的应用所调用的模型，在 Router 的 **Default models** 页面选择其他聊天或嵌入模型。Hindsight 会使用新的默认模型，无需额外修改。

要仅为 Hindsight 指定模型，从 Router 的 **How to call this model** 窗口获取完整模型名称。聊天模型位于 **LLM** 页面，嵌入模型位于 **Tools** 页面。参见[从 Router 获取连接信息](olares-router.md#从-router-获取连接信息)。

也可以将 Hindsight 连接到其他兼容 OpenAI API 的服务。聊天和嵌入请求共用同一个基础 URL，因此该端点需同时支持这两种请求。

1. 打开**设置** > **应用** > **Hindsight**。
2. 点击**管理环境变量**，按需修改以下值：

   | 环境变量 | 默认值 | 填写内容 |
   | --- | --- | --- |
   | `HINDSIGHT_API_LLM_BASE_URL` | 留空。应用会生成你的 Router 地址。 | 使用 Router 时保持为空，或填写其他兼容 OpenAI API 的基础 URL。 |
   | `HINDSIGHT_API_LLM_MODEL` | `default-chat` | Router 或模型服务中的完整聊天模型名称。 |
   | `HINDSIGHT_API_EMBEDDINGS_MODEL` | `default-embedding` | Router 或模型服务中的完整嵌入模型名称。 |
   | `HINDSIGHT_API_LLM_API_KEY` | 留空。应用会填入 `not-required`。 | 使用 Router 时保持为空，或填写模型服务要求的 API 密钥。 |

3. 点击**确认**，再点击**应用**。

要重新使用 Router 的默认模型，将以上变量恢复为**默认值**列中的值。

## 连接 OpenClaw 或 OpenCode

也可以将 Olares 上的 [OpenClaw](openclaw.md) 或 [OpenCode](opencode.md) 连接到同一个 Hindsight 服务。连接时使用 [Hindsight API 地址](#获取-hindsight-api-地址)，安装前检查插件的兼容要求。

如果 Hindsight 要求 API token，该 token 用于验证智能体的身份，与 Router API 密钥和 Olares 登录密码不同。

### 连接 OpenClaw

该插件会替换 OpenClaw 当前使用的记忆插件。自动记忆设置与 Hermes 相互独立。操作前，查看 [OpenClaw 插件兼容要求](https://github.com/vectorize-io/hindsight/blob/v0.10.2/hindsight-integrations/openclaw/README.md#openclaw-compatibility)。

1. 从启动台打开 OpenClaw CLI，安装插件：

   ```bash
   openclaw plugins install @vectorize-io/hindsight-openclaw
   ```

2. 运行配置向导：

   ```bash
   npx --package @vectorize-io/hindsight-openclaw hindsight-openclaw-setup
   ```

3. 选择 **External API**，粘贴 [Hindsight API 地址](#获取-hindsight-api-地址)。使用 Olares 默认配置时，token 留空。
4. 打开**设置** > **应用** > **OpenClaw**，点击**停止**，再点击**恢复**，让插件加载生效。
5. 在 OpenClaw 中开始对话，告诉它一条信息。打开 Hindsight，确认这条信息出现在为当前对话场景创建的记忆库中。

记忆库选择和自动记忆设置，参见 [OpenClaw 的 Hindsight 插件指南](https://hindsight.vectorize.io/sdks/integrations/openclaw#configuration)。

### 连接 OpenCode

以下步骤为一个项目配置 `@vectorize-io/opencode-hindsight` 插件。该插件提供记忆工具，也能自动保存和检索对话。

1. 在文件管理器中打开项目文件夹，找到 `opencode.json`。使用 OpenCode 默认工作区时，路径为 `Home/Code/<project>/opencode.json`。如果文件不存在，在项目文件夹中创建它。
2. 在 `plugin` 数组中添加以下配置，保留已有插件和设置。将示例地址替换为你的 [Hindsight API 地址](#获取-hindsight-api-地址)：

   ```json
   {
     "plugin": [
       [
         "@vectorize-io/opencode-hindsight",
         {
           "hindsightApiUrl": "https://your-hindsight-api-host",
           "bankId": "opencode"
         }
       ]
     ]
   }
   ```

   使用 `opencode` 作为独立记忆库，或填写已有记忆库 ID，以共享其中的记忆。

3. 保存文件。从启动台打开 OpenCode Terminal，切换到项目目录，运行 `opencode`。启动时会自动安装插件。
4. 让 OpenCode 使用 `hindsight_retain` 保存一条信息，再在新会话中使用 `hindsight_recall` 检索。在所选 Hindsight 记忆库中查看结果。

更多配置，以及后续替代方案 Coding Agents 插件，参见 [OpenCode 的 Hindsight 插件指南](https://hindsight.vectorize.io/sdks/integrations/opencode)。

## 常见问题

### 为什么 Hindsight 提示 LLM API 密钥必填？

Hindsight 收到的模型 API 密钥为空。

1. 打开**设置** > **应用** > **Hindsight** > **管理环境变量**。
2. 使用 Router 时，将 `HINDSIGHT_API_LLM_API_KEY` 设为 `not-required`。使用需要认证的模型服务时，填写有效密钥。
3. 点击**确认**，再点击**应用**。

### 为什么 API 健康检查会显示登录提示？

检查 Hindsight API 入口的[认证级别](../manual/olares/settings/manage-entrance.md#设置认证级别和模式)，确认设为**内部**。按[获取 Hindsight API 地址](#获取-hindsight-api-地址)中的步骤，在 Hermes CLI 中运行检查。从电脑访问入口时，开启 LarePass VPN。

### 为什么 Hindsight 启动时报 `upstream_circuit_open`？

模型服务连续报错后，Router 会暂时停止转发请求。检查 `HINDSIGHT_API_EMBEDDINGS_MODEL` 指定的模型。发送一段短文本，并在[控制面板](../manual/olares/controlhub/manage-container.md)中查看模型日志。

如果日志包含 `infer failed on device GPU.0: general error`，先修复嵌入模型的推理问题，再重试 Hindsight。增加 Hindsight 的资源配额无法解决独立模型服务中的错误。

### 为什么 Hermes 检索不到刚保存的信息？

先查看信息是否已出现在记忆库中，再检查两段对话是否[使用同一个记忆库 ID](#选择-hermes-使用的记忆库)。Hindsight 插件可能只检索 `observation`（归纳后的记忆）。新信息完成归纳前，可能暂时查不到。

要同时检索原始事实，在插件配置中将 `recall_types` 设为 `"observation,world,experience"`，然后开始新会话。此设置对自动检索和手动检索均生效。

### 为什么记忆处理较慢？

Hindsight 需要提取事实、计算嵌入向量，并在后台整理记忆。这些任务可能在 Hermes 回复后继续运行。再次提交相同内容前，先查看记忆库和工具返回结果。

记忆处理和聊天请求会共用模型资源。可以先保存一条简短信息，查看 Hindsight 和模型日志。[关闭自动记忆](#关闭自动记忆)后，只在主动要求时保存信息。增加应用的 CPU 或内存配额前，先检查资源使用情况。

## 了解更多

- [Hermes Agent](hermes.md)：配置智能体及其模型连接。
- [Olares Router](olares-router.md)：配置模型路由，了解认证方式。
- [Hindsight 文档](https://hindsight.vectorize.io/)：了解记忆库、集成方式和 API 功能。
