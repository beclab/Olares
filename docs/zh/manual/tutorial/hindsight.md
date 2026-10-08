---
outline: [2, 3]
description: 在 Olares 上配置 Hindsight，连接智能体，保存、检索记忆并分析历史信息。
---

# 在 Olares 上使用 Hindsight

本指南适用于已在 Olares 上安装 Hindsight、准备连接智能体的用户。文档基于 Hindsight 0.10.2 和 Olares Chart 0.0.16 的配置。

Hindsight 为智能体保存跨对话的长期记忆。记忆库（bank）用一个 ID 组织记忆。智能体可以保存经历（retain）、检索事实（recall），也可以基于记忆分析问题（reflect）。

:::warning 连接前先检查模型
请先在 Olares Router 中配置可正常调用的聊天模型和嵌入模型。Hindsight 启动时会检查模型连接，嵌入服务报错可能导致启动失败。
:::

## 理解服务连接

Hindsight 涉及两类地址：

| 连接 | 用途 | 地址格式 |
| --- | --- | --- |
| 智能体 → Hindsight | 保存和查询记忆 | 本次安装的 Hindsight API 入口地址 |
| Hindsight → Router | 调用聊天模型和嵌入模型 | `https://router.<你的 Olares 域名>/v1` |

配置智能体时，不要把 Router 地址填成 Hindsight API 地址。Hindsight API 入口地址末尾也不需要添加 `/v1`。

网页控制台用于查看记忆库和数据，API 用于连接智能体。记忆数据存放在 Olares 系统 PostgreSQL 的专用数据库中，并启用 `vector` 扩展。本应用不额外部署 PostgreSQL，也不附带模型权重。

## 准备并安装

1. 打开 Router，检查 `default-chat` 是否指向聊天模型，`default-embedding` 是否指向嵌入模型。
2. 分别用简短输入测试两个模型。模型应用显示 Running，不代表推理一定正常。
3. 在市场中从你可用的来源安装 Hindsight。应用仍在测试阶段时，使用 Local Sources → Upload 中已上传的 Chart。
4. 按下表填写安装参数。使用 Router 默认路由时，可以保留默认设置。

| 安装参数 | 默认行为 | 何时修改 |
| --- | --- | --- |
| `HINDSIGHT_API_LLM_BASE_URL` | 留空后自动生成 `https://router.<你的 Olares 域名>/v1` | 使用同时提供聊天和嵌入模型的其他 OpenAI 兼容端点 |
| `HINDSIGHT_API_LLM_MODEL` | `default-chat` | 指定聊天模型或路由 |
| `HINDSIGHT_API_EMBEDDINGS_MODEL` | `default-embedding` | 指定嵌入模型或路由 |
| `HINDSIGHT_API_LLM_API_KEY` | 留空后自动使用非空占位值 `not-required` | Router 或模型服务开启密钥校验时，填写有效密钥 |

占位值只用于通过 Hindsight 的非空检查，不能通过服务端的真实密钥认证。聊天和嵌入请求共用这里设置的密钥和地址。

5. 等待 Hindsight 状态变为 Running，打开网页控制台。
6. 从智能体所在环境检查 API 健康状态：

   ```bash
   curl -fsS "${HINDSIGHT_API_URL}/health"
   ```

   先将 `HINDSIGHT_API_URL` 设为实际 API 入口地址。正常响应包含 `"status":"healthy"` 和 `"database":"connected"`。

:::tip 升级后的旧配置
升级可能保留之前填写的服务地址。要使用动态生成的 Router 地址，请清空 `HINDSIGHT_API_LLM_BASE_URL` 并应用更改。
:::

## 获取 API 地址并选择记忆库

1. 在 Olares 设置中打开 Hindsight 的应用设置，找到 API 入口。网页控制台入口和 API 入口是两个不同地址。
2. 复制 API 入口地址，保留 HTTPS。作为客户端基础地址使用时，去掉 `/health` 等路径。
3. 检查入口访问策略是否允许目标智能体访问。Chart 默认将 API 入口设为内部访问。请从智能体所在环境测试 `/health`，不能只看浏览器是否能打开。
4. 在 Hindsight 控制台创建记忆库，复制它的 ID，例如 `hermes-personal` 或 `opencode-project-a`。

