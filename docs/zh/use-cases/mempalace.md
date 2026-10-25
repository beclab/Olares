---
outline: [2, 3]
description: 在 Olares 上使用 MemPalace 保存项目笔记和对话记忆。导入文档、查回原文，并通过 MCP 让 Hermes Agent 和 OpenClaw 使用同一个记忆库。
head:
  - - meta
    - name: keywords
      content: Olares, MemPalace, AI 记忆, MCP, Hermes Agent, OpenClaw, 语义搜索, 本地记忆
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-25"
---

# 使用 MemPalace 在不同 AI Agent 之间共享项目记忆

MemPalace 是面向 AI Agent 的本地记忆服务。它保留笔记和对话的原始措辞，通过语义检索找回相关内容，并按项目和主题组织记录。它的知识图谱还可以跟踪事实随时间的变化。

在 Olares 上，导入的记忆独立于 Agent 的聊天会话持久保存。Hermes Agent 和 OpenClaw 可以通过 MCP 使用同一个记忆库，原始文件则保存在 Files 中。默认的文本嵌入和搜索在本地运行，无需云端 API 密钥。

本指南以一个名为 Lakeview 的虚构厨房翻新项目为例。你也可以用同样的方法管理自由职业客户的需求、家庭事务，或软件团队的设计决策。

::: info 翻译说明
本文为对应英文指南的中文译稿，使用 AI 辅助翻译并结合实际操作核对。
:::

## 学习目标

- 从 Files 导入项目笔记，并查回原始内容。
- 连接 Hermes Agent，在新的会话中找回已经保存的决定。
- 让 OpenClaw 使用同一份项目记忆。
- 添加办公文档，并了解日记、时序事实和记忆整理工具的用途。

## 适用场景

| 你想完成的事 | 对应能力 |
| --- | --- |
| 在多次客户会议之间保留偏好和决定 | 原文记录、语义搜索、来源筛选 |
| 将项目从一个助手交接给另一个助手 | MCP 共享记忆和 Agent 日记 |
| 从方案、电子表格或会议导出记录中查找细节 | 项目、对话和办公文档导入 |
| 跟踪预算、供应商或技术方案的变化 | 时序知识图谱、主题连接和组织审查 |

打开 MemPalace 后会进入 Web 终端。你可以直接运行命令，也可以让连接到 MCP 的 Agent 使用记忆工具。

## 前提条件

