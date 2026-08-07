---
outline: [2, 3]
description: 使用 Olares 上的 OnlyOffice 在浏览器中编辑文档，并了解内置测试客户端及其账号功能限制。
head:
  - - meta
    - name: keywords
      content: Olares, OnlyOffice, Document Server, 文档编辑, 自托管办公
app_version: "1.1.14"
doc_version: "2.0"
doc_updated: "2026-08-04"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/onlyoffice.md)为准。
:::

# 使用 OnlyOffice 编辑文档

Olares 上的 OnlyOffice 提供文档、电子表格、演示文稿及其他支持格式的浏览器编辑器。你可以通过内置 Web 客户端上传、编辑和下载文件。

:::warning Olares 应用包含的组件
OnlyOffice 应用包含 ONLYOFFICE Document Server 和官方 Node.js 测试客户端 `onlyofficeclient`，不包含 ONLYOFFICE Workspace 或 DocSpace。

测试客户端不提供用户账号，也不是可用于生产环境的文档管理门户。**Username** 菜单中的身份是预设测试身份，并非真实账号。应用中没有关闭测试客户端并创建 OnlyOffice 用户的设置。
:::

## 了解应用组成

ONLYOFFICE Document Server 提供文档编辑器和编辑后端。通常还需要 ONLYOFFICE Workspace、DocSpace 或兼容的文档管理系统来提供文件存储、用户账号和权限管理。

Olares 应用内置的测试客户端提供了一个体验编辑器的简单界面。你可以用它：

- 从电脑上传文件。
- 在浏览器中创建和编辑文档。
- 下载编辑后的文件。

如需体验多人编辑，可以在 **Username** 中选择不同的预设身份，并在不同浏览器会话中打开同一文档。此功能仅用于模拟和测试协作。

它不提供：

- 用户身份验证、真实账号或权限管理。
- 可用于生产环境的共享文档库或协作门户。

## 安装 OnlyOffice

1. 打开应用市场，搜索 "OnlyOffice"。

   ![OnlyOffice](/images/manual/use-cases/onlyoffice.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 使用 OnlyOffice

从启动台打开 OnlyOffice。主页提供新建和上传办公文件的入口。

![OnlyOffice 主页](/images/manual/use-cases/onlyoffice-main.png#bordered)

### 创建文件

1. 在主页的 **Create new** 区域，选择要创建的文件类型：

   - **Document**：创建 `.docx` 文本文档。
   - **Spreadsheet**：创建 `.xlsx` 电子表格。
   - **Presentation**：创建 `.pptx` 演示文稿。
   - **PDF form**：创建 `.pdf` 文件。

2. 在 OnlyOffice 编辑器中开始编辑。文件会保存在文档列表中，之后可以再次打开查看或编辑。

### 导出文件

编辑完成后，选择 **File** > **Download As**，按所需格式下载文件副本。

| 文件类型 | 下载格式 |
|:--|:--|
| 文档 | `DOCX`、`PDF`、`ODT`、`DOTX`、`PDF/A`、`OTT`、`RTF`、`TXT`、`FB2`、`EPUB`、<br>`HTML`、`JPG`、`PNG` |
| 电子表格 | `XLSX`、`ODS`、`CSV`、`PDF`、`XLTX`、`OTS`、`XLSB`、`PDF/A`、`JPG`、`PNG` |
| 演示文稿 | `PPTX`、`PPSX`、`PDF`、`ODP`、`POTX`、`PDF/A`、`OTP`、`JPG`、`PNG` |
| PDF | `DOCX`、`PDF`、`ODT`、`DOTX`、`OTT`、`RTF`、`TXT`、`FB2`、`EPUB`、`HTML`、<br>`JPG`、`PNG` |

### 上传文件

1. 在主页点击 **Upload file**。

2. 从电脑中选择文件。

   OnlyOffice 支持 `.docx`、`.xlsx`、`.pptx`、`.pdf`、`.odt`、`.ods` 和 `.odp` 等常见办公文件格式。

3. 在 **File upload** 对话框中，等待文件完成加载、转换和编辑器脚本加载，然后选择打开方式：

   - 点击 **Edit**，打开并编辑文件。
   - 点击 **View**，以只读方式打开文件。
   - 点击 **Embedded View**，使用嵌入式查看器打开文件。

   ![OnlyOffice 文件上传对话框](/images/manual/use-cases/onlyoffice-file-upload-dialog.png#bordered)

4. 上传完成后，文件会显示在文档列表中，之后可以再次打开。

:::info
ONLYOFFICE 使用 Office Open XML 格式进行编辑。上传或打开其他格式的文件进行编辑时，OnlyOffice 可能会先转换文件。转换过程中部分格式可能发生变化，尤其是原格式与目标格式支持的文档功能不同时。
:::

:::warning
ONLYOFFICE 将内置测试客户端作为集成示例提供。它不使用用户身份验证来保护存储的文件。未经适当修改，请勿将其用作生产环境中的文档管理系统。
:::

## 常见问题

### 可以通过 ONLYOFFICE 移动端或桌面端应用登录吗？

不可以。Olares 上的 OnlyOffice 不提供移动端或桌面端应用所需的账号，这些客户端也无法直接连接内置的 Document Server。

### 可以关闭演示界面并创建用户吗？

不可以。内置 Web 界面是 Document Server 的测试客户端，不是可配置的 Workspace 或 DocSpace 门户。你可以用它完成简单的浏览器文档编辑；如需账号和文档管理功能，请将 Document Server 接入兼容的平台。

## 了解更多

- [将 OnlyOffice 迁移到新架构](onlyoffice-migration.md)：升级到 Olares 1.12.6 后迁移已有文档。
- [ONLYOFFICE Docs API 集成示例](https://api.onlyoffice.com/docs/docs-api/samples/language-specific-examples/)：了解测试客户端的用途和安全限制。
- [ONLYOFFICE 文档](https://helpcenter.onlyoffice.com/docs)：查看 ONLYOFFICE 编辑器的官方帮助。
