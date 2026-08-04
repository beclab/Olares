---
outline: [2, 3]
description: 在 Olares 上使用 TREK 协作规划旅行。创建行程、管理预算、与朋友分享，并将旅行计划导出为 PDF。
head:
  - - meta
    - name: keywords
      content: Olares, TREK, NOMAD, trip planner, travel planning, collaborative, itinerary, budget, packing list, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-16"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/trek.md)为准。
:::

# 使用 TREK (NOMAD) 协作规划旅行

TREK（前身为 NOMAD）是一个自托管的实时协作旅行规划器。它将交互式地图、详细行程、预算、打包清单和团队功能整合到一个应用中。在 Olares 上运行 TREK 可确保您的所有旅行数据保持私密，同时让您与朋友和家人一起规划旅行。

## 学习目标

在本指南中，您将学习如何：
- 在 Olares 上安装和设置 TREK。
- 构建旅行计划，包括每日日程、预算和打包清单。
- 邀请朋友并实时协作规划旅行。
- 保护您的账户安全并备份旅行数据。
- 配置高级设置，例如第三方单点登录（SSO）和地图 API 密钥。

## 安装 TREK

1. 打开 Market 并搜索 "TREK"。

   ![TREK](/images/manual/use-cases/trek.png#bordered)

2. 点击 **Get**，然后点击 **Install**。
3. 当提示时，设置环境变量：
   - **ADMIN_EMAIL**：您的管理员邮箱地址。
   - **ADMIN_PASSWORD**：您的管理员密码。
   
   :::info 密码要求
   密码必须至少 8 个字符，并包含大写字母、小写字母和数字。
   :::

4. 点击 **Confirm** 并等待安装完成。

## 设置 TREK

1. 从 Launchpad 打开 TREK，然后使用安装期间设置的邮箱和密码登录。
2. 首次登录时，TREK 要求您重置密码。输入新密码，然后点击 **Update password**。

   :::warning
   由于 TREK 是一个私有的自托管应用，它不使用自动电子邮件密码恢复系统。如果您忘记了更新后的管理员密码，您的账户将无法恢复。为防止永久失去对工作空间和旅行数据的访问权限，请确保将您的管理员密码安全存放，例如存放在密码管理器中。
   :::

## 使用 TREK

### 创建旅行计划

1. 在主页上，点击 **Create First Trip**。

   ![Create first trip](/images/manual/use-cases/trek-create-trip.png#bordered)

2. 指定旅行详情。
   - **Cover Image**：上传旅行的封面图片。
   - **Title**：指定旅行名称，例如 `Paris Summer 2026`。
   - **Description**：输入旅行描述，例如整体主题或目标。
   - **Dates**：选择旅行的开始和结束日期。
   - **Number of Days**：选择旅行持续时间。

3. 点击 **Create New Trip**。行程将显示在 **My Trips** 页面上。

   ![First trip created](/images/manual/use-cases/trek-trip-created.png#bordered)

### 规划每日行程

通过添加地点并将它们组织到每日日程中来构建逐日计划。

1. 点击新创建的行程以打开旅行规划器，在那里您可以开始添加地点和活动。

   ![Trip planner](/images/manual/use-cases/trek-trip-planner.png#bordered)

2. 点击 **Add Place/Activity**。
3. 输入要搜索的地点，例如 `Eiffel Tower`，点击 <i class="material-symbols-outlined">search</i>，从结果列表中选择目标地点，然后点击 **Add**。

   该地点将显示在旅行规划器的右侧面板中。

   ![Add a place](/images/manual/use-cases/trek-place-added.png#bordered)   

4. 将地点拖入行程中的特定日期。

   例如：
      - **Day 1**：Eiffel Tower, Trocadero Gardens
      - **Day 2**：Louvre Museum, Tuileries Garden
      - **Day 3**：Notre-Dame Cathedral, Sainte-Chapelle, Latin Quarter

5. 通过拖放在一天内重新排序地点。
6. 将活动跨天拖动以将其移到新日期。
7. 点击地点以添加备注或在交互式地图上查看它。

   ![Itinerary view](/images/manual/use-cases/trek-itinerary.png#bordered)

:::tip 路线优化
选择 **Optimize** 以自动重新排序一天内的地点，实现最高效的路径。您还可以将路线导出到 Google Maps 进行导航。

   ![Optimize route](/images/manual/use-cases/trek-optimize-route.png#bordered){width=40%}
:::

### 添加旅行笔记

在行程中记录每日提醒、旅行想法或具体计划。

1. 在您的旅行规划器中，点击 **Plan** 选项卡。
2. 找到您想要添加笔记的特定日期，然后点击 <i class="material-symbols-outlined">docs</i>。
3. 选择与笔记主题匹配的图标。
4. 在 **Note** 字段中，输入简短的标题或摘要，例如 `Buy Metro tickets`。
5. 在 **Daily Note** 字段中，输入更多详情，例如 `Get a carnet of 10 tickets at the station before heading to the Louvre`。
6. 点击 **Add**。

   ![Add notes to days](/images/manual/use-cases/trek-add-note.png#bordered){width=40%}

### 查看天气预报

点击行程中的日期以查看该目的地的天气预报。TREK 通过 Open-Meteo 提供最长 16 天的预报（无需 API 密钥），对于更远的日期则提供历史气候平均值作为备用。

![Weather forecast](/images/manual/use-cases/trek-weather.png#bordered)

### 记录预订信息

在一个地方跟踪您的航班、住宿、餐厅和旅行预订。

1. 在您的旅行规划器中，点击 **Book** 选项卡。
2. 点击 **Manual Booking** 打开 **New Reservation** 窗口。
3. 选择 **BOOKING TYPE**，例如 **Flight**。
4. 指定预订详情。例如，对于酒店住宿：

   - **TITLE**：输入预订名称，例如 Hotel Le Meurice。
   - **LINK TO DAY ASSIGNMENT**：选择行程中的特定日期以关联此预订。
   - **DATE and END DATE**：指定您的入住和退房日期。
   - **STATUS**：选择预订的当前状态，例如 Pending 或 Confirmed。
   - **LOCATION / ADDRESS**：输入酒店地址。
   - **BOOKING CODE**：输入您的确认号码。
   - **FILES**：选择 **Attach file** 上传您的预订确认或电子机票。
   - **PRICE** 和 **BUDGET CATEGORY**：输入总费用以自动将此预订与您的旅行预算同步。

   <!--![Reservations](/images/manual/use-cases/trek-reservations.png#bordered)-->

5. 点击 **Add**。

### 附加旅行文件

通过将预订确认、电子机票和旅行保险文件直接附加到行程项目、地点或预订上来保持有序。每个文件最大支持 50 MB。

1. 在您的旅行规划器中，点击 **Files** 选项卡。
2. 上传要附加的文件。
3. 在 **Assign File** 窗口中，为文件添加备注，然后选择要关联文档的位置，例如特定日期或地点。

   ![Assign file](/images/manual/use-cases/trek-documents.png#bordered)

4. 关闭窗口。

### 跟踪旅行费用

使用基于类别的预算和多币种支持来跟踪旅行费用。

1. 在您的旅行规划器中，点击 **Budget** 选项卡。
2. 输入费用类别名称，例如 `Food`、`Transport`、`Accommodation` 或 `Activities`。

   ![Create budget category](/images/manual/use-cases/trek-budget-category.png#bordered)

3. 点击 <i class="material-symbols-outlined">add_2</i>。将显示预算规划器。

   ![Budget planner](/images/manual/use-cases/trek-budget-table.png#bordered)

4. 从右上角的下拉菜单中选择您的首选货币。
5. 指定费用详情：

   - **NAME**：输入项目名称，例如 `Dinner cruise on the Seine`。
   - **TOTAL**：输入总费用。
   - **PERSONS**：输入分摊费用的人数。
   - **DAYS**：输入费用的持续时间。
   - **DATE**：输入费用日期。
   - **NOTE**：输入额外上下文。

6. 选择行末的 <i class="material-symbols-outlined">add</i> 添加条目。

   TREK 自动计算 **PER PERSON**、**PER DAY** 和 **P. P / DAY** 金额，并更新右侧的总预算。

7. 要添加更多费用类别，在右侧面板中输入类别名称，然后点击其旁边的 <i class="material-symbols-outlined">add</i>。

   TREK 显示按类别划分的支出饼图分解。

   ![Budget management](/images/manual/use-cases/trek-budget.png#bordered)

### 构建打包清单

创建分类打包清单，分配责任并跟踪打包进度。

1. 在您的旅行规划器中，点击 **Lists** 选项卡。
2. 点击 **Add category**，输入类别名称，例如 `Clothing`、`Electronics` 或 `Toiletries`，然后点击行末的 <i class="material-symbols-outlined">check</i>。
3. 在您的类别下，输入要打包的物品，例如 `Walking shoes`，并为每个物品指定数量。
4. 要将类别分配给特定的旅行同伴，点击 <i class="material-symbols-outlined">person_add</i>。
5. 打包物品时选择其旁边的复选框。TREK 会在页面顶部更新您的整体打包进度。

   ![Packing list](/images/manual/use-cases/trek-packing-list.png#bordered)

6. 要为未来的旅行节省时间，选择右上角的 **Save as template** 保存当前清单。规划下一次旅行时，点击 **Apply template** 加载已保存的模板，以预填充的清单开始。

### 将行程导出为 PDF

计划准备好后，将其导出为 PDF 以与旅行同伴分享或打印以供离线参考。

1. 打开您要导出的行程。
2. 点击行程顶部的 **PDF**。

   ![Export plan as a PDF](/images/manual/use-cases/trek-export-pdf.png#bordered){width=40%}

3. 在弹出窗口中，点击 **Save as PDF**。

   TREK 生成包含封面页、逐日行程、图片和备注的 PDF。

## 与他人协作

### 邀请成员加入行程

与朋友和家人分享您的行程：生成只读查看的公开链接，或为您的旅行同伴设置用户账户以协作规划行程。

:::info 外部访问和安全
- 要邀请 Olares 网络之外的人，首先将应用的 **Authentication level** 设置为 **Public**，路径为 **Settings** > **Applications** > **TREK**。

   ![Authentication level of TREK](/images/manual/use-cases/trek-auth-level.png#bordered){width=70%}

- 将入口级别设置为 Public 使您的 TREK 登录页面可以从互联网上的任何地方访问。您的数据保持私密，但完全依赖 TREK 账户凭据进行保护。确保所有用户设置强密码。
:::

<Tabs>
<template #Option-1:-Share-an-invite-link>

生成只读链接，以便朋友或家人无需登录即可查看您的行程。

1. 打开一个行程，然后点击右上角的 **Share**。
2. 在 **Public Link** 下，选择您想要可见的行程模块，例如 **Map & Plan**、**Bookings** 或 **Packing**。
3. 点击 **Create link**。
4. 复制生成的链接并发送给您的旅行同伴。

![Invite link](/images/manual/use-cases/trek-invite-link.png#bordered)
</template>
<template #Option-2:-Add-collaborators>

为您的旅行同伴设置用户账户，然后邀请他们与您一起积极编辑和规划行程。

1. 点击右上角的用户头像，然后点击 **Admin**。
2. 在 **Users** 选项卡上，点击 **Create User**。

   ![Create user](/images/manual/use-cases/trek-create-user.png#bordered)

3. 在 **Create Users** 窗口中：

   a. 输入新成员的姓名、邮箱和密码。

   b. 选择要分配的角色。

   c. 点击 **Create User**。

4. 点击左上角的 **My Trips**，然后打开您想要分享的行程。
5. 点击右上角的 **Share**。
6. 在 **Share Trip** 窗口中，从 **Invite User** 列表中选择用户，然后点击 **Invite**。

   ![Invite user](/images/manual/use-cases/trek-invite-user.png#bordered)

   被邀请的成员登录后即可立即查看共享的行程。

   <!-- ![Share trip](/images/manual/use-cases/trek-share-trip.png#bordered) -->

   <!-- ![Synced trip](/images/manual/use-cases/trek-synced-trip.png#bordered) -->
</template>
</Tabs>

### 实时协作

当成员加入行程时，所有更改都会即时同步。进入行程的 **Collab** 选项卡以访问您的团队仪表板：
- **Chat**：向您的旅行群组发送实时消息。
- **Notes**：发布所有行程成员可见的笔记。
- **Polls**：创建投票以就团队决策进行表决。
- **What's next**：查看您的即将到来的行程。

![Team collaboration](/images/manual/use-cases/trek-collaboration.png#bordered)

## 下一步

- [配置 TREK 高级设置](trek-advanced-settings.md)。

## 常见问题

### 我忘记了 TREK 密码。如何重置？

恢复流程取决于您账户的角色。
- **对于成员**
   
   联系您的 TREK 管理员。管理员可以登录 TREK 并通过 **Admin** > **Users** 为您分配新密码。

- **对于管理员**
   - 如果您尚未更改初始密码，可以在 Control Hub 中查看安装期间设置的原始凭据：
   
      a. 进入 **Browse** > **trek-{username}** > **Deployments** > **trek**，然后点击 <i class="material-symbols-outlined">edit_square</i>。

      ![Trek in Control Hub](/images/manual/use-cases/trek-control-hub.png#bordered)
      
      b. 在 YAML 编辑器中，找到 `containers` 部分并定位 `ADMIN_EMAIL` 和 `ADMIN_PASSWORD` 环境变量。

      ![Trek credentials in Control Hub](/images/manual/use-cases/trek-env-vars.png#bordered)

   - 如果您已更改初始密码，则无法恢复。为防止失去对工作空间和旅行数据的访问权限，请确保将您的管理员密码安全存放，例如存放在密码管理器中。

### 地图搜索没有返回结果

TREK 默认使用 OpenStreetMap。要获得更全面的搜索结果，请在 **Admin** > **Settings** > **API Keys** 下添加 Google Places API 密钥。更多信息，请参阅[使用 Google API 密钥改进地图搜索](../use-cases/trek-advanced-settings.md#improve-map-search-with-google-api-keys)。

### 文件上传大小限制是多少？

每个文件最大支持 50 MB。

支持的格式包括 `.jpg`、`.jpeg`、`.png`、`.gif`、`.webp`、`.heic`、`.pdf`、`.doc`、`.docx`、`.xls`、`.xlsx`、`.txt` 和 `.csv`。

## 了解更多

- [TREK on GitHub](https://github.com/mauriceboe/NOMAD)
