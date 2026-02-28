---
outline: [2, 4]
description: 了解如何在 LarePass 应用中创建、导入和管理 Olares ID。
---

# 创建、管理账户

管理 Olares 账户是 LarePass 的核心功能。如果你是新用户，需要先创建一个 Olares ID。本指南将介绍创建流程及常用的账户操作。

::: tip 提示
Olares ID 仅可在 **LarePass** 移动端创建。
:::

## 创建 Olares ID

开始之前，请先在手机上安装 [LarePass](https://olares.cn/larepass)。根据个人需求，可选择以下两种方式之一：

- **快速创建**：输入符合要求且可用的名称即可创建 Olares ID（默认方式）。  
- **高级创建**：使用可验证凭证（VC）将现有可信身份（如社交账号）与 Olares ID 绑定，适用于需要更高安全性或专属域名的个人或组织用户。

### 快速创建

快速创建个人 Olares ID：

1. 在 LarePass 中点击**创建账户**。  
2. 输入想要的 Olares ID（至少 8 个字符，仅限小写字母和数字）。  
3. 点击**继续**完成创建。  

   ![快速创建](/images/manual/larepass/create-olares-id.png)

获得 Olares ID 后，等待 [安装 Olares](../get-started/install-olares.md) 完成，然后执行 [激活](../get-started/activate-olares.md)。

### 高级创建

::: tip VC 支持
Olares 目前通过 Gmail 提供 VC 支持，详情参见 [Gmail Issuer Service](/developer/contribute/olares-id/verifiable-credential/olares.md#gmail-issuer-service)。
:::

<Tabs>
<template #个人-Olares-ID>

1. 在 LarePass 中点击**创建账户**。  
2. 在创建页面右上角点击 <i class="material-symbols-outlined">display_settings</i>。  
3. 在**高级账户创建**页面选择 **个人 Olares ID**。  
   ![高级创建](/images/manual/larepass/advanced_creation.png)  
4. 选择 Gmail VC 选项，按提示完成 Gmail 身份验证后点击**继续**。  
5. 绑定完成后点击**继续**，即可查看你的 Olares ID 信息。  
   ![绑定 VC 后的 Olares ID](/images/manual/larepass/individual_olares_id_vc.png)
</template>
<template #组织-Olares-ID>

::: tip 提示
需先在 Olares Space 中 [配置自定义域名](/zh/space/host-domain.md#add-your-domain) 并在 LarePass 创建对应组织。
:::

1. 在 LarePass 中点击 **创建账户**。  
2. 在创建页面右上角点击 <i class="material-symbols-outlined">display_settings</i>。  
3. 在**高级账户创建**页面选择**组织 Olares ID** > **加入现有组织**。  
   ![高级创建（组织）](/images/manual/larepass/advanced_creation_org.png)  
4. 输入组织的域名并点击**继续**。  
5. 通过邮箱绑定 VC，目前仅支持 Gmail 与 Google Workspace 邮箱。  
   ![组织 ID VC](/images/manual/larepass/organization_olares_id.png)  

完成后，你将获得组织 Olares ID。
</template>
</Tabs>

## 导入现有账户

你可以通过导入已有的 Olares ID 来设置 LarePass。

::: tip 备份助记词
确保已 [备份助记词](back-up-mnemonics.md)，否则无法完成账户导入。
:::

### 首次导入账户

如果当前设备上尚未添加任何账户：

1. 打开 LarePass。
2. 根据提示输入 Olares ID 对应的 12 个助记词。
3. 按提示完成设置。

### 添加其他账户

如果你已经登录了一个账户，希望再添加一个账户：

<Tabs>
<template #iOS-&-Android>

1. 打开 LarePass 应用。
2. 点击你的个人头像。
3. 在**切换账户**页面底部，点击**添加新账户**。
4. 点击**导入账户**。
5. 输入 Olares ID 对应的 12 个助记词。
6. 按提示完成设置。

</template>
<template #macOS-&-Windows>

1. 打开 LarePass 桌面客户端。
2. 点击你的个人头像。
3. 点击**切换账户**。
4. 点击底部的**添加新账户**。
5. 输入 Olares ID 对应的 12 个助记词。
6. 按提示完成设置。

</template>
<template #Chrome-extension>

1. 在 Chrome 中打开 LarePass 扩展。
2. 点击个人头像上方的选项图标。
3. 点击**添加新账户**。
4. 输入 Olares ID 对应的 12 个助记词。
5. 按提示完成设置。

</template>
</Tabs>