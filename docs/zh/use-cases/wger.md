---
outline: [2, 3]
description: 在 Olares 上使用 wger 记录训练与营养，并通过 MCP 连接 Lares 或 Hermes Agent，查询日志和记录体重。
head:
  - - meta
    - name: keywords
      content: "Olares, wger, 训练日志, 体重, 营养记录, MCP, Lares, Hermes Agent, 自托管"
app_version: "0.0.10"
doc_version: "1.2"
doc_updated: "2026-10-31"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/wger.md)为准。
:::

# 用 wger 记录训练与营养

wger 是一款健身日志应用，将训练计划、训练记录、身体测量和营养记录集中在浏览器中。在 Olares 上，你可以将计划与记录保存在自己的设备上，在每次训练之间随时查阅。

通过 MCP 集成，你还可以让 Lares 或 Hermes Agent 用自然语言查询日志和添加记录。

:::info 应用包版本
客户端同步和 MCP 功能需要 wger Chart 0.0.10 或更新版本。
:::

## 学习目标

- 创建包含训练日和动作的训练计划。
- 记录实际完成的组数，跟踪体重变化。
- 添加食材并记录营养日记。
- 将桌面或手机客户端连接到 Olares 上的健身日志。
- 将 AI 助手连接到 wger，汇总日志并保存体重记录。

## 安装并登录

1. 打开 Olares Market，搜索 **wger**。
2. 点击 **Get**，然后点击 **Install**。
3. 为 **ADMIN_PASSWORD** 设置至少 12 个字符的密码，并妥善保存，用于首次登录。
4. 安装完成后，从 Launchpad 打开 wger。
5. 使用用户名 `admin` 和安装时设置的密码登录。

安装密码仅用于初始化空数据库。更新应用会保留已有账号与记录。本应用包关闭了公开注册和邮件发送。

## 连接桌面或手机客户端

你可以在电脑浏览器中使用 wger，也可以将官方客户端连接到同一台 Olares。各端登录同一个 wger 账号，即可使用同一份健身日志。

### 下载客户端

