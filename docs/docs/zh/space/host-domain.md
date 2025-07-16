---
outline: [2, 3]
description: 介绍在 Olares Space 中添加自定义域名的步骤，包含域名验证方法、DNS 解析配置、组织 ID 创建和邮箱关联流程。
---

# 配置自定义域名

无论是企业用户希望员工使用公司特定域名登录，还是个人想使用自己的域名，Olares Space 都支持设置自定义域名以便访问。

本指南将帮助你在 Olares Space 上为 Olares 系统添加自己的域名。

## 前提条件

::: info 注意
只有账户处于 DID 阶段时才能绑定新域名。如果账户已经绑定了 Olares ID，说明该账户已经关联了一个域名。
:::

在创建和配置自定义域名之前，请确保：

- **DID 账户状态**：确认账户处于 DID 状态（即尚未绑定 Olares ID）。

- **域名**：通过域名注册商注册了域名。该域名不能已经绑定在 Olares Space 的其他账户上。

- **LarePass 移动端**：安装了 LarePass 移动端，因为可验证凭证和域名管理任务需要使用此应用。

- **域名 DNS 设置权限**：可以配置 TXT 记录和 NS 记录。

## 添加域名

准备就绪后，按照以下步骤在 Olares Space 中添加域名。

1. 在 Olares Space 控制台中，进入**域名** > **域名设置**，按照提示输入你的自定义域名。

   ![alt text](/images/how-to/space/submit_a_domain.jpg#bordered)

2. 为域名添加 TXT 记录以验证域名所有权。系统会验证你的配置。验证通过后，域名设置状态会自动更新为**等待配置域名 NS 记录**。

   ![alt text](/images/how-to/space/txt.jpg#bordered)

3. 添加 NS 记录，允许 Olares Space 为你的域名配置 DNS。

   ![alt text](/images/how-to/space/ns.jpg#bordered)

   系统会验证你的配置。验证通过后，域名状态会更新为**等待申请域名可验证凭证**。

   ![alt text](/images/how-to/space/awaiting_domain.jpg)

4. 打开 LarePass 移动端，进入**组织 Olares ID** > **创建组织**。你应该能看到你的域名。

5. 点击域名将其存储在区块链上。完成后，Olares Space 上的域名设置状态应变为**等待配置规则**。

至此，你已经成功将域名与 DID 关联。现在可以继续[设置邮件邀请规则](manage-domain.md#设置邮箱邀请规则)并使用该域名创建组织 Olares ID。

## 创建组织 Olares ID

现在你的组织已有经过验证的域名，你或被邀请的其他成员可以使用这个域名创建 Olares ID。

![org-olares-id](/images/how-to/larepass/organization_olares_id.png)

1. 在 LarePass 移动端，进入**组织 Olares ID** > **加入已有组织**
2. 输入你的组织域名，点击**继续**。如果出现错误，请检查域名是否已验证和配置
3. 通过邮箱账号绑定 VC。目前仅支持 Gmail 和 Google Workspace 邮箱

完成后，你将获得组织 Olares ID。现在可以继续[激活 Olares](/manual/get-started/activate-olares)。

## 域名状态说明

提交域名后，需要完成几个步骤来验证输入的域名。

下表说明了不同的域名状态及相应需要执行的操作：

| 状态          | 需执行的操作                 |
|-------------|------------------------|
| 等待配置 TXT 记录 | 添加 TXT 记录              |
| 等待配置 NS 记录  | 添加 NS 记录               |
| 等待申请域名可验证凭证 | 在移动端完成区块链域名申请          |
| 等待提交域名可验证呈现 | 在移动端完成区块链域名申请          |
| 等待配置规则      | 为组织成员设置邮件邀请规则          |
| 绑定中         | 等待与 Olares 绑定，可以访问详情页面 |
| 已分配         | 已绑定 Olares，可以访问详情页面    |