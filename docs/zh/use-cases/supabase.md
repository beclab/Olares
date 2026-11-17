---
outline: [2, 3]
description: 在 Olares 上使用 Lares 为私有笔记应用构建 Supabase 后端，验证用户间的数据隔离，并通过 Metabase 查询项目费用。
head:
  - - meta
    - name: keywords
      content: Olares, Supabase, Lares, MCP, PostgreSQL, RLS, Metabase, Code Server
app_version: "0.0.15"
doc_version: "2.1"
doc_updated: "2026-11-17"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/supabase.md)为准。
:::

# 用 Supabase 和 Lares 构建应用后端

Supabase 在你的 Olares 设备上为应用提供 PostgreSQL 数据库、用户认证和 API。将 Lares 连接到 Supabase Model Context Protocol（MCP）服务器后，你可以通过对话创建数据表和访问策略。

在本教程中，你将构建并运行一个命令行笔记应用，通过行级安全（RLS）保护每个用户的私有笔记。你还将使用 Metabase 查询存储在另一张数据表中的示例项目费用。

## 学习目标

通过本教程，你将学会：

- 将 Lares 连接到 Supabase MCP 服务器。
- 创建用于存储用户笔记的数据表，并通过 RLS 配置每个用户的访问权限。
- 使用 Lares 生成笔记应用，在 Code Server 中运行，并使用在 Supabase 中创建的用户账号登录。
- 验证每个用户只能访问自己的笔记。
- 将 Metabase 连接到 Supabase，并查询示例项目费用。

## 前置条件

