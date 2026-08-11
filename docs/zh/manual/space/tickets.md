---
outline: [2, 3]
description: 在 Olares Space 提交支持工单、使用 Olares CLI 上传日志，并管理你的支持请求。
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, tickets, 工单, 支持, Olares CLI, 日志
---
# 在 Olares Space 中创建和管理支持工单

使用 Olares Space 的**工单**页面报告问题、寻求帮助并跟踪支持请求的状态。

## 创建工单

使用以下任一方式在 Olares Space 中创建支持工单。

### 手动创建工单

通过网页表单提交新的支持请求。描述你的问题，并附加截图、导出的系统日志或其他文件，帮助 Olares 支持团队更快地诊断和解决问题。

:::tip 提前准备系统日志
如需附加系统日志，请先从 Olares **设置 > 高级 > 导出系统日志**导出。导出所需时间取决于日志量大小。详细步骤参阅[导出系统日志](../olares/settings/developer.md#导出系统日志)。
:::

1. 在浏览器中打开 [Olares Space](https://space.olares.com/)，使用 LarePass 扫描二维码登录。
2. 在左侧边栏中，选择**工单**。
3. 点击 **+ 新建工单**，然后填写工单详情：

   ![在 Olares Space 创建工单](/images/zh/manual/space/create_ticket.png#bordered)

   - **工单类型**：选择与问题最匹配的请求类型。
   - **产品或功能**：选择与问题相关的产品或区域。
   - **具体问题**：可选。选择更具体的主题。
   - **标题**：输入问题的简短摘要。
   - **问题描述**：详细描述问题，包括触发原因、复现步骤以及你期望的结果。
   - **附件**（可选）：点击上传区域或拖拽文件到该区域。

4. 在表单底部，选择一个操作：
   - **新建工单**：将工单发送给支持团队。
   - **保存草稿**：保存当前进度，暂不提交。
   - **删除草稿**：删除当前草稿。

### 通过 Olares CLI 自动创建工单

直接从 Olares 设备上传系统日志。系统会自动创建工单并附加收集到的日志。

1. 在**工单**页面，点击 **+ 新建工单** 旁边的 <i class="material-symbols-outlined">terminal</i>。

   ![使用 Olares CLI 上传日志](/images/zh/manual/space/upload-log-cli.png#bordered)

2. 在**使用 Olares CLI 上传日志**窗口中，复制**命令**。

   :::info 验证码
   验证码有效期有限，倒计时结束后会自动刷新。请在验证码过期前复制命令。
   :::

   ![上传日志对话框](/images/zh/manual/space/upload_logs_window.png#bordered){width=60%}

3. 访问 Olares 设备上的终端。根据你的环境选择合适的方式：
   - [通过控制面板访问终端](/one/access-terminal-control-hub.md)：在浏览器中使用网页终端。
   - [通过 SSH 访问](/one/access-terminal-ssh.md)：从你的电脑通过网络连接。
   - [直接物理访问](/one/access-physical-console.md)：在设备上直接登录。
4. 运行从窗口中复制的命令：

   ```bash
   sudo olares-cli logs upload --code <verification-code> --olares-id <your-olares-id>
   ```

   示例：

   ```bash
   sudo olares-cli logs upload --code 37116007 --olares-id laresprime@olares.com
   ```

5. 命令执行完成后，会创建一个附带日志的工单。记下工单编号以便后续跟进，例如 `TKT-10306`。

   以下示例展示在控制面板的终端中运行命令：

   ![使用 olares-cli 上传日志](/images/zh/manual/space/cli-upload-log.png#bordered)

6. 返回**工单**页面，找到你在第 5 步中记下的工单编号对应的工单。该工单名为 **Olares CLI logs {creation-date}**，状态为**待处理**。

   ![Olares CLI 日志工单](/images/zh/manual/space/cli-logs-ticket.png#bordered)

## 管理工单

提交工单后，你可以追踪进度、查看 Olares 支持团队的回复，并通过回复沟通以及时解决问题。

### 查看工单

1. 要查看所有工单，打开**工单**页面。列表会显示每个工单的标题、状态、工单编号、工单类型和创建时间。
2. 点击目标工单打开详情页，查看 Olares 支持团队的回复。
3. 使用状态下拉列表筛选工单：

   - **全部**：所有工单。
   - **草稿**：已保存但未提交的工单。
   - **待处理**：等待首次回复的工单。
   - **处理中**：Olares 支持团队正在处理的工单。
   - **已解决**：已解决的工单。
   - **已关闭**：已关闭的工单。

### 回复工单

通过回复与 Olares 支持团队沟通、补充详情、提供最新进展或询问后续问题。

1. 在**工单**页面，点击目标工单。
2. 在工单详情页底部，点击**添加消息**。
3. 在**新消息**面板中，输入消息内容，并根据需要附加文件。
4. 点击**发送消息**。

### 关闭或解决工单

当问题已解决或不再需要跟进时，将工单标记为 **已解决** 或 **关闭**：

1. 打开工单详情页。
2. 在页面底部点击**标记为已解决**或**关闭工单**。

### 重新打开工单

如果问题再次出现或未完全修复，你可以重新打开**已关闭**或**已解决**状态的工单：

1. 打开工单详情页。
2. 在页面底部点击**重新打开工单**。
3. 添加回复，说明重新打开的原因。

工单状态会变回**待处理**，Olares 支持团队将继续处理。

## 常见问题

### 工单与我的 Olares ID 有什么关系？

Olares Space 的**工单**页面仅展示当前登录的 Olares ID 所创建的工单。要查看某个工单，请使用创建该工单时所用的 Olares ID 登录 Olares Space。

### 提交工单后可以编辑吗？

不可以。提交后无法编辑标题或描述。你可以通过添加回复来补充信息。

### 可以删除工单吗？

不可以。已提交的工单无法删除。你可以选择关闭它们。
