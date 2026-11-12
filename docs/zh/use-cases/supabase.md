---
outline: [2, 3]
description: 在 Olares 上通过 Lares 和 Supabase 官方 MCP 配置应用的私有后端，再连接应用、自动化工作流、报表和资料问答。
head:
  - - meta
    - name: keywords
      content: Olares, Supabase, Lares, MCP, PostgreSQL, RLS, n8n, Metabase, Flowise, Code Server
app_version: "0.0.14"
doc_version: "2.1"
doc_updated: "2026-11-12"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/supabase.md)为准。
:::

# 用 Supabase 和 Lares 配置应用后端

Supabase 集成了 PostgreSQL、用户认证、文件存储、实时更新和 Edge Functions。你可以通过 Studio 在自己的 Olares 设备上管理这些服务。

Lares 可以调用 Supabase 官方的模型上下文协议（MCP）服务器，通过对话查看数据表、执行数据库迁移和配置行级安全策略（RLS）。下面以虚构的家庭装修项目为例，先创建私有笔记表，再连接一个在 Code Server 中运行的小应用。你也可以将 n8n、Metabase 和 Flowise 接入同一后端。

## 学习目标

- 将 Lares 接入 Supabase 官方 MCP 服务器。
- 让 Lares 为应用创建数据表和按用户隔离的访问策略。
- 使用普通用户凭据连接应用，并检查数据访问范围。
- 为其他集成选择正确的地址和凭据。

## 前提条件

