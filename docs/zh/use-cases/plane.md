---
outline: [2, 3]
description: 在 Olares 上自托管 Plane，作为开源的 Jira 和 Asana 替代方案。用模块、迭代周期和多种视图组织工作，实现私有的团队项目管理。
head:
  - - meta
    - name: keywords
      content: Olares, Plane, open source jira alternative, jira alternative, self-hosted project management, kanban, team collaboration, plane on olares
app_version: "1.0.9"
doc_version: "1.0"
doc_updated: "2026-04-14"
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/plane.md)为准。
:::

# 使用 Plane 管理项目

Plane 是一个开源的项目管理平台，将可视化看板与敏捷任务工作流相结合。它帮助团队在一个地方规划冲刺、跟踪任务并协作编写文档。

在 Olares 上自托管 Plane，可以将所有项目数据置于你的掌控之下。

## 学习目标

在本指南中，你将学习如何：
- 安装 Plane 及其依赖的后台服务。
- 设置工作区并邀请团队成员加入。
- 通过分类工作、安排冲刺和分配任务来运行项目。
- 使用不同的可视化布局洞察项目进度。

## 安装 Plane 及所需依赖

在安装 Plane 之前，你必须先安装 RabbitMQ（V4.0.0 或更高版本）和 MinIO（V1.0.0 或更高版本）。

:::info
Plane 依赖多个中间件组件才能平稳运行。虽然 PostgreSQL 和 Redis 已经预装在你的 Olares 系统中，但 RabbitMQ 和 MinIO 需要手动安装。
:::

