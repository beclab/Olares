---
outline: [2, 3]
description: 在 Olares 上使用 Goose 和 AI Router，连接 Goose Desktop，将项目笔记整理成 Markdown 简报。也可通过浏览器中的 Goose Terminal 进行对话和文件处理。
head:
  - - meta
    - name: keywords
      content: "Olares, Goose, AI 智能体, AI Router, Goose Desktop, 终端, 文件自动化"
app_version: "0.0.13"
doc_version: "1.0"
doc_updated: "2026-10-13"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/goose.md)为准。
:::

# 使用 Goose 在 Olares 上将项目笔记整理成简报

Goose 是一个开源 AI 智能体，可通过自然语言读取文件、编辑文档和执行命令。在 Olares 上，智能体运行于你的 Olares 设备，你可以通过电脑上的 Goose Desktop 或浏览器中的 Goose Terminal 与它交互。

应用默认使用 AI Router 的默认聊天模型。本示例在 Goose 的持久化工作目录中生成项目简报，无需先开放个人目录访问权限即可完成文件任务。

## 学习目标

通过本指南，你将学会：

- 将 Goose Desktop 连接到 Olares 上的 Goose 服务。
- 通过对话创建 Markdown 文档，并验证保存结果。
- 使用 Goose Terminal，并按需开放已有文件的访问权限。

## 前提条件

- Olares 1.12.7 或更高版本。
- [AI Router](olares-router.md) 已配置可正常使用且支持工具调用的默认聊天模型。

## 安装 Goose