| 平台 | 下载方式 |
| --- | --- |
| Windows | [GitHub Releases](https://github.com/wger-project/flutter/releases)，选择 `windows-x64.zip`。 |
| macOS | [GitHub Releases](https://github.com/wger-project/flutter/releases)，选择 `wger.app.zip`。 |
| Linux | [GitHub Releases](https://github.com/wger-project/flutter/releases)，选择 `linux-x64.tar.gz`，也可以从 [Flathub](https://flathub.org/apps/de.wger.flutter) 安装。 |
| Android | [GitHub Releases](https://github.com/wger-project/flutter/releases)，选择 `app-release.apk`，也可以使用 [Google Play](https://play.google.com/store/apps/details?id=de.wger.flutter) 或 [F-Droid](https://f-droid.org/packages/de.wger.flutter/)。 |
| iPhone | [App Store](https://apps.apple.com/app/id6502226792)，是否可下载取决于商店地区。 |

在 GitHub 发布页展开对应版本的 **Assets**，即可查看安装包。Windows、macOS、Linux 和 Android 均可从这里下载。iPhone 请使用 App Store，`Runner.app.zip` 和 `Runner.xcarchive.zip` 属于开发产物，不能作为普通 iPhone 的直接安装包使用。

### 设置服务器并登录

1. 在运行 wger 客户端的电脑或手机上安装 LarePass，使用该 Olares 的 ID 登录并开启专用网络（VPN）。桌面端点击头像，在菜单中打开 **VPN connection**。手机端进入 **Settings**，在 **My Olares** 中开启 VPN。
2. 在 Olares 设置中打开 **Applications** > **wger** > **Entrances**，复制 **wger Mobile** 的 HTTPS 地址。虽然名称为 Mobile，桌面端和手机端都使用这个入口。保持其访问级别为 **internal**。
3. 打开 wger 客户端，在登录页展开 **ADVANCED**，将自定义服务器设为刚复制的地址。只填写服务器根地址，例如 `https://YOUR-WGER-CLIENT-HOST`，不要附加 `/api`、`/mcp` 或 `/ps`。
4. 使用 wger 用户名和密码登录。初始账号为 `admin`，密码是安装时设置的密码。wger 账号与 LarePass 登录账号分别使用。

连接和同步期间保持 LarePass VPN 开启。如果不修改默认服务器，客户端会连接官方 wger 服务，而非你的 Olares。不要将独立的 **wger MCP** 地址填作客户端服务器。

### 在各端使用同一份日志

登录后，等待首次同步完成，再打开已有记录。要检查是否连接了正确的服务器，可以在客户端新增一个自定义测量分类，然后用同一账号打开网页版，在测量分类中查找它。

本应用包包含 PowerSync，用于在已连接的客户端与 Olares 之间同步健身日志。电脑和手机使用相同的服务器与 wger 账号。同步期间保持 LarePass VPN 开启，并在同步完成后切换设备。

## 创建第一个训练计划

截图中的记录均为虚构演示数据。使用时请替换为自己的训练计划和测量值。

1. 打开 **Training**，进入训练计划概览。
2. 点击 **+**，输入名称，例如 `每周力量训练`，并设置起止日期。
3. 保持 **Fixed weekly schedule** 开启，使训练日按周重复，然后点击 **Save**。
4. 打开计划菜单，选择 **Edit**，点击 **Add training day**，为训练日命名，例如 `全身训练 A`。
5. 为训练日添加动作。在动作库中搜索并选择动作，填写组数、重复次数和重量，然后保存。

计划页面按训练日显示动作。下面的示例包含深蹲，设置为 3 组，每组 8 次，重量 40 kg。这些数值仅用于演示界面，不代表训练建议。

![包含一个动作的训练计划](/images/manual/use-cases/wger-routine.jpg)

首次启动时会同步动作库。如果搜索没有结果，请检查应用的网络连接，并等待首次同步完成。

## 记录一次训练

1. 在训练计划中，点击训练日旁的 **Add training log**。
2. 检查日期，按组填写实际完成的次数和重量。
3. 提交动作记录。如需填写训练备注和感受，在训练会话区域填写后，再提交该区域。

将目标值保留在计划中，将实际结果写入日志。这样可以重复使用同一计划，同时保留之前的训练记录。

## 跟踪体重

1. 打开 **Body weight** > **Weight overview**。
2. 点击 **+**，输入测量值和日期，然后保存。
3. 后续测量时继续添加记录。使用图表上方的时间筛选查看不同时间段。

图表显示单次测量值、平均值和趋势。下方表格保留每条记录的日期和备注。

![体重记录与趋势](/images/manual/use-cases/wger-weight.jpg)

## 记录营养摄入

### 添加食材

全新安装后，本地食材库可能为空。你可以根据食品营养标签添加常用食材。

1. 打开 **Nutrition** > **Ingredient overview**，点击 **Add**。
2. 输入食材名称，以及每 100 g 的能量、蛋白质、碳水化合物和脂肪含量。
3. 补充已知的其他数值，注明来源，然后保存。

练习时可使用明确标注为虚构的名称，例如 `演示燕麦（虚构）`。截图中的演示食材每 100 g 含 380 kcal 能量、12 g 蛋白质、60 g 碳水化合物和 8 g 脂肪。记录真实食品时，请使用其营养标签上的实际数值。

### 添加营养日记

1. 进入营养计划概览，新建一个计划，例如 `饮食日记`。
2. 打开计划，点击 **Add nutrition diary entry**。
3. 搜索并选择刚添加的食材，填写数量和日期。
4. 提交后，向下滚动到 **Logged**，查看能量和宏量营养素的合计。

![已记录的能量与宏量营养素](/images/manual/use-cases/wger-nutrition.jpg)

计划中的预设餐食与 **Logged** 区域相互独立。添加日记会更新已记录的摄入量，不会同时添加预设餐食。

## 通过 AI 助手使用日志

MCP 为 AI 助手提供操作 wger 记录的工具。下面的示例使用前文在浏览器中创建的训练计划、训练日志、营养日记和体重记录。你可以选择同一台 Olares 上的 Lares 或 Hermes Agent。

先完成 [Lares 设置](lares.md#开始使用)或 [Hermes Agent 设置](hermes.md#配置-hermes-agent)，并确保 [Router](olares-router.md) 中有可用的聊天模型。Hermes 可以使用 `default-chat`。如果 Lares 的模型选择器只显示具体模型，请选择 Router 默认聊天模型对应的完整模型名称。

模型连接和 wger 连接分别配置。Router 提供模型，下面的 MCP 连接用于让助手访问你的健身日志。

### 获取 MCP 连接信息

1. 在 Olares 设置中，打开 **Applications** > **wger** > **Entrances**，找到 **wger MCP** 入口。复制其地址。如果地址末尾尚未包含 `/mcp`，请补上该路径。下文用 `WGER_MCP_URL` 表示这个完整地址。它与 wger 网页入口不是同一个地址。
2. 打开文件管理器，进入 **Data** > **wger** > **data** > **state** > **mcp-credentials**。打开 `auth.private`，只复制 `MCP_STATIC_TOKEN` 的值，下文用 `WGER_MCP_TOKEN` 表示。

将令牌填入客户端的连接设置，不要粘贴到聊天中。这个文件还包含 MCP 服务内部使用的 wger API 密钥，下文的客户端配置不需要该密钥。

:::info 日志访问范围
预览版连接使用 wger 的 `admin` 账号。持有 MCP 令牌的客户端可以访问该账号的记录，并使用已开放的新增工具。这不是为每个 wger 用户分别登录的授权方式。MCP 入口使用 Olares internal 访问方式和 Bearer 认证，以下示例中的助手均运行在同一台 Olares 上。
:::

### 连接 Lares

1. 打开 Lares，进入 **Settings** > **MCP**。
2. 点击 **Add server**，填写连接信息：

   | 字段 | 填写内容 |
   | --- | --- |
   | **Server name** | `wger_lab` |
   | **Transport** | `Streamable HTTP` |
   | **MCP URL** | 你的 `WGER_MCP_URL` |
   | **Headers** | 下方的 JSON 对象，将占位符替换为你的 MCP 令牌 |

   ```json
   {
     "Authorization": "Bearer WGER_MCP_TOKEN"
   }
   ```

3. 点击 **Save**，并保持服务器启用。Lares 会在保存后加载工具。如果显示加载错误，请先检查 MCP 地址和令牌，再继续操作。
4. 新建会话，发送下文的示例请求。Lares 请求授权时，先检查调用的 wger 工具及其参数，再批准执行。

### 连接 Hermes Agent

1. 在文件管理器中，进入 **Data** > **hermesagent** > **home**，备份 `config.yaml`。
2. 在 `mcp_servers` 下添加以下 `wger_lab` 配置，将 `WGER_MCP_URL` 和 `WGER_MCP_TOKEN` 替换为前面获取的地址和令牌。保留已有服务器和模型设置。如果文件中已有 `mcp_servers`，请将条目添加到其下方，不要重复创建该键。

   ```yaml
   mcp_servers:
     wger_lab:
       url: "WGER_MCP_URL"
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

3. 保存文件。在当前 Hermes 终端会话中输入 `/exit` 退出，再运行 `hermes`，使新连接生效。这个示例使用[终端聊天](hermes.md#与-hermes-agent-对话)，不需要启用 Gateway API。
4. 每次发送一个请求，并在会话中查看 wger 工具返回的结果。

这份配置开放了下文使用的读取工具，以及一个写入工具 `log_body_weight`。如果希望 Hermes 只能读取 wger 数据，从列表中移除 `log_body_weight` 即可。预览版 MCP 服务也没有开放更新和删除工具。

### 查询训练与营养记录

两个助手都可以使用以下请求。请将计划名称和日期替换为你自己的数据：

```text
帮我看看 wger 里“每周力量训练”这个计划在 2026 年 10 月 27 日到 31 日
练得怎么样。每天做了多少组、多少次、用了多重？这几天的训练量合起来
是多少？其他计划不用算，我现在只想回顾一下记录。
```

你不需要知道工具名称或记录 ID，助手会查找计划并选择工具。提供计划名称和日期范围，可以帮助它排除其他计划的记录。请对照 wger 的训练日志检查日期和组数。不同单位的重量需要先换算，不能直接相加。

查询营养摄入时，先按名称选择计划，再要求逐日汇总：

```text
按 wger 里“饮食日记”的记录，我在 2026 年 10 月 29 日、30 日和 31 日
每天吃了多少热量、蛋白质、碳水和脂肪？三天合起来也帮我算一下。
```

请使用你实际记录饮食的日期。助手会选择对应的营养计划并读取每天的记录。打开 wger 中对应计划的 **Logged** 区域，将回答与原始记录对照。

汇总只反映已经录入的食物。如果助手额外评价是否达到营养目标，也不能仅因日记漏记就判断实际摄入不足。

### 保存体重记录

使用你想记录的测量值，并提供日期、单位和一条便于查找的备注，让助手先检查是否已有记录。例如：

```text
我 2026 年 10 月 31 日称了 72.2 公斤，帮我记到 wger，
备注写“体重日记 2026-10-31”。如果这条已经记过，就别重复添加了。
记好后给我看看。
```

在 Lares 中，先核对 `log_body_weight` 调用中的日期、体重、单位和备注，再批准执行。随后打开 wger 的 **Body weight** > **Weight overview**，在表格中查找该记录。只提供日期而未指定时间时，记录时间为应用时区的当天中午 12 点。

如果写入后请求迟迟没有结束，先让助手读取历史，再决定是否重试。备注只是查找记录的依据，服务端不会强制保证备注唯一。

### 把日志变成复盘和挑战

积累几次训练和饮食记录后，可以让助手把不同记录联系起来，帮你理解变化。下面的示例使用前面 Hermes 配置中已有的读取工具，不需要额外开放写入权限。使用时替换为自己的计划名称和日期。

#### 看看进步来自哪里

```text
我最近深蹲有进步吗？帮我看看 wger 里“每周力量训练”这个计划截至
2026 年 10 月 31 日最近三次深蹲的记录，把日期、组数、次数、重量和
训练量放在一起比较。从第一次到最后一次变化了多少？主要是加了重量、
多做了次数，还是多做了组数？别把其他计划混进来，也不用估算最大力量。
```

这个请求可以把分散的训练日志整理成一条进步脉络。参考变化百分比前，先核对各次训练记录。如果不足三次，让助手展示已有记录并说明缺少哪些数据，不要补造。

#### 对照训练日和其他日子的饮食

```text
帮我把 2026 年 10 月 29 日到 31 日“每周力量训练”的训练日期，和 wger 里
“饮食日记”的记录对起来看看。有训练记录和没有训练记录的日子，各有几天
记了饮食？平均每天多少热量、多少蛋白质？没记饮食的日子别当成没吃饭。
我只想看看记录里有什么规律，不用推断这些习惯导致了体重变化。
```

助手需要关联训练日期和饮食条目，再分别计算两类日期的平均值。没有训练日志，只能说明当天没有记录训练。时间太短或日记不完整时，也可能不足以看出有意义的差异。

#### 生成一张本周成就卡

```text
用 wger 里“每周力量训练”和“饮食日记”在 2026 年 10 月 27 日到 31 日的
记录，给我做一张简短的“本周成就卡”吧。最多三个小成就，每个都附上
数字和日期，没有记录支持就别硬发奖。再想一个有趣的下周记录挑战，
让我把日记记得更完整。重点是记录习惯，不是加重量或者少吃。
原来的日志不要改。
```

成就卡是助手在会话中生成的文字总结。例如，它可以根据真实日期，表扬你完成了几次训练记录或连续记录了饮食。助手可以提出挑战，但这个请求不会自动设置提醒，也不会在 wger 中创建新的挑战功能。

请将成就卡与各次训练明细对照。助手可能在对比中算对数据，却在最后的总结中数错总数。发现对不上时，可以继续这样问：

```text
成就卡里的总组数和前面的训练表格对不上。帮我重新核对原始记录，
把成就卡修正一下，只保留能从记录里确认的成就。
```

## 持续使用日志

账号和日志记录保存在 PostgreSQL 中。上传的媒体与应用状态保存在应用存储中。更新应用时，请保留这些数据。

浏览器与已连接的客户端使用同一份账号记录。备份或更新应用时，请保留独立 PostgreSQL 数据、上传的媒体和应用状态。

## 官方资源

- [wger 官方网站](https://wger.de/)
- [官方文档](https://wger.readthedocs.io/en/latest/)
- [服务端 GitHub 仓库](https://github.com/wger-project/wger)
- [客户端 GitHub 仓库](https://github.com/wger-project/flutter)
- [GitHub 客户端下载页](https://github.com/wger-project/flutter/releases)
