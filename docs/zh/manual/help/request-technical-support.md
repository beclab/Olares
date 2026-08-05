---
outline: [2, 3]
description: 了解如何通过 Assist Hub 应用、Olares Space 或 GitHub 获取 Olares 技术支持。
head:
  - - meta
    - name: keywords
      content: Olares, 技术支持, 系统日志, GitHub Issue, Assist Hub, 帮助
---
# 获取技术支持

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../../manual/help/request-technical-support.md)。
:::

如果你无法通过故障排查指南解决问题，可以通过以下任一渠道联系 Olares 团队获取帮助。

## 在 Assist Hub 应用中提交工单

Olares 设备上的 Assist Hub 应用是获取帮助的推荐方式。它内置了日志收集功能，你可以直接提交工单，无需先手动导出日志。

:::info 先关联 Olares Space
提交工单前，请确保你的 Olares Space 账户已与 Olares 设备关联。如果未关联，系统会提示你在 LarePass 的**设置** > **集成**中完成绑定。具体操作参见[在 Olares Space 中监控 Olares](../space/manage-olares.md)。
:::

### 安装 Assist Hub

在 Olares 设备上安装 Assist Hub 应用，并通过 LarePass 登录，以便提交工单。

1. 打开应用市场，搜索 “Assist Hub”，并安装该应用。
2. 打开 Assist Hub 应用，屏幕上会显示登录二维码。
3. 在你的移动设备上，打开 LarePass 应用，进入**设置**标签页。
4. 点击右上角的扫描图标，扫描屏幕上的二维码，然后点击**确认**登录。

### 提交工单

创建工单并描述你的问题。Assist Hub 可以从 Olares 设备自动收集系统信息和日志。

1. 点击左侧菜单或页面底部的**创建工单**。
2. 填写表单：
   - **Issue Type**：选择与问题最匹配的类别。
   - **Subcategory**：可选。选择更具体的子类别。
   - **Issue Title**：输入问题的简短摘要。
   - **Description**：描述问题、触发条件以及复现步骤。可以直接粘贴或拖拽图片到编辑器中。
   - **Attachments**（可选）：点击上传区域或拖拽文件到该区域上传附件。
3. 查看**System Information**。该区域的信息会从你的 Olares 实例自动收集。如果不想让该信息包含在工单中，请关闭**Attach in ticket**开关。
4. 如需附加完整系统日志，展开**Collect Logs**，开启**Full logs**，然后点击**Collect**。等待收集完成。
5. 点击**Submit**提交工单。

### 管理工单

提交工单后，你可以查看工单状态、添加回复并更新工单状态。

#### 查看工单状态

查看工单状态，或打开某个工单查看详情。

1. 在左侧边栏选择一个状态来筛选工单。**Home** 显示所有工单，其他分类按状态显示工单。
2. 点击工单查看详情。

#### 回复工单

通过回复补充更多详情，或回答支持团队的问题。

1. 打开工单详情页。
2. 点击**添加回复**。
3. 输入消息后点击**发送回复**。

#### 关闭或解决工单

当问题已解决或不再需要跟进时，将工单标记为关闭或已解决。

1. 打开工单详情页。
2. 在页面底部点击**关闭**或**已解决**。

:::info
你无法删除已提交的工单。如果不再需要，请关闭它。
:::

#### 重新打开工单

如果问题再次出现或未完全修复，你可以重新打开已关闭或已解决的工单。

1. 打开工单详情页。
2. 点击**重新打开**。
3. 添加回复，说明重新打开的原因。

工单状态会变回 **In progress**，支持团队将继续处理。

## 在 Olares Space 中提交工单

你也可以直接通过网页浏览器在 Olares Space 中创建和管理支持工单。这种方式无需安装单独的 Assist Hub 应用，也无需能够访问 Olares 设备，除非你需要使用 `olares-cli` 上传日志。

详细步骤参见[在 Olares Space 中管理支持工单](../space/tickets.md)。

## 在 GitHub 上报告

如果你希望公开报告问题，或无法访问 Olares 设备，可以使用 Olares GitHub 仓库。

1. 在 Olares 设置中导出并下载系统日志：

   <!--@include: ../../reusables/export-system-logs.md#export-system-logs-steps-->

2. 打开 [Olares GitHub 仓库](https://github.com/beclab/Olares)，选择以下方式之一：
   - 创建一个新的 **Discussion**（适合一般问题或需求帮助）。
   - 提交一个 **Issue**（适合报告 Bug 或技术问题）。
3. 描述问题并附上导出的系统日志文件。视情况提供以下信息：
   - 复现问题的具体步骤
   - 错误信息或异常行为的描述
   - 使用环境的信息（操作系统、Olares 版本等）

## 常见问题

### 工单与我的 Olares ID 有什么关系？

工单与创建时使用的 Olares ID 绑定。Assist Hub 应用只展示当前 Olares ID 所创建的工单。

如果你重装了 Olares 并切换到新的 Olares ID，就无法在 Assist Hub 中看到之前 ID 创建的工单。此时可以使用原来的 Olares ID 登录 Olares Space，然后在 **Tickets** 页面中找到这些工单。

### 提交工单后可以编辑吗？

不可以。提交后无法编辑标题或描述。你可以通过添加回复来补充信息。

### 可以删除工单吗？

不可以。已提交的工单无法删除。你可以选择关闭它。
