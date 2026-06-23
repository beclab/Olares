---
outline: [2, 3]
description: 了解如何通过 Control Hub 内置的 Olares CLI 访问 Olares One 主机终端以进行命令行操作。
head:
  - - meta
    - name: keywords
      content: Olares One, 终端, Olares CLI, Control Hub
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../one/access-terminal-control-hub.md)为准。
:::

# 通过 Control Hub 访问 Olares One 终端

如果你需要运行快速、偶尔的命令，请使用 Olares CLI。这个基于 Web 的终端内置于 Control Hub 应用中。它默认以 `root` 身份运行，无需 SSH 客户端或 IP 地址配置。

## 前提条件

- Olares One 已完成设置并正在运行。
- 你可以在浏览器中打开 Olares desktop。

## 通过 Olares CLI 访问

1. 打开浏览器并访问你的 Olares desktop：

    ```text
    https://desktop.<username>.olares.com
    ```
    
2. 从 Launchpad 打开 Control Hub。
3. 在左侧边栏的 **Terminal** 部分下，点击 **Olares**。

    ![打开终端](/images/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

## 了解更多

- [Olares CLI](../developer/install/cli/olares-cli.md)
- [使用 Control Hub 管理 Olares](../manual/olares/controlhub/index.md)
