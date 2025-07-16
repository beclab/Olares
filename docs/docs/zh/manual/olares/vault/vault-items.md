---
outline: [2, 3]
description: 了解 Olares 系统应用 Vault 的基本操作。学习创建和管理 Vault 项目，使用标签组织内容，通过快速筛选功能高效管理敏感数据。
---

# Vault 基本操作

本文档将带你了解 Vault 的基础使用方法，从设置第一个 Vault 到高效地组织你的敏感数据。

## 了解 Vault 组件

### Vault 类型

Olares 为用户提供两种主要类型的 Vault：

* 主 Vault（**我的 Vault**）：账户激活时自动创建，作为用户的私人 Vault。使用用户助记词加密以确保最大安全性。
* 共享 Vault（**Team Vault**）：Olares 内的协作 Vault，支持团队成员或家庭之间安全共享信息。

### Vault 项目

Vault 项目是存储敏感信息的独立安全容器。每个 Vault 项目包含以下字段：

* **名称**：便于识别的标题。
* **标签**：用于组织和快速筛选。
* **字段**：存储不同类型信息的数据区域。
* **历史记录**：记录项目修改信息。
* **附件**：添加相关文件。
* **过期时间**：为时间敏感的信息设置过期日期。

目前，Vault 支持以下项目类型：
- 网站/应用
- 计算机
- 信用卡
- 银行账户
- Wi-Fi 密码
- 护照
- 验证器
- 文档
- 自定义

### 字段

字段是 Vault 项目的核心组件，支持存储多种数据类型，包括：

* 用户名
* 密码
* 助记词
* 电子邮件地址
* URL
* 日期和月份
* 信用卡号
* 电话号码
* PIN
* 明文
* 一次性密码 (OTP)

## 使用密码保护 Vault

首次在 Olares 中使用 Vault 时，系统会提示你设置本地密码。此密码不应与你的 Olares 登录密码相同。

1. 为 Vault 设置本地密码。
2. 使用助记词短语导入已与 Olares 服务器关联的 Olares ID。

![Vault password](/images/manual/olares/vault-local-password.png)

:::tip 提示
如果你不知道助记词短语的位置，请参阅[备份助记词短语](../../larepass/back-up-mnemonics)。
:::

## 管理 Vault 项目

:::tip 提示
从一开始就使用描述性名称和相关标签来组织 Vault 项目。随着 Vault 项目数量增加，这种做法会越来越有价值。
:::

### 添加

要添加 Vault 项目：
<tabs>
<template #Olares>

1. 打开 **Vault**，在右上角点击 <i class="material-symbols-outlined">add</i>。
2. 选择一种类型（例如**网站/应用**），点击**创建**。

   ![Add vault item](/images/manual/olares/add-vault-item.png#bordered)

3. 填写必填字段，例如项目名称、用户名、密码和 URL。

   ![Fill item fields](/images/manual/olares/fill-item-fields.png#bordered)

4. 点击**保存**创建新的 Vault 项目。

</template>

<template #LarePass-桌面端、移动端>

1. 在设备上打开 **LarePass**，导航到应用中的 **Vault** 页面。
2. 点击右上角的 <i class="material-symbols-outlined">add</i>。
3. 选择一种类型（例如**网站/应用**），点击**创建**。
4. 填写必填字段，例如项目名称、用户名、密码和 URL。
5. 点击**保存**创建新的 Vault 项目。

</template>

<template #LarePass-浏览器扩展>

:::info 提示
LarePass 浏览器扩展目前仅支持 Google Chrome。请访问 [LarePass 页面](https://olares.cn/larepass) 下载扩展。
:::
:::tip 提示
为方便访问，你可以将扩展固定到工具栏。
:::

1. 点击浏览器工具栏或扩展菜单中的 LarePass 图标，在浏览器窗口右侧打开 LarePass。
2. 导航到扩展中的 **Vault** 页面。
3. 点击右上角的 <i class="material-symbols-outlined">add</i>。
4. 选择一种类型（例如**网站/应用**），点击**创建**。
5. 填写必填字段，例如项目名称、用户名、密码和 URL。URL 字段会自动填充为当前网页地址。
6. 点击**保存**创建新的 Vault 项目。

</template>
</tabs>

### 编辑

:::info 提示
LarePass 浏览器扩展不支持编辑 Vault 项目。如需完整编辑功能，请使用 Olares 的 Vault 应用或 LarePass 的移动或桌面版本。
:::

在编辑模式下，你可以：
- 更新必填字段。
- 为项目添加标签以便组织和筛选。
- 设置过期时间。
- 添加文件附件（每个文件不得超过 1 MB）。
- 查看并恢复历史记录。Vault 为每个项目保留多达 10 条记录，超过限制时会删除较旧的记录。

**编辑 Vault 项目**：

1. 在 Vault 中选择需要编辑的项目。
2. 在项目的详情窗口或页面中，点击右上角的 <i class="material-symbols-outlined">edit_note</i> 进入编辑模式。
3. 对项目详情进行必要的修改。
4. 点击**保存**。

### 收藏 Vault 项目

重要项目可以标记为收藏，便于快速访问。

<tabs>
<template #Olares>

1. 在 Vault 中，点击项目，右侧打开其详情窗口。
2. 点击右上角的 <i class="material-symbols-outlined">star_border</i> 将该项目标记为收藏。

</template>

<template #LarePass-桌面端、移动端>

1. 在设备上打开 LarePass，导航到应用中的 **Vault** 页面。
2. 点击项目以进入其详情页面。
3. 点击右上角的 <i class="material-symbols-outlined">star_border</i> 将该项目标记为收藏。

</template>
</tabs>

## 筛选 Vault 项目

你可以使用快速筛选或搜索框找到所需的 Vault 项目。

### 快速筛选

* 按主 Vault 和共享 Vault：选择**我的 Vault** 或**团队 Vault** 快速找到项目。
* 按标签：点击标签名称轻松定位已标记的 Vault 项目。
* 按收藏：点击**收藏**列出所有收藏项目。
* 按最近使用：点击**最近使用**显示你的最近项目。
* 按附件：点击**附件**显示所有带有附件的项目。

### 关键词搜索

点击 <i class="material-symbols-outlined">search</i> 使用关键词直接搜索目标项目。