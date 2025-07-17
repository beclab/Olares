---
description: 设置和使用 LarePass 的密码自动填充功能，在所有设备上实现安全便捷的密码管理体验。
---
# 使用 LarePass 自动填充密码

密码自动填充让你无需手动输入凭据，既方便又安全。通过 LarePass，你可以将密码安全地存储在密码库中，并在所有设备上自动填充。
## 开始之前

确保你的设备上已安装 LarePass 移动客户端或 Chrome 扩展程序，并使用 Olares ID 登录。

:::tip 提示
要获取 LarePass 的不同下载选项，请访问 [LarePass 页面](https://olares.cn/larepass)。
:::

## 启用自动填充服务
<tabs>
<template #Android>

1. 打开 LarePass，进入**设置** > **自动填充**。
2. 打开自动填充，并选择 LarePass 作为自动填充提供程序。
3. 按提示查看并接受安全提示。

</template>
<template #iOS>

由于 iOS 系统限制，必须手动启用 LarePass 自动填充：

1. 打开 iOS 设备上的**设置**应用。
2. 使用搜索功能快速找到自动填充设置。
3. 确保自动填充服务已开启，然后激活 LarePass 作为自动填充提供程序。

</template>
<template #Chrome-扩展>

登录浏览器扩展程序时会自动启用自动填充。
</template>
</tabs>

## 保存密码

当你在应用或网站中输入凭据时，LarePass 会检测该操作并提示保存密码。

:::info 提示
在 iOS 上，密码无法自动保存。可以手动添加 Vault 项目或使用 Chrome 扩展。Vault 项目将在所有平台同步。
:::

1. 登录应用或网站。
2. 出现提示时，点击**保存**将密码存储到 LarePass。
3. 在详情页面或窗口中输入此 Vault 项目的名称，点击**保存**。

## 使用自动填充

<tabs>
<template #Android>

1. 打开尚未登录的应用或网站。
2. 点击用户名或密码字段。
3. 在弹出窗口中，点击**使用 LarePass 自动填充**。
4. 解锁 Vault 以访问保存的凭据。
5. 选择匹配的 Vault 项目自动填充登录信息。

</template>
<template #iOS>

1. 打开尚未登录的应用或网站。
2. 点击用户名或密码字段，键盘将上滑显示匹配的登录项，或显示**密码**选项。
3. 如果显示匹配的登录项，点击它进行自动填充。
4. 如果显示**密码**选项，点击它并解锁 Vault 以访问可用的 Vault 项目。
   :::info 提示
   如果其他自动填充服务（如 iCloud 钥匙串）处于激活状态，请在提供程序列表中选择 **LarePass**。
   :::
5. 选择匹配的 Vault 项目自动填充登录信息。

</template>
<template #Chrome-extension>

1. 打开尚未登录的网站。
2. 在文本字段中点击 LarePass 图标。
3. 在弹出窗口中，选择匹配的登录项进行自动填充。
4. 如果未保存该网站的凭据，选择**新建项目**添加新的 Vault 项目。

</template>
</tabs>