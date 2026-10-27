---
outline: [2, 3]
description: 在 Olares 上把 Goose 用作本地 AI 智能体。通过浏览器中的 Goose Terminal 对话，或从电脑上的 Goose Desktop 连接。
head:
  - - meta
    - name: keywords
      content: "Olares, Goose, AI 智能体, AI Router, Goose Desktop, 终端, 文件自动化"
app_version: "0.0.13"
doc_version: "1.0"
doc_updated: "2026-10-22"
---

# 在 Olares 上把 Goose 用作本地 AI 智能体

Goose 是一个开源的通用 AI 智能体，可用于研究、写作、编程、自动化，以及更多场景。在 Olares 上，Goose 运行在你的设备上，通过 Router 使用本地模型，默认使用 default-chat 聊天模型。你可以通过浏览器中的 Goose Terminal 或电脑上的 Goose Desktop 与它交互。

## 学习目标

通过本指南，你将学会：

- 在 Goose Terminal 中运行 Goose。
- 将 Goose Desktop 连接到 Olares 上的 Goose 服务。
- 可选：开放 Home 或 External 文件的访问权限。
- 可选：将 Goose 连接到其他模型服务商。

## 前提条件

开始之前，你需要：

- Olares 1.12.7 或更高版本。
- 可正常使用 [Router](olares-router.md) 进行本地模型连接。
- 已从 Market 安装以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.8-27B (llama.cpp) | 从 Market 安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 Goose

1. 打开 Market，搜索“Goose”。
2. 点击**获取**，然后点击**安装**，等待安装完成。

## 使用 Goose Terminal

Goose Terminal 在浏览器中打开，直接连接 Goose 服务，无需安装客户端，也无需配置连接。

### 运行一次性请求

使用 `goose run -t` 发送单个 prompt，无需进入交互式对话。

1. 从 Launchpad 打开 Goose Terminal。
2. 运行 `goose run -t`，后跟你的 prompt，并将 `Your_Prompt` 替换为你的问题或任务：

   ```bash
   goose run -t "Your_Prompt"
   ```

   例如，提问：

   ```bash
   goose run -t "Goose Terminal 和 Goose Desktop 有什么区别？请用三句话回答。"
   ```

3. Goose 显示回答后返回终端提示符。

### 开始对话

想要来回多轮对话时，使用交互式会话。

1. 从 Launchpad 打开 Goose Terminal。
2. 在工作目录中启动交互式会话：

   ```bash
   cd /config/workspace
   goose session
   ```

3. 首次运行此命令时，Goose 会询问是否分享匿名使用数据。选择 **Yes** 或 **No** 继续。
4. 发送你的 prompt。以下示例在 `/config/workspace/goose-demo` 中创建一个示例项目简报：

   ```text
   仅在 /config/workspace/goose-demo 中操作，如目录不存在请创建。
   创建 notes.md，包含以下事实：
   - 项目：优化客户首次使用流程
   - 目标：缩短完成首次配置所需的时间
   - 交付物：配置检查清单和故障排查指南
   - 目标日期：2026-11-30
   - 待确认事项：谁负责故障排查指南？
   然后读取 notes.md，生成 brief.md，包含以下章节：
   项目概述、交付物、目标日期、待确认事项。
   保留所有事实，不要编造负责人或时间安排。
   最后返回两个文件的绝对路径，并概述你创建的内容。
   ```

5. 如终端提示批准工具，按提示操作。
6. 在同一终端中验证结果：

   ```bash
   # 列出目录，确认文件存在
   ls -l /path/to/directory
   # 查看文件内容
   cat /path/to/file
   ```

   本例中：

   ```bash
   ls -l /config/workspace/goose-demo
   cat /config/workspace/goose-demo/brief.md
   ```

   简报应包含项目目标、两项交付物、目标日期，以及尚未确定负责人的问题。模型生成的措辞可能不同，但事实应保持一致。

7. 按 `Ctrl+C` 退出会话。如果 Goose 提示再次按键，再按一次 `Ctrl+C`。这会结束当前终端交互。

:::info 文件位置
Goose 将文件保存在 Olares 设备的 `/config/workspace` 目录中。重启应用不会删除这些文件。重要文档请在 Files 中另行保存一份。
:::

