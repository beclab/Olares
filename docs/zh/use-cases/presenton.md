---
connectionVersion: 1.12.7
connectionLatestPath: /zh/use-cases/presenton
outline:
- 2
- 3
description: 在 Olares 上通过 AI Router 创建可编辑的演示文稿
head:
- - meta
  - name: keywords
    content: Olares, Presenton, AI, 演示文稿, PowerPoint, PDF
app_version: "0.0.10"
doc_version: "1.2"
doc_updated: "2026-11-01"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/presenton.md)为准。
:::

# 使用 Presenton 创建可编辑的演示文稿

Presenton 根据提示词或文档创建幻灯片。选择设计、检查生成内容后，可导出可编辑的 PowerPoint 文件或 PDF。本 Olares 应用使用 AI Router 的默认对话模型，并将演示文稿数据保存在自己的设备中。

## 学习目标

- 根据提示词生成简短演示文稿。
- 检查并编辑大纲和幻灯片。
- 导出结果，继续编辑或分享。
- 通过已连接的助手浏览模板、上传文本参考资料。

## 前提条件

- 运行 Olares 1.12.7 或更高版本的 amd64 设备。
- 已连接 [Olares 专用网络](../manual/larepass/private-network.md)的浏览器。
- AI Router 中可用的 **default-chat** 路由。

应用固定使用上游 0.9.12-beta 容器镜像。本应用关闭图片生成，基础流程无需配置单独的图片服务。模板可能使用占位图片。分享前请检查。

## 安装 Presenton

1. 打开应用市场并安装 **Presenton**。
2. 从启动台打开 Presenton。
3. 等待首次启动完成。提示时创建实例账号，然后登录。

应用已预先配置为通过 AI Router 使用 **default-chat**。此流程无需粘贴外部模型密钥。首次使用可能需要下载本地索引资源，之后会重复使用已保存的缓存。

## 创建简短演示文稿

在新建演示文稿页面选择 **Standard**，将幻灯片数量设置为 **1**，语言选择 **English**，再输入具体说明。例如：

> Create one introductory slide for the Portland Library Community Garden. Volunteers meet on Saturdays at 9 AM. Maya Chen coordinates the group. The garden welcomes new volunteers. Use a short title and readable English text.

点击 **Generate presentation**。检查生成的大纲、页数和主题顺序，选择合适的内置模板后，再继续生成完整幻灯片。

![Presenton 中的演示文稿流程](/images/manual/use-cases/presenton-create.png#bordered)

## 检查并编辑

逐页阅读幻灯片。双击文字元素即可编辑，完成后点击元素外部。在顶部点击演示文稿标题可修改名称。导出前请修正姓名、日期和事实。模型可以提供初稿，但可能遗漏细节，或生成不适合当前设计长度的文字。

请对照来源检查幻灯片中可见的姓名、日期、金额和链接，并检查模板自带的数字。助手的生成任务完成后，幻灯片仍可能出现无关的占位指标或遗漏来源细节。分享前请在编辑器中修正。

先用一页幻灯片检查默认模型能否完成流程。较长演示文稿和复杂排版取决于模型返回结构化内容的能力。如果生成较慢，先查看当前任务，不要重复提交相同提示词。

![Presenton 中编辑后的幻灯片](/images/manual/use-cases/presenton-editor.png#bordered)

## 导出并保存结果

点击 **Export**。如需在演示软件中继续编辑，选择 **PPTX**。如需阅读和分享，选择 **PDF**。分享前，打开导出文件并检查排版。

演示文稿记录使用 Olares PostgreSQL。上传文档、素材和导出文件保存在应用的私有数据目录中。下载重要导出文件，并在备份计划中同时包含数据库和应用数据。

## 通过助手准备参考资料

Presenton 提供原生 MCP 服务，已连接的助手可以浏览模板、上传参考文件和发起演示文稿任务。Lares 和 Hermes Agent 均可使用该服务。你可以让助手整理需求和参考资料，再在 Presenton 中检查和编辑幻灯片。

### 连接助手

独立的 internal MCP 入口需要 Presenton 0.0.10 或更高版本。在已配置当前设备连接的 Olares CLI 电脑终端中运行：

```bash
olares-cli settings apps get presenton -o json
```

在 `entrances` 中找到名为 `mcp` 的条目。将 `https://` 与该条目的 `url` 拼接，再添加 `/mcp`。使用此地址，而非应用的主网页地址。

以实例管理员身份登录 Presenton。在应用网页入口下打开 `/admin`，选择 **API keys**。选择账号，填写名称和有效期，然后创建并复制密钥。在助手的 MCP 连接中配置：

| 配置项 | 内容 |
| --- | --- |
| 连接方式 | Streamable HTTP |
| 服务地址 | 当前 Presenton 实例的 internal MCP 地址，以 `/mcp` 结尾 |
| 请求头 | `Authorization: Bearer <presenton-api-key>` |

将密钥保存在助手的私有连接设置中。这里使用 Presenton 自己的密钥，与 AI Router 密钥不同。从电脑连接时，先开启 [LarePass VPN](../manual/larepass/private-network.md)。internal 入口不是公开分享地址。如果请求跳转到 Olares 网页登录，应检查是否误用了应用的网页入口。

连接后，可以对助手说：

> 列出 Presenton 的内置模板，显示模板名称，先不要生成演示文稿。

助手应调用 `list_templates`。从返回的模板中选择，不要让助手自行编造模板标识。

### 上传文本参考资料

将以下虚构示例保存为 `garden-brief.txt`，或换成你自己的简短资料：

```text
Portland Library Community Garden
Coordinator: Maya Chen
Volunteer meeting: every Saturday at 9 AM
New volunteers are welcome.
```

让助手可以读取实际文件，然后提出请求：

> 将 garden-brief.txt 上传到 Presenton，作为参考资料。告诉我上传是否成功，先不要生成演示文稿。

助手通过 `upload_files` 上传内容，并取得 Presenton 返回的文件路径。后续生成任务应直接使用这些路径。如果助手只能看到文件名或附件标识，需要先让它取得文件内容。

Presenton 的文档处理服务可以读取 TXT。后续使用参考资料生成演示文稿时，请对照原始资料检查幻灯片。

:::info PDF 参考资料
上传 PDF 和提取 PDF 内容是两个独立步骤。当前应用可以上传 PDF，但内容解析尚不稳定。建议先将需要的文字整理成 TXT，同时保留原 PDF 供人工核对。
:::

### 检查助手返回的结果

对于演示文稿任务，助手应持续查询同一个任务，直到完成或返回错误。使用返回的编辑链接检查文稿，再导出已审阅的结果。任务仍在处理中时，不要重复发起。

上传参考资料是生成幻灯片的准备步骤，不代表已经完成演示文稿，也不代表最终幻灯片保留了原文的每项事实。

## 了解更多

- [Presenton 文档](https://docs.presenton.ai/)
- [Olares Router](./olares-router.md)

- [Lares](./lares.md)：在 Olares 上使用助手。
- [Hermes Agent](./hermes.md)：使用具有持久对话和工具的智能体。
