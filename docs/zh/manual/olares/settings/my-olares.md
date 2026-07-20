---
outline: [2, 3]
description: 了解如何利用“我的 Olares”页面管理账户、设备、安全设置及网络访问策略。 
---

# 我的 Olares

你可以通过设置中的**我的 Olares** 功能管理 Olares 账户、连接设备和安全策略。点击**设置**页面左上角的个人头像即可进入**我的 Olares**。

![My Olares](/images/zh/manual/olares/my-olares-1.12.6.png#bordered)

## 硬件

查看和管理 Olares 硬件。你可以查看**型号**、**设备状态**、**设备标识符**、 **CPU** 和 **GPU** 信息。

![硬件](/images/zh/manual/olares/my-hardware-1.12.6.png#bordered)

可执行的操作包括：

* **关机**：点击以关闭 Olares 设备。请使用 LarePass 完成后续操作。关机后 Olares 状态将在 LarePass 上显示为 `Olares 已关机`。无法远程操作，需手动开机。
* **重启**：点击以重启 Olares 设备，请使用 LarePass 完成后续操作。重启过程中 Olares 状态将在 LarePass 上显示为 `正在重启`，约 5–8 分钟后恢复为 `Olares 运行中`。
<a id="reset-ssh"></a>
* **重置 SSH 密码** <Badge type="tip" text="Olares One 专有" />

  修改默认 SSH 密码以避免非授权的远程访问，步骤如下：
   1. 在**硬件**页面点击**重置 SSH 密码**。
   2. 在弹窗中输入新的 SSH 密码（需满足强度要求），点击**重置**。
   3. 使用 LarePass 应用扫描页面上的二维码。
   4. 在 LarePass 中点击**确认**以完成重置。

* **工作模式**<Badge type="tip" text="Olares One 专有" />

  切换 Olares One 的性能档位，支持以下两种模式：
  - **静音模式**：限制 CPU/GPU 功耗，满足日常负载并保持安静。
  - **性能模式**：释放 CPU/GPU 最大性能，适合 AI 推理、游戏等高负载场景。

* **限制 CPU 频率** <Badge type="tip" text="Olares One 专有" /> <br>
开启后，将 CPU 频率上限从 5.4 GHz 降至 5.0 GHz。关闭后恢复原有最高频率。

* **自动开机** <Badge type="tip" text="Olares One 专有" />：开启后，设备在接通电源或停电后恢复供电时会自动开机。

  :::info
  需要 Olares OS 1.12.6 或更高版本，以及 EC 固件 1.03 或更高版本。如果任一条件未满足，开关会显示但无法操作。
  
  如需查看当前 EC 固件版本或升级 EC 固件，可参考[管理 BIOS 和 EC](/zh/one/update-firmware.md)。
  :::
  

## Olares Space
在 Olares Space 中查看你的订阅计划详情和使用情况，包括反向代理方案、
备份、流量消耗等。首次使用时需按提示登录 Olares Space。

## 更改密码

更新 Olares 登录密码以增强账户安全性。

## 设置网络访问策略

定义 Olares 系统级身份验证策略，控制用户如何登录到你的 Olares。
  * **双因素**（推荐）：需要你的登录密码和 LarePass 生成的两步验证码，安全性更高。
  * **单因素**：仅需要你的登录密码。

## Olares OS 版本

查看当前 Olares 系统的版本。如果提示有可用的新版本，可转至 LarePass 手机客户端**设置** > **Olares 管理**页面下完成系统升级。详细步骤请参考[升级 Olares](../../larepass/manage-olares.md#升级-olares)。

## 沟通与反馈

获取帮助资源，或向 Olares 团队提交反馈，改进系统使用体验。

## 致谢

查看 Olares 使用的开源组件许可证及相关致谢信息。

## 设备

查看已授权访问 Olares 的设备。每个设备条目会显示设备名称和客户端类型。点击某台设备可查看更多详情或管理其访问权限。

## 退出登录

点击**退出登录**可退出当前 Olares 会话。退出后，如需从该设备继续访问 Olares，需要重新登录。
