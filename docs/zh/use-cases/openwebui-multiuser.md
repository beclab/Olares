---
outline: deep
description: 通过配置公开访问、添加用户并共享本地模型，在 Olares 设备上与其他用户共享 Open WebUI。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, 多用户, 共享访问, 本地 LLM
app_version: "1.0.38"
doc_version: "1.1"
doc_updated: "2026-07-30"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/openwebui-multiuser.md)。
:::

# 在 Open WebUI 中设置多用户访问

在 Open WebUI 中配置多用户账号，以共享本地模型和 AI 资源。此项配置允许多人使用各自独立的账号，访问 Olares 设备上的同一 Open WebUI 实例。

## 学习目标

在本指南中，你将学习如何：

- 将 Open WebUI 入口改为公开。
- 添加额外用户账号。
- 与指定用户或所有用户共享已配置的模型。
- 验证用户是否可以访问共享模型。

## 前提条件

开始前，请确保已满足以下条件：

- 已安装并配置 [Open WebUI](openwebui.md)，且至少连接了一个本地模型。
- 拥有 Open WebUI 实例的管理员权限。

## 将入口改为公开

默认情况下，Open WebUI 只允许 Olares 所有者访问。要允许其他用户访问，请将应用入口类型改为公开。

1. 打开 Olares 设置，然后前往**应用** > **Open WebUI** > **入口** > **Open WebUI**。
2. 将**认证级别**从**私有**改为**公开**，然后点击**提交**。

   ![Entrance public](/images/manual/use-cases/openwebui-entrance-public.png#bordered){width=70%}

:::warning 安全提示
将入口设置为**公开**会把 Open WebUI 直接暴露到互联网。此时 Open WebUI 的账号系统会成为主要保护措施。需确保管理员账号和所有子用户账号都使用强密码。
:::

## 添加用户

为共享 Olares 设备的成员创建专属账号。

1. 在 Open WebUI 中，点击你的头像图标，选择 **Admin Panel**。
2. 在 **Users** 标签页中，点击右上角的 **Add User**。
3. 在 **Add User** 窗口中，选择角色并填写用户信息。

   ![Add user in Open WebUI](/images/manual/use-cases/openwebui-add-user1.png#bordered)

4. 点击 **Save**。新用户现在可以使用你指定的凭据登录。

## 共享本地模型

管理员添加的模型默认为私有。在你明确共享之前，其他用户不可见且不可用。

1. 在 Open WebUI 中，点击你的头像图标，选择 **Admin Panel**。
2. 选择 **Settings** 标签页，然后从左侧边栏选择 **Models**。
3. 点击要共享的模型右侧的 <span class="material-symbols-outlined">edit</span>。
4. 点击右上角的 **Access**。
5. 选择以下访问控制选项之一：

   - **Public**：所有已登录用户都可以访问该模型。
   - **Private**：只有指定用户可以使用该模型。
   - 点击 **Add Access** 将用户添加到访问列表。

   ![Model access list](/images/manual/use-cases/openwebui-model-access-list1.png#bordered)

6. 关闭 **Access Control** 窗口。
7. 向下滚动并点击 **Save & Update**。

## 验证访问

为确保配置生效，需测试新用户的使用体验。

1. 使用刚创建的用户账号登录 Open WebUI。
2. 在聊天区域，查看模型下拉列表。你应该能看到共享模型。

   ![User model dropdown](/images/manual/use-cases/openwebui-subuser-model-dropdown1.png#bordered)
