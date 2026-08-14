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

## 选择合适的支持渠道

为了最快解决问题，请选择最适合你情况的渠道：

| 渠道 | 适用场景 |
| --- | --- |
| **Ticket 应用** | 可以访问设备时使用。提供图形界面，并自动收集系统日志。 |
| **[Olares Space](https://www.olares.com/space/login?redirect=/)** | 无法访问设备（如系统崩溃或无响应）时使用，或希望通过 `olares-cli` 上传日志时使用。 |
| **GitHub** | 报告技术问题或提出需要公开技术讨论和持续跟进的功能请求。<br>• OS 问题：[beclab/Olares](https://github.com/beclab/Olares)<br>• 应用问题：[beclab/apps](https://github.com/beclab/apps) |
| **[论坛](https://www.olares.com/forum/)** | 社区讨论、提问和交流经验。 |
| **[Discord](https://discord.gg/olares)** | 实时交流和快速提问。 |

## 提交工单

你可以通过以下两种方式之一提交支持工单，具体取决于当前是否可以访问 Olares 设备。

### 通过 Ticket 应用提交

如果你希望通过图形界面自动收集设备上的系统信息和日志，请使用 Ticket 应用。

:::warning 前提条件
- **系统版本**：确保 Olares 系统已更新至 1.12.6 或更高版本。
- **登陆账号一致**：登录 LarePass 的 Olares ID 必须与当前 Olares 设备上登录的账号一致。
- **工单与账号绑定**：工单与创建时使用的 Olares ID 绑定。Ticket 应用仅显示当前 ID 关联的工单；如果切换了新 ID，必须使用原 ID 登录 Olares Space 才能查看之前的历史工单。
:::

1. 在应用市场中安装 Ticket 应用，然后打开它。
2. 使用 LarePass 扫描二维码登录。
3. 创建工单并填写必要的问题详情。
4. 查看自动收集的系统信息。如果不想包含在工单中，请关闭**随工单提交**。
5. 展开**系统日志**，点击**采集日志**，即可自动收集并附加到工单中。

   :::tip 希望手动收集日志？
   你也可以在 Olares 中前往**设置 > 高级 > 导出系统日志**手动导出系统日志，然后通过**附件**上传。详细步骤，参见[导出系统日志](/zh/manual/olares/settings/developer.md#导出系统日志)。
   :::

提交后，你可以追踪工单进度、查看回复，并直接与 Olares 支持团队沟通，直到问题解决。

### 通过 Olares Space 提交

如果你无法访问 Olares 设备（例如系统崩溃或无响应），请使用 Olares Space。此前通过 Ticket 应用提交的工单也可以在这里查看。

在 Olares Space 中，你可以通过以下两种方式创建工单：

- **手动创建**：填写网页表单，并上传导出的系统日志作为附件。
- **自动创建（CLI）**：直接在 Olares 终端使用 `olares-cli` 上传日志。系统会自动生成一个工单。

详细步骤，参见[在 Olares Space 中创建和管理支持工单](../space/tickets.md)。

## 在 GitHub 上报告问题

对于适合公开技术讨论的 Olares OS 系统问题或功能请求，建议前往 GitHub 提交 Issue。

选择对应的仓库：
- **Olares OS 问题：**[beclab/Olares](https://github.com/beclab/Olares)
- **应用问题：**[beclab/apps](https://github.com/beclab/apps)

在这两个仓库中，你都可以创建 **Discussion** 进行一般性交流，或创建 **Issue** 报告具体的技术问题。

1. 在 Olares 设置中[导出系统日志](/zh/manual/olares/settings/developer.md#导出系统日志)。
2. 打开对应的 GitHub 仓库，创建新的 Issue 或 Discussion。
3. 详细描述问题并附上导出的系统日志文件。请始终包含：
   - 复现问题的具体步骤
   - 错误信息或异常行为
   - 环境信息（操作系统、Olares 版本等）
