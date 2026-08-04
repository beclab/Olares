---
outline: [2, 3]
description: 了解如何使用显示器和键盘直接登录 Olares One 主机终端以进行命令行操作。
head:
  - - meta
    - name: keywords
      content: Olares One, 终端, 物理控制台, 显示器, 键盘
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/access-physical-console.md)。
:::

# 物理访问 Olares One 终端

如果网络访问或 SSH 不可用，你可以使用显示器和键盘物理登录 Olares One 设备。

## 前提条件

- Olares One 已完成设置并已开机。
- 一台连接到 Olares One 的显示器和键盘。
- 如果 Olares OS 已激活，需要一台安装了 LarePass 应用的移动设备，以便从 Vault 获取登录密码。

## 步骤 1：准备登录密码

:::info 与 Olares Desktop 密码不同
此密码用于登录 Olares One 主机系统。它与你在浏览器中登录 Olares Desktop 时使用的密码不同。
:::

默认登录密码为 `olares`。激活 Olares OS 后，系统会提示你在 LarePass 应用中重置 SSH 密码，新密码将自动生成并保存到你的 Vault 中。

根据你的激活状态确定密码。

<tabs>
<template #未激活系统>

如果你尚未激活 Olares OS，请使用默认密码 `olares`。

</template>
<template #已激活系统>

如果你已经激活了 Olares OS，请从 LarePass 移动应用中获取已保存的密码。

<!--@include: ./reusables-reset-ssh.md#view-saved-ssh-password-->

</template>
</tabs>

## 步骤 2：登录

:::info
出于安全考虑，输入时屏幕上不会显示字符。
:::

1. 在连接显示器上显示的文本登录提示中，输入用户名 `olares`，然后按 **Enter**。

    ```text
    olares login:
    ```

2. 输入步骤 1 中获取的密码，然后按 **Enter**。
