---
outline: [2, 3]
description: 了解如何通过 Ticket 应用、Olares Space 或 GitHub 获取 Olares 支持。
head:
  - - meta
    - name: keywords
      content: Olares, 支持, 帮助, Ticket 工单, Olares Space 工单, 系统日志, GitHub Issue
---
# 获取支持

如果你无法通过故障排查指南解决问题，可以联系 Olares 支持团队获取帮助。

## 在哪里获取支持

为了最快解决问题，请选择最适合你情况的方式：

| 方式 | 适用场景 |
| --- | --- |
| **[Ticket](#提交工单)** | 私密的一对一技术支持，可安全地与 Olares 支持团队共享系统日志。 |
| **[GitHub Olares 仓库](https://github.com/beclab/Olares)** | 报告 Olares OS 系统问题或提出 OS 功能请求。 |
| **[GitHub apps 仓库](https://github.com/beclab/apps)** | 报告 Olares Market 中应用相关的技术问题或功能请求。 |
| **安全报告** | 报告系统漏洞。请遵循我们的[安全策略](https://github.com/beclab/Olares/blob/main/SECURITY.md)。 |
| **[论坛](https://www.olares.com/forum/)** | 分享知识、查找或发布教程、讨论功能以及获取社区帮助。 |

## 提交工单

你可以通过 Ticket 应用或 Olares Space 提交支持工单，具体取决于当前是否可以访问 Olares 设备。

:::tip 前提条件
- **系统版本**：确保 Olares 系统已更新至 1.12.6 或更高版本。
- **登陆账号一致**：登录 LarePass 的 Olares ID 必须与当前 Olares 设备上登录的账号一致。
:::

### 通过 Ticket 应用提交

当设备可访问时，使用 Ticket 应用。它提供图形界面，并直接从设备自动收集系统日志。

工单与创建时使用的 Olares ID 绑定。Ticket 应用仅显示当前 ID 关联的工单；如果切换了新 ID，必须使用原 ID 登录 Olares Space 才能查看之前的历史工单。

1. 在应用市场中安装 Ticket 应用，然后打开它。
2. 使用 LarePass 扫描二维码登录。
3. 创建工单并填写必要的问题详情。
4. 查看自动收集的系统信息。如果不想包含在工单中，请关闭**随工单提交**。
5. 展开**系统日志**，点击**采集日志**，即可自动收集并附加到工单中。

提交后，你可以追踪工单进度、查看回复，并与 Olares 支持团队沟通，直到问题解决。

### 通过 Olares Space 提交

如果设备无法访问，或者你无法使用 Ticket 应用，请使用 [Olares Space](https://www.olares.com/space/login?redirect=/)。此前通过 Ticket 应用提交的工单也可以在这里查看。

在 Olares Space 中，你可以通过以下两种方式创建工单：

- **手动创建**：填写网页表单，并上传导出的系统日志作为附件。
- **自动创建（CLI）**：在设备终端运行 Olares Space 中提供的命令。该命令使用 Olares CLI 自动收集和上传系统日志，然后自动创建工单。

   ![上传日志对话框](/images/zh/manual/space/upload_logs_window1.png#bordered)

详细信息，参见[在 Olares Space 中创建和管理支持工单](../space/tickets.md)。