- Lares 已选择可用的模型。参阅 [Lares 入门](lares.md#开始使用)。
- 已从应用市场安装 Code Server，用于运行笔记应用。

## 安装 Supabase

1. 打开应用市场，搜索“Supabase”。
2. 点击**获取**，然后点击**安装**，等待安装完成。应用会以 **Supabase Studio** 的名称出现在启动台。

每次安装提供一个自托管项目，无需 Supabase Cloud 账户。

## 在 Lares 中接入官方 MCP 服务器

1. 打开 Olares 设置，进入**应用** > **Supabase**，复制 **Supabase Studio** 的**访问地址**。
2. 在 Studio 地址后添加 `/api/mcp`，例如：

   ```text
   https://YOUR_STUDIO_HOST/api/mcp
   ```

   此连接使用 Studio 入口。后续 SDK 和其他应用使用独立的 Supabase API 入口。
3. 打开 Lares，进入 **Settings** > **MCP**，点击 **Add server**。
4. 填写以下配置：

   | 字段 | 值 |
   | --- | --- |
   | Server name | `supabase_official` |
   | Transport | `Streamable HTTP` |
   | MCP URL | 以 `/api/mcp` 结尾的 Studio URL |
   | Headers | `{}` |

5. 点击 **Save**。Lares 会连接服务器并加载工具。
6. 在 Lares 中新建对话，发送以下请求：

   ```text
   使用 supabase_official 直接调用 MCP 工具，列出 public schema 中的表，
   并返回这个项目的 API URL。不要修改数据。
   ```

   Lares 应返回数据表列表和 Supabase API URL。新项目的列表可能为空。

同一 Olares 实例上的 Lares 通过 Olares 应用身份完成认证。这个 MCP 连接不需要 Supabase Cloud 的个人访问令牌，也不需要 `service_role` key。保持 Studio 入口受 Olares 保护，不要为了连接 agent 而开放为公共访问。

:::warning 数据库管理权限
MCP 服务器拥有项目的数据库管理权限，可以修改表结构和访问策略。Lares 请求确认时，请检查 SQL，只批准预期的操作。应用用户将使用独立的 SDK 连接，并受 RLS 限制。
:::

仅需查看数据时，在 MCP URL 后添加 `?read_only=true`。服务器会禁用迁移并拒绝数据库写入。如需同时将工具范围限制为数据库操作，使用 `?read_only=true&features=database`。

## 通过 Lares 创建私有笔记表

应用使用笔记所有者的 Supabase Auth 用户 ID 保存数据。RLS 让登录用户只能读取和修改自己的记录。

1. 在 Lares 对话中描述数据表和策略：

   ```text
   使用 supabase_official 为我的应用准备私有笔记表。
   先检查 public schema，再创建 public.user_notes：
   - id：UUID 主键，默认 gen_random_uuid()
   - owner_id：UUID，必填
   - note：text，必填
   启用 RLS，将 SELECT、INSERT、UPDATE、DELETE 权限授予 authenticated。
   创建 own_notes 策略，FOR ALL TO authenticated，USING 和 WITH CHECK
   都使用 auth.uid() = owner_id。
   用 apply_migration 修改表结构，执行前先展示计划中的 SQL。
   只修改这张表，不删除现有数据，也不授予匿名访问权限。
   如果表已存在，先解释差异，再提出迁移方案。
   ```

2. 检查迁移方案，批准预期的工具调用。在 Lares 默认权限模式下，迁移和可写连接的 `execute_sql` 都需要确认，包括通过该工具发出的查询。
3. 打开 Supabase Studio 的 **Table Editor**，找到 `user_notes`，检查 `id`、`owner_id`、`note` 列和 RLS 标记。

   也可以让 Lares 检查数据表并概述策略。确认读取条件和写入检查都使用 `auth.uid()` 与 `owner_id` 进行比较。

::: details 替代方式：在 SQL Editor 中创建同一张表
如果不通过 Lares 创建，打开 Supabase Studio 的 **SQL Editor**，粘贴以下 SQL，点击 **Run**。两种方式选择一种即可。

```sql
create table public.user_notes (
  id uuid default gen_random_uuid() primary key,
  owner_id uuid not null,
  note text not null
);
alter table public.user_notes enable row level security;
grant select, insert, update, delete on public.user_notes to authenticated;
create policy own_notes on public.user_notes
for all to authenticated
using (auth.uid() = owner_id)
with check (auth.uid() = owner_id);
```
:::

为其他应用配置后端时，可以把所需表结构或迁移文件交给 Lares，并说明各张表的访问范围。先让它检查现有数据，再提出修改方案。官方 MCP 负责数据库操作，应用设置、Auth 用户创建、SMTP 和登录跳转 URL 仍需通过相应设置页面或 API 配置。

## 在 Code Server 中连接应用

可以先让 Lares 帮你准备示例：

```text
帮我将私有笔记应用接入 Supabase。使用 supabase_official 检查 user_notes
及其 RLS 策略，获取 API URL 和 anon key。
按照下面的 SDK 示例准备 notes.mjs 和 .env 模板。
不要在对话或源码中填写密码。准备好文件并说明后续步骤后就停止。
不要创建 Auth 用户、运行应用或测试双用户隔离；这些操作由我按下文完成。
```

此阶段 Lares 只检查现有数据表和策略，并准备文件内容。将文件复制到 Code Server，然后按下文自行创建用户、运行应用并验证隔离。直接写入文件或执行命令需要另外授权的工作区工具。

### 获取 API 连接信息并创建用户

1. 使用 Lares 返回的 API URL 和 legacy `anon` key。对应工具是 `get_project_url` 和 `get_publishable_keys`。SDK 使用 Supabase API URL，不添加 `/rest/v1` 或 `/api/mcp`。不要使用 `service_role` key。
2. 打开 Supabase Studio，进入 **Authentication** > **Users** > **Add user** > **Create new user**。
3. 填写邮箱和自定密码，为本地示例保持 **Auto confirm user?** 开启，然后创建用户。

   邮箱可使用 `demo@example.com`。这会创建一个 Supabase Auth 应用用户，自动确认让示例无需发送邮件即可运行。

::: details 替代方式：从 Studio 和设置中复制连接信息
在 Olares **设置** > **应用** > **Supabase** > **Supabase API** 中复制**访问地址**。

在 Supabase Studio 中进入 **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**，复制 `anon` key。
:::

### 运行笔记应用

1. 从启动台打开 Code Server，选择 **Terminal** > **New Terminal**。编辑器不包含 Node.js 运行环境，通过 nvm 安装 Node.js 22 或更高版本：

   ```bash
   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.8/install.sh | bash
   \. "$HOME/.nvm/nvm.sh"
   nvm install 22
   ```

   如果 `node -v` 提示 `libatomic.so.1: cannot open shared object file`，运行 `sudo apt-get install -y libatomic1`，然后再次检查 `node -v`。
2. 在项目终端中安装 SDK：

   ```bash
   npm install @supabase/supabase-js@2.118.0
   ```

3. 在项目目录创建 `.env`，填入 API URL、`anon` key 和刚创建的用户凭据：

   ```dotenv
   SUPABASE_URL=https://YOUR_SUPABASE_API_HOST
   SUPABASE_ANON_KEY=YOUR_LEGACY_ANON_KEY
   SUPABASE_USER_EMAIL=YOUR_USER_EMAIL
   SUPABASE_USER_PASSWORD=YOUR_USER_PASSWORD
   ```

4. 将 `.env` 加入 `.gitignore`，然后创建 `notes.mjs`：

   ::: details 示例：notes.mjs
   ```js
   import { createClient } from '@supabase/supabase-js';

   const client = createClient(
     process.env.SUPABASE_URL,
     process.env.SUPABASE_ANON_KEY,
     {
       auth: { persistSession: false },
       realtime: { transport: WebSocket }
     }
   );

   const { data: session, error: loginError } =
     await client.auth.signInWithPassword({
       email: process.env.SUPABASE_USER_EMAIL,
       password: process.env.SUPABASE_USER_PASSWORD
     });
   if (loginError) throw loginError;

   if (process.env.NOTES_READ_ONLY !== '1') {
     const { error: insertError } = await client.from('user_notes').insert({
       owner_id: session.user.id,
       note: process.env.NOTES_TEXT ||
         'Ask the Lakeview contractor to confirm the delivery window.'
     });
     if (insertError) throw insertError;
   }

   const { data: notes, error: readError } =
     await client.from('user_notes').select('id, note');
   if (readError) throw readError;
   console.table(notes);

   const { data: greeting, error: functionError } =
     await client.functions.invoke('hello', { body: { name: 'Lakeview' } });
   if (functionError) throw functionError;
   console.log(greeting);
   await client.auth.signOut();
   ```
   :::

5. 运行应用：

   ```bash
   node --env-file=.env notes.mjs
   ```

   终端会显示当前用户的笔记，以及内置 `hello` Edge Function 返回的结果。
### 验证双用户数据隔离

1. 在 Studio 中创建第二个 Auth 用户，使用不同的邮箱和密码，并保持 **Auto confirm user?** 开启。
2. 将 `.env` 中的 `SUPABASE_USER_EMAIL` 和 `SUPABASE_USER_PASSWORD` 换成第二个用户的凭据。在该用户插入任何笔记之前运行：

   ```bash
   NOTES_READ_ONLY=1 node --env-file=.env notes.mjs
   ```

   如果第二个用户此前没有笔记，表格应为空；无论如何都不应出现第一个用户的 Lakeview 笔记。
3. 以第二个用户插入一条可区分的笔记：

   ```bash
   NOTES_TEXT='Riverside user note' node --env-file=.env notes.mjs
   ```

   输出应包含 `Riverside user note`，且不包含第一个用户的 Lakeview 笔记。
4. 在 `.env` 中恢复第一个用户的凭据，再次运行第 2 步的只读命令。输出应包含 Lakeview 笔记，且不包含 `Riverside user note`。这两次查询共同表明每个登录用户只能读取自己的记录；只读运行不会重复插入笔记。

不要通过把 `anon` 换成 `service_role` 来解决 RLS 错误，后者会绕过应用依赖的访问策略。

## 将更多应用接入同一后端

以下集成使用虚构的家庭装修项目数据。开始对应章节前，先从应用市场安装相应应用。Lares 可以准备数据库对象，各应用内的连接配置仍需单独完成。

每个示例都提供了可以先发给 Lares 的请求。Supabase MCP 负责数据库侧操作，直接配置 n8n、Metabase 或 Flowise 还需要对应应用已授权的工具或 API 连接。没有这些连接时，让 Lares 准备配置，再自行填写下文中的字段。服务端凭据保存在目标应用的凭据管理中。

### 收集连接信息

根据连接用途选择地址和凭据：

| 连接 | 地址 | 凭据 |
| --- | --- | --- |
| Lares MCP | Studio HTTPS URL + `/api/mcp` | 同一实例上的 Olares 应用身份，无需额外请求头 |
| SDK 应用 | Supabase API HTTPS URL | `anon` key 和登录用户 |
| n8n 和 Flowise | Supabase API HTTPS URL | 保存在应用服务端凭据中的 legacy `service_role` key |
| Metabase | Supabase 应用主机和导出的 session 端口 | `postgres.supabase` 和数据库密码 |

#### 获取连接地址

1. 打开 Olares **设置** > **应用** > **Supabase**，复制 **Supabase API** 的**访问地址**，供 n8n 和 Flowise 使用。不添加 `/rest/v1`。
2. 为 Metabase 记下**权限** > **导出端口** > `postgres-session` 的导出端口。
3. 在 Control Hub 中打开 **Terminal** > **Olares**，运行：

   ```bash
   olares-cli settings apps get supabase
   ```

   使用顶层 **URL** 作为数据库主机，去掉 `https://` 和路径。**Ports** 列出映射端口，**Entrances** 列出独立的 Studio 和 API URL。没有配置 Olares profile 时，先[登录 Olares CLI](https://www.olares.com/docs/zh/developer/cli-install#登录-olares)，再运行命令。

数据库 TCP 映射需要可达的局域网或 LarePass VPN 连接。以当前实例显示的端口为准，此安装包的 session 和 transaction 连接池分别映射到 `45432` 和 `46543`。

#### 获取连接凭据

1. 在 Studio 中打开 **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**。
2. 点击 **Reveal**，复制 n8n 或 Flowise 所需的 `service_role` key。
3. 为 Metabase 打开 Control Hub，进入 **Browse** > **supabase-{user-name}** > **Secrets** > **supabase-secrets**，在 **Data** 中查看解码后的 `POSTGRES_PASSWORD`。

`service_role` key 会绕过 RLS，只应保存在可信的服务端应用中。不要放入浏览器代码、截图或代码仓库。官方 MCP 提供的可公开使用的 `anon` key 无法替代这些服务端凭据。

### 通过 Lares 准备任务和费用表

让 Lares 创建共享的示例数据：

```text
使用 supabase_official 检查 public schema，并准备两张表。
project_tasks：id bigint 自增主键（generated by default as identity），
project text 必填，title text 必填，status text 必填且默认 'todo'。
project_costs：id integer 主键，project text 必填，amount integer 必填。
两张表都启用 RLS，不添加匿名访问策略。
向 project_costs 插入以下虚构数据：
1, Lakeview, 100
2, Riverside, 200
3, Maple, 300
使用迁移修改表结构，使用 execute_sql 插入示例数据。
先展示拟执行的修改，保留现有数据，避免重复插入。
```

批准预期操作后，在 Studio 的 **Table Editor** 中检查。`project_costs` 应包含三条记录，`project_tasks` 在 n8n 工作流添加任务之前为空。这些示例使用服务端凭据或直接数据库登录，因此不受应用用户的 RLS 策略限制。

::: details 替代方式：在 SQL Editor 中准备数据表
如果不让 Lares 创建数据表，可运行以下 SQL。

```sql
create table public.project_tasks (
  id bigint generated by default as identity primary key,
  project text not null,
  title text not null,
  status text not null default 'todo'
);
alter table public.project_tasks enable row level security;

create table public.project_costs (
  id integer primary key,
  project text not null,
  amount integer not null
);
alter table public.project_costs enable row level security;

insert into public.project_costs values
  (1, 'Lakeview', 100),
  (2, 'Riverside', 200),
  (3, 'Maple', 300);
```
:::

### 用 n8n 记录交付任务

n8n 在应用之间运行自动化工作流。此示例通过手动触发器和 Supabase 节点向 `project_tasks` 插入交付任务。

向 Lares 发送：

```text
帮我完成下面的 n8n 示例。使用 supabase_official 检查 project_tasks，
获取 API URL，准备手动触发的工作流和 Supabase Create a row 节点配置，
用于添加 Lakeview 交付任务。优先使用 n8n 中已有的 Supabase 凭据，
不要让我把 service_role key 粘贴到对话中。说明还需要完成哪些设置，
再帮我验证插入结果，避免重复创建任务。
```

按以下步骤完成连接，或核对生成的配置：

1. 打开 n8n，创建带手动触发器的工作流。
2. 点击触发器旁的 **+**，搜索 **Supabase**，选择 **Create a row** 添加节点。
3. 在 **Credential** 下点击 **Connect to Supabase**。
4. 在 **Supabase account** 窗口中配置：

   | 字段 | 值 |
   | --- | --- |
   | Host | Supabase API URL |
   | Secret Key | `service_role` key |

5. 点击 **Save**。
6. 将 **Table Name or ID** 设为 `project_tasks`。
7. **Data to Send** 保持 **Define Below for Each Column**。
8. 在 **Fields to Send** 下点击 **Add Field**，逐一添加以下字段：

   | 字段 | 值 |
   | --- | --- |
   | `project` | `Lakeview` |
   | `title` | `Confirm cabinet delivery for 10 November 2026` |
   | `status` | `todo` |

9. 点击 **Execute step**。
10. 返回 Supabase Studio 的 **Table Editor**，找到 `project_tasks` 并检查新增任务。

### 用 Metabase 汇总项目费用

Metabase 根据数据库创建报表和仪表盘。此示例连接 Supabase 并汇总项目费用。

向 Lares 发送：

```text
使用 supabase_official 检查 project_costs 示例数据，准备三行费用明细
及总额查询的 SQL，帮我配置下面的 Metabase PostgreSQL 连接。
列出无法通过 Supabase MCP 获取的主机、端口或凭据，
让我直接在 Metabase 中填写敏感信息，再帮我检查是否为 3 条记录、总额 600。
```

使用以下设置完成 Metabase 连接：

1. 打开 Metabase，进入 **Admin** > **Databases** > **Add database**。
2. 选择 PostgreSQL 并填写：

   | 字段 | 值 |
   | --- | --- |
   | Display name | `Home project costs` |
   | Host | [获取连接地址](#获取连接地址)第 3 步的顶层 URL，去掉 `https://` |
   | Port | `postgres-session` 对应的导出端口 |
   | Database name | `postgres` |
   | Username | `postgres.supabase`，安装包固定的连接池用户名 |
   | Password | [获取连接凭据](#获取连接凭据)中的已解码 `POSTGRES_PASSWORD` |
   | SSL | 此安装包的局域网连接设为关闭 |

3. 保存连接。
4. 点击右上角头像，选择 **Main app**。
5. 点击右上角 **+ New** > **SQL query**，选择 `Home project costs` 数据库。
6. 运行以下查询，列出项目费用：

   ```sql
   select project, amount
   from public.project_costs
   order by id;
   ```

   结果包含 `project` 和 `amount` 两列，三行分别为 `Lakeview` 100、`Riverside` 200 和 `Maple` 300。
7. 运行以下查询计算总额：

   ```sql
   select count(*) as records, sum(amount) as total_usd
   from public.project_costs;
   ```

   结果为一行，显示 3 条记录，总额为 600。

:::warning 数据库权限和网络范围
此示例在你控制的 Metabase 中使用安装管理员的数据库账户。共享报表访问权限前，应配置专用的只读数据库账户及适当的 Metabase 权限。TCP 映射允许局域网或 VPN 连接，访问范围不只限于 Metabase 应用。不要用 API key 代替数据库凭据，也不要把浏览器 HTTPS 入口作为数据库连接。
:::

### 用 Flowise 查询项目笔记

Flowise 通过连接节点构建 AI 对话流程。此示例将项目资料存入 Supabase，并使用 Router 中的对话模型和嵌入模型回答问题。Flowise 保留自己的应用数据库，此流程不需要直接连接 Supabase 的 PostgreSQL。

向 Lares 发送：

```text
帮我准备下面的 Flowise 示例。使用 supabase_official 检查数据库，
为 1024 维嵌入准备 project_documents 表和匹配函数的迁移方案。
保留现有数据，执行前先展示修改。准备使用 default-chat 和
default-embedding 的流程配置，嵌入 batch size 为 1，encoding 为 float。
列出我仍需完成的 Router、Flowise 凭据及特定版本兼容设置，
再帮我将回答与原始资料核对。
```

运行流程前，先完成以下模型和应用设置：

#### 准备 Flowise 和 Router

流程需要两个可通过 Router 调用的模型：对话模型负责回答问题，嵌入模型负责将资料向量化。先安装模型、配置默认路由、复制连接地址，再将路由名称加入 Flowise 模型下拉列表。

1. 从应用市场安装 **Qwen3.8-27B (llama.cpp)** 和 **Jina CLIP v2** 嵌入模型，等待下载完成。

   :::info
   Jina CLIP v2 输出 1024 维向量。资料导入和检索必须使用同一嵌入模型。更换维度需要更换向量列并重新建立索引。
   :::
2. 在 Router 的**默认模型**页面，将 **Qwen3.8-27B (llama.cpp)** 设为默认对话模型，将 **Jina CLIP v2** 设为默认嵌入模型。`default-chat` 和 `default-embedding` 路由随后指向这两个模型。
3. 从 Router 连接示例中复制供 **Olares 内的应用**使用的 OpenAI 兼容 base URL，包含 `/v1`。其他连接位置和密钥要求参阅[连接 AI 应用](https://www.olares.com/docs/zh/manual/best-practices/connect-ai-apps)。
4. Flowise 的 OpenAI 节点使用内置模型目录，其中没有这两个 Router 路由。在 Control Hub 中进入 **Browse** > **flowise-{user-name}** > **Deployments** > **flowise**，打开 flowise 容器的 **Terminal**，运行以下脚本，复制目录并添加路由：

   ::: details 将 Router 路由加入 Flowise 模型目录
   ```bash
   node <<'NODE'
   const fs = require('fs');
   const base = '/usr/local/lib/node_modules/flowise/node_modules/flowise-components';
   const source = [base + '/models.json', base + '/dist/models.json']
     .find(path => fs.existsSync(path));
   if (!source) throw new Error('Bundled model catalog not found');
   const catalog = JSON.parse(fs.readFileSync(source, 'utf8'));
   for (const [category, provider, name] of [
     ['chat', 'chatOpenAI', 'default-chat'],
     ['embedding', 'openAIEmbeddings', 'default-embedding']
   ]) {
     const entry = catalog[category].find(item => item.name === provider);
     if (!entry) throw new Error('Model provider not found: ' + provider);
     if (!entry.models.some(model => model.name === name)) {
       entry.models.push({ label: name, name });
     }
   }
   fs.writeFileSync('/home/node/.flowise/olares-models.json', JSON.stringify(catalog, null, 2));
   NODE
   ```
   :::

5. 在 Control Hub 中打开 **ConfigMaps > flowise-config**，点击右上角 **Edit** 图标，在 YAML 编辑器的 `data` 下添加：

   ```yaml
   MODEL_LIST_CONFIG_JSON: "/home/node/.flowise/olares-models.json"
   ```

6. 保存并重启 Flowise。

   :::info
   直接修改的 ConfigMap 可能在应用升级时被覆盖。如果升级后模型列表不再显示这两个路由，请重新添加该配置。
   :::

#### 可选：修复 Flowise 3.1.4 镜像的 WebSocket 兼容问题

应用市场的 Flowise 3.1.4 镜像使用 Node.js 20 运行 Flowise 3.1.2，其 Supabase 依赖可能提示 `Node.js 20 detected without native WebSocket support`。使用此镜像时按以下步骤处理。

1. 在电脑上创建 `supabase-websocket.cjs`，内容如下：

   ```js
   globalThis.WebSocket ||= require('/usr/local/lib/node_modules/flowise/node_modules/ws');
   ```

2. 通过文件管理器上传到 `Data/flowise/`。在 Flowise 中，该文件挂载为 `/home/node/.flowise/supabase-websocket.cjs`。
3. 在 Control Hub 中打开 **ConfigMaps > flowise-config**，点击右上角 **Edit** 图标，在 YAML 编辑器的 `data` 下添加：

   ```yaml
   NODE_OPTIONS: "--require=/home/node/.flowise/supabase-websocket.cjs"
   ```

   如果 `NODE_OPTIONS` 已有值，在原值后追加此选项。
4. 保存并重启 Flowise。

   :::info
   直接修改的 ConfigMap 可能在应用升级时被覆盖。保留已上传文件，如果 Supabase 节点再次提示 WebSocket 错误，请重新添加 `NODE_OPTIONS` 配置。
   :::

#### 创建资料存储

1. 将以下 SQL 交给 Lares，让它先检查现有对象，展示迁移方案，再通过 `apply_migration` 创建数据表和函数，不替换现有数据。批准预期迁移。
2. 在 Studio 中确认 `project_documents` 已创建且启用了 RLS。也可直接在 **SQL Editor** 中运行 SQL，代替 Lares 操作。

   ::: details 示例：project_documents 表和匹配函数
   ```sql
   create extension if not exists vector with schema extensions;
   create table public.project_documents (
     id uuid default gen_random_uuid() primary key,
     content text,
     metadata jsonb default '{}',
     embedding extensions.vector(1024)
   );
   alter table public.project_documents enable row level security;

   create function public.match_project_documents(
     query_embedding extensions.vector,
     match_count integer default 4,
     filter jsonb default '{}'
   ) returns table(id uuid, content text, metadata jsonb, similarity float)
   language sql stable security invoker
   set search_path = public, extensions
   as $$
     select id, content, metadata,
            1 - (embedding operator(extensions.<=>) query_embedding)
     from public.project_documents
     where metadata @> filter
     order by embedding operator(extensions.<=>) query_embedding
     limit match_count
   $$;
   ```
   :::

#### 构建对话流程

1. 在 Flowise 中创建 chatflow，添加以下节点：

   | 节点 | 配置 |
   | --- | --- |
   | Plain Text | <ul><li>**Text**：输入示例资料 `Lakeview cabinet delivery is on 10 November 2026. Confirmation reference: LAKEVIEW-6728.`</li><li>**Additional Metadata**：`{"project":"Lakeview"}`</li></ul> |
   | OpenAI Embeddings | <ul><li>**Connect Credential**：新建凭据，填写任意值，例如 `olares`，Router 不校验 Olares 内应用的 key</li><li>**Model Name**：`default-embedding`</li><li>**Batch Size**：`1`，Jina CLIP v2 每次请求只接受一条文本</li><li>**Base Path**：Router base URL</li><li>**Encoding Format**：`float`</li></ul> |
   | Supabase | <ul><li>**Supabase Project URL**：Supabase API URL</li><li>**Connect Credential**：`service_role` key</li><li>**Table Name**：`project_documents`</li><li>**Query Name**：`match_project_documents`</li></ul> |
   | OpenAI | <ul><li>**Connect Credential**：新建凭据，填写任意值，例如 `olares`，Router 不校验 Olares 内应用的 key</li><li>**Model Name**：`default-chat`</li><li>**Base Path**：Router base URL</li></ul> |
   | Conversational Retrieval QA Chain | 结合检索资料和对话模型回答问题，无需配置字段 |

2. 从节点输出圆点拖动到目标节点输入圆点，依次连接：

   a. Plain Text 的 **Document** 输出连接 Supabase 的 Document 输入。

   b. OpenAI Embeddings 的输出连接 Supabase 的 Embeddings 输入。

   c. Supabase 的 **Supabase Retriever** 输出连接 Conversational Retrieval QA Chain 的 **Vector Store Retriever** 输入。

   d. OpenAI 对话模型的输出连接 Conversational Retrieval QA Chain 的 **Chat Model** 输入。

3. 保存流程，点击画布上的 **Upsert Vector Database**，将资料转换为向量并写入 `project_documents`。
4. 提问 `What is the Lakeview cabinet delivery date and confirmation reference?`，将回答与原文核对。

## 添加文件和实时更新

要保存收据或项目文件，在 Studio 的 **Storage** 页面创建私有 bucket，配置访问策略后再从应用使用。同一 Supabase API URL 支持上传、下载和生成短时签名 URL。参阅 [Storage 访问控制](https://supabase.com/docs/guides/storage/security/access-control)。

要接收笔记实时更新，将 `user_notes` 加入 `supabase_realtime` publication，并通过 SDK 订阅。保留 RLS 策略，在数据库订阅就绪后再依赖事件通知。参阅 [Postgres Changes](https://supabase.com/docs/guides/realtime/postgres-changes)。

官方自托管 MCP 工具覆盖数据库迁移、SQL、表结构检查、类型生成、文档搜索以及安全和性能建议。Auth 用户、Storage bucket、SMTP、登录跳转 URL 及其他应用设置仍通过各自界面或 API 配置。此安装包未启用 `query_logs` 所需的分析服务。

## 故障排查

| 现象 | 检查方法 |
| --- | --- |
| Lares 无法连接 MCP | 使用受保护的 Studio HTTPS URL，以 `/api/mcp` 结尾，传输方式为 `Streamable HTTP`，请求头为 `{}`。独立 API 网关的 `/mcp` 路由被阻止 |
| 外部 MCP 客户端被重定向到登录页 | 此连接依赖同一实例上 Lares 的 Olares 应用身份。不要开放 Studio，也不要将空请求头配置照搬到未认证的外部客户端 |
| Lares 执行 SQL 前请求确认 | 可写 MCP 连接将 `execute_sql` 视为可写工具。逐次检查操作，或使用单独的 `?read_only=true` 连接查看数据 |
| `query_logs` 提示日志未启用 | 此安装包未启用日志分析，请在 Control Hub 中查看对应容器日志 |
| 查询提示 `password authentication failed for user "supabase_read_only_user"` | 请管理员将该数据库账户密码与实例的 Studio 数据库配置同步。修改 MCP 请求头无法修复数据库凭据 |
| 登录成功但 REST 提示 `role "" does not exist` | 在 SQL Editor 中检查该用户的 `auth.users.role`。普通应用用户应为 `authenticated`。修正空角色后，重新登录或刷新会话获取新 token |
| SDK 查询没有返回记录 | 先检查登录用户、记录归属和 RLS 策略，再考虑更换凭据 |
| n8n 或 Flowise 认证失败 | 使用 Supabase API URL 和 legacy `service_role` key |
| Metabase 无法连接 | 检查导出端口、主机可达性、数据库密码和 `postgres.supabase` 用户名 |
| Flowise 提示缺少 WebSocket 支持 | 应用上文对应版本的兼容修复 |
| Flowise 嵌入请求提示只能处理单条文本 | 将嵌入节点 **Batch Size** 设为 `1` |

规划备份时，同时包含 PostgreSQL 数据和上传的存储对象。

## 了解更多

- [用 Lares 管理 Olares](lares.md)：选择模型并开始对话。
- [用 n8n 自动化工作流](n8n.md)：配置 n8n 并创建自动化流程。
- [使用 Olares Router 作为 AI 网关](olares-router.md)：管理模型路由和凭据。
- [启用自托管 Supabase MCP 服务器](https://supabase.com/docs/guides/self-hosting/enable-mcp)：了解上游服务器的网络访问要求。
