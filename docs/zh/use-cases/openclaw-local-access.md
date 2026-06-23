---
outline: [2, 3]
description: 了解如何授予 OpenClaw 访问 Olares 上本地文件和外部存储的权限。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, file access, local files, external storage, NAS
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/openclaw-local-access.md)。
:::

# 可选：启用本地文件访问

默认情况下，OpenClaw 在安全的沙箱环境中运行，只能访问自己的工作空间。为了让你的 AI 助手真正个性化，你可以授予它访问 Olares 文件系统的权限。这允许 OpenClaw 读取、组织和处理你的文档、媒体以及连接的外部存储（如 NAS 或 USB 驱动器）。

:::tip OS 版本要求
你必须将 Olares OS 升级到 V1.12.5 或更高版本才能使用此功能。
:::

## 学习目标

在本指南中，你将学习如何：
- 为 Home 和 External 目录启用文件访问权限。
- 通过更新核心指令来告诉你的助手文件存储位置。
- 验证 OpenClaw 是否可以读取和写入你的 Olares 文件系统。

## 步骤 1：启用文件访问设置

设置相应的环境变量以授予 OpenClaw 管理文件的权限。

1. 打开 **Settings**，然后前往 **Applications** > **OpenClaw** > **Manage environment variables**。

    ![Enable file access settings](/images/manual/use-cases/openclaw-file-access.png#bordered){width=70%}

    两个变量控制文件访问：

    - **ALLOW_HOME_DIR_ACCESS**：允许访问 Files 应用中的 **Home** 目录。
    - **ALLOW_EXTERNAL_DIR_ACCESS**：允许访问 Files 应用中的 **External** 目录。

2. 对于每个变量，点击 <i class="material-symbols-outlined">edit_square</i>，将值设置为 **true**，然后点击 **Confirm**。
3. 点击 **Apply**。
4. 等待大约一分钟，让 OpenClaw 容器重启并应用新权限。

## 步骤 2：指示助手

授予权限后，Olares 将你的 **Home** 和 **External** 目录挂载到 OpenClaw 容器中的以下路径：
- `/home/userdata/home/`
- `/home/userdata/external/`

你必须告诉助手这些位置。有两种方法：
- （推荐）直接与助手聊天，让它记住这些位置。
- 使用以下步骤手动将访问规则添加到其核心文件中。

1. 打开 Control UI。
2. 前往 **Agents** > **Files**，然后在 **Core Files** 区域点击 **TOOLS**。
3. 点击 **Content** 区域以显示文本。

    ![Edit core files](/images/manual/use-cases/openclaw-edit-tools.png#bordered)
4. 在文件末尾 `## Why Separate?` 部分之后，复制并粘贴以下内容：

    ```markdown
    ## File Access Protocol on Olares
    You have access to user's files on Olares.
    
    ### Primary Storage: files on Olares primary stored in the directory `/home/userdata/home`
    - This is your default search directory.
    - Priority Trigger: When the user mentions keywords like "my Olares", prioritize searching this directory.
    
    ### External Storage: You can also access external storage devices (e.g., NAS, USB drives) connected to Olares in the directory `/home/userdata/external`.
    - Priority Trigger: When the user mentions keywords like "my NAS", "on my USB", or similar, prioritize searching this directory.
    
    ### Critical Security Rule
    Any deletion or modification of files within the `/home/userdata` requires explicit prior approval from the user before execution.
    ```

    ![Core files edited](/images/manual/use-cases/openclaw-tools-edited.png#bordered)    
5. 点击 **Save**。

## 步骤 3：测试连接

验证 OpenClaw 是否可以读取和写入你的 Olares 文件系统。

1. 前往 Control UI 中的 **Chat** 页面，或使用你连接的 Discord 机器人。
2. 向助手发送类似的消息：

    ```text
    In my Olares documents directory, create a txt file introducing yourself.
    ```

    助手将回复确认，指示文件名和创建路径。例如：

    ```text
    Done! Created your intro file at `/home/userdata/home/john-intro.txt`. Let me know if you want me to tweak it!
    ```
3. 打开 **Files** 应用并检查新创建文件的路径。

    ![Local access success](/images/manual/use-cases/openclaw-local-access-success.png#bordered)
