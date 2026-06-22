---
outline: [2, 3]
description: 了解如何在 Olares One 上添加用户、分配角色和资源限制，以及管理现有账户。
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../one/create-users.md)为准。
:::

# 添加和管理用户 <Badge text="5 min"/>

在 Olares One 上，你可以创建多个用户账户来安全地共享设备。每个用户都有自己的空间、应用和资源限制。

## 开始之前

根据以下权限确定要分配的角色：

| **权限** | **Super admin** | **Admin** | **Members** |
|---|---|---|---|
| 创建用户 | Admin 和 Members | Members | — |
| 移除用户 | Admin 和 Members | Members | — |
| 资源访问 | 使用所有资源 | 使用分配的资源 | 使用分配的资源 |

## 前提条件

**硬件**
- 你的 Olares One 有足够的可用 CPU 和内存资源。

**用户权限**
- 你以 **Super admin** 或 **Admin** 身份登录。

**Olares ID**
- 新用户拥有有效的 Olares ID，且未在其他 Olares 设备上激活。
- 新用户 Olares ID 的域名部分与当前域名一致。

## 添加用户

1. 前往 **Settings** > **Users**。
2. 点击 **Create account**。
3. 在对话框中填写所需信息：

   - **Olares ID**：仅输入用户名（`@` 前面的部分）。
   - **Role**：选择 **Members** 或 **Admin**。
   - **CPU**：分配 CPU 核心数。最少 1 核。
   - **Memory**：分配内存。最少 3 GB。

4. 点击 **Save**。

   账户创建后，系统会生成一个临时的激活 wizard URL 和一次性密码。

5. 复制并分享激活凭据给用户。

:::tip 远程激活
被邀请的用户可以远程激活其访问权限。完整步骤请参阅[加入 Olares](../../manual/get-started/join-olares)。
:::

6. 要查看用户的激活状态，请前往 **Users** 页面。
   ![查看用户列表](/images/one/settings-create-users.png#bordered){width=85%}

## 管理现有用户

用户创建后，你可以查看账户详情、调整资源限制、重置密码或移除用户。

1. 前往 **Settings** > **Users**。
2. 选择一个用户打开 **Account info** 页面。
3. 要调整资源限制，点击 **Modify limits**。更新 CPU 或内存数值，然后点击 **OK**。
4. 要重置密码，点击 **Reset password**，然后将生成的密码分享给用户。

   Super admin 可以重置 Admin 和 Members 的密码。Admin 可以重置 Members 的密码。
5. 要移除用户，点击 **Delete user**，然后点击 **OK** 确认。

![管理用户](/images/one/settings-manage-user.png#bordered){width=90%}

## 资源

- [角色和权限](../../manual/olares/settings/roles-permissions.md)：了解更多关于 Olares 中的角色及对应权限。
