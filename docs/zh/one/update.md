---
outline: [2, 3]
description: 了解如何使用 LarePass 应用检查并安装 Olares One 的系统更新。
head:
  - - meta
    - name: keywords
      content: Olares One, 更新 Olares
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/update.md)。
:::

# 更新操作系统 <Badge type="tip" text="15 min" />

使用 LarePass 应用保持 Olares One 更新到最新功能和安全补丁。

:::warning 需要管理员权限
只有 Olares 管理员才能执行系统更新。请注意，更新适用于整个 Olares 集群，并会影响所有成员。
:::

## 学习目标

完成本教程后，你将学会如何：
- 在 LarePass 中检查是否有新的 Olares OS 版本可用。
- 在 **仅下载** 和 **下载并升级** 之间进行选择。
- 安装更新并确认系统成功重启。

## 前提条件

确保：
- Olares One 已开启并连接到网络。
- 你的手机可以通过 LarePass 访问 Olares One。
- 你使用具有管理员权限的账户登录。

## 检查更新

1. 在手机上打开 LarePass，前往 **Settings**。
2. 在 **My Olares** 卡片中，点击 **System** 进入 **Olares management** 页面。
3. 点击 **System update**。
4. 如果 **New version** 字段中显示新版本，点击 **Upgrade**。
  ![检查可用版本](/images/one/check-version1.png#bordered)

## 选择升级方式

将弹出对话框询问你如何处理更新包。选择适合你日程安排的方式：

- **Download only**：Olares 在后台下载包。你可以继续使用系统，稍后手动安装更新。
- **Download and upgrade**：Olares 下载包，下载完成后提示你重启并安装。
    ![选择升级方式](/images/one/olares-upgrade1.png#bordered)

## 安装并重启

1. 根据你的选择开始安装：
   - **Download only**：下载完成后，返回 **System update** 并点击 **Upgrade now**。确认重启提示以开始安装。
   - **Download and upgrade**：下载完成后，确认重启提示以开始安装。
2. 等待更新和重启完成。成功消息表示升级已完成。
    ![升级成功消息](/images/one/upgrade-success.png#bordered)
3. 刷新 Olares 桌面以同步最新的系统更改。
