---
outline: [2, 3]
description: 将 Olares Space 与第三方服务连接，扩展系统功能。了解如何集成、授权并管理已连接的服务，实现数据的无缝同步。
---

# 在 LarePass 中管理集成

**LarePass** 是连接 Olares 与外部服务的集中枢纽，支持接入 **Olares Space**，以及 Google Drive、Dropbox、AWS S3、腾讯 COS 等三方服务。通过这些集成，你可以实现文件同步、安全备份等功能，进一步扩展 Olares 的能力。

:::info
我们正持续支持更多第三方集成，敬请关注！
:::

## 连接 Olares Space

**Olares Space** 是 Olares 的云托管服务，使用与 LarePass、Olares 相同的账户体系。

### 步骤 1：登录 Olares Space

1. 在浏览器访问 <https://space.olares.com/login>  
2. 打开手机上的 LarePass。  
3. 在**设置**页面点击右上角 **扫码**图标。  
4. 扫描 Olares Space 登录页的二维码。  
5. 确认风险提示并完成登录。  

### 步骤 2：授权 Olares Space

1. 于 LarePass 中进入**设置** > **集成**。  
2. 点击右上角 <i class="material-symbols-outlined">add</i>，选择 **Space** 以添加 Olares Space 账户。  

### 步骤 3：关联 Olares ID

关联 Olares ID 后，可导入区块链钱包，在个人资料中使用 NFT 头像。

1. 打开 Dock / 启动台中的**设置**应用。  
2. 选择左侧**集成**。  
3. 在右侧点击 Olares Space 卡片查看详情。  
4. 点击**绑定**，LarePass 会弹出确认提示。  
5. 在手机端打开 LarePass：  
   - 若已弹出提示，点击**确认**；  
   - 如未出现提示，可手动进入**设置** > **集成**，再点击 Olares Space 卡片并确认。  
6. 返回 Olares，点击**确认**完成关联。  

## 通过 OAuth 添加云盘

Google Drive、Dropbox 等通过 OAuth 登录的服务集成需在 LarePass 移动端完成授权：

1. 在手机上打开 LarePass。  
2. 依次点击**设置** > **集成**，再点右上角 <i class="material-symbols-outlined">add</i>。  
3. 选择 **Google Drive** 或 **Dropbox**。  
4. 按提示登录并授权。  

授权成功后，集成将显示在列表中，并可在**文件管理器** > **云存储**中访问存储。

## 通过 API 密钥添加云盘

AWS S3、腾讯云 COS 等服务需使用 Access Key & Secret Key 手动配置，可在 LarePass 手机端或 Olares **设置**中完成：

1. 在手机上打开 LarePass。  
2. 进入**设置** > **集成**，点击右上角 <i class="material-symbols-outlined">add</i>。  
3. 选择 **AWS S3** 或 **Tencent COS**。  
4. 输入 Access Key、Secret Key 及其他必要信息，点击**确认**。  

配置成功后，服务将显示在集成列表，可通过**文件管理器** > **云存储**应用访问。  
你也可以在 [Olares 设置](/zh/manual/olares/settings/integrations.md) 中直接配置。

## 断开集成

:::warning
断开 Olares Space 可能影响设备管理及云备份访问，你可在需要时重新连接。
:::

1. 打开 LarePass，进入**设置** > **集成**。  
2. 点击要移除的集成。  
3. 点击右上角 <i class="material-symbols-outlined">more_horiz</i>，选择**删除**。
