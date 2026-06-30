---
outline: [2, 3]
description: 在 Olares 上运行 Karakeep，将链接、笔记、图片和 PDF 保存在一个自托管工作空间中。通过手机访问你的收藏，并使用本地模型添加 AI 自动标签。
head:
  - - meta
    - name: keywords
      content: Olares, Karakeep, Hoarder, bookmark manager, self-hosted, AI auto-tag, mobile app, Ollama, video download
app_version: "1.0.4"
doc_version: "1.0"
doc_updated: "2026-05-15"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/karakeep.md)为准。
:::

# 使用 Karakeep 保存和整理书签

Karakeep（前身为 Hoarder）是一个自托管的书签和内容管理应用，将链接、笔记、图片和 PDF 集中存储在一处。它自动获取页面元数据，为全文搜索索引内容，支持共享列表，并可以使用本地 AI 模型自动为条目添加标签。

在 Olares 上运行 Karakeep 可将你保存的内容和 AI 处理完全保留在你的设备上。

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Karakeep。
- 从 Web 界面保存书签。
- 从手机登录 Karakeep。
- 使用本地模型启用 AI 自动标签。
- 为书签链接启用视频下载。
- 配置 SMTP 以发送邮件邀请。

## 安装 Karakeep

1. 打开 Market 并搜索 "Karakeep"。

   ![Karakeep in Market](/images/manual/use-cases/karakeep.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

安装完成后，从 Launchpad 打开 Karakeep 并创建第一个账户。第一个注册的用户将成为 Karakeep 管理员。

## 保存你的第一个书签

Karakeep 会在你粘贴 URL 时自动获取页面标题、内容和元数据。

1. 从 Launchpad 打开 Karakeep。
2. 将任意 URL 粘贴到仪表板顶部的输入框中，然后点击 **保存**。Karakeep 会保存链接并在后台开始获取页面。

   ![在 Karakeep 中保存 URL](/images/manual/use-cases/karakeep-save-url.png#bordered)

有关完整功能集，包括笔记、列表和协作，请参阅 [Karakeep 文档](https://docs.karakeep.app/)。

## 从移动设备访问 Karakeep

Karakeep 移动应用允许你直接从手机将图片和链接保存到你的 Olares 托管实例。要连接，你需要一个可访问的端点和一个 API 密钥。

### 安装 Karakeep 移动应用

从 App Store（iOS）或 Google Play（Android）安装 Karakeep 应用。你可以在 Market 中的 Karakeep 页面找到直接下载链接。

![Karakeep 移动应用下载](/images/manual/use-cases/karakeep-mobile-stores.png#bordered)

### 允许移动应用访问

默认情况下，Karakeep 仅对你的 Olares 账户私有。要让移动应用登录，请更改认证级别。

1. 打开 **设置**，然后前往 **应用** > **Karakeep** > **入口** > **Karakeep**。
2. 将 **认证级别** 设置为以下之一，然后点击 **提交**：

   - **内部**：需要 LarePass VPN。推荐个人使用。
   - **公开**：无需 LarePass VPN 即可访问，但会将 Karakeep 暴露到互联网并使用 FRP 流量。

   ![设置 Karakeep 认证级别](/images/manual/use-cases/karakeep-auth-level.png#bordered)

:::warning 公开访问
使用强账户密码并妥善保管你的 API 密钥。
:::

### 获取你的 Karakeep 端点 URL

在同一 **入口** 页面上，复制 Karakeep 显示的端点 URL。例如：

```text
https://abc123.{username}.olares.com
```

保存此 URL 供后续使用。

### 生成 API 密钥

1. 返回浏览器中的 Karakeep。
2. 点击你的个人资料图标，然后选择 **用户设置**。
3. 前往 **API 密钥** 标签页，然后点击 **新建 API 密钥**。
4. 输入密钥名称（例如，`mobile`），然后点击 **创建**。复制生成的密钥。你只能看到它一次。

   ![在 Karakeep 中生成 API 密钥](/images/manual/use-cases/karakeep-api-key.png#bordered)

### 从移动应用登录

1. 如果你在认证步骤中选择了 **内部**，请在手机上打开 LarePass 并开启 VPN。

   ![在移动设备上启用 LarePass VPN](/images/manual/get-started/larepass-vpn-mobile.png#bordered)

2. 打开 Karakeep 应用，然后选择 **改用 API 密钥**。
3. 输入你的详细信息：

   - **服务器地址**：粘贴你之前复制的端点 URL。
   - **API 密钥**：粘贴你生成的 API 密钥。

4. 点击 **登录**。你现在可以从手机保存图片和链接了。

   ![Karakeep 移动登录](/images/manual/use-cases/karakeep-mobile-signin.png#bordered)

Karakeep 还支持浏览器扩展和其他客户端。有关完整列表，请参阅 [karakeep.app/apps](https://karakeep.app/apps/)。

## 使用本地模型自动为书签添加标签

Karakeep 可以使用 Olares 上托管的本地模型为你的保存内容生成标签。本指南使用 Qwen3.5 27B Q4_K_M (Ollama) 作为文本模型。

### 前提条件

- 从 Market 安装的本地模型应用，且模型已完全下载。对于图像标签，还需安装 [Ollama](ollama.md) 并拉取视觉模型，如 `llava`。

### 获取模型端点和名称

1. 从 Launchpad 打开你的模型应用。模型名称显示在页面上（例如，`qwen3.5:27b-q4_K_M`）。记下它供后续使用。

   ![获取模型名称](/images/manual/use-cases/deerflow2-get-model-name.png#bordered)

2. 打开 **设置**，然后前往 **应用** > **Qwen3.5 27B Q4_K_M (Ollama)**。
3. 在 **共享入口** 下，选择模型应用以查看端点 URL。

   ![获取共享端点](/images/manual/use-cases/ollama-shared.png#bordered){width=70%}

4. 复制共享端点。例如：

   ```text
   http://94a553e00.shared.olares.com
   ```

### 将 Karakeep 连接到本地模型

1. 打开 **设置**，然后前往 **应用** > **Karakeep** > **管理环境变量**。
2. 点击每个变量旁边的 <i class="material-symbols-outlined">edit_square</i>，输入值，然后点击 **确认**：

   - **OLLAMA_BASE_URL**：上一步中的共享端点 URL。例如：
     ```text
     http://94a553e00.shared.olares.com
     ```
   - **INFERENCE_TEXT_MODEL**：上一步中的模型名称。例如：
     ```text
     qwen3.5:27b-q4_K_M
     ```
   - **INFERENCE_IMAGE_MODEL**（可选）：在 Ollama 中拉取的视觉模型名称。仅当你希望 Karakeep 为图像添加标签时才设置此项。

3. 点击 **应用**，等待 Karakeep 重启。

   ![管理 Karakeep 环境变量](/images/manual/use-cases/karakeep-manage-env-vars.png#bordered)

:::info 仅文本标签
如果你只需要文本标签，请将 `INFERENCE_IMAGE_MODEL` 留空。
:::

### 为现有书签生成标签

1. 打开 Karakeep，点击你的个人资料图标，然后选择 **用户设置**。
2. 前往 **AI 设置**。默认提示适用于大多数情况。
3. 点击你的个人资料图标，选择 **管理员设置**，然后前往 **后台任务**。
4. 在 **推理任务** 中，点击 **为所有书签重新生成 AI 标签**。队列大小会增加。

   ![在 Karakeep 中重新生成标签](/images/manual/use-cases/karakeep-regenerate-tags.png#bordered)

5. 队列清空后，刷新仪表板。标签将出现在你的书签上。

   ![生成的标签](/images/manual/use-cases/karakeep-generated-tags.png#bordered)

## 启用视频下载

Karakeep 可以使用 `yt-dlp` 自动从保存的链接下载视频到 Olares。视频下载后，你可以在 Karakeep 中离线观看，或从书签附件中将视频文件下载到你的本地计算机。

1. 打开 **设置**，然后前往 **应用** > **Karakeep** > **管理环境变量**。
2. 点击 `CRAWLER_VIDEO_DOWNLOAD` 旁边的 <i class="material-symbols-outlined">edit_square</i>，将值设置为 `true`，然后点击 **确认**。
3. 点击 **应用**，等待 Karakeep 重启。
4. Karakeep 重启后，像添加任何其他书签一样添加视频链接。

   Karakeep 会在后台将视频下载到 Olares。根据视频大小和网络条件，这可能需要一些时间。

5. 在书签卡片上，点击 <i class="material-symbols-outlined">pan_zoom</i> 打开书签详情。

   视频下载到 Olares 后，**视频** 将出现在详情页面顶部的内容类型下拉菜单中。

6. 可选：要将视频文件下载到你的本地计算机，请在右侧边栏的 **附件** 下找到 **视频**，然后点击其旁边的 <i class="material-symbols-outlined">download</i>。

   ![视频下载](/images/manual/use-cases/karakeep-video-download.png#bordered)

:::warning 下载失败
由于网络问题或来源网站的反机器人保护，视频下载可能会失败。如果下载失败，请参阅 [为什么我的视频下载失败](#为什么我的视频下载失败) 进行故障排除。
:::

## 配置 SMTP 以发送邮件邀请

要通过电子邮件邀请用户，请在 Olares 中配置系统 SMTP 设置。Karakeep 默认使用系统 SMTP 设置。

1. 打开 **设置**，然后前往 **高级** > **系统环境变量**。
2. 根据邮件提供商提供的 SMTP 设置配置以下系统 SMTP 变量：

   - `OLARES_USER_SMTP_ENABLED`：设置为 `true`。
   - `OLARES_USER_SMTP_SERVER`：SMTP 服务器主机，例如 `smtp.gmail.com`。
   - `OLARES_USER_SMTP_PORT`：SMTP 端口，例如 TLS 使用 `587`。
   - `OLARES_USER_SMTP_USERNAME`：SMTP 用户名。对于 Gmail，使用你的完整 Gmail 地址。
   - `OLARES_USER_SMTP_PASSWORD`：SMTP 密码或应用密码。
   - `OLARES_USER_SMTP_FROM_ADDRESS`：发件人电子邮件地址。

   对于大多数邮件提供商，这些变量已足够。如果你的提供商需要特定的 SSL/TLS 设置，请按照提供商的说明配置相关的 SMTP 安全变量。

3. 点击 **应用**。Karakeep 将重启以使新的 SMTP 设置生效。

Karakeep 重启后，你可以从管理员用户管理页面发送电子邮件邀请。

## 常见问题

### 为什么我的视频下载失败？

#### 原因

视频网站应用反机器人措施，如 CAPTCHA、IP 封锁和 headless 浏览器检测。这些措施可以阻止 Karakeep 用于视频下载的工具 `yt-dlp`。网络不稳定也可能导致部分下载失败。

#### 解决方案

1. 打开 Control Hub，点击 **浏览**，选择你的 Karakeep 项目，展开 **部署**，然后选择正在运行的 Pod。
2. 在右侧面板中，向下滚动到 **容器** 部分，找到 `karakeep` 容器，然后点击其旁边的 <i class="material-symbols-outlined">article</i>。

   ![检查容器日志](/images/manual/use-cases/karakeep-container-logs.png#bordered)

3. 搜索 `[VideoCrawler]` 以找到具体错误。

   例如，如果你看到类似以下的错误，说明来源网站已阻止下载请求：

   ```text
   ERROR: [youtube] mZp8yCueuKU: Sign in to confirm you're not a bot
   ```

4. 如果你的网络使用动态 IP 地址，请等待 IP 地址更改后重试。

某些失败是由来源网站的限制引起的。在这些情况下，Karakeep 可能无法下载视频，直到限制解除或 `yt-dlp` 支持更新的网站行为。

## 了解更多

- [Karakeep 文档](https://docs.karakeep.app/)：官方功能参考、API 文档和第三方客户端集成。
- [通过 Ollama 下载和运行本地 AI 模型](ollama.md)：安装 Ollama 以托管用于图像标签的视觉模型。
- [设置 Open WebUI 进行本地 AI 聊天](openwebui.md)：Olares 上共享模型端点的参考工作流。
