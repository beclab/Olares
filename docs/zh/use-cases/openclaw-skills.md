---
outline: [2, 3]
description: 学习如何管理、安装 OpenClaw 技能与插件，以及如何排除相关故障。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw 教程, OpenClaw 使用指南, 安装技能, 安装插件
app_version: "1.0.3"
doc_version: "1.1"
doc_updated: "2026-05-29"      
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/openclaw-skills.md)为准。
:::

# 管理 OpenClaw 技能和插件

OpenClaw 可以通过技能和插件进行扩展：
- 技能为 AI 添加新能力。例如，管理模型上下文协议（MCP）服务器。
- 插件可扩展系统以支持额外的通道或社区功能。例如，通过 BlueBubbles 添加 iMessage。

:::info 为什么需要手动安装
为了保护你的设备，OpenClaw 在一个受限的非 root 环境中运行，没有管理权限。智能体无法自行安装软件或修改系统，因此安装技能和依赖项需要由你手动完成。
:::

## 学习目标

通过本指南，你将学会：
- 理解技能加载机制，以及某些技能在 Olares 中被阻止的原因。
- 通过配置向导、ClawHub 或手动上传安装技能。
- 识别并安装技能缺失的依赖项。
- 安装并启用插件，以使用新通道或功能扩展 OpenClaw。

## 了解技能

了解技能的来源和加载方式有助于你有效地管理它们。

### 位置和优先级

技能从三个位置加载。如果多个位置存在同名的技能，OpenClaw 会使用优先级最高的那个，以便你轻松自定义或覆盖内置技能。

优先级从高到低的顺序如下：
1. 工作区技能（`Data/clawdbot/config/workspace/skills`）：每个智能体独有的技能，覆盖所有其他技能。
2. 托管/本地技能（`Data/clawdbot/config/skills`）：同一台机器上所有智能体共享的技能。
3. 捆绑技能：随 OpenClaw 安装一起提供的默认技能。

:::tip 查看所有可用技能
要查看智能体可用的完整技能列表，包括捆绑的、共享的和工作区技能，请在 OpenClaw CLI 中运行 `openclaw skills list` 命令。
:::

### Olares 上的兼容性

并非所有技能都能在 Olares 环境中运行。OpenClaw 会根据技能声明的要求，主动阻止无法正常运行的技能。

技能可能因以下原因被阻止：
- 操作系统不兼容：技能需要特定的操作系统（例如 macOS），而 Olares 运行在 Linux 上。例如，Apple 生态系统技能（如 apple-reminders）无法在 Olares 中使用。
- 缺少可执行文件（`bins`）：环境缺少必需的命令行工具，例如用于管理 GitHub 问题的 `gh`。
- 缺少配置（`config`）：`openclaw.json` 中未启用必需的设置。
- 缺少环境变量（`env`）：未提供必需的 API 密钥或身份验证令牌。

## 安装技能

可通过以下三种方式为 OpenClaw 安装新技能：
- 通过 `openclaw config` 向导安装技能。
- 从 ClawHub（OpenClaw 的包管理器）安装技能。
- 通过本地上传手动安装技能。

### 通过 `openclaw config` 安装

你可以使用内置配置向导安装默认或官方支持的技能。本示例将安装 `clawhub` 技能，供后续方法使用。

1. 打开 OpenClaw CLI。
2. 输入以下命令启动向导：

    ```bash
    openclaw config
    ```
3. 按照提示配置安装。使用方向键导航，按**回车**键确认。

    | 设置 | 选项 |
    |:---------|:-------|
    | Where will the Gateway run | Local (this machine) |
    | What do you want to configure | Skills |
    | Configure skills now | Yes |
    | Install missing skill dependencies | 导航到 **clawhub** 技能，按**空格**键<br>选中它，然后按**回车**。 |
    | Preferred node manager for skill installs | npm<br>等待 `Installed clawhub` 消息出现<br>后再继续。 |
    | Set [API_KEY] for [skill] | 对所有这些设置选择 **No**。|

4. 最后，对 **What do you want to configure** 选择 **Done**。出现 `Configure complete` 消息时，表示设置已完成。

### 从 ClawHub 安装