- Olares 1.12.7 或更高版本。
- 已[配置 Hermes Agent](hermes.md#配置-hermes-agent)，通过 Router 使用 `default-chat` 模型。

模型地址和身份验证规则参见[连接 AI 应用](../manual/best-practices/connect-ai-apps.md)。本流程使用本地终端会话，无需配置消息平台。

::: info 应用上架状态
本指南对应 MemPalace 3.11.0 测试包。请在你的 Market 中可以找到该版本后按下文安装，或使用应用提交中提供的测试包。
:::

## 安装 MemPalace

1. 打开 Market，搜索 `MemPalace`。
2. 点击 **Get**，再点击 **Install**，等待安装完成。
3. 从启动台打开 MemPalace Terminal，运行：

   ```bash
   mempalace status
   ```

首次导入或写入记忆时，会下载约 80 MB 的默认本地嵌入模型。应用安装可以先完成，模型在需要使用时再下载。如果下载超时，待网络恢复后重试原操作，已经下载好的模型会在后续操作中复用。

## 从 Files 导入项目资料

1. 在电脑上将下面的虚构项目笔记保存为 `brief.md`：

   ```markdown
   # Lakeview kitchen renovation

   The approved budget is EUR 18400.
   Cabinet delivery is booked for 10 November 2026.
   Installation starts on 12 November 2026.
   The delivery contact is Olivia Carter.
   The client prefers oak cabinetry and warm white, dimmable lighting.
   Northstar Joinery will supply the cabinets.
   ```

2. 打开 Files，在 **Home** 中新建 `MemPalace-demo` 文件夹，将 `brief.md` 上传到其中。
3. 在 MemPalace Terminal 中导入该文件夹：

   ```bash
   mempalace mine /home/app/userdata/Home/MemPalace-demo --wing lakeview-home
   ```

   等待输出 `Done.` 和已归档的 Drawer 数量。Home 通过只读方式挂载，导入的记忆会保存在 MemPalace 自己的应用数据中。

4. 使用笔记中的主题提出问题：

   ```bash
   mempalace search "When is cabinet delivery?" --wing lakeview-home --results 3
   ```

   阅读返回的原文和来源文件名。你可以使用与原文不同的表达方式查询，再根据来源内容作出项目决定。

   ![通过项目问题查回 Lakeview 笔记原文](https://cdn.olares.com/images/2026/10/ee7fefa1275966f278391220fd52869a.jpg)

Wing 用于组织项目，Room 用于组织项目中的主题，Drawer 保存单条记录。重复导入会跳过未变化的文件。后续继续导入同一项目时，请沿用原来的 Wing。

## 让 Hermes Agent 使用同一个记忆库

### 获取 MCP 地址

1. 打开 **Settings** > **Applications** > **MemPalace**。
2. 在 **Entrances** 下选择 **MemPalace MCP**。
3. 复制 **Endpoint** 地址，在末尾添加一次 `/mcp`。

   ![从 MemPalace MCP 入口复制 Endpoint 地址](https://cdn.olares.com/images/2026/10/dd2af69c620cb23c62900ad5f868787f.jpg)

本连接使用 MCP 入口，保留其 **Internal** 身份验证级别。下面的示例适用于运行在同一个 Olares 上的 Agent。

### 在 Hermes 中添加连接

1. 打开 Files，找到 **Data** > **hermesagent** > **home** > `config.yaml`。
2. 在 `mcp_servers` 下添加 `mempalace`，将示例 URL 替换为刚才复制的地址：

   ```yaml
   mcp_servers:
     mempalace:
       url: "https://<your-mempalace-mcp-endpoint>/mcp"
       tools:
         include:
           - mempalace_status
           - mempalace_search
           - mempalace_add_drawer
           - mempalace_get_drawer
           - mempalace_diary_write
   ```

3. 在现有的 `platform_toolsets.cli` 列表中追加 `mcp-mempalace`：

   ```yaml
   platform_toolsets:
     cli:
       - hermes-cli
       - mcp-mempalace
   ```

   请将这些内容合并到已有配置段中，保留当前模型、其他 MCP 服务和已经使用的工具集。

4. 保存文件，在 Hermes CLI 中启动新的会话：

   ```bash
   hermes chat
   ```

### 保存决定，再在新会话中召回

向 Hermes 提出请求：

```text
请使用 MemPalace，将下面的项目笔记原文保存到
wing lakeview-agent、room client-meeting：
"Lakeview client meeting: cabinet delivery is booked for 10 November 2026.
Delivery contact is Olivia Carter. Project reference: LV-7421."
请在记忆工具执行成功后确认。
```

保存成功后结束当前会话，再开始一个新会话，要求它查找共享记忆：

```text
请先使用 MemPalace 搜索 Lakeview client meeting。
我们记录的送货日期、联系人和项目编号是什么？
请附上来源 Drawer ID。
```

打开返回的记录，或将引用内容与原笔记核对。需要供后续会话使用的事实，可以让 Agent 通过 MemPalace 保存，也可以导入其来源文件。

## 从 OpenClaw 使用这份记忆

下面的可选流程使用支持原生 MCP 的 OpenClaw 2026.10.5。请先[配置 OpenClaw](openclaw.md#初始化-openclaw)，通过 Router 使用 `default-chat` 模型。

1. 打开 OpenClaw CLI，添加同一个 MemPalace 地址：

   ```bash
   openclaw mcp add mempalace \
     --url "https://<your-mempalace-mcp-endpoint>/mcp" \
     --transport streamable-http \
     --include mempalace_search,mempalace_add_drawer,mempalace_get_drawer,mempalace_status,mempalace_diary_write
   ```

2. 运行 `openclaw mcp probe`，连接结果中应列出 `mempalace` 的五个工具。
3. 在 Market 中重启 OpenClaw，待其运行后，从 Control UI 开始一个新会话。
4. 要求 OpenClaw 使用 MemPalace 搜索 Lakeview client meeting，并返回送货日期、联系人、项目编号和 Drawer ID。

现在两个 Agent 使用的是同一份记录。你也可以让 OpenClaw 保存安装预约，再让 Hermes 在新的会话中查询。

## 添加文档和对话导出记录

将更多来源文件上传到 Files 中的文件夹，再在 MemPalace Terminal 中选择对应的导入模式：

| 来源 | 命令 |
| --- | --- |
| 项目笔记、Markdown 和代码 | `mempalace mine <folder> --wing <project>` |
| 对话导出记录 | `mempalace mine <folder> --mode convos --wing <project>` |
| PDF、DOCX、PPTX、XLSX、RTF 和 EPUB | `mempalace mine <folder> --mode extract --wing <project>` |

将 `<folder>` 替换为挂载后的路径，例如 `/home/app/userdata/Home/MemPalace-demo`，将 `<project>` 替换为 Wing 名称。本应用包已包含文档提取所需的依赖。

办公文档导入会提取其中的文本供检索。原文档继续保存在 Files 中，用于查看版式和图片。扫描件需要先完成 OCR，再使用这个文本导入流程。查询一个主题时，可以添加 `--wing` 和 `--room`。来源及日期筛选参数可通过 `mempalace search --help` 查看。

## 整理长期积累的记忆

对于持续推进的项目，MemPalace 还提供以下工具：

| 需要处理的事 | 对应工具 |
| --- | --- |
| 会议结束后留下交接记录 | Agent 日记写入和读取 |
| 修改预算，同时保留之前的事实 | 知识图谱新增、替换和时间线 |
| 将供应商主题连接到项目简报 | Room 隧道和关系遍历 |
| 交换任务进度或完整的文本附件记录 | 任务事件、确认和 Artifact 工具 |

要从 Hermes 使用这些功能，可将相应工具名加入 `tools.include`，或移除该筛选段以发现完整工具集。例如，时序事实相关工具为 `mempalace_kg_add`、`mempalace_kg_query`、`mempalace_kg_supersede` 和 `mempalace_kg_timeline`。修改配置后开始新的会话。

在 MemPalace Terminal 中运行以下命令，可以查看记忆库的组织情况：

```bash
mempalace audit
```

::: details 可选：通过 Router 生成并应用主题分类
当某个 Wing 已积累较多笔记时，可以让模型建议主题分类。按照[连接 AI 应用](../manual/best-practices/connect-ai-apps.md)中的说明，复制 Router 为 **Apps in Olares** 提供的 **Base URL**，保留 `/v1`。

先结束当前 Agent 任务。维护命令会在执行期间暂停 MCP，并在完成后自动恢复服务。

```bash
mempalace-maintenance rooms propose --wing lakeview-home \
  --llm-provider openai-compat \
  --llm-model default-chat \
  --llm-endpoint "https://<your-router-endpoint>/v1" \
  --accept-external-llm
```

最后一个参数表示允许把抽样笔记发送到你指定的模型地址。模型在本地还是云端运行，取决于 Router 当前的 `default-chat` 路由。

查看建议的分类后，先预览重新归类的结果：

```bash
mempalace-maintenance rooms apply --wing lakeview-home --from all
```

确认方案合适后再应用：

```bash
mempalace-maintenance rooms apply --wing lakeview-home --from all --yes
```
:::

## 保存源文件和记忆数据

源文件保留在 Files 中，导入的记忆、日记和图谱数据保存在 MemPalace 的应用存储中。手动复制备份时，先在 Market 中停止 MemPalace，再通过 Files 将 **Data** > **mempalace** > **mempalace** 复制到备份位置，完成后恢复应用运行。源文档也应纳入平时的 Files 备份。

默认嵌入模型和搜索在 Olares 上运行。Agent 对话及可选的模型辅助整理，会使用通过 Router 配置的模型服务商。需要让这些模型调用也保留在设备上时，请在 Router 中选择本地模型。