不同项目或用户可以使用独立记忆库。多个智能体使用同一个 bank ID 时，可以共享其中的记忆。bank ID 用于组织数据，不能替代访问控制。

## 连接 Hermes

以下命令在 Hermes 环境中执行。本次使用的 Hermes 镜像已包含 `hindsight-client` 0.10.2；客户端包和 Hermes 的记忆提供者插件是两个不同组件。

1. 如果找不到 Hindsight 记忆提供者，从 Hermes 插件目录安装：

   ```bash
   hermes plugins install hindsight
   ```

2. 启动记忆配置向导：

   ```bash
   hermes memory setup
   ```

3. 选择 Hindsight，再选择 Local External（`local_external`，连接已运行的服务）。填写 Hindsight API 地址和 bank ID。仅当 Hindsight 入口要求 API token 时，填写该 token。
4. 如果向导提供初始记忆模板，可选择适合的角色，也可以选择 Blank（空白）。使用已有记忆库时，注意不要误覆盖其配置。
5. 开始新的 Hermes 对话：

   ```bash
   hermes
   ```

Local External 连接已经部署在 Olares 上的服务。Local Embedded（`local_embedded`，内嵌服务）会启动另一套 Hindsight，需要安装 `hindsight-all` 等额外依赖。

### 按需调用，关闭自动记忆

如果不希望每轮对话都读写记忆，优先在 Plugins 页面修改 Hindsight 提供者设置。

通过 UI 修改：

1. 打开 Hermes Web UI 的 Plugins（插件）页面，确认当前 profile 是你要使用的配置。
2. 在 Runtime Provider Plugins 区域，确认 Memory Provider 为 `hindsight`，状态为 ready / active。保留 Mode、API URL 和 Bank ID。
3. 在 Hindsight 设置中找到 Auto Recall，关闭它。这会停止每轮对话前自动检索记忆。
4. 找到 Auto Retain 并关闭它，停止自动保存对话。如果当前插件版本没有显示 Auto Retain，请使用下方文件方式设置 `auto_retain: false`；不要用 Retain Indicator 代替。
5. 点击 Save，刷新页面检查开关状态已保存，再开始新的 Hermes 会话。

:::info 自动调用与提示显示
Recall Indicator 和 Retain Indicator 只控制“已检索记忆”和“saving to memory”提示的显示。关闭 Indicator 不会停止实际读写。Recall Sync 控制自动检索的同步方式，也不是关闭自动检索的开关。

通用 Memory 页面中的 Memory Enabled、User Profile Enabled 控制 Hermes 自带记忆，不控制 Hindsight 的自动保存和检索。
:::

如果当前 UI 缺少自动记忆开关，可以直接编辑配置文件。也可以在 Files 页面访问该文件，下载并备份，在电脑上修改后上传到同一目录。上传同名文件会覆盖原文件，请保留其他字段并检查目录。Files 无法访问配置目录时，使用直接编辑方式。

直接编辑配置文件：在默认 Hermes 环境中打开 `~/.hermes/hindsight/config.json`；本 Olares Chart 默认使用 `/opt/data/hindsight/config.json`。把以下字段合并到已有配置：

```json
{
  "auto_retain": false,
  "auto_recall": false
}
```

保留原有连接地址和记忆库配置，然后开始新的 Hermes 会话。这两个开关属于 Hindsight 提供者；通用 Memory 设置中的开关控制 Hermes 自带记忆。

只要 Hindsight 提供者已激活、工具已启用，关闭自动调用后仍可手动使用：

| 目标 | 示例提示词 |
| --- | --- |
| 保存事实 | 使用 `hindsight_retain` 记录：我主要在 Olares 上开发和部署应用。 |
| 查询事实 | 使用 `hindsight_recall` 查询我主要在哪个平台开发和部署应用。 |
| 分析历史决策 | 使用 `hindsight_reflect` 比较已记录的 embedding 配置和变更原因，分别列出当前事实、历史事实和待验证事项。 |

检查工具返回结果，并在记忆库中查看。聊天回复“已记录”并不能证明处理完成。测试时可以开启另一段对话，从同一记忆库检索刚才保存的事实。

