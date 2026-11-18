---
outline: [2, 3]
description: 在 Olares 上使用 wger 记录训练与营养，并通过 MCP 连接 Lares 或 Hermes Agent，查询日志和记录体重。
head:
  - - meta
    - name: keywords
      content: "Olares, wger, 训练日志, 体重, 营养记录, MCP, Lares, Hermes Agent, 自托管"
app_version: "0.0.10"
doc_version: "1.3"
doc_updated: "2026-11-18"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/wger.md)为准。
:::

# 用 wger 记录训练与营养

wger 将训练计划、训练日志、身体测量和营养记录集中在一款应用中。在 Olares 上，你可以把计划和记录保存在自己的设备上，随时查看。

先在浏览器中安排训练、记录实际完成的内容，并跟踪体重与饮食。之后还可以通过桌面或手机客户端访问同一份日志，或让 Lares 和 Hermes Agent 帮你查询记录。

## 安装 wger

1. 打开应用市场，搜索“wger”。
2. 点击 **Get**，然后点击 **Install**。按提示为 **ADMIN_PASSWORD** 设置至少 12 个字符的密码并妥善保存，等待安装完成。

## 使用 wger

从 Launchpad 打开 wger，在浏览器中登录。

:::tip 登录信息
用户名为 `admin`，密码是安装时设置的 **ADMIN_PASSWORD**。
:::

### 创建训练计划

1. 打开 **Training** > **Routines**，点击 **+**。
2. 在 **Add** 面板中设置：
   - **Name**：输入计划名称。
   - **Start** 和 **End**：选择计划的起止日期。
   - **Fixed weekly schedule**：如果训练日每周重复，保持开启。
3. 点击 **Save**，进入计划编辑页面。
4. 点击 **Add training day**，为训练日命名，并在训练日面板中点击 **Save**。
5. 点击 **Add exercise**。在新的动作区块中设置：
   - **Search by exercise name**：搜索并选择动作。
   - **Sets**：输入计划完成的组数。
   - **Weight** 和 **Repetitions**：填写每组的目标重量和次数。

返回 **Training** > **Routines**，重新打开该计划，确认训练日、动作和计划组数都已保留。

![包含一个动作的训练计划](/images/manual/use-cases/wger-routine.jpg)

### 记录一次训练

1. 打开 **Training** > **Routines**，选择要记录的计划。
2. 找到已完成的训练日，点击旁边的 <i class="material-symbols-outlined">add_chart</i>，打开训练日志。
3. 在训练日志中检查：
   - **Date**：设置实际训练日期。
   - **Exercises**：将预填的重量和次数改为实际完成的数值。只有填写了重量或次数的行会保存。
4. 点击动作列表下方的 **Submit** 保存组数。出现 **Success!** 即表示提交成功。
5. 在计划中检查已保存的训练会话和动作记录。
6. 如需添加备注或训练感受，使用 **Exercises** 上方的独立表单：
   - 填写备注或选择 **General impression**。
   - 点击该表单的 **Submit**。

将目标值留在训练计划中，将实际完成的内容记入日志。这样既能重复使用计划，也能保留之前的训练记录。

### 跟踪体重

1. 打开 **Body weight** > **Weight overview**。
2. 点击 **+** 打开 **Add** 面板，设置：
   - **Weight**：输入测量值，选择 **KG** 或 **LB**。
   - **Date** 和 **Time**：核对测量时间。