1. 打开 Market，搜索 "RabbitMQ"。

   ![RabbitMQ](/images/manual/use-cases/rabbitmq.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。
3. 搜索 "MinIO" 并安装。

   ![MinIO](/images/manual/use-cases/minio.png#bordered)

4. 搜索 "Plane" 并安装。

   ![Plane](/images/manual/use-cases/plane.png#bordered)

## 创建工作区

安装完成后，注册你的账号并创建一个工作区，所有项目都将存放在此处。

1. 从 Launchpad 打开 Plane，然后在欢迎页面点击 **Get started**。
2. 在 **Setup your Plane Instance** 页面，填写所需信息，然后点击 **Continue**。

   ![Register for Plane](/images/manual/use-cases/plane-register.png#bordered){width=60%}

3. 点击右下角的 **Create workspace**。

   ![Create workspace](/images/manual/use-cases/plane-create-workspace.png#bordered)

4. 指定你的工作区名称，选择团队规模，然后点击 **Create workspace**。
5. 点击新创建的工作区，然后使用你刚刚设置的邮箱和密码登录。
6. 完善个人资料设置，然后点击 **Continue**。你将进入新的工作区。

## 邀请团队

邀请团队成员加入工作区，以便他们与你协作。

:::info
在发送邀请之前，你必须在 Olares 中将 Plane 设为公开访问。否则，没有 Olares 密码的用户将无法加入。
:::

1. 打开 Settings，进入 **Applications** > **Plane** > **Entrances** > **Plane**，然后将 **Authentication level** 更改为 **Public**。

   ![Change authentication level for Plane](/images/manual/use-cases/plane-auth-level.png#bordered){width=80%}

2. 返回 Plane，点击左上角的工作区名称，然后选择 **Invite members**。
3. 在 **Members** 页面，点击 **Add member**。
4. 输入团队成员的邮箱地址，然后选择要分配给该成员的角色。
5. 如需同时邀请多位成员，点击 **Add more**。

   ![Invite team members](/images/manual/use-cases/plane-invite-members.png#bordered){width=80%}

6. 点击 **Send invitations**。被邀请者将出现在 **Pending invites** 面板中。

   当被邀请的成员接受邀请后，他们将自动加入工作区。

   :::tip
   被邀请的成员可以在自己的 Plane 界面中查看 **Workspace invites** 部分来接受邀请。
   :::

## 使用 Plane

为了解如何使用 Plane 管理一个多阶段项目，让我们通过一个示例场景进行演示：你的团队需要执行一个 "Product Page Revamp" 项目，以提高网站转化率。

### 创建项目

首先，为这个特定项目创建一个项目，并添加将执行工作的团队成员。

1. 打开你的工作区并创建一个新项目：
   - 对于第一个项目，在 **Home** 页面点击 **Get started**。
   - 对于后续项目，从左侧边栏选择 **Projects**，然后点击 **Add Project**。
2. 定义项目的核心详情：
   - **Project name**: `Product Page Revamp`
   - **Project ID**: `WEB`
   - **Description**: `Improve UX/UI and messaging for the core product landing page`
3. 选择图标，设置项目可见性，并指定负责人。
4. 点击 **Create project**。

   ![Create project](/images/manual/use-cases/plane-create-project.png#bordered){width=70%}

5. 点击 **Open project**。
6. 在左侧边栏中，点击新项目名称，点击 <span class="material-symbols-outlined">more_horiz</span>，然后点击 **Settings**。
7. 在左侧边栏中，在项目名称下选择 **Members**。
8. 点击 **Add member**，选择参与该项目的成员及其角色，然后点击 **Add members**。
9. 点击左上角的 **Back to workspace**。

### 分类工作

为了保持大型项目的条理性，将相关任务分组到逻辑类别中会有所帮助。在 Plane 中，这些类别称为 "Modules"。

在本场景中，我们创建三个模块："Visual assets"、"Copywriting" 和 "Technical SEO"。

1. 从左侧边栏点击新项目以展开，然后点击 **Modules**。
2. 点击 **Build your first module** 或 **Add Module**。
3. 定义模块的核心详情：
   - **Title**: `Visual assets`
   - **Description**: `Focus on photography, iconography, and UI design elements`
4. 设置日期范围、状态标签、负责人和成员。
5. 点击 **Create Module**。

   ![Create module](/images/manual/use-cases/plane-create-module.png#bordered){width=70%}

6. 重复这些步骤来创建另外两个模块。

### 将工作安排到冲刺中

不要一次性处理所有事情，而是将时间线拆分为专注的、固定时间段的冲刺。在 Plane 中，冲刺称为 "Cycles"。

在本场景中，我们将创建两个迭代周期来展示从规划到执行的过渡。

1. 在左侧边栏中，点击 **Cycles**。
2. 点击 **Set your first cycle** 或 **Add cycle**。
3. 定义阶段的核心详情：
   - **Title**: `Phase 1: Discovery`
   - **Description**: `Research, wireframe, and define the core value proposition; the goal is to finalize the skeleton of the new product page`
4. 选择开始和结束日期。
5. 点击 **Create cycle**。

   ![Create cycle](/images/manual/use-cases/plane-create-cycle.png#bordered){width=70%}

6. 重复这些步骤来创建 Phase 2 的迭代周期：
   - **Title**: `Phase 2: Execution`
   - **Description**: `Hi-Fi UI design, final copy production, and SEO auditing; the goal is to complete final visual assets and prepare for development`

### 创建并分配工作项

现在你的结构已经就位，详细说明完成改版所需的具体行动。将这些行动项分配给你的团队，设置优先级，并将它们映射到你刚刚创建的类别和阶段。

在本场景中，我们将为该项目创建以下任务项。

| Task title | Module | Cycle | Priority |
|:---|:---|:---|:---|
| Conduct UX audit | Technical SEO | Phase 1: Discovery | High |
| Draft eye-catching headlines | Copywriting | Phase 1: Discovery | Urgent |
| Create Low-Fi sketches | Visual assets | Phase 1: Discovery | Medium |
| Design final UI mockups | Visual assets | Phase 2: Execution | High |
| Write meta descriptions | Technical SEO | Phase 2: Execution | Medium |

1. 在左侧边栏中，点击 **Work items**。
2. 点击 **Create your first work item** 或 **Add work item**。
3. 定义任务的核心详情：
   - **Title**: `Conduct UX audit`
   - **Description**: `Review the current homepage for friction points, and focus on mobile navigation and Add to Cart button visibility`。
4. 设置状态和优先级，分配给团队成员，设置日期范围，并关联到迭代周期和模块。

   ![Create work items](/images/manual/use-cases/plane-create-work-item.png#bordered){width=70%}

5. 点击 **Save**。
6. 重复这些步骤来创建本场景中剩余的任务工作项。
7. 如需为工作项添加更多上下文，点击它以打开详情页，在那里你可以附加文件、添加子工作项或发表评论。

   ![Work item details page](/images/manual/use-cases/plane-work-item-details.png#bordered)

### 跟踪进度

无论你是进行每日站会、检查即将到来的截止日期，还是寻找日程冲突，你都需要以不同的方式查看数据。

使用 **Work items** 页面右上角的布局图标来切换视图，获取你需要的洞察。

![Layout views](/images/manual/use-cases/plane-layouts.png#bordered)

以下布局可用：
- **List Layout**：将任务分组为可折叠的部分（如 Todo 和 In Progress），让你可以快速一目了然地了解所有事项的状态。
- **Board Layout**：以看板卡片形式显示任务，让你可以轻松地将工作从一列拖放到另一列，随着进度推进。
- **Calendar Layout**：将任务绘制在传统的月度日历网格上，让你可以准确看到交付物的截止日期。
- **Table Layout**：提供类似电子表格的界面，带有独立的列，便于批量查看和更新优先级、负责人和标签。
- **Timeline Layout**：以甘特图风格的水平条形图映射任务持续时间，帮助你把握项目节奏并发现重叠的工作。

### 起草和共享项目资源

在团队中共享知识。通过在工作本身旁边编写和存储项目资源，让每个人保持一致。

1. 在左侧边栏中，点击 **Pages**。
2. 点击 **Create your first page** 或 **Add page**。
3. 输入新文档的标题，例如 `Revamp strategy for homepage 2026`。
4. 点击 **Create Page**。
5. 使用编辑器与团队协作起草文档。

   ![Document collaboration](/images/manual/use-cases/plane-documents.png#bordered)

6. 要将文档关联到工作项，点击右上角的复制链接图标，然后将其粘贴到工作项的描述中，以便负责人获得所需的上下文。
7. 要保存文档的本地副本，点击右上角的 <span class="material-symbols-outlined">more_horiz</span>，然后点击 **Export**。

## 了解更多

- [Official Plane documentation](https://docs.plane.so/)