近期 Hermes 插件可能默认只检索汇总后的 observation（归纳记忆）。如果新保存的事实还未完成归纳，暂时查不到，请检查提供者的 `recall_types`。官方插件支持设置为 `"observation,world,experience"`，同时检索原始事实。

## 连接其他智能体

以下接入方式由上游支持。本次 Olares 环境已测试 Hermes；OpenClaw 和 OpenCode 还需要在你的环境中验证。

### OpenClaw（龙虾）

1. 在 OpenClaw 环境中安装插件并启动配置向导：

   ```bash
   openclaw plugins install @vectorize-io/hindsight-openclaw
   npx --package @vectorize-io/hindsight-openclaw hindsight-openclaw-setup
   ```

2. 选择 External API，填写 Hindsight API 地址；如入口要求认证，再填写 Hindsight API token。
3. 如需固定的共享或独立记忆库，设置 bank ID。配置完成后启动或重启 OpenClaw gateway。

该插件会占用 OpenClaw 的 memory 插件槽位。它的自动记忆行为与 Hermes 分开配置。版本要求和完整参数见 [OpenClaw 官方接入说明](https://github.com/vectorize-io/hindsight/blob/v0.10.2/hindsight-integrations/openclaw/README.md)。

### OpenCode

在已有的 `opencode.json` 中加入官方插件：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["@vectorize-io/opencode-hindsight"]
}
```

如果已有其他插件，把该条目合并到列表。通过 Linux shell 启动 OpenCode 前，设置自托管服务地址：

```bash
export HINDSIGHT_API_URL="https://your-hindsight-api-host"
export HINDSIGHT_BANK_ID="opencode-project-a"
opencode
```

将示例主机名替换为实际 API 入口地址。仅当入口要求 token 时，设置 `HINDSIGHT_API_TOKEN`。未指定 API 地址时，该插件默认连接 Hindsight Cloud。

插件提供 retain、recall、reflect 工具和自动记忆钩子。更多 coding harness 接入方式见 [官方集成列表](https://hindsight.vectorize.io/integrations)。客户端插件独立发布，版本和配置不必与服务端相同。

## 常见问题

### 启动时报 LLM API key 必填

进程收到的 key 为空。Chart 0.0.16 会在安装输入留空时自动提供 `not-required`。请检查已安装 Chart 是否包含该处理。如果模型端点开启认证，应填写它要求的有效密钥。

### 启动时报 `upstream_circuit_open`

Router 在上游连续失败后暂时停止转发请求。先检查 `default-embedding` 指向哪个模型，用短文本测试，再查看模型应用日志。

本次测试中，EmbeddingGemma 曾报 `infer failed on device GPU.0: general error`。推理错误触发 Router 熔断，随后 Hindsight 的 embedding 启动检查失败。先恢复模型推理，再重试 Hindsight。GPU 错误持续出现时，可以将同一嵌入模型切换到 CPU 测试。调大 Hindsight 资源配额无法修复独立的模型服务。

### 回答后仍显示 saving to memory

保存记忆可以异步执行，还包括事实提取、嵌入计算和后台处理。先检查记忆库和操作结果。如果保存反复失败，再查看 Hindsight 和模型服务日志。只想主动保存时，可以关闭 `auto_retain`。

### retain 或 reflect 很慢

这些操作需要模型推理。检查聊天请求与后台记忆处理是否争用并发能力较低的模型路由。先用短事实和手动调用测试，处理尚未完成时不要反复提交同一段长对话。

API 容器上限为 1 核 CPU / 2Gi 内存，控制台上限为 0.5 核 / 512Mi。实测存在资源压力时再提高配额。聊天模型、嵌入模型和系统 PostgreSQL 使用各自独立的资源配置。

## 延伸阅读

- [Hindsight 0.10.2 发布说明](https://github.com/vectorize-io/hindsight/releases/tag/v0.10.2)
- [Hermes 接入与提供者配置](https://github.com/vectorize-io/hindsight/blob/v0.10.2/hindsight-integrations/hermes/README.md)
- [OpenCode 接入说明](https://github.com/vectorize-io/hindsight/blob/v0.10.2/hindsight-integrations/opencode/README.md)
- [Olares 文档](https://docs.olares.cn)