1. 打开 Market，搜索“Goose”。
   <!-- ![Market 中的 Goose](/images/manual/use-cases/goose.png#bordered) -->

2. 点击**获取**，然后点击**安装**，等待安装完成。

Goose 已默认配置 AI Router，无需为这一默认连接填写模型服务商的 API 密钥。模型是否可用以及响应性能取决于 AI Router 中的模型配置。

## 配置 Goose Desktop

1. 按照[官方安装指南](https://goose-docs.ai/docs/getting-started/installation/)在电脑上安装 Goose Desktop。使用 1.54.0 版本，与本指南对应的服务端保持一致。

2. 在电脑上开启 LarePass VPN，以连接 Olares。

   :::tip 电脑与 Olares 在同一局域网？
   也可以使用 Olares 本地访问。域名和 Windows 配置方法请参阅[在本地访问 Olares 服务](../manual/best-practices/local-access.md)。下文使用入口的 HTTPS 地址。
   :::

3. 在 Olares 中打开**设置** > **应用** > **Goose**，在**入口**中选择 **Goose ACP**，复制其 HTTPS 地址。此入口用于连接客户端，不是浏览器聊天页面。

4. 从 Launchpad 打开 Goose Terminal，运行：

   ```bash
   printenv GOOSE_SERVER__SECRET_KEY
   ```

   复制输出值。这是连接你自己的 Goose 服务所需的密钥，不是模型 API 密钥，请妥善保管。

5. 在 Goose Desktop 中打开 **Settings**，找到 **External Backend (ACP)**，填写：

   | 设置项 | 值 |
   |:---|:---|
   | **Use external backend** | 开启 |
   | **Backend Base URL** | 从 Olares 设置中复制的 Goose ACP HTTPS 地址，不要追加 `/acp` 或 `/v1` |
   | **Secret Key** | `printenv GOOSE_SERVER__SECRET_KEY` 的输出值 |
   | **Remote Working Directory (optional)** | `/config/workspace` |
   | **Certificate Fingerprint (optional)** | 首次连接时留空 |

6. 应用设置，然后新建对话。如果已有窗口仍使用原来的连接或工作目录，请重启 Goose Desktop。

7. 发送以下消息，确认 Goose 能正常回复：

   ```text
   请仅回复：Goose 已就绪。
   ```

启用外部服务端后，模型请求和智能体命令均从 Olares 发起。电脑上的文件路径不会自动映射到远程智能体。连接字段的说明可参阅 Goose 的[远程服务端配置指南](https://goose-docs.ai/docs/guides/remote-goose-server/)。

## 创建项目简报

### 生成文档

1. 在已连接的对话中发送：

   ```text
   仅在 /config/workspace/goose-demo 中操作，如果目录不存在，请创建。
   创建 notes.md，包含以下事实：
   - 项目：优化客户首次使用流程
   - 目标：缩短完成首次配置所需的时间
   - 交付物：配置检查清单和故障排查指南
   - 目标日期：2026-11-30
   - 待确认事项：谁负责故障排查指南？
   然后读取 notes.md，生成 brief.md，包含以下章节：
   项目概述、交付物、目标日期、待确认事项。
   保留所有事实，不要编造负责人或时间安排。
   最后返回两个文件的绝对路径，并概述已完成的工作。
   ```

2. 如果 Goose 请求批准工具调用，先确认操作内容，再允许写入文件。等待任务完成。

3. 确认回复中包含 `/config/workspace/goose-demo/notes.md` 和 `/config/workspace/goose-demo/brief.md`。

简报应包含项目目标、两项交付物、目标日期，以及尚未确定负责人的问题。模型生成的措辞可能不同，但事实应保持一致。

### 验证结果

1. 从 Launchpad 打开 Goose Terminal，检查已保存的文件：

   ```bash
   ls -l /config/workspace/goose-demo
   cat /config/workspace/goose-demo/brief.md
   ```

2. 将简报与原始笔记对照，确认 Goose 实际写入了文件，而不只是把建议内容显示在对话中。

3. 返回同一对话，继续发送：

   ```text
   读取 /config/workspace/goose-demo/brief.md，追加“后续行动”章节，
   提出如何确认故障排查指南负责人的建议。将建议明确标注为提议，
   保留现有项目事实。
   ```

4. 在 Goose Terminal 中再次读取文件，检查修改结果。

`/config/workspace` 中的文件会在应用正常重启后保留。对需要长期保存的文档，请另行备份。

## 使用 Goose Terminal

### 开始对话

1. 从 Launchpad 打开 Goose Terminal。

2. 在工作目录中启动交互式会话：

   ```bash
   cd /config/workspace
   goose session
   ```

3. 在提示符后输入消息，也可使用上面的项目简报任务。如果终端请求批准工具调用，按提示操作。

4. 按 `Ctrl+C` 退出会话。如果 Goose 先中断当前请求并提示再次按键退出，再按一次 `Ctrl+C`。这会结束当前终端交互，Goose 服务仍可供 Desktop 客户端连接。

如需执行一次性请求，可运行：

```bash
cd /config/workspace
goose run -t "读取 goose-demo/brief.md，概括待确认事项，不要修改任何文件。"
```

### 查找并恢复会话

列出已保存的会话：

```bash
goose session list
```

使用列表中的会话 ID 恢复会话：

```bash
goose session --resume --session-id SESSION_ID
```

将 `SESSION_ID` 替换为实际会话 ID。避免在两个正在运行的会话中同时编辑相同文件。

## 配置目录访问和模型

### 处理 Home 或 External 中的文件

Home 和 External 目录访问为可选项。如需处理已有文件：

1. 打开**设置** > **应用** > **Goose**，编辑应用环境变量。

2. 开启所需目录的访问权限：

   | 环境变量 | 默认值 | 开启后 Goose 可访问的目录 |
   |:---|:---|:---|
   | `ALLOW_HOME_DIR_ACCESS` | `false` | Home：`/config/Home` |
   | `ALLOW_EXTERNAL_DIR_ACCESS` | `false` | External：`/config/External` |

3. 保存修改，等待 Goose 恢复为**运行中**，然后重新连接客户端或打开 Goose Terminal。

4. 向 Goose 提供待处理文件的准确路径。建议先处理文档副本，确认输出后再用于原文件。

开启访问后，所选目录会开放给智能体，而不仅是提示词中指定的某个文件。你可以通过同一环境变量设置再次关闭访问。

### 使用其他模型服务商

应用默认使用 AI Router。如需使用第三方模型服务商，请在 Olares 设置中修改应用环境变量。已连接的 Desktop 客户端和 Goose Terminal 都会使用这些服务端设置。

对于兼容 OpenAI 的服务商，配置：

| 环境变量 | 值 |
|:---|:---|
| `GOOSE_PROVIDER` | `openai` |
| `GOOSE_MODEL` | 服务商提供的准确模型 ID |
| `OPENAI_HOST` | 服务商的根地址，例如 `https://api.openai.com`，不包含 `/v1` |
| `OPENAI_BASE_PATH` | `v1/chat/completions`，或服务商指定的聊天补全请求路径 |
| `OPENAI_API_KEY` | 服务商的 API 密钥 |

保存后，等待 Goose 恢复为**运行中**，再开始新对话。使用 Anthropic 时，将 `GOOSE_PROVIDER` 设置为 `anthropic`，在 `GOOSE_MODEL` 中填写模型 ID，并设置 `ANTHROPIC_API_KEY`。

如需切回 AI Router，将 `GOOSE_PROVIDER` 设置为 `custom_olares_router`，将 `GOOSE_MODEL` 设置为 `default-chat`，并移除不再需要的第三方密钥。客户端的服务商设置不会覆盖 Olares 服务端已经配置的环境变量。

## 常见问题

### 客户端为什么无法连接？

确认 Goose 正在运行、电脑已开启 LarePass VPN，并且客户端填写的是从 Olares 设置中复制的 Goose ACP 地址。连接密钥应使用 Goose Terminal 中查询到的值。如果提示密钥被拒绝，说明客户端填写的值与服务端不一致。

### 模型请求为什么返回 503？

先确认 AI Router 的默认聊天模型已就绪。如果错误提示模型正在处理其允许数量的请求，请等待当前请求完成后再试。避免同时从 Desktop 和 Terminal 发送请求。

### 为什么对话名称都是 New Chat？

自动生成 AI 会话标题默认关闭，你可以手动重命名会话。如需开启，将应用环境变量 `GOOSE_DISABLE_SESSION_NAMING` 设置为 `false`。标题生成会额外调用模型，可能与聊天请求争用并发有限的模型资源。

### Goose 为什么找不到文件？

请使用 Olares 上的路径，例如 `/config/workspace/goose-demo/brief.md`。处理 Home 或 External 文件前，先开启对应目录访问，再使用 `/config/Home` 或 `/config/External` 路径。电脑上的路径不是服务端路径。

## 了解更多

- [Olares Router](olares-router.md)：配置和管理 Goose 使用的模型。
- [Goose 官方安装指南](https://goose-docs.ai/docs/getting-started/installation/)：安装适用于你电脑操作系统的 Desktop 客户端。
