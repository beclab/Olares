---
outline: [2, 3]
description: 在 Olares 上使用 MemPalace 保存项目笔记和对话记忆。导入文档、查找原文，并通过 MCP 让 Hermes Agent 和 OpenClaw 共享记忆。
head:
  - - meta
    - name: keywords
      content: Olares, MemPalace, AI 记忆, MCP, Hermes Agent, OpenClaw, 语义搜索, 本地记忆
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-30"
---

# 使用 MemPalace 在不同 AI Agent 之间共享项目记忆

MemPalace 是面向 AI Agent 的本地记忆服务。它按原文保存笔记和对话，并帮助你通过语义搜索找到相关内容。

在 Olares 上，导入的记忆独立于 Agent 的聊天会话持久保存。Hermes Agent 和 OpenClaw 可以通过 MCP 使用同一个记忆库，源文档则保存在 Olares Files 中。默认的文本嵌入和搜索在本地运行，无需云端 API 密钥。

本指南以虚构的 Lakeview 厨房翻新项目为例。你将导入项目简报，让 Hermes Agent 查找文件并保存会议笔记，再在新会话和 OpenClaw 中找回这条笔记。MCP 连接为 Agent 提供记忆工具。本流程不会配置自动保存或自动召回，你需要明确要求 Agent 搜索或保存记录。

## 学习目标

- 从 Files 导入项目简报并搜索原文。
- 让 Hermes Agent 阅读简报并保存会议笔记。
- 在新的 Hermes 会话中找回会议笔记。
- 让 OpenClaw 找回 Hermes 保存的笔记。

## 前提条件

- **Olares OS：** 1.12.7 或更高版本。

如需借助 AI Agent 整理和查找项目信息，请在阅读相应章节前配置好以下应用：