使用 ClawHub CLI 从 [ClawHub](https://clawhub.ai/) 搜索和安装技能。通过 ClawHub 安装技能会自动处理必要的包依赖。

:::tip 前提条件
确保已安装 [clawhub 技能](#通过-openclaw-config-安装)。该技能会启用 `openclaw skills` 命令，如 `list`、`search` 和 `install`，让你可以安装更多 ClawHub 支持的技能。
:::

1. 打开 OpenClaw CLI。
2. 要查看官方预设技能列表，请运行以下命令：

    ```bash
    openclaw skills list
    ```
3. 要搜索特定技能，请使用 `search` 命令。

    例如，要搜索日历技能，请运行以下命令：
    ```bash
    openclaw skills search Caldav Calendar
    ```

    终端返回搜索结果，开头显示技能 ID，后面是其描述。在本例中，技能 ID 为 `caldav-calendar`。

    ![ClawHub 中的技能 ID](/images/manual/use-cases/openclaw-skill-id.png#bordered)
    
    :::warning 安全建议
    强烈建议在安装技能之前，先在官方 ClawHub 网站上搜索并阅读该技能的详细信息。这可以确保你安装的是正确的技能，并防止安装恶意包。
    :::

4. 使用技能 ID 安装目标技能。
    
    例如，要安装此日历技能，应运行以下命令：

    ```bash
    openclaw skills install caldav-calendar
    ```

5. 等待终端指示技能已安装，然后通过运行以下命令验证：

    ```bash
    openclaw skills list
    ```

    **caldav-calendar** 的状态为 **ready**，表示安装成功。

6. 打开 **Control UI**，从左侧边栏选择 **Skills**，然后点击 **Ready** 标签页。你会看到新安装的技能已启用。

    ![新安装的技能已启用](/images/manual/use-cases/skill-enabled.png#bordered)

### 上传技能

1. 打开文件管理器，然后进入**应用** > **数据** > **clawdbot** > **config**。
2. 创建一个名为 `skills` 的新文件夹。
3. 将你的技能包（例如解压后的 `.zip` 文件）上传到此 `skills` 文件夹中。
4. 如有缺失，安装所需的包依赖。

## 安装缺失的依赖项

如果技能被阻止或无法使用，你需要识别并安装其缺失的依赖项。

1. 打开 OpenClaw CLI 并运行以下命令：

    ```bash
    openclaw skills check
    ```
    
    终端列出所有不可用的技能，并在括号中显示其缺失的要求。

    ![从 CLI 安装缺失的依赖项](/images/manual/use-cases/missing-dependency-cli1.png#bordered){width=70%}

2. 使用 `npm` 或 `brew` 手动安装依赖项。有关安装要求的详细信息，请参阅 `skills.md` 文件。

    - 示例：`gh-issues` 技能需要安装 `gh`。
    - 运行以下命令安装它：
        ```bash
        npm i -g gh
        ```
3. 缺失组件安装完成后，重启 OpenClaw 容器使更改生效：

    a. 从启动台打开控制面板。
    
    b. 点击**部署**下的 **clawdbot**，然后点击**重启**。

4. 验证安装：

    a. 打开 Control UI。
    
    b. 进入 **Skills** 页面。该技能现在应标记为 **eligible**。
    
    c. 如有需要，配置必需的 API 密钥，然后智能体将能够使用该技能。

## 安装插件

1. 在 OpenClaw CLI 中，输入以下命令检查兼容插件列表：

    ```bash
    openclaw plugins list
    ```

2. 在 **Name** 列中找到目标插件名称，然后输入以下命令安装它：

    ```bash
    openclaw plugins install {Name}
    ```
    例如，要安装 BlueBubbles，应输入以下命令：

    ```bash
    openclaw plugins install @openclaw/bluebubbles
    ```

    :::warning 插件安装被阻止
    如果安装失败并显示 `Plugin "{Name}" installation blocked` 错误，你可以在命令后附加 `--dangerously-force-unsafe-install` 来绕过此安全限制。务必先确认插件来源可信且安全，再使用此参数。
    
    例如：
    ```bash
    openclaw plugins install @openclaw/nextcloud-talk --dangerously-force-unsafe-install
    ```
    :::

3. 安装完成后，关闭 OpenClaw CLI 并重新打开它以加载新插件。
4. 通过检查插件状态进行验证：

    ```bash
    openclaw plugins list
    ```

    现在插件的状态为 **enabled**。

5. 打开 Control UI，从左侧边栏选择 **Settings**，然后进入 **Automation** > **Plugins**。
6. 找到 **@openclaw/bluebubbles** 并点击它以展开面板：

    - 如果已启用，关闭切换开关，然后再打开，以强制系统显式保存配置。
    - 如果已禁用，打开切换开关。

    ![开启插件](/images/manual/use-cases/toggle-plugin2.png#bordered)

7. 点击右上角的 **Save**。系统验证配置并自动应用更改。

    ::: tip 手动重启
    如果你需要手动重启 OpenClaw，请不要使用 OpenClaw CLI。请使用以下方法之一：
    - **从设置或应用市场重启应用**：
        - 打开**设置**，进入**应用** > **OpenClaw**，点击**暂停**，然后点击**恢复**。
        - 打开**应用市场**，进入**我的 Olares**，找到 **OpenClaw**，点击操作按钮旁边的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，选择**暂停**，然后选择**恢复**。
    - **重启容器**：打开控制面板，点击**部署**下的 `clawdbot`，然后点击**重启**。
    :::
