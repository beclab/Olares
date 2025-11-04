---
outline: [2, 3]
description: 将 Olares 与第三方服务连接，扩展系统功能。集中管理 Cookie，用于订阅、网页访问及自动化任务。
---

# 在设置中管理集成

**设置**中的**集成**页面集中展示所有已连接至 Olares 的服务与访问认证凭证（Cookies）。目前可进行以下两类操作：

- **集成第三方服务**：连接 Olares Space、 AWS S3、腾讯云 COS 等云存储服务，扩展 Olares 的文件存储能力；
- **管理 Cookie**：集中保存、导入或删除网站 Cookie，以支持 Olares在 Olares 中访问订阅服务或其他受限内容。

## 查看与管理服务

通过以下步骤查看或管理集成的服务：

1. 通过 Dock 或启动台打开**设置**。
2. 在左侧菜单选择**集成** > **关联您的账户与数据** ，即可看到已授权的服务列表。
3. 点击任一集成卡片查看连接状态和操作选项。
4. 在**账户设置**页面点击**删除**可移除该集成。

## 通过 API 密钥添加云对象存储

Olares 支持通过 API 密钥手动配置 **AWS S3** 和**腾讯云 COS**：

1. 进入**设置** > **集成** > **关联您的账户与数据**，点击右上角**添加账户**。
2. 选择 **AWS S3** 或 **Tencent COS**，点击**确认**。
3. 在弹出的挂载对话框输入以下信息：
   - **Access Key**
   - **Secret Key**
   - **Region**
   - **Bucket name**
4. 点击**下一步**。凭证验证通过后将显示成功提示。

配置完成后，挂载的存储服务将出现在**文件管理器** > **云存储**下。

你也可以在 [LarePass](../../larepass/integrations.md#通过-api-密钥添加云盘) 中完成同样的配置。

:::tip 需要在 LarePass 应用操作的集成
OAuth 类型的集成以及 Olares Space 需在**LarePass**应用中完成连接，详见 [LarePass 集成文档](../../larepass/integrations.md)。
:::

## 管理 Cookie 

你可以在 **设置** > **集成** > **Cookie 管理** 中打开 Cookie 管理页面。

Cookie 管理页会按网站域名分组展示所有已保存的 Cookie。每个条目显示 Cookie 名称、值、到期时间及所属域名。

![Cookie 管理](/images/zh/manual/olares/cookie-management.png#bordered)

可执行的操作包括：

- **导入 Cookie**：目前支持手工导入，复制并粘贴 Cookie 内容至导入对话框。支持格式为 **Netscape**、**Json**、以及 **Head String**。
   :::tip 注意
   第一次使用仅支持导入。你也可以在 [LarePass 插件](../../larepass/manage-knowledge.md#通过-larepass-扩展收集内容)中直接从浏览器导入当前页面 Cookie。
   :::
- **编辑 Cookie**：点击具体的 Cookie 条目，可在下方输入框中编辑对应的值。
- **删除 Cookie**：移除无效或过期条目。支持对单个域名或全部 Cookie 进行删除。

:::tip 定期维护 Cookie
当 Cookie 过期时，系统会以红色标识提醒。若 Cookie 已过期或缺失，部分订阅或网页访问任务可能失败。建议定期检查 Cookie 有效期并及时更新。
:::