3. 点击 **Submit**，确认测量值出现在图表下方的表格中。

   ![添加体重记录](/images/manual/use-cases/wger-weight-entry.png#bordered)

4. 后续测量时继续添加记录，并使用图表上方的筛选器查看变化。

### 记录营养日记

营养日记需要食材和营养计划。先创建这两项，再记录实际摄入。

**添加食材**

1. 打开 **Nutrition** > **Ingredient overview**，点击 **Add**。
2. 填写食材每 100 g 的营养数据：
   - **Name**：输入食材名称。
   - **Energy**、**Protein**、**Carbohydrates** 和 **Fat**：根据食品营养标签填写。
   - **License**：为食材选择许可协议。
3. 点击 **Save**。

**记录饮食**

1. 打开 **Nutrition** > **Nutrition plans**，点击 **+**。在计划面板中设置：
   - **Description**：输入计划名称。
   - **Only track calories**：保持默认开启，以便在不设置预定餐食的情况下记录饮食。
2. 提交计划并打开它，点击 **Add nutrition diary entry**。
3. 在日记面板中搜索并选择刚添加的食材，填写摄入量和日期。
4. 点击 **Submit**。在 **Logged** 区域核对日期和营养合计。
5. 点击日期打开当日饮食日记，确认食材和摄入量已记录。

   ![显示食材、摄入量和营养数据的饮食日记](/images/manual/use-cases/wger-nutrition-diary.png#bordered)

## 连接桌面或手机客户端（可选）

要在其他设备上查看日志，将 wger 客户端连接到你的 Olares。

### 准备工作

- 已安装 wger Chart 0.0.10 或更新版本。
- 已在客户端设备上安装 LarePass。

### 下载客户端

| 平台 | 下载方式 |
| --- | --- |
| Windows 和 macOS | [wger 客户端发布页](https://github.com/wger-project/flutter/releases) |
| Linux | [wger 客户端发布页](https://github.com/wger-project/flutter/releases)或 [Flathub](https://flathub.org/apps/de.wger.flutter) |
| Android | [Google Play](https://play.google.com/store/apps/details?id=de.wger.flutter)、[F-Droid](https://f-droid.org/packages/de.wger.flutter/)或 [wger 客户端发布页](https://github.com/wger-project/flutter/releases) |
| iPhone | [App Store](https://apps.apple.com/app/id6502226792) |

从 GitHub 发布页下载时，选择与你的操作系统对应的安装包。

### 连接 Olares 上的日志

1. 在客户端设备上启用 LarePass VPN。
2. 在 Olares 设置中，打开 **Applications** > **wger** > **Entrances** > **wger Mobile**。保持 **Authentication level** 为 **Internal**，然后复制 **Endpoint**。

   ![Olares 设置中的 wger Mobile Endpoint](/images/manual/use-cases/wger-mobile-endpoint.jpg#bordered)

3. 在 wger 客户端的登录页面展开 **Advanced**，设置：
   - **Server**：选择 **Self-hosted**。
   - **URL of the wger instance**：填入刚复制的 Endpoint。
   - **Sign-in method**：选择 **Username and password**。
   - **Allow self-signed certificates**：保持关闭。

   ![wger 客户端的自托管服务器设置](/images/manual/use-cases/wger-client-advanced.jpg#bordered){width=70%}

4. 点击 **Done**，使用 `admin` 和安装时设置的密码登录。
5. 客户端同步期间保持 LarePass VPN 开启，确认浏览器中创建的训练计划也出现在客户端。

## 使用 AI 助手查询 wger（可选）

通过 MCP 将 AI 助手连接到 wger，查询日志或记录体重。以下示例使用 Lares 和 Hermes Agent。MCP 需要 wger Chart 0.0.10 或更新版本。

:::info MCP 使用哪个账号？
这个预览版连接始终访问 wger 的 `admin` 账号，与客户端使用的 wger 账号无关。要查询上面创建的记录，请将它们保存在 `admin` 账号中。持有 MCP 令牌的人可以访问该账号的记录及已开放的新增工具。
:::

先完成 [Lares 设置](lares.md#开始使用)或 [Hermes Agent 设置](hermes.md#配置-hermes-agent)，并确保 [Router](olares-router.md) 中有可用的聊天模型。

### 获取连接信息

连接助手需要两个值：

- **MCP endpoint** (`WGER_MCP_ENDPOINT`)：在 Olares 设置中打开 **Applications** > **wger** > **Entrances** > **wger MCP**，复制 **Endpoint**。
- **MCP token** (`WGER_MCP_TOKEN`)：在文件管理器中打开 **Data** > **wger** > **data** > **state** > **mcp-credentials** > `auth.private`，复制 `MCP_STATIC_TOKEN` 的值。不要使用同一文件中的 wger API 密钥。

### 连接 AI 助手

选择下面的一种助手。后续查询和写入示例均适用。

<Tabs>
<template #Lares>

1. 打开 Lares，进入 **Settings** > **MCP**。
2. 点击 **Add server**，填写连接信息：

   | 字段 | 填写内容 |
   | --- | --- |
   | **Server name** | `wger_lab` |
   | **Transport** | `Streamable HTTP` |
   | **MCP URL** | `WGER_MCP_ENDPOINT/mcp` |
   | **Headers** | 下方的 JSON 对象，将占位符替换为 MCP 令牌 |

   ```json
   {
     "Authorization": "Bearer WGER_MCP_TOKEN"
   }
   ```

   :::tip
   替换 `WGER_MCP_TOKEN` 时，请保留前面的 `Bearer` 和空格。
   :::

3. 点击 **Save**，保持服务器启用。Lares 会在保存后加载工具。如果出现加载错误，请检查 MCP 地址和令牌。
4. 新建会话，发送后面的示例请求。Lares 请求授权时，先核对工具和参数，再批准执行。

</template>
<template #Hermes-Agent>

1. 在文件管理器中打开 **Data** > **hermesagent** > **home**，备份 `config.yaml`。
2. 在 `mcp_servers` 下添加以下 `wger_lab` 配置。将 `WGER_MCP_ENDPOINT` 和 `WGER_MCP_TOKEN` 替换为刚获取的值，保留 Endpoint 后面的 `/mcp`。保留现有的服务器和模型设置；如果文件中已有 `mcp_servers`，请将条目添加到该键下，不要重复创建。

   ```yaml
   mcp_servers:
     wger_lab:
       url: "WGER_MCP_ENDPOINT/mcp"
       headers:
         Authorization: "Bearer WGER_MCP_TOKEN"
       connect_timeout: 30
       timeout: 60
       tools:
         include:
           - list_routines
           - get_routine
           - search_exercises
           - get_exercise
           - list_workout_logs
           - list_workout_sessions
           - list_nutrition_plans
           - list_log_items
           - nutrition_summary
           - get_body_weight_history
           - log_body_weight
   ```

3. 保存文件。如果使用[终端聊天](hermes.md#与-hermes-agent-对话)，请在当前会话中输入 `/exit`，再运行 `hermes`，加载新连接。
4. 每次发送一个请求，并在会话中查看 wger 工具返回的结果。

</template>
</Tabs>

### 查询日志

连接助手后，请它读取上面创建的记录。将方括号内的名称替换为自己的计划名称：

```text
找到 wger 中的“[训练计划名称]”，展示目前的训练会话和已完成的组数。
然后查看“[营养计划名称]”中记录的食物。如果某项没有记录，请直接说明，
不要估算数值。
```

![Lares 展示从 wger 读取的训练会话和已完成组数](/images/manual/use-cases/wger-lares-journal-result.png)

将回答与 wger 中的训练日志，以及营养计划的 **Logged** 区域对照。

### 记录体重

让助手保存一次测量。发送前，将方括号中的内容替换为自己的数据：

```text
我在[日期]称重[体重][公斤或磅]。请将这条记录添加到 wger，备注为
“[独特的备注]”。如果这条记录已经存在，就不要重复添加。完成后展示保存的记录。
```

在 wger 的 **Body weight** > **Weight overview** 中核对保存的记录。

## 了解更多

- [wger 官方文档](https://wger.readthedocs.io/en/latest/)：了解训练计划、营养记录等功能。