### 查找并恢复会话

会话保存在 Goose 服务上。你可以列出会话，并在之后于 Goose Terminal 或 Goose Desktop 中恢复任意会话。

- 列出已保存的会话，运行 `goose session list`。
- 恢复会话，运行 `goose session --resume --session-id SESSION_ID`，ID 取自列表。

   :::tip
   两个同时运行的会话编辑同一文件时可能互相覆盖修改。在另一个会话中编辑同一文件前，请先退出其中一个。
   :::

## 使用 Goose Desktop

Goose Desktop 运行在你的电脑上，通过 Goose ACP 入口连接到 Olares 上的 Goose 服务。

### 准备连接信息

1. 从 Launchpad 打开 Goose Terminal，运行：

   ```bash
   printenv GOOSE_SERVER__SECRET_KEY
   ```

2. 复制返回值并保存，稍后配置 Goose Desktop 时使用。

   这是 Goose Desktop 与 Goose 服务之间的连接密钥，不是模型 API 密钥，请妥善保管。

3. 打开 Olares 设置，进入**应用** > **Goose** > **入口** > **Goose ACP**，复制 **Endpoint** URL。
4. 打开 Router，进入 **API keys** 页面，为 `goose` 创建一个可访问 **default-chat** 模型的密钥。复制密钥值，它只显示一次，请妥善保管。
5. 在 Router 中进入 **LLM** 页面，找到被指定为 `default-chat` 的模型，点击右侧的 **View connection example**。
6. 根据电脑的连接方式选择：
   - 如果电脑与 Olares 设备在同一局域网，选择 **Device in LAN**，复制 Router URL。
   - 如果电脑通过互联网连接，选择 **Remote**，复制 Router URL。

### 连接到 Goose 服务

:::tip
如果电脑与 Olares 设备不在同一局域网，连接前请先开启 LarePass VPN。
:::

