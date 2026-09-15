---
outline: [2, 3]
description: 了解如何在重新安装 Olares 后，使用现有 Olares ID 通过 LarePass 手机端重新激活。
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, 重新激活, 重装, Olares ID
---

# 重新激活 Olares

如果你重新安装了 Olares，原来的实例将无法使用。你可以使用现有的 Olares ID 重新激活新安装的实例，无需创建新的 ID。

:::warning 需要同一网络
重新激活时，请确保手机和 Olares 设备处于同一网络，以便 LarePass 能够发现设备。
:::

根据你重新安装 Olares 的方式，选择对应的激活方法：

<Tabs>
<template #通过脚本重新安装>

如果你使用一键脚本重新安装了 Olares，并且已经完成了初始配置，可按照以下步骤使用现有的 Olares ID 重新激活：

![激活](/images/manual/larepass/activate-olares1.png#bordered)

1. 打开 LarePass。
2. 点击**扫描二维码**，扫描安装向导中的二维码。
3. 按照 LarePass 指引重置 Olares 登录密码。

激活成功后，LarePass 将返回主页，安装向导将跳转至登录页。
</template>
<template #通过-ISO-或-Docker-重新安装>

如果你使用 ISO 文件或 Docker 镜像重新安装了 Olares，可按照以下步骤使用现有的 Olares ID 重新激活：

1. 打开手机上的 LarePass 应用，会出现错误提示“未发现运行中的 Olares”。

    ![未发现运行中的 Olares](/images/manual/larepass/no-active-olares-found.png#bordered)

2. 点击错误提示旁边的**了解更多**。
3. 选择**重新激活**。

    ![重新激活 Olares](/images/manual/larepass/reactivate-olares.png#bordered)

4. 在账号激活页面，点击**发现附近的 Olares**。LarePass 将列出同一网络中检测到的 Olares 实例。
5. 从列表中选择目标 Olares 实例，并点击**立即安装**。
6. 安装完成后，点击**立即激活**。
7. 在**选择反向代理**对话框中，选择一个地理位置离你较近的节点并点击**确认**。安装程序会自动为 Olares 配置 HTTPS 证书和 DNS。

    :::tip 提示
    - 你可以稍后在 Olares 中的[更改反向代理](../olares/settings/change-frp.md)页面调整此设置。
    - 如果你的 Olares 设备连接的是公网 IP 网络，此步骤会自动跳过。
    :::

8. 按照屏幕上的说明重置 Olares 的登录密码，然后点击**完成**。

    ![重置密码](/images/manual/larepass/docker-reset-password.png#bordered)

激活完成后，LarePass 会显示你的 Olares 设备的桌面地址，如 https://desktop.marvin123.olares.com。
</template>
</Tabs>
