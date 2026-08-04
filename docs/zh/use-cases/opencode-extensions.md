---
outline: [2, 3]
description: 了解如何通过技能和插件扩展 Olares 上的 OpenCode。使用预安装的技能进行包管理和 Web 预览，或添加社区插件以获得额外功能。
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, skills, plugins, extensions, AI coding agent, self-hosted
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/opencode-extensions.md)为准。
:::

# 使用技能和插件扩展 OpenCode

OpenCode 支持两种类型的扩展：

- **Skills**：Markdown 指令文件，教 OpenCode 如何处理特定领域的任务。OpenCode 根据上下文自动加载它们。
- **Plugins**：JavaScript 或 TypeScript 模块，添加运行时功能。它们在启动时运行，并可以接入 OpenCode 的执行管道。

两者都可以全局范围或单个项目范围配置。

## 学习目标

在本教程结束时，你将学习如何：
- 使用预安装的 `system-admin` 和 `web-preview` 技能。
- 在全局或项目范围列出、加载和定位技能文件。
- 将插件安装为 npm 包或本地 `.js` / `.ts` 文件。
- 识别扩展 OpenCode 的热门社区插件。

## 技能

OpenCode 在对话期间自动加载相关技能，因此你很少需要自己管理它们。

### 预安装技能

Olares 上的 OpenCode 附带两个技能：

| 技能 | 描述 |
|:------|:------------|
| `system-admin` | 通过 `pkg-install` 进行系统包管理 |
| `web-preview` | 通过内置反向代理进行开发服务器预览 |

#### system-admin

在聊天中让 OpenCode 安装或删除系统包。例如，输入 "Install ffmpeg" 或 "Remove the curl package"，`system-admin` 技能将运行相应的 `pkg-install` 命令。

如果技能未激活，使用 `/skill load system-admin` 手动加载。

有关完整的命令参考，请参阅 [管理包](opencode-packages.md)。

#### web-preview

`web-preview` 技能在容器内启动开发服务器，并通过内置反向代理暴露它。

在聊天中描述你想要的内容：

```text
Start the web project in this folder on port 5544
```

OpenCode 启动服务器，确认它正在运行，并返回预览 URL：

```text
https://<your-OpenCode-domain>/__preview/<port>/
```

域名与你访问 OpenCode 时浏览器地址栏中显示的相同。

<!-- ![Web preview in browser](/images/manual/use-cases/opencode-web-preview.png#bordered) -->

如果技能未激活，使用 `/skill load web-preview` 手动加载。

### 管理技能

列出可用技能或手动加载一个：

```text
/skill list
/skill load <skill-name>
```

<!-- ![Skill list output](/images/manual/use-cases/opencode-skill-list.png#bordered) -->

技能文件是存储在以下位置的 Markdown 文件：

| 范围 | Olares Files 中的路径 |
|:------|:-----|
| 全局（所有项目） | `Application/Data/opencode/.config/opencode/skills/` |
| 项目级别 | `Home/Code/<project>/.opencode/skills/` |

## 插件

插件是 npm 包或本地脚本，可在运行时扩展 OpenCode。

### 作为 npm 包安装

在 OpenCode 配置文件中声明插件：

| 范围 | Files 中的配置文件 |
|:------|:-----|
| 全局 | `Application/Data/opencode/.config/opencode/opencode.json` |
| 项目级别 | 项目根目录的 `opencode.json` |

示例：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": [
    "opencode-helicone-session",
    "opencode-wakatime"
  ]
}
```

OpenCode 在启动时解析并安装声明的包，并将它们缓存在 `~/.cache/opencode/node_modules/` 中。

### 作为本地文件安装

将 `.js` 或 `.ts` 文件放在插件目录中。OpenCode 在启动时自动加载它们。

| 范围 | Olares Files 中的路径 |
|:------|:-----|
| 全局插件 | `Application/Data/opencode/.config/opencode/plugins/` |
| 项目级别插件 | `Home/Code/<project>/.opencode/plugins/` |

### 热门社区插件

| 插件 | 描述 |
|:-------|:------------|
| `opencode-wakatime` | 跟踪 OpenCode 使用时间 |
| `opencode-firecrawl` | 网页爬取和搜索 |
| `oh-my-opencode` | 后台代理、LSP/AST 工具和预设代理 |
| `opencode-supermemory` | 跨会话持久记忆 |
| `opencode-pty` | 让 AI 在 PTY 中运行并与后台进程交互 |

## 了解更多

- [管理包](opencode-packages.md)
- [OpenCode 插件文档](https://opencode.ai/docs/plugins/)
- [OpenCode 官方文档](https://opencode.ai/docs)
