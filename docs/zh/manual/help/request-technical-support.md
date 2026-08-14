---
outline: [2, 3]
description: 了解如何通过 Ticket 应用、Olares Space 或 GitHub 获取 Olares 技术支持。
head:
  - - meta
    - name: keywords
      content: Olares, 技术支持, 系统日志, GitHub Issue, Ticket, 帮助
---
# 获取技术支持

如果你无法通过故障排查指南解决问题，可以通过以下任一渠道联系 Olares 支持团队获取帮助。

## 在 Ticket 应用中提交工单

安装并使用 Ticket 应用，你可以在 Olares 设备上直接创建和管理支持工单、追踪工单进度，并与 Olares 支持团队沟通。它还可以自动收集系统信息和日志，无需手动导出。

如果你更喜欢图形界面，并需要自动收集当前 Olares 设备上的日志和系统信息，请选择此方式。

:::info 前提条件
- 确保 Olares 系统已更新到 1.12.7 或更高版本。
- 登录 LarePass 的 Olares ID 必须与当前 Olares 设备上登录的 Olares 账号一致。
:::

### 安装 Ticket

在 Olares 设备上安装 Ticket 应用，并通过 LarePass 登录，以便提交工单。

1. 打开应用市场，搜索 “Ticket”，并安装该应用。

   ![Ticket app in Market](/images/zh/manual/help/ticket.png#bordered)

2. 打开 Ticket 应用，屏幕上会显示登录二维码。
3. 在你的移动设备上，打开 LarePass 应用。
4. 进入**设置**标签页。
5. 点击右上角的扫描图标，扫描屏幕上的二维码，然后点击**确认**登录。

### 提交工单

创建工单并描述你的问题。Ticket 可以从 Olares 设备自动收集系统信息和日志。

1. 点击主页面上的**新建工单**，或点击左侧边栏中**全部工单**旁边的 <i class="material-symbols-outlined">add_circle</i>。
2. 在**工单详情**中，填写表单：
   - **工单类型**：选择与问题最匹配的请求类型。
   - **产品或功能**：选择与问题相关的产品或区域。
   - **具体问题**：可选。选择更具体的主题。
   - **标题**：输入问题的简短摘要。
   - **问题描述**：详细描述问题，包括触发原因、复现步骤以及你期望的结果。
   - **附件**（可选）：点击上传区域或拖拽文件到该区域上传附件。
3. 在**系统信息**中，查看从 Olares 设备自动收集的设备信息。如果不想将该信息包含在工单中，请关闭**随工单提交**。
4. 在**系统日志**中，展开该区域并点击**采集日志**，即可自动收集并附加系统日志。

   采集完成后，日志文件会作为附件显示在该区域。

   ![收集日志](/images/zh/manual/help/collect-logs.png#bordered)

   :::tip 不想使用自动收集？
   你也可以从 Olares **设置 > 高级 > 导出系统日志**手动导出系统日志，然后在**附件**中上传。详细步骤，参见[导出系统日志](/zh/manual/olares/settings/developer.md#导出系统日志)。
   :::

5. 点击**新建工单**提交。

### 管理工单

提交工单后，你可以追踪进度、查看 Olares 团队的回复，并通过回复沟通以及时解决问题。

#### 查看工单

查看工单状态，并打开目标工单查看回复和详情。

1. 在左侧边栏中选择一个状态来筛选工单。**全部工单**会显示你的所有工单，其他分类按状态显示工单。
2. 点击工单查看详情和 Olares 支持团队的回复。

#### 回复工单

通过回复与 Olares 支持团队沟通、补充详情、提供最新进展或询问后续问题。

1. 打开工单详情页。
2. 点击**添加消息**。
3. 在**新消息**面板中输入消息。你也可以附加文件，或点击**采集日志**收集并上传系统日志。
4. 点击 **发送消息**。

#### 关闭或解决工单

当问题已解决或不再需要跟进时，将工单标记为关闭或已解决。

1. 打开工单详情页。
2. 在页面底部点击**关闭工单**或**标记为已解决**。

:::info
你无法删除已提交的工单。如果不再需要，可关闭工单。
:::

#### 重新打开工单

如果问题再次出现或未完全修复，你可以重新打开已关闭或已解决的工单。

1. 打开工单详情页。
2. 点击**重新打开**。
3. 添加回复，说明重新打开的原因。

工单状态会变回**待处理**，Olares 团队将继续处理该工单。

## 在 Olares Space 中提交工单

你也可以直接通过网页浏览器在 Olares Space 中创建和管理支持工单。这种方式无需安装单独的 Ticket 应用，即使无法访问 Olares 设备（例如系统崩溃或无响应）也可以使用。通过 Ticket 应用提交的工单也可以在这里查看，方便集中跟踪所有请求。

如果你无法访问 Olares 设备，请选择此方式。你可以手动填写网页表单，也可以使用 `olares-cli` 上传日志并自动创建工单。

在 Olares Space 中，你可以通过以下任一方式创建工单：

- **手动创建**：填写网页表单，并根据需要附加导出的系统日志。
- **通过 Olares CLI 自动创建**：直接从 Olares 终端上传日志，系统会自动创建一个工单以便跟进。

详细步骤，参见[在 Olares Space 中创建和管理支持工单](../space/tickets.md)。

## 在 GitHub 上报告

对于需要持续技术讨论和跟进的 Olares OS 系统问题或功能请求，建议公开提交到 GitHub。

- **Olares OS** 相关问题，提交到 [beclab/Olares](https://github.com/beclab/Olares)。
- **Olares Market 中的应用**相关问题，提交到 [beclab/apps](https://github.com/beclab/apps)。

在每个仓库中，你都可以创建 Discussion 进行一般性问题交流，或提交 Issue 报告 Bug 和技术问题。

1. 在 Olares 设置中[导出系统日志](/zh/manual/olares/settings/developer.md#导出系统日志)。
2. 打开对应的 GitHub 仓库，创建新的 Discussion 或 Issue。
3. 描述问题并附上导出的系统日志文件。视情况提供以下信息：
   - 复现问题的具体步骤
   - 错误信息或异常行为的描述
   - 使用环境的信息（操作系统、Olares 版本等）

## 常见问题

### 工单与我的 Olares ID 有什么关系？

工单与创建时使用的 Olares ID 绑定。Ticket 应用只展示当前 Olares ID 所创建的工单。

如果你重装了 Olares 并切换到新的 Olares ID，就无法在 Ticket 中看到之前 ID 创建的工单。此时可以使用原来的 Olares ID 登录 Olares Space，然后在**工单**页面中找到这些工单。

### 提交工单后可以编辑吗？

不可以。提交后无法编辑标题或描述。你可以通过添加回复来补充信息。

### 可以删除工单吗？

不可以。已提交的工单无法删除。你可以选择关闭它。

### 什么时候该用哪个支持渠道？

根据你的情况选择最合适的渠道：

- **Ticket 应用**：当你希望使用图形界面，并需要自动收集当前 Olares 设备上的日志和系统信息时使用。
- **Olares Space**：当你想通过浏览器创建工单，或使用 `olares-cli` 上传日志时使用。
- **GitHub**：适用于需要公开技术讨论和持续跟进的 Olares OS 系统问题或应用相关问题。
  - Olares OS 问题：[beclab/Olares](https://github.com/beclab/Olares)
  - 应用问题：[beclab/apps](https://github.com/beclab/apps)
- **论坛**：访问 [Olares 论坛](https://www.olares.com/forum/) 参与社区讨论、提问和交流经验。
- **Discord**：加入 [Olares Discord](https://discord.gg/olares) 社区，进行实时交流和快速提问。
