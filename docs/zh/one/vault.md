---
outline: [2, 3]
description: 了解 Olares 中 Vault 的基础知识。学习如何设置保险库和管理保险库条目。
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, Vault, 密码管理器, 存储凭据, 安全敏感数据
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../one/vault.md)为准。
:::

# 使用 Vault 安全存储密码 <Badge type="tip" text="15 min" />

Vault 是 Olares 中专用的密码和敏感数据管理器。使用它可以安全地存储密码、密钥、数字身份和敏感文档。

本指南涵盖使用 Vault 的基础知识，包括设置你的第一个保险库、导入凭据以及高效管理敏感数据。

## 学习目标

- 首次初始化你的个人 Vault。
- 创建并保存一个安全条目。本指南以 Wi-Fi 密码为例。
- 编辑并整理你的凭据。
- 使用搜索和筛选定位条目。

## 开始之前

熟悉相关的 Vault 概念以及 Vault 如何组织你的数据。

### 保险库类型

Olares Vault 提供两种主要的保险库类型：

* My vault：这是你的私人保险库，在账户激活时自动创建。该保险库使用你的助记词加密，仅对你可见。
* Team vaults：这些是协作保险库，用于与团队成员或家人安全共享凭据。

### 保险库条目

保险库条目是特定信息的加密容器。虽然通常用于登录凭据，但保险库条目也可以存储信用卡、安全笔记、护照和网络详情。

## 设置 Vault

首次在设备上打开 Vault 时，必须出于安全目的进行初始化。

1. 从 Launchpad 打开 Vault 应用。

   ![查找 Vault](/images/one/find-vault-app.png#bordered)

2. 为 Vault 设置本地密码，该密码仅用于在当前设备上解锁 Vault。此本地 Vault 密码作为第二层防御。

   ![首次打开 Vault](/images/one/open-vault1.png#bordered){width=35%}

   :::tip 安全最佳实践
   不要使用与 Olares 登录相同的密码。如果有人猜到了你的登录密码，这个二级密码可以确保你的敏感数据仍然处于锁定状态。
   :::

3. 点击 **Confirm**。

4. 使用你的助记词导入已链接到服务器的 Olares ID。

   ![输入助记词](/images/one/vault-enter-mnemonic-phrase.png#bordered){width=35%}

   :::tip
   关于如何获取助记词的信息，请参阅[显示并备份你的助记词](../../manual/larepass/back-up-mnemonics.md#reveal-and-back-up-your-mnemonic-phrase)。
   :::

5. 点击 **Next**。

## 添加保险库条目

创建一个新的保险库条目来存储"我的公司 Wi-Fi 密码"。

1. 从 Dock 或 Launchpad 打开 Vault 应用。
2. 在 **All vaults** 面板中点击 <i class="material-symbols-outlined">add</i>。
3. 从 **Select Vault** 列表中，选择 **My Vault** 用于私人用途。

   ![选择 My Vault](/images/one/select-my-vault.png#bordered){width=50%}

4. 从类别中选择 **WIFI Password**，然后点击 **Confirm**。

   ![添加保险库条目 Wi-Fi](/images/one/select-wifi-vault.png#bordered){width=50%}

   右侧将打开详情面板。它显示与你所选类别相关的默认字段。

   ![配置 Wi-Fi 保险库条目设置](/images/one/new-vault-item-wifi.png#bordered){width=50%}

5. 在字段中填写详情。

   ![添加保险库条目 Wi-Fi](/images/one/fill-info-vault.png#bordered){width=50%}

6. 点击 **Save**。该保险库条目将被加密并添加到 **All vaults** 列表中。

## 管理保险库条目

保持你的保险库条目有序且最新。

### 编辑保险库条目

修改现有的保险库条目以更新详情。

1. 打开 Vault 应用。
2. 在 **All vaults** 面板中，点击目标保险库条目。
3. 在右侧的详情面板中，点击 <i class="material-symbols-outlined">edit_note</i>。

   ![编辑 Wi-Fi 保险库条目](/images/one/edit-vault-item.png#bordered){width=50%}

4. 根据需要进行更改。例如，移除过期日期。
5. 点击 **Save**。

### 标记收藏保险库条目

标记常用的保险库条目以便快速访问。

1. 打开 Vault 应用。
2. 在 **All vaults** 面板中，点击目标保险库条目。
3. 点击右上角的 <i class="material-symbols-outlined">star_border</i>。

   ![标记收藏](/images/one/mark-favourite.png#bordered){width=50%}

   该条目将在 **All vaults** 面板中以星标标记。

   ![收藏的保险库条目](/images/one/favourite-vault-item.png#bordered){width=70%}

## 查找保险库条目

使用关键词搜索或筛选快速定位保险库条目。

### 搜索

在 **All vaults** 面板中，点击 <i class="material-symbols-outlined">search</i> 并输入关键词（例如 `Wi-Fi` 或 `company`）以查找特定的保险库条目。

![搜索保险库条目](/images/one/search-vault.png#bordered){width=70%}

### 筛选

使用左侧边栏缩小列表范围：

- 保险库类别：在 My Vault 和 Team Vaults 之间切换以更改范围。
- 标签：点击标签名称查看所有相关的保险库条目。
- 收藏：点击 **Favorites** 查看所有带星标的条目。
- 最近使用：点击 **Recently used** 查看你的访问历史。
- 附件：点击 **Attachments** 查看包含文件的条目。

## 资源

- [导入保险库条目](../../manual/olares/vault/vault-items.md#import)
- [管理共享保险库](../../manual/olares/vault/share-vault-items.md)
- [使用 LarePass 自动填充密码](../../manual/larepass/autofill.md)
- [生成双因素认证码](../../manual/larepass/two-factor-verification.md)

<!--<template #LarePass-desktop-or-mobile>

1. 打开 LarePass 桌面端或移动端，然后前往 **Vault** 标签页。
2. 点击右上角的 <i class="material-symbols-outlined">add</i>。
3. 从 **Select Vault** 列表中，选择 **My Vault** 用于私人用途，或选择 **Team Vault** 用于共享。
4. 选择保险库条目的类别，然后点击 **Confirm**。
5. 指定保险库条目的详细设置，然后点击 **Save**。
</template>

<template #LarePass-browser-extension>

:::info 开始之前
- LarePass 浏览器扩展目前仅适用于 Google Chrome。
- 从 Chrome Web Store 或[官方页面](https://www.olares.com/larepass)安装。
- 为快速访问，将扩展固定到浏览器工具栏。
:::

1. 点击浏览器工具栏上的 LarePass 图标以打开侧边栏。
2. 点击 **Vault**。
3. 在 **All vaults** 面板中点击 <i class="material-symbols-outlined">add</i>。
4. 选择保险库条目的类别，然后点击 **Confirm**。
5. 指定保险库条目的详细设置，然后点击 **Save**。
</template>
</tabs>
![添加保险库条目](/images/one/new-vault-item.png#bordered){width=50%}
![配置保险库条目设置](/images/one/configure-vault-item.png#bordered){width=50%}-->
