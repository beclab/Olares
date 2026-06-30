---
outline: [2, 3]
description: 了解如何通过 Control Hub、SSH 或直接物理访问来访问 Olares One 主机终端以进行命令行操作。
head:
  - - meta
    - name: keywords
      content: Olares One, 终端, Olares CLI, SSH 访问, 直接物理访问
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/access-overview.md)为准。
:::

# 访问 Olares One 终端

连接到 Olares One 终端是执行系统任务（如集群设置、系统配置和维护）所必需的。

根据你的当前任务、网络环境和可用工具选择最适合的方法：
- [从 Control Hub 访问](access-terminal-control-hub.md)：使用 Control Hub 应用内置的 Olares CLI 直接获得 `root` 访问权限。推荐用于快速或偶尔的任务。
- [通过网络访问（SSH）](access-terminal-ssh.md)：使用标准的安全外壳（SSH）协议从本地网络或远程网络连接。
- [直接在设备上访问](access-physical-console.md)：使用连接的显示器和键盘在设备本身上登录。当 SSH 访问不可用时，此方法很有用。
