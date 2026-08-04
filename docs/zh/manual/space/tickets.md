---
outline: [2, 3]
description: 在 Olares Space 提交支持工单、使用 olares-cli 上传日志，并管理你的支持请求。
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, tickets, 工单, 支持, olares-cli, 日志
---
# 在 Olares Space 管理支持工单

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../manual/space/tickets.md)。
:::

在 Olares Space 的 **Tickets** 页面中，你可以报告问题、寻求帮助并跟踪支持请求的状态。你可以通过网页界面提交工单，也可以直接在 Olares 设备上使用 `olares-cli` 上传日志。

## 创建工单

提交新的支持请求，并附上详情和可选附件。

1. [登录 Olares Space](manage-accounts.md#log-in-to-olares-space)。
2. 在左侧边栏中，点击 **Tickets**。
3. 点击 **+ New ticket**。

   ![在 Olares Space 创建工单](/images/how-to/space/create_ticket.png#bordered)

4. 填写表单：
   - **Issue Type**：选择与问题最匹配的类别。
   - **Subcategory**：可选。选择更具体的子类别。
   - **Issue Title**：输入问题的简短摘要。
   - **Description**：描述问题、触发条件以及复现步骤。可以使用工具栏设置文本格式，也可以直接粘贴或拖拽图片到编辑器中。
   - **Attachments**：可选。点击上传区域或拖拽文件到该区域。最多可附加 50 个文件，单个文件不超过 2 GB。
5. 在表单底部，选择一个操作：
   - **Submit**：将工单发送给支持团队。
   - **Save draft**：保存当前进度，暂不提交。
   - **Delete draft**：删除当前草稿。

## 使用 olares-cli 上传日志

直接从 Olares 设备收集并上传系统日志，系统会自动创建一个工单。

1. 在 **Tickets** 页面，点击 **+ New ticket** 旁边的 <i class="material-symbols-outlined">terminal</i>。
2. 在 **Upload logs via olares-cli** 窗口中，复制 **Full command line**。

   :::info 验证码
   验证码有效期有限，倒计时结束后会自动刷新。请在验证码过期前复制完整的命令行。
   :::

   ![上传日志对话框](/images/how-to/space/upload_logs_window.png#bordered){width=60%}

3. 在 Olares 设备上打开终端。根据你的环境选择合适的方式：
   - [通过 Control Hub 访问终端](/one/access-terminal-control-hub.md)：使用网页终端快速访问。
   - [通过 SSH 访问](/one/access-terminal-ssh.md)：通过网络安全连接。
   - [直接物理访问](/one/access-physical-console.md)：在设备上直接登录。

4. 运行从窗口中复制的命令：

   ```bash
   sudo olares-cli logs upload --code <verification-code> --olares-id <your-olares-id>
   ```

   示例：

   ```bash
   sudo olares-cli logs upload --code 75753956 --olares-id laresprime@olares.com
   ```

5. 命令执行完成后，会创建一个附带日志的工单。记下工单编号以便后续跟进。

   ![使用 olares-cli 上传日志](/images/how-to/space/cli-upload-log.png#bordered)

6. 返回 Olares Space 并打开 **Tickets** 页面。你会看到一个名为 **Olares CLI logs {creation-date}** 的新工单，状态为 **Open**。

   ![CLI 日志工单](/images/how-to/space/cli-logs-ticket.png#bordered)

## 查看工单

使用工单列表跟踪现有支持请求的状态和详情。

1. 要查看所有工单，打开 **Tickets** 页面。列表会显示每个工单的标题、状态、工单编号、问题类型和创建时间。
2. 要按状态筛选工单，点击列表上方的状态下拉列表，然后选择一个状态：

   - **All**：所有工单。
   - **Draft**：已保存但未提交的工单。
   - **Open**：等待首次回复的工单。
   - **In progress**：支持团队正在处理的工单。
   - **Resolved**：已解决的工单。
   - **Closed**：已关闭的工单。

## 管理工单

提交工单后，你可以添加回复并更新工单状态，以推进支持请求的处理。

### 回复工单

通过回复补充更多详情、回答问题或提供最新进展。

1. 在 **Tickets** 页面，点击要更新的工单。
2. 在工单详情页底部，点击 **Add a Reply**。
3. 在 **Add a Reply** 对话框中，在 **Description** 字段输入你的消息。你也可以附加文件。
4. 点击 **Send reply**。

### 关闭或解决工单

当问题已解决或不再需要跟进时，将工单标记为 **Resolved** 或 **Closed**：

1. 打开工单详情页。
2. 在页面底部点击 **Resolved** 或 **Close**。

:::info
你无法删除已提交的工单。如果不再需要，请关闭它。
:::

### 重新打开工单

如果问题再次出现或未完全修复，你可以重新打开 **Closed** 或 **Resolved** 状态的工单：

1. 打开工单详情页。
2. 在页面底部点击 **Reopen**。
3. 添加回复，说明重新打开的原因。

工单状态会变回 **Open** 或 **In progress**，支持团队将继续处理。

## 常见问题

### 工单与我的 Olares ID 有什么关系？

每个工单都与创建时使用的 Olares ID 关联。要查看某个工单，请使用对应的 Olares ID 登录 Olares Space。如果你切换到其他 Olares ID，之前 ID 创建的工单不会显示在列表中。

### 提交工单后可以编辑吗？

不可以。提交后无法编辑标题或描述。你只能通过添加回复来补充信息。

### 可以删除工单吗？

不可以。已提交的工单无法删除。你可以选择关闭它。