1. 按照[官方安装指南](https://goose-docs.ai/docs/getting-started/installation/)在电脑上安装 Goose Desktop。
2. 打开 Goose Desktop，选择 **Connect to a Provider** > **Add a custom provider** > **Configure manually**。
3. 在 **Add Custom Provider** 窗口中配置以下设置：

   | 设置项 | 值 |
   |:---|:---|
   | **Provider Type** | 选择 **OpenAI Compatible**。 |
   | **Display Name** | 填写便于识别的名称，例如 `olares-router`。 |
   | **API URL** | 填写之前复制的 Router URL。 |
   | **This provider requires an API key** | 勾选此选项。 |
   | **API Key** | 填写在 Router 中为 `goose` 创建的 API 密钥。 |
   | **Available Models** | 填写 `default-chat`。 |

4. 点击 **Create Provider**。
5. 按需选择隐私选项，然后点击 **Get Started**。
6. 在 Goose Desktop 中，从左侧边栏进入 **Settings** > **Agent**。
7. 找到 **External Backend (ACP)**，开启 **Use external backend**，并配置以下设置：

   | 设置项 | 值 |
   |:---|:---|
   | **Backend Base URL** | 在**准备连接信息**中复制的 Goose ACP **Endpoint** URL，不要追加 `/acp` 或 `/v1`。 |
   | **Remote Working Directory** (optional) | `/config/workspace`<br><br>**Note**：虽然 Goose Desktop 将此字段标记为可选，但这里必须填写。智能体需要一个服务端目录来工作。 |
   | **Secret Key** | 在**准备连接信息**中保存的连接密钥。 |

8. 重启 Goose Desktop。
9. 新建对话。在聊天框下方确认模型为 `default-chat`，工作目录为 `/config/workspace`。
10. 发送以下消息，确认 Goose 回复：

      ```text
      请仅回复：Goose is ready。
      ```

### 继续 Terminal 中的会话

你在 Goose Terminal 中创建的会话也会以 **CLI Session** 条目出现在 Goose Desktop 左侧边栏的 **CHATS** 下。打开它即可从上次中断处继续对话，工作目录和文件保持不变。

## 处理 Home 或 External 中的文件

默认情况下，Goose 只能访问自己的工作目录。要让 Goose 处理 Home 或 External 目录中的文件，请先开启访问：

1. 打开 Olares 设置，进入**应用** > **Goose** > **管理环境变量**。
2. 将所需目录的访问变量设为 `true`：

   | 环境变量 | 默认值 | 开启后 Goose 可访问的目录 |
   |:---|:---|:---|
   | `ALLOW_HOME_DIR_ACCESS` | `false` | Home：`/config/Home` |
   | `ALLOW_EXTERNAL_DIR_ACCESS` | `false` | External：`/config/External` |

3. 点击**应用**，等待 Goose 恢复为**运行中**。
4. 重启 Goose Desktop 或 Goose Terminal。
5. 向 Goose 提供待处理文件的准确路径。

   :::warning
   智能体可以修改已开放目录中的文件。建议先处理文档副本，确认输出后再用于原文件。
   :::

   :::info 访问范围
   开启变量后，整个目录都会开放给智能体，而不仅是提示词中提到的单个文件。如需收回访问权限，将该变量改回 `false` 并应用修改。
   :::

## 使用其他模型服务商

默认的模型连接是 Router。如需使用第三方模型服务商，在 Olares 设置中修改应用环境变量。这些设置同时控制 Goose Desktop 和 Goose Terminal 使用的服务端。

1. 打开 Olares 设置，进入**应用** > **Goose** > **管理环境变量**。
2. 对于兼容 OpenAI 的服务商，设置以下变量：

   | 环境变量 | 值 |
   |:---|:---|
   | `GOOSE_PROVIDER` | `openai` |
   | `GOOSE_MODEL` | 服务商提供的准确模型 ID |
   | `OPENAI_HOST` | 服务商的服务器根地址，例如 `https://api.openai.com`，不包含 `/v1` |
   | `OPENAI_BASE_PATH` | `v1/chat/completions`，或服务商指定的聊天补全路径 |
   | `OPENAI_API_KEY` | 服务商的 API 密钥 |

3. 对于 Anthropic，设置以下变量：

   | 环境变量 | 值 |
   |:---|:---|
   | `GOOSE_PROVIDER` | `anthropic` |
   | `GOOSE_MODEL` | 服务商提供的准确模型 ID |
   | `ANTHROPIC_API_KEY` | 服务商的 API 密钥 |

4. 点击**应用**，等待 Goose 恢复为**运行中**。
5. 重启 Goose Desktop 或 Goose Terminal。
6. 切回 Router：

   a. 将 `GOOSE_PROVIDER` 设为 `custom_olares_router`。

   b. 将 `GOOSE_MODEL` 设为 `default-chat`。

   c. 移除第三方密钥。

:::tip
如需同时更改 Goose Desktop 和 Goose Terminal 使用的模型，请编辑 Olares 服务端的环境变量。Goose Desktop 中的服务商设置不影响此配置。
:::

## 常见问题

### Goose Desktop 为什么连接失败？

确认 Goose 正在运行，且 Goose Desktop 使用的是从 Olares 设置中复制的 Goose ACP 地址。连接密钥使用 Goose Terminal 中查询到的值。如果提示密钥被拒绝，说明 Goose Desktop 中填写的值与服务端不一致。

### 模型请求为什么返回 503？

先确认 Router 的默认聊天模型已就绪。如果错误提示模型正在处理其允许数量的请求，请等待当前请求完成后再试。避免同时从 Goose Desktop 和 Goose Terminal 发送请求。

### 模型返回了空响应。请重新发送消息以继续。

未开启或未配置 External Backend (ACP)。在 Goose Desktop 中打开 **Settings** > **Agent**，检查 External Backend (ACP) 设置。

### Goose 为什么在 Olares 上找不到我的文件？

使用 Olares 上的路径，例如 `/config/workspace/goose-demo/brief.md`。处理 Home 或 External 文件前，先开启对应目录的访问，再使用 `/config/Home` 或 `/config/External` 路径。电脑上的路径不是服务端路径。

## 了解更多

- [Olares Router](olares-router.md)：配置和管理 Goose 使用的模型。
- [Goose 安装指南](https://goose-docs.ai/docs/getting-started/installation/)：安装适用于你的操作系统的 Goose Desktop。