- Lares 中已选择可用的模型。参阅 [Lares 入门](lares.md#开始使用)。
- 已从应用市场安装 Code Server，用于运行笔记应用。
- 已从应用市场安装 Metabase，用于生成报表。

## 安装 Supabase

1. 打开应用市场，搜索“Supabase”。
2. 点击**获取**，然后点击**安装**，等待安装完成。

## 将 Lares 连接到 Supabase MCP 服务器

将 Lares 连接到 Supabase，以便在对话中使用数据库工具。

:::info 连接认证
Lares 通过同一实例上的 Olares 身份认证进行连接，因此不需要额外提供访问令牌或 API key。在 Olares 设置中，将 **Supabase Studio** 入口保持为 **Internal**。
:::

1. 打开 Olares 设置，进入**应用** > **Supabase** > **Supabase Studio**。
2. 复制 **Endpoint** 地址，并在末尾添加 `/api/mcp`。例如：

   ```text
   https://304244530.laresprime.olares.com/api/mcp
   ```

3. 打开 Lares，进入 **Settings** > **MCP**，点击 **Add server**。
4. 填写以下设置：

   | 设置项 | 值 |
   | --- | --- |
   | Server name | `supabase_official` |
   | Transport | `Streamable HTTP` |
   | MCP URL | 复制的入口地址，以 `/api/mcp` 结尾 |

5. 点击 **Save**。Lares 会建立连接并加载服务器的工具。
6. 在 Lares 中开始新对话，发送以下提示词检查 MCP 连接：

   ```text
   使用 supabase_official 列出 public schema 中的数据表，并返回
   此 Supabase 实例的 API URL。直接使用 MCP 工具，不要修改数据。
   ```

   Lares 应返回 Supabase API URL 和 `public` schema 中的数据表列表。如果尚未创建任何数据表，列表将为空。

## 使用 Supabase 构建并运行笔记应用

使用 Lares 准备 Supabase 数据表、配置访问策略并生成笔记应用。然后在 Code Server 中运行应用，检查每个用户是否只能访问自己的笔记。

### 使用 Lares 创建笔记表和访问策略

MCP 连接就绪后，使用 Lares 为笔记应用创建数据表和访问策略，确保每个用户只能访问自己的笔记。

每条笔记的 `owner_id` 字段保存其所属用户的 ID。策略将该值与当前登录用户的 ID 比较，以控制访问权限。

:::warning 数据库管理权限
MCP 服务器可以修改数据表和访问策略。请检查拟执行的 SQL，并仅批准预期的修改。
:::

1. 在 Lares 对话中描述所需的数据表和策略：

   ```text
   使用 supabase_official 为我的应用准备一个私有笔记表。
   先检查 public schema。创建 public.user_notes，包含：
   - id：UUID 主键，默认 gen_random_uuid()
   - owner_id：UUID，必填
   - note：text，必填
   启用 RLS，并向 authenticated 授予 SELECT、INSERT、UPDATE、DELETE 权限。
   创建 own_notes 策略，使用 FOR ALL TO authenticated，
   将 USING 和 WITH CHECK 都设为 auth.uid() = owner_id。
   使用 apply_migration 修改表结构，执行前展示计划运行的 SQL。
   只修改此表，不删除现有数据，也不授予匿名访问权限。
   如果表已存在，先说明差异，再提出迁移方案。
   ```

2. 检查拟执行的 SQL。Lares 请求审批时，仅批准预期的修改，并等待执行完成。
3. 打开 Supabase Studio，进入 **Table Editor**。找到 `user_notes`，检查 `id`、`owner_id` 和 `note` 列，并确认 **RLS policy** 指示器显示一条策略。

4. 让 Lares 检查 `user_notes` 并总结其 RLS 策略：

   ```text
   使用 supabase_official 检查 public.user_notes。确认已启用 RLS，
   并总结策略名称、适用角色，以及 USING 读取条件和 WITH CHECK 写入检查条件。
   ```

   总结应显示 RLS 已启用，且有一条适用于 `authenticated` 角色的 `own_notes` 策略。两个条件都将 `auth.uid()` 与 `owner_id` 比较。

::: details 替代方式：在 SQL Editor 中创建笔记表
如果尚未使用 Lares 创建 `user_notes`，打开 Supabase Studio 的 **SQL Editor**，粘贴以下查询并点击 **Run**。然后按照上述步骤检查数据表和策略。

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

### 使用 Lares 创建笔记应用

这个命令行应用用于用户登录、添加笔记，以及从 Supabase 的 `user_notes` 表中列出自己的笔记。

1. 让 Lares 生成程序及其配置文件：

   ```text
   帮我将私有笔记应用连接到 Supabase。使用 supabase_official
   检查 user_notes 及其 RLS 策略，并获取 API URL 和 anon key。
   使用 @supabase/supabase-js v2 和 Node.js 22 创建 notes.mjs。
   从 process.env 读取 SUPABASE_URL、SUPABASE_ANON_KEY、NOTES_EMAIL
   和 NOTES_PASSWORD。我会使用 node --env-file=.env 加载 .env。
   使用 signInWithPassword 登录，并支持以下命令：
   - whoami：显示已登录用户的邮箱和 ID，不显示会话令牌
   - add "text"：向 user_notes 插入笔记，将 owner_id 设为该用户的 ID，
     然后显示保存的笔记及其 ID
   - list：查询 user_notes 的 id 和 note，依靠 RLS 隔离用户数据
   报告 SDK 错误，并在每条命令执行后退出。
   将 notes.mjs、.env、.env.example 和 .gitignore 直接保存在
   工作区根目录中，不要放在子文件夹里。
   在 .env 中填写 API URL 和 anon key。将 NOTES_EMAIL 和
   NOTES_PASSWORD 留空，并在 Git 中忽略 .env 和 node_modules/。
   在 .env.example 中使用占位值。不要将密码放在对话或源代码中。
   告诉我文件的具体位置。不要创建 Auth 用户、修改表结构或策略，也不要运行应用。
   准备好文件后停止。
   ```

   Lares 会为同一个应用准备四个文件：

   | 文件 | 用途 |
   | --- | --- |
   | `notes.mjs` | 要运行的程序，用于登录 Supabase、添加笔记和列出当前登录用户的笔记。 |
   | `.env` | 运行程序时使用的配置。Lares 填写 API URL 和 `anon` key，你需要补充 Supabase Auth 用户的邮箱和密码。 |
   | `.env.example` | 使用占位值的配置模板，可供参考或用于配置另一份应用。 |
   | `.gitignore` | 告诉 Git 不要将 `.env` 和 `node_modules` 依赖文件夹纳入版本控制。 |

2. 打开 Olares 文件管理器，进入 **Data** > **lares** > **data** > **workspace**，复制 `notes.mjs`、`.env`、`.env.example` 和 `.gitignore`。
3. 在 Olares 文件管理器中，进入 **Home** > **Code**，创建名为 `notes-app` 的文件夹，并将四个文件全部粘贴到该文件夹中。Code Server 可以访问此文件夹。

### 为笔记应用创建用户账号

在 Supabase 中创建用户账号，用于登录笔记应用。此账号与你的 Olares 账号不同。

1. 打开 Supabase Studio，进入 **Authentication** > **Users** > **Add user** > **Create new user**。
2. 输入邮箱（例如 `demo@example.com`），并自行设置密码。保持 **Auto confirm user?** 启用，然后点击 **Create user**。自动确认允许你在此示例中直接登录，无需接收确认邮件。

### 在 Code Server 中运行笔记应用

1. 从启动台打开 Code Server。
2. 通过 **File** > **Open Folder** 打开 `/home/coder/notes-app`，该路径对应 Olares 文件管理器中的 **Home** > **Code** > **notes-app**。
3. 如果出现提示，确认信任此文件夹。检查 Explorer 中是否显示 `notes.mjs` 和 `.env`。
4. 打开 **Terminal** > **New Terminal**。
5. 运行 `node -v`，检查已安装的 Node.js 版本。此示例使用 Node.js 22。
6. 如果未安装 Node.js 或版本低于 22，依次运行以下四条命令，通过 nvm 安装：

   ```bash
   touch "$HOME/.bashrc"
   curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.8/install.sh | bash
   . "$HOME/.nvm/nvm.sh"
   nvm install 22
   ```

7. 运行 `node -v` 和 `npm -v`，确认两条命令都显示版本号后再继续。如果在上一步安装了 Node.js，`node -v` 应显示 `v22.x.x`。
8. 在 Code Server 终端中，切换到项目文件夹并安装 Supabase JavaScript SDK：

   ```bash
   cd /home/coder/notes-app
   npm install @supabase/supabase-js@2.118.0
   ```

9. 在 Code Server 中打开 `.env`。保留 `SUPABASE_URL` 和 `SUPABASE_ANON_KEY` 不变，将 `NOTES_EMAIL` 和 `NOTES_PASSWORD` 设为你创建的笔记应用账号的邮箱和密码，然后保存文件：

   ```dotenv
   NOTES_EMAIL="your-email@example.com"
   NOTES_PASSWORD="your-password"
   ```

   ::: warning 妥善保管密码
   将密码保存在 `.env` 中，不要粘贴到 Lares 对话中，也不要写入 `notes.mjs`。
   :::

   ::: details 替代方式：从 Supabase Studio 和 Olares 设置复制连接信息
   打开 Olares 设置，进入**应用** > **Supabase** > **Supabase API**。将其 **Endpoint** 复制到 `.env` 的 `SUPABASE_URL` 中，不要包含 `/rest/v1` 或 `/api/mcp`。

   在 Supabase Studio 中，打开 **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**，将 `anon` key 复制到 `SUPABASE_ANON_KEY`。不要使用 `service_role`，它会绕过 RLS，允许访问其他用户的笔记。
   :::

10. 在同一终端中运行以下命令，登录、添加笔记并读取保存的笔记。`--env-file=.env` 选项会从 `.env` 加载连接设置和用户凭据：

    ```bash
    node --env-file=.env notes.mjs whoami
    node --env-file=.env notes.mjs add "Ask the Lakeview contractor to confirm the delivery window."
    node --env-file=.env notes.mjs list
    ```

    终端会显示已登录用户、新笔记的 ID 和保存的笔记。每次执行 `add` 命令都会创建一条新记录。

### 检查用户是否只能访问自己的笔记

1. 在 Supabase Studio 中，重复[为笔记应用创建用户账号](#为笔记应用创建用户账号)的步骤，使用不同的邮箱（例如 `demo2@example.com`）和单独的密码。
2. 在 Code Server 中打开 `.env`，将 `NOTES_EMAIL` 和 `NOTES_PASSWORD` 更新为第二个用户的邮箱和密码。保留 `SUPABASE_URL` 和 `SUPABASE_ANON_KEY` 不变，然后保存文件。
3. 运行以下命令：

   ```bash
   node --env-file=.env notes.mjs whoami
   node --env-file=.env notes.mjs list
   ```

   当前登录用户应为第二个用户。虽然第一个用户的笔记仍在数据库中，但新用户的笔记列表应为空。
4. 为第二个用户添加一条笔记，然后列出其笔记：

   ```bash
   node --env-file=.env notes.mjs add "Compare the Riverside flooring quotes."
   node --env-file=.env notes.mjs list
   ```

   列表应包含 Riverside 笔记，不包含第一个用户的 Lakeview 笔记。
5. 在 `.env` 中恢复第一个用户的邮箱和密码，保存文件，然后再次运行 `node --env-file=.env notes.mjs list`。你应看到该用户的 Lakeview 笔记，而看不到 Riverside 笔记。

## 使用 Metabase 查询 Supabase 数据

Metabase 直接连接 PostgreSQL 查询项目费用。通过 Lares 使用 Supabase MCP 创建示例数据后，你需要在 Metabase 中配置数据库连接并运行报表查询。

### 完成 Metabase 首次设置

如果首次打开 Metabase，请先完成设置向导，再添加 Supabase 连接：

1. 打开 Metabase，点击 **Let's get started**。
2. 输入姓名、邮箱、团队名称，并为 Metabase 管理员账号设置密码。点击 **Next**，回答使用情况问题。
3. 在 **Add your data** 页面，点击 **Continue with sample data**，继续设置，暂不连接 Supabase。你将在[将 Metabase 连接到 Supabase](#将-metabase-连接到-supabase)部分添加连接。
4. 选择使用数据偏好，点击 **Finish**，然后点击 **Take me to Metabase**。

### 获取数据库连接信息

为 Metabase 准备 Supabase 数据库主机、导出的会话端口和数据库密码。此连接使用数据库用户名 `postgres.supabase`。

#### 获取数据库主机和端口

1. 打开 Olares 设置，进入**应用** > **Supabase** > **权限** > **导出端口** > `postgres-session`。记录 **Exported port**。

2. 如果尚未完成，请在电脑上[安装 Olares CLI](../developer/cli-install.md#先安装-olares-cli)，并[登录你的 Olares 实例](../developer/cli-install.md#登录-olares)。在电脑的终端中运行：

   ```bash
   olares-cli settings apps get supabase
   ```

3. 记录命令输出中的顶层 **URL**，用作数据库主机，不要包含 `https://` 或路径。
4. 在 **Ports** 列表中，检查 `postgres-session` 的 **EXPOSE** 值是否与第 1 步记录的导出端口一致。在 Metabase 中使用此端口。

#### 获取数据库密码

1. 在 Control Hub 中，进入 **Browse** > **supabase-{user-name}** > **Deployments** > **db**，打开其 Pod，然后打开 **db** 容器的终端。
2. 运行以下命令显示数据库密码：

   ```sh
   . /runtime/keys.env
   printf '%s\n' "$POSTGRES_PASSWORD"
   ```

3. 保留此终端，添加连接时需要将密码复制到 Metabase。不要将密码粘贴到 Lares 对话中。

### 使用 Lares 创建项目费用表

1. 在 Lares 中发送以下提示词，创建 `project_costs` 并添加三条虚构的装修费用。此示例中 `amount` 的单位为美元。

   ```text
   使用 supabase_official 检查 public schema，并准备一张表。
   project_costs：id integer 主键，project text 必填，amount integer 必填。
   启用 RLS，不添加匿名访问策略。
   插入以下虚构数据：
   1, Lakeview, 100
   2, Riverside, 200
   3, Maple, 300
   使用迁移修改表结构，使用 execute_sql 插入示例数据。
   先展示拟执行的修改，保留现有数据，避免重复插入。
   ```

2. 检查拟执行的修改，并在 Lares 中确认。
3. 当工具调用显示 **Waiting for approval** 时，检查其 SQL，并为预期操作点击 **Allow once**。迁移、插入示例数据和验证查询可能都需要审批，请按需重复操作。
4. 等待 Lares 完成后，打开 Supabase Studio 的 **Table Editor**，选择 `project_costs`。检查其中是否包含提示词中的三条记录。

::: details 替代方式：在 SQL Editor 中准备数据表
如果尚未使用 Lares 创建 `project_costs`，在 Supabase Studio 的 **SQL Editor** 中运行以下查询。

```sql
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

### 将 Metabase 连接到 Supabase

1. 在 Metabase 中，打开右上角的 **Settings** 菜单，选择 **Admin**。进入 **Databases** > **Add database**。
2. 选择 PostgreSQL，并填写：

   | 字段 | 值 |
   | --- | --- |
   | Display name | `Home project costs` |
   | Host | [获取数据库主机和端口](#获取数据库主机和端口)第 3 步记录的顶层 URL，去掉 `https://` |
   | Port | [获取数据库主机和端口](#获取数据库主机和端口)第 1 步记录的 **Exported port** |
   | Database name | `postgres` |
   | Username | `postgres.supabase` |
   | Password | [获取数据库密码](#获取数据库密码)中的 `POSTGRES_PASSWORD` |
   | Use a secure connection (SSL) | 此安装包的局域网连接设为关闭 |

   ::: warning 数据库管理员权限
   此连接会授予 Metabase 对 Supabase 的数据库管理员级别访问权限，包括在没有 RLS 读取策略时访问 `project_costs`。仅在你控制的 Metabase 实例中使用这些凭据。
   :::

3. 点击 **Save**。

4. 在 **Home project costs** 数据库页面，检查 **Connection and sync** 是否显示 **Connected**。

### 在 Metabase 中查询项目费用

1. 打开右上角的 **Settings** 菜单，选择 **Main app**。
2. 点击右上角的 **New** > **SQL query**。检查 SQL 编辑器上方的数据库选择器，如果尚未选择 `Home project costs`，请选择该数据库。
3. 将以下查询粘贴到 SQL 编辑器，然后点击运行按钮，列出示例项目费用：

   ```sql
   select project, amount
   from public.project_costs
   order by id;
   ```

   查询返回包含 `project` 和 `amount` 两列的表格，三行分别为 `Lakeview` 100、`Riverside` 200 和 `Maple` 300。

4. 将编辑器中的 SQL 替换为以下查询，再次点击运行按钮计算总额：

   ```sql
   select count(*) as records, sum(amount) as total_usd
   from public.project_costs;
   ```

   查询返回一行，显示 3 条记录，总额为 600。

## 故障排查

| 现象 | 检查方法 |
| --- | --- |
| Lares 无法连接 MCP | 使用以 `/api/mcp` 结尾的 Supabase Studio HTTPS URL，传输方式为 `Streamable HTTP`。将 **Supabase Studio** 入口保持为 **Internal**。 |
| Code Server 找不到 `notes.mjs` 或 `.env` | 检查两个文件是否都在 Olares 文件管理器的 **Home** > **Code** > **notes-app** 中，然后在 Code Server 的 `/home/coder/notes-app` 目录下运行命令 |
| 终端提示 `node: command not found` | 在同一终端中运行 `. "$HOME/.nvm/nvm.sh"`，然后运行 `nvm use 22`。再次运行应用前，先检查 `node -v`。如果尚未安装 nvm 或 Node.js 22，请按照[在 Code Server 中运行笔记应用](#在-code-server-中运行笔记应用)中的安装步骤操作。 |
| `node -v` 提示缺少 `libatomic.so.1` | 在 Code Server 终端中运行 `sudo apt-get install -y libatomic1`，然后再次检查 `node -v` |
| 笔记应用返回空列表 | 如果当前登录用户尚未添加笔记，返回空列表属于正常情况。如果缺少已有笔记，运行 `node --env-file=.env notes.mjs whoami` 检查用户，再检查笔记的 `owner_id` 是否与该用户 ID 一致，以及 RLS 策略是否与[使用 Lares 创建笔记表和访问策略](#使用-lares-创建笔记表和访问策略)中的设置一致。 |
| Metabase 无法连接 | 检查[获取数据库主机和端口](#获取数据库主机和端口)中记录的数据库主机和导出的 `postgres-session` 端口，确认 Metabase 可以访问该主机和端口。用户名使用 `postgres.supabase`，密码使用[获取数据库密码](#获取数据库密码)中取得的值。 |

## 了解更多

- [用 Lares 管理 Olares](lares.md)：选择模型并开始对话。
- [启用自托管 Supabase MCP 服务器](https://supabase.com/docs/guides/self-hosting/enable-mcp)：了解上游服务器的网络访问要求。
