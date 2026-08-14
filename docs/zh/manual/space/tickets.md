---
outline: [2, 3]
description: 在 Olares Space 提交支持工单、使用 olares-cli 上传日志，并管理你的支持请求。
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, tickets, 工单, 支持, Olares CLI, 日志
---
# 在 Olares Space 中创建和管理支持工单

使用 Olares Space 的**工单**页面报告问题、寻求帮助并跟踪支持请求的状态。通过 Olares CLI 上传日志自动创建工单，或通过网页表单手动创建。

**工单**页面仅展示当前登录的 Olares ID 所创建的工单。要查看某个工单，请使用创建该工单时所用的 Olares ID 登录 Olares Space。

:::tip 前提条件
需要 Olares OS v1.12.6 或更高版本。
:::

## 通过 Olares CLI 自动创建工单

在设备终端运行 Olares Space 中显示的命令，即可自动上传系统日志并创建工单。

1. 在浏览器中打开 [Olares Space](https://space.olares.com/)，使用 LarePass 扫描二维码登录。
2. 在左侧边栏中，选择**工单**，然后点击 <i class="material-symbols-outlined">terminal</i>。

   ![使用 Olares CLI 上传日志](/images/zh/manual/space/upload-log-cli.png#bordered)

3. 在**使用 Olares CLI 上传日志**窗口中，复制**命令**。

   :::info 验证码
   验证码有效期有限，倒计时结束后会自动刷新。请在验证码过期前复制命令。
   :::

   ![上传日志对话框](/images/zh/manual/space/upload_logs_window.png#bordered){width=70%}

4. 访问 Olares 设备上的终端。例如，[通过 SSH 访问](/zh/one/access-terminal-ssh.md)。
5. 在终端中运行复制的命令。它会使用 Olares CLI 自动收集和上传系统日志。
6. 命令执行完成后，会创建一个附带日志的工单。记下工单编号以便后续跟进，例如 `TKT-10324`。
7. 返回**工单**页面，找到你在第 6 步中记下的工单编号对应的工单。该工单名为 **Olares CLI logs {creation-date}**，状态为**待处理**。

   ![Olares CLI 日志工单](/images/zh/manual/space/cli-logs-ticket.png#bordered)

提交后，你可以追踪工单进度、查看回复，并与 Olares 支持团队沟通，直到问题解决。

## 手动创建工单

通过网页表单提交新的支持请求。

1. 在浏览器中打开 [Olares Space](https://space.olares.com/)，使用 LarePass 扫描二维码登录。
2. 在左侧边栏中，选择**工单**，然后点击 **+ 新建工单**。
3. 填写必要的问题详情，然后点击**新建工单**。

   ![在 Olares Space 创建工单](/images/zh/manual/space/create_ticket.png#bordered)

   :::tip 导出系统日志
   如需附加系统日志，请从 Olares **设置** > **高级** > **[导出系统日志](../olares/settings/developer.md#导出系统日志)** 导出。
   :::

## 了解更多

- [获取支持](/zh/manual/help/request-technical-support.md)：了解其他支持选项，例如 Ticket 应用或 GitHub。