- **Hermes Agent：** 按照 [Hermes Agent 配置指南](hermes.md#配置-hermes-agent)完成配置，并选择 Router 的 `default-chat` 模型。
- **OpenClaw：** 按照 [OpenClaw 配置指南](openclaw.md#初始化-openclaw)完成配置，使用 OpenClaw 2026.10.5 或更高版本，并选择 Router 的 `default-chat` 模型。

## MemPalace 如何组织记忆

MemPalace 将保存的文本分为三个层级：

| 层级 | 保存什么 | 本指南中的例子 |
| --- | --- | --- |
| **Wing** | 同一项目、人物或其他主题的记录 | 导入的简报放在 `lakeview-home`，Hermes 将会议笔记保存在 `lakeview-agent`。 |
| **Room** | Wing 中某个话题的记录 | 简报放在 `general`，会议笔记放在 `client-meeting`。 |
| **Drawer** | 单条记录或导入文件一部分的原始文本 | 搜索结果可以通过 Drawer ID 标明来自 `brief.md` 的段落或保存的会议笔记。 |

两个 Wing 都属于同一个 MemPalace 记忆库，连接的 Agent 均可访问。使用 `--wing` 可以只搜索指定 Wing。限定为 `lakeview-home` 时，结果不会包含 `lakeview-agent` 中的会议笔记。

## 安装 MemPalace

1. 打开 Market，搜索“MemPalace”。
2. 点击 **Get**，再点击 **Install**，等待安装完成。

::: tip 首次导入可能需要更长时间
首次导入文件或写入记忆时，会下载约 80 MB 的本地嵌入模型。如果下载超时，请在网络恢复后重试。MemPalace 会缓存模型，供后续操作使用。
:::

## 从 Files 导入项目资料

1. 在电脑上将下面的示例笔记保存为 `brief.md`：

   ```markdown
   # Lakeview kitchen renovation

   The approved budget is EUR 18400.
   Cabinet delivery is booked for 10 November 2026.
   Installation starts on 12 November 2026.
   The delivery contact is Olivia Carter.
   The client prefers oak cabinetry and warm white, dimmable lighting.
   Northstar Joinery will supply the cabinets.
   ```

2. 打开 Olares Files，在 **Home** 中新建 `MemPalace-demo` 文件夹，并将 `brief.md` 上传到其中。
3. 从启动台打开 MemPalace Terminal。Olares Files 中的 `/Files/Home/MemPalace-demo/` 在终端内挂载为 `/home/app/userdata/Home/MemPalace-demo`。运行以下命令导入：

   ```bash
   mempalace mine /home/app/userdata/Home/MemPalace-demo --wing lakeview-home
   ```

   成功导入后，输出会显示 `brief.md` 已处理，并归档了一条 Drawer：

   ![MemPalace 导入 brief.md 并归档一条 Drawer](/images/manual/use-cases/mempalace-import.png#bordered)

   Home 以只读方式挂载。MemPalace 会将导入的记忆保存在自己的应用数据中。

4. 搜索导入笔记中的橱柜送货日期：

   ```bash
   mempalace search "When is cabinet delivery?" --wing lakeview-home --results 3
   ```

   检查返回的段落是否写明送货日期为 2026 年 11 月 10 日，并将 `brief.md` 列为来源。

::: tip 关于这次搜索
- `--wing lakeview-home` 将搜索范围限定在导入简报的 Wing。
- `--results 3` 最多返回三个匹配结果。你可以将 `3` 改成其他数字。

后续导入 Lakeview 简报的其他资料时，请沿用同一个 Wing。
:::

## 让 Hermes Agent 使用同一个记忆库

无需配置消息平台，也可以使用 Hermes CLI。本节先让 Hermes 阅读导入的简报，再保存一条供后续会话查找的笔记。

### 复制 MCP Endpoint

1. 打开 **Settings** > **Applications** > **MemPalace**。
2. 在 **Entrances** 下选择 **MemPalace MCP**。
3. 确认 **Authentication level** 为 **Internal**。
4. 复制 **Endpoint** URL。

   ![从 MemPalace MCP 入口复制 Endpoint](/images/manual/use-cases/mempalace-endpoint.png#bordered)

### 在 Hermes 中添加连接

1. 打开 Olares Files，找到 **Data** > **hermesagent** > **home** > `config.yaml`。
2. 点击 <i class="material-symbols-outlined">edit_square</i> 进入编辑。
3. 在 `mcp_servers` 下添加 `mempalace`。将 `<mempalace-endpoint-url>` 替换为上一步复制的完整 URL：

   ```yaml
   mcp_servers:
     mempalace:
       url: "<mempalace-endpoint-url>/mcp"
       tools:
         include:
           - mempalace_status
           - mempalace_search
           - mempalace_add_drawer
           - mempalace_get_drawer
           - mempalace_diary_write
   ```

4. 在现有的 `platform_toolsets.cli` 列表中添加 `mcp-mempalace`：

   ```yaml
   platform_toolsets:
     cli:
       - hermes-cli
       - mcp-mempalace
   ```

   将上述内容加入 `config.yaml` 中已有的列表，不要删除其他 MCP 服务或 CLI 工具集。

5. 点击 <i class="material-symbols-outlined">save</i> 保存更改。

### 阅读导入的简报

1. 在 Hermes CLI 中开始新会话：

   ```bash
   hermes chat
   ```

2. 让 Hermes 搜索 `lakeview-home` Wing：

   ```text
   Search MemPalace's lakeview-home wing for the Lakeview kitchen brief.
   What's the approved budget and who supplies the cabinets? Include the source.
   ```

   Hermes 应从 `brief.md` 中找出预算和供应商，并标明来源。

   ![Hermes 查找预算、供应商和来源](/images/manual/use-cases/mempalace-hermes-chat1.png#bordered)

### 保存会议笔记并在新会话中找回

这条会议笔记重复了简报中的送货信息，并增加了项目编号。

1. 在当前 Hermes CLI 会话中，让 Hermes 保存会议笔记：

   ```text
   Use MemPalace to save this exact project note in wing lakeview-agent,
   room client-meeting:
   "Lakeview client meeting: cabinet delivery is booked for 10 November 2026.
   Delivery contact is Olivia Carter. Project reference: LV-7421."
   Confirm after the memory tool succeeds.
   ```

   Hermes 可能会显示类似下面的保存确认：

   ![Hermes 确认已将 Lakeview 会议笔记保存到 MemPalace](/images/manual/use-cases/mempalace-hermes-chat2.png#bordered)

2. 输入 `/exit` 并按 **Enter** 退出 Hermes CLI，然后运行 `hermes chat` 开始新会话。

3. 在新会话中，让 Hermes 搜索共享记忆：

   ```text
   Search MemPalace for the Lakeview client meeting.
   What delivery date, contact, and project reference did we record?
   Include the source drawer ID.
   ```

4. Hermes 可能会像下面这样返回会议详情和来源 Drawer ID：

   ![Hermes 找回 Lakeview 会议详情和来源 Drawer ID](/images/manual/use-cases/mempalace-hermes-chat3.png#bordered)

## 从 OpenClaw 使用这份记忆

OpenClaw 可以连接同一个 MemPalace 服务，找回 Hermes 保存的会议笔记。

1. 打开 OpenClaw CLI，注册 MemPalace MCP Endpoint。将 `<mempalace-endpoint-url>` 替换为[复制 MCP Endpoint](#复制-mcp-endpoint)中获得的完整 URL：

   ```bash
   openclaw mcp add mempalace \
     --url "<mempalace-endpoint-url>/mcp" \
     --transport streamable-http \
     --include mempalace_search,mempalace_add_drawer,mempalace_get_drawer,mempalace_status,mempalace_diary_write
   ```

2. 运行 `openclaw mcp probe`，输出中应显示 `mempalace` 有五个工具。
3. 在 Market 中打开 **My Olares**，找到 OpenClaw。打开操作菜单，选择 **Stop**，然后选择 **Resume**。等待 OpenClaw 恢复运行。
4. 从启动台打开 Control UI，选择 **New Session**。
5. 向 OpenClaw 提问：

   ```text
   Search MemPalace for the Lakeview client meeting.
   What delivery date, contact, and project reference did we record?
   Include the source drawer ID.
   ```

OpenClaw 应返回 Hermes 保存的会议笔记中的送货日期、联系人和项目编号。

## 添加文档和对话导出记录

将文件上传到 Olares Files 中的文件夹，然后在 MemPalace Terminal 中按文件类型运行对应命令：

| 来源 | 命令 |
| --- | --- |
| 项目笔记、Markdown 和代码 | `mempalace mine <folder> --wing <project>` |
| 对话导出记录 | `mempalace mine <folder> --mode convos --wing <project>` |
| PDF、DOCX、PPTX、XLSX、RTF 和 EPUB | `mempalace mine <folder> --mode extract --wing <project>` |

运行前替换占位符：

- `<folder>`：文件夹在 MemPalace Terminal 中的挂载路径。
- `<project>`：保存导入记录的 Wing。要将文件加入简报所在的 Wing，请使用 `lakeview-home`。

::: tip 文档导入
`extract` 模式导入可搜索的文本，不保留文档的版式或图片。原文件请保留在 Olares Files 中。扫描件需要先进行 OCR，再导入。本应用已包含所需的文本提取工具。
:::

## 进阶用法

### 检查记忆的组织情况

在 MemPalace Terminal 中运行以下命令，查看整个记忆库的组织情况：

```bash
mempalace audit
```

评分衡量的是记忆的组织情况，并非搜索准确率。查看分数下方的具体发现，了解哪些问题值得检查。`n/a` 表示该项没有可报告的分数。

:::: details 可选：通过 Router 整理项目记忆

随着项目记录增多，你可以按主题将它们归入不同 Room。MemPalace 可以通过 Olares Router 调用模型，为指定 Wing 建议 Room。应用更改前，请查看建议并预览记录将移动到哪里。

Lakeview 示例只导入了一份简报，因此建议中可能包含没有记录的 Room。如果一个 Wing 中有多个不同主题的笔记，MemPalace 就有更多记录可供分类。

<!--@include: ../../reusables/ai-service-connections.md#app-endpoint-overview-->

1. 获取 Router Endpoint。打开 Olares **Settings** > **Applications** > **Router** > **Entrances**。选择 **Router**，然后复制 **Endpoint** URL。

   在下面的命令中，用该 URL 替换 `<router-endpoint-url>`。命令会添加 `/v1`，并使用 Router 中选为默认聊天模型的模型。

2. 先完成当前运行中的 Agent 任务。维护命令会在执行期间暂停 MCP，并在完成后自动恢复服务。在 MemPalace Terminal 中运行以下命令，为 `lakeview-home` 建议 Room：

   ```bash
   mempalace-maintenance rooms propose --wing lakeview-home \
     --llm-provider openai-compat \
     --llm-model default-chat \
     --llm-endpoint "<router-endpoint-url>/v1" \
     --accept-external-llm
   ```

   `--accept-external-llm` 允许将抽样笔记发送到指定的模型 Endpoint。模型在本地还是云端运行，取决于 Router 当前的 `default-chat` 路由。

3. `rooms propose` 完成后，在输出的 `Saved to` 行找到建议文件的路径。在 MemPalace Terminal 中运行 `cat`，后面接该路径，查看建议的 Room 名称和描述。确认它们适合项目后再继续。

4. 预览记录将如何归入建议的 Room：

   ```bash
   mempalace-maintenance rooms apply --wing lakeview-home --from all --show 3
   ```

   此命令只预览，不会应用更改。`--from all` 包含 `lakeview-home` 所有 Room 中的记录，`--show 3` 最多显示每个目标 Room 的三个例子。检查建议的归类是否符合记录内容。

5. 确认预览结果后应用更改：

   ```bash
   mempalace-maintenance rooms apply --wing lakeview-home --from all --yes
   ```

::::

### 备份记忆

源文件仍保存在 Olares Files 中。导入的记忆、日记和图谱数据保存在 MemPalace 的应用存储中。如需手动复制备份，请先在 Market 中停止 MemPalace，通过 Files 将 **Data** > **mempalace** > **mempalace** 复制到备份位置，然后恢复应用运行。源文档也应纳入日常 Files 备份。

默认的嵌入模型和搜索在 Olares 上运行。Agent 对话和可选的模型辅助整理使用 Router 中配置的模型服务商。如果希望这些模型调用也留在设备上，请在 Router 中选择本地模型。

## 了解更多

- [MemPalace 官方指南](https://mempalaceofficial.com/guide/getting-started)
