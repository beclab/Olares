---
outline: [2, 3]
description: Use Lares and the official Supabase MCP server to prepare a private app backend on Olares, then connect your application, workflows, reports, and document search.
head:
  - - meta
    - name: keywords
      content: Olares, Supabase, Lares, MCP, PostgreSQL, RLS, n8n, Metabase, Flowise, Code Server
app_version: "0.0.14"
doc_version: "2.1"
doc_updated: "2026-11-12"
---

# Build an app backend with Supabase and Lares

Supabase combines PostgreSQL, user authentication, file storage, realtime updates, and Edge Functions. Its Studio interface lets you manage these services on your own Olares device.

Lares can use Supabase's official Model Context Protocol (MCP) server to inspect tables, apply database migrations, and configure row-level security (RLS) from a conversation. The example below prepares private notes for a fictional home renovation project, then connects a small application in Code Server. You can also connect n8n, Metabase, and Flowise to the same backend.

## Learning objectives

- Connect Lares to the official Supabase MCP server.
- Ask Lares to prepare an app's tables and per-user access policy.
- Connect an application with ordinary user credentials and check its data access.
- Choose the correct addresses and credentials for other integrations.

## Prerequisites

- Lares with a working model selected. See [Get started with Lares](lares.md#get-started).
- Code Server installed from Market for the notes application.

## Install Supabase

1. Open Market and search for "Supabase".
2. Click **Get**, then **Install**, and wait for installation to complete. The app appears on the Launchpad as **Supabase Studio**.

Each installation provides one self-hosted project. You do not need a Supabase Cloud account.

## Connect the official MCP server in Lares

1. Open Olares Settings, go to **Applications** > **Supabase**, and copy the **Supabase Studio** > **Endpoint** address.
2. Append `/api/mcp` to the Studio address. For example:

   ```text
   https://YOUR_STUDIO_HOST/api/mcp
   ```

   Use the Studio entrance for this connection. The separate Supabase API entrance is used by SDKs and other apps later in this guide.
3. Open Lares, go to **Settings** > **MCP**, and click **Add server**.
4. Enter the following values:

   | Field | Value |
   | --- | --- |
   | Server name | `supabase_official` |
   | Transport | `Streamable HTTP` |
   | MCP URL | The Studio URL ending in `/api/mcp` |
   | Headers | `{}` |

5. Click **Save**. Lares connects and loads the server's tools.
6. Start a conversation in Lares and ask:

   ```text
   Use supabase_official to list the tables in the public schema and return
   this project's API URL. Use the MCP tools directly and do not modify data.
   ```

   Lares should return the table list and the Supabase API URL. A new project can have an empty table list.

Olares authenticates the connection from Lares on the same Olares instance. You do not need a Supabase Cloud personal access token or a `service_role` key for this MCP connection. Keep the Studio entrance protected by Olares. Do not make it public to connect an agent.

:::warning Database administration access
The MCP server manages the project with database administration privileges. It can change schema and access policies. Review the SQL and approve only the intended changes when Lares asks. The application's own users will use a separate SDK connection with RLS.
:::

For a connection that can only inspect data, append `?read_only=true` to the MCP URL. The server then disables migrations and rejects database writes. To limit the available tools to database operations as well, use `?read_only=true&features=database`.

## Prepare a private notes table with Lares

The application stores each note with its owner's Supabase Auth user ID. RLS allows a signed-in user to read and change only their own rows.

1. In the Lares conversation, describe the required table and policy:

   ```text
   Use supabase_official to prepare a private notes table for my app.
   Inspect the public schema first. Create public.user_notes with:
   - id: UUID primary key, default gen_random_uuid()
   - owner_id: UUID, required
   - note: text, required
   Enable RLS and grant SELECT, INSERT, UPDATE, DELETE to authenticated.
   Create an own_notes policy FOR ALL TO authenticated with both USING
   and WITH CHECK set to auth.uid() = owner_id.
   Use apply_migration for the schema changes. Show the planned SQL before
   applying it. Only change this table, and do not drop existing data or
   grant anonymous access. If the table already exists, explain any
   differences before proposing a migration.
   ```

2. Review the proposed migration and approve the intended tool calls. In Lares's default permission mode, migrations and the writable connection's `execute_sql` tool require approval, including SQL queries sent through that tool.
3. Open Supabase Studio and go to **Table Editor**. Find `user_notes` and check the `id`, `owner_id`, and `note` columns and the RLS badge.

   You can also ask Lares to inspect the table and summarize its policy. Check that both the read condition and the write check compare `auth.uid()` with `owner_id`.

::: details Alternative: create the same table in SQL Editor
Use this alternative instead of asking Lares to create the table. Open Supabase Studio's **SQL Editor**, paste the following query, and click **Run**.

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

For another application, give Lares its required schema or migration files and describe who should access each table. Ask it to inspect existing data before proposing changes. The official MCP server handles database work, but app settings, Auth user creation, SMTP, and login redirect URLs still need their corresponding application settings or APIs.

## Connect an application in Code Server

Start by asking Lares to prepare the example for you:

```text
Help me connect a private notes app to Supabase. Use supabase_official
to inspect user_notes and its RLS policy, and get the API URL and anon key.
Prepare notes.mjs and a .env template for the SDK example below.
Keep passwords out of the conversation and source code. Stop after
preparing the files and explaining the next steps. Do not create Auth
users, run the app, or test two-user isolation; I will do those below.
```

Lares only checks the existing table and policy and prepares file contents at this stage. Copy the files into Code Server, then create users, run the app, and verify isolation yourself in the steps below. Writing files or running commands directly requires a separate authorized workspace tool.

### Get the API connection and create a user

1. Use the API URL and legacy `anon` key returned by Lares. They come from `get_project_url` and `get_publishable_keys`. The SDK connects to the Supabase API URL, without `/rest/v1` or `/api/mcp`. Do not use a `service_role` key.
2. Open Supabase Studio and go to **Authentication** > **Users** > **Add user** > **Create new user**.
3. Enter an email and password, keep **Auto confirm user?** enabled for this local example, and create the user.

   Use an address such as `demo@example.com` and a password of your own. This creates an application user in Supabase Auth. Auto confirmation lets you run this example without sending email.

::: details Alternative: copy the connection details from Studio and Settings
Copy the API URL from Olares **Settings** > **Applications** > **Supabase** > **Supabase API** > **Endpoint**.

In Supabase Studio, open **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**, and copy the `anon` key.
:::

### Run the notes application

1. Open Code Server from the Launchpad, then open **Terminal** > **New Terminal**. The editor does not include a Node.js runtime. Install Node.js 22 or later with nvm:

   ```bash
   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.8/install.sh | bash
   \. "$HOME/.nvm/nvm.sh"
   nvm install 22
   ```

   If `node -v` reports `libatomic.so.1: cannot open shared object file`, run `sudo apt-get install -y libatomic1` and check `node -v` again.
2. In your project terminal, install the SDK:

   ```bash
   npm install @supabase/supabase-js@2.118.0
   ```

3. Create `.env` in the project directory and fill in your API URL, `anon` key, and the user credentials you just created:

   ```dotenv
   SUPABASE_URL=https://YOUR_SUPABASE_API_HOST
   SUPABASE_ANON_KEY=YOUR_LEGACY_ANON_KEY
   SUPABASE_USER_EMAIL=YOUR_USER_EMAIL
   SUPABASE_USER_PASSWORD=YOUR_USER_PASSWORD
   ```

4. Add `.env` to `.gitignore`, then create `notes.mjs`:

   ::: details Example: notes.mjs
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

5. Run the application:

   ```bash
   node --env-file=.env notes.mjs
   ```

   The terminal displays the signed-in user's notes and a response from the included `hello` Edge Function.
### Verify two-user isolation

1. In Studio, create a second Auth user with a different email and password, keeping **Auto confirm user?** enabled.
2. In `.env`, replace `SUPABASE_USER_EMAIL` and `SUPABASE_USER_PASSWORD` with the second user's credentials. Before inserting anything as this user, run:

   ```bash
   NOTES_READ_ONLY=1 node --env-file=.env notes.mjs
   ```

   The notes table should be empty if this user has no existing notes. In all cases, it must not contain the first user's Lakeview note.
3. Insert a distinct note as the second user:

   ```bash
   NOTES_TEXT='Riverside user note' node --env-file=.env notes.mjs
   ```

   The output should include `Riverside user note` and exclude the first user's Lakeview note.
4. Restore the first user's credentials in `.env` and run the read-only command from step 2 again. The output should include the Lakeview note and exclude `Riverside user note`. Together, these results confirm that each signed-in user reads only their own rows. The read-only runs do not add another note.

Do not replace `anon` with `service_role` to resolve an RLS error. That key bypasses the policy the application relies on.

## Connect more apps to the same backend

The following integrations use fictional home renovation projects. Install the corresponding app from Market before following its section. Lares can prepare the database objects, while you configure the connection inside each app.

Each example includes a request you can send to Lares first. Supabase MCP covers the database side. Configuring n8n, Metabase, or Flowise directly also requires an authorized tool or API connection to that app. Without one, ask Lares to prepare the configuration and complete the fields below yourself. Keep server credentials in the target app's credential store.

### Collect connection details

Choose the address and credential for the connection:

| Connection | Address | Credential |
| --- | --- | --- |
| Lares MCP | Studio HTTPS URL + `/api/mcp` | Olares app identity on the same instance, no additional headers |
| SDK application | Supabase API HTTPS URL | `anon` key and a signed-in user |
| n8n and Flowise | Supabase API HTTPS URL | Legacy `service_role` key in the app's server-side credential store |
| Metabase | Supabase application host and exported session port | `postgres.supabase` and the database password |

#### Get connection addresses

1. Open Olares **Settings** > **Applications** > **Supabase**. Copy the **Supabase API** > **Endpoint** URL for n8n and Flowise. Do not append `/rest/v1`.
2. For Metabase, note the **Exported port** under **Permissions** > **Exported ports** > `postgres-session`.
3. In Control Hub, open **Terminal** > **Olares**, and run:

   ```bash
   olares-cli settings apps get supabase
   ```

   Use the top-level **URL** as the database host, without `https://` or a path. The **Ports** list shows the mapped ports, and **Entrances** shows the separate Studio and API URLs. If no Olares profile is configured, [log in to Olares CLI](https://www.olares.com/docs/developer/cli-install#log-in-to-olares) and run the command again.

Database TCP mappings require a reachable LAN or LarePass VPN connection. Use the ports displayed by your instance. The session and transaction pools are mapped to `45432` and `46543` in this package.

#### Get connection credentials

1. In Studio, open **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**.
2. Click **Reveal** and copy the `service_role` key for n8n or Flowise.
3. For Metabase, open Control Hub and go to **Browse** > **supabase-{user-name}** > **Secrets** > **supabase-secrets**. Reveal the decoded `POSTGRES_PASSWORD` in **Data**.

The `service_role` key bypasses RLS. Keep it in trusted server-side applications. Do not put it in browser code, screenshots, or a code repository. The official MCP connection exposes the publishable `anon` key, not a replacement for these server credentials.

### Prepare task and cost tables with Lares

Ask Lares to create the shared example data:

```text
Use supabase_official to inspect the public schema and prepare two tables.
project_tasks: id bigint generated by default as identity primary key,
project text not null, title text not null, status text not null default 'todo'.
project_costs: id integer primary key, project text not null, amount integer not null.
Enable RLS on both tables. Do not add anonymous access policies.
Insert these fictional rows into project_costs:
1, Lakeview, 100
2, Riverside, 200
3, Maple, 300
Use migrations for the schema and execute_sql for the sample rows.
Show the proposed changes first. Preserve existing data and avoid duplicate rows.
```

Approve the intended changes, then open **Table Editor** in Studio. `project_costs` should show the three rows, and `project_tasks` is empty until the n8n workflow adds a task. These examples use server credentials or a direct database login, so they can access data without the application user's RLS policy.

::: details Alternative: prepare the tables in SQL Editor
Run this instead of asking Lares to create the tables.

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

### Record a delivery task with n8n

n8n runs automated workflows between apps. This example uses a manual trigger to insert a delivery task into `project_tasks` through the Supabase node.

Ask Lares:

```text
Help me build the n8n example below. Use supabase_official to check
project_tasks and get the API URL. Prepare the manual-trigger workflow
and Supabase Create a row settings for the Lakeview delivery task.
Use an existing n8n Supabase credential if available. Do not ask me to
paste its service_role key into chat. Tell me which setup steps remain,
then help me verify the inserted row without creating duplicate tasks.
```

Use the following steps to complete the connection or check the generated configuration:

1. Open n8n and create a workflow with a manual trigger.
2. Click the **+** next to the trigger, search for **Supabase**, and select **Create a row** to add the node.
3. Click **Connect to Supabase** under **Credential**.
4. In the **Supabase account** window, configure:

    a. **Host**: Enter the Supabase API URL.

    b. **Secret Key**: Enter the `service_role` key.

5. Click **Save**.
6. Set **Table Name or ID** to `project_tasks`.
7. Keep **Data to Send** as **Define Below for Each Column**.
8. Under **Fields to Send**, click **Add Field** and enter the following columns one by one:

   | Field | Value |
   | --- | --- |
   | `project` | `Lakeview` |
   | `title` | `Confirm cabinet delivery for 10 November 2026` |
   | `status` | `todo` |

9. Click **Execute step**.
10. Go to Supabase Studio's **Table Editor**, find `project_tasks`, and check for the new added task.

### Summarize project costs with Metabase

Metabase builds reports and dashboards from a database. This example connects it to Supabase and summarizes the sample project costs.

Ask Lares:

```text
Use supabase_official to check the project_costs example data and prepare
SQL for the three-row report and total. Help me configure the Metabase
PostgreSQL connection below. List any host, port, or credentials that
cannot be obtained through Supabase MCP, and have me enter secrets
directly in Metabase. Help me check that the total is 600 across 3 records.
```

Complete the Metabase connection with the following settings:

1. Open Metabase, and go to **Admin** > **Databases** > **Add database**.
2. Choose PostgreSQL, and enter:

   | Field | Value |
   | --- | --- |
   | Display name | `Home project costs` |
   | Host | The top-level URL collected in step 3 of [Get connection addresses](#get-connection-addresses), without `https://` |
   | Port | The **Exported port** next to `postgres-session` |
   | Database name | `postgres` |
   | Username | `postgres.supabase`, the package's fixed pooler username |
   | Password | The decoded `POSTGRES_PASSWORD` from [Get connection credentials](#get-connection-credentials) |
   | SSL | Off for this package's LAN connection |
3. Save the connection.
4. Click your profile icon in the top-right corner, and select **Main app**.
5. In the top right, click **+ New** > **SQL query**, select the `Home project costs` database.
6. To list the sample project costs, run the following query:

   ```sql
   select project, amount
   from public.project_costs
   order by id;
   ```

   The query returns a table with the `project` and `amount` columns and three rows: `Lakeview` 100, `Riverside` 200, and `Maple` 300.

7. To calculate the total, run the following query:

   ```sql
   select count(*) as records, sum(amount) as total_usd
   from public.project_costs;
   ```

   The query returns one row showing 3 records and a total of 600.

:::warning Database privileges and network scope
This example uses the installation administrator's database login in a Metabase instance you control. Before sharing reporting access, configure a dedicated read-only database account and appropriate Metabase permissions. TCP mappings allow LAN/VPN connections, rather than only the Metabase app. Do not substitute an API key or use a browser HTTPS entrance as the database connection.
:::

### Ask Flowise about project notes

Flowise builds AI chatflows from connected nodes. This example stores project notes in Supabase and answers questions about them, using chat and embedding models from Router. Flowise keeps its own application database and does not need a direct PostgreSQL connection to Supabase for this workflow.

Ask Lares:

```text
Help me prepare the Flowise example below. Use supabase_official to inspect
the database and propose the project_documents table and match function
for 1024-dimensional embeddings. Preserve existing data and show changes
before applying them. Prepare the chatflow settings for default-chat and
default-embedding, with embedding batch size 1 and encoding float.
List the Router, Flowise credential, and version-specific setup I still
need to complete, then help me verify the answer against the source note.
```

Complete the model and application setup below before running the chatflow:

#### Prepare Flowise and Router

The chatflow needs two models that Flowise can call through Router: a chat model to answer questions and an embedding model to vectorize notes. In this part, you install both models, point Router's default routes at them, copy the endpoint that Flowise connects to, and make the route names appear in Flowise's model dropdowns.

1. Install **Qwen3.8-27B (llama.cpp)** and the **Jina CLIP v2** embedding model from Market, and wait for both to finish downloading.

   :::info
   Jina CLIP v2 outputs 1024-dimensional vectors. Keep the same embedding model for document ingestion and retrieval. A different dimension requires a different vector column and a fresh index.
   :::
2. In Router, open the **Default models** page, and assign **Qwen3.8-27B (llama.cpp)** as the chat default and **Jina CLIP v2** as the embedding default. The `default-chat` and `default-embedding` route names then point to these models.
3. Note down the OpenAI-compatible base URL for **Apps in Olares** from Router's connection example, including `/v1`. For other connection locations and key requirements, see [Connect AI apps](https://www.olares.com/docs/manual/best-practices/connect-ai-apps).
4. Flowise's OpenAI nodes pick models from a bundled catalog, which does not include the two Router routes. In Control Hub, go to **Browse** > **flowise-{user-name}** > **Deployments** > **flowise**, open the flowise container's **Terminal**, and run the following script to copy the catalog and add `default-chat` and `default-embedding` to it:

   ::: details Add Router routes to Flowise's model catalog
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
   
5. In Control Hub, open **ConfigMaps > flowise-config**, click the **Edit** icon in the top-right corner, and add the following entry under `data` in the YAML editor.

   ```yaml
   MODEL_LIST_CONFIG_JSON: "/home/node/.flowise/olares-models.json"
   ```

6. Save the change and restart Flowise.

    :::info
    A direct ConfigMap edit can be replaced during an app upgrade. Reapply the entry after upgrading Flowise if the model dropdowns lose the two routes.
    :::

#### (Optional) Apply a WebSocket fix for Flowise image 3.1.4

The Market's Flowise 3.1.4 image runs Flowise 3.1.2 on Node.js 20, and its Supabase dependency can fail with `Node.js 20 detected without native WebSocket support`. Apply the following steps if your installation uses this image.

1. On your computer, create a file named `supabase-websocket.cjs` with the following content:

   ```js
   globalThis.WebSocket ||= require('/usr/local/lib/node_modules/flowise/node_modules/ws');
   ```

2. Upload it with Files to `Data/flowise/`. The file is mounted in Flowise at `/home/node/.flowise/supabase-websocket.cjs`.
3. In Control Hub, open **ConfigMaps > flowise-config**, click the **Edit** icon in the top-right corner, and add the following entry under `data` in the YAML editor:

   ```yaml
   NODE_OPTIONS: "--require=/home/node/.flowise/supabase-websocket.cjs"
   ```

   Append the option if `NODE_OPTIONS` already has a value.

4. Save the change and restart Flowise.

    :::info
    A direct ConfigMap edit can be replaced during an app upgrade. Keep the uploaded file, and reapply the `NODE_OPTIONS` entry if the Supabase node reports the WebSocket error again.
    :::

#### Create a document store

1. Give the following SQL to Lares. Ask it to inspect existing objects, show the planned migration, and use `apply_migration` to create the table and function without replacing existing data. Approve the intended migration.
2. In Studio, check that `project_documents` exists with RLS enabled. You can also run the SQL directly in **SQL Editor** instead of using Lares.

   ::: details Example: project_documents table and match function
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

#### Build the chatflow

1. In Flowise, create a chatflow and add the following nodes:

   | Node | Configuration |
   | --- | --- |
   | Plain Text | <ul><li>**Text**: Enter this project note as an example `Lakeview cabinet delivery is on 10 November 2026. Confirmation reference: LAKEVIEW-6728.`</li><li>**Additional Metadata**: `{"project":"Lakeview"}`</li></ul> |
   | OpenAI Embeddings | <ul><li>**Connect Credential**: Create a credential with any value such as `olares`, Router does not verify the key for apps inside Olares</li><li>**Model Name**: `default-embedding`</li><li>**Batch Size**: `1`, the Jina CLIP v2 service only accepts one text per request</li><li>**Base Path**: Router base URL</li><li>**Encoding Format**: `float`</li></ul> |
   | Supabase | <ul><li>**Supabase Project URL**: The Supabase API URL</li><li>**Connect Credential**: `service_role` key</li><li>**Table Name**: `project_documents`</li><li>**Query Name**: `match_project_documents`</li></ul> |
   | OpenAI | <ul><li>**Connect Credential**: Create a credential with any value such as `olares`, Router does not verify the key for apps inside Olares</li><li>**Model Name**: `default-chat`</li><li>**Base Path**: Router base URL</li></ul> |
   | Conversational Retrieval QA Chain | Answers the question with the chat model using the retrieved notes. No field configuration is needed |

2. Connect the nodes by dragging from each node's output dot to the target node's input dot:

    a. Drag Plain Text's **Document** output to the Supabase node's Document input.

    b. Drag OpenAI Embeddings' output to the Supabase node's Embeddings input.

    c. Drag the Supabase node's **Supabase Retriever** output to Conversational Retrieval QA Chain's **Vector Store Retriever** input.

    d. Drag the OpenAI chat model's output to Conversational Retrieval QA Chain's **Chat Model** input.

3. Save the flow, and click **Upsert Vector Database** on the canvas. This embeds the note and writes it to the `project_documents` table.
4. Ask `What is the Lakeview cabinet delivery date and confirmation reference?`, and compare the answer with the source note.


## Add files and realtime updates

For receipts or project files, create a private bucket in Studio's **Storage** page and configure access policies before using it from the app. The same Supabase API URL supports file upload, download, and short-lived signed URLs. See [Storage access control](https://supabase.com/docs/guides/storage/security/access-control).

For live note updates, add `user_notes` to the `supabase_realtime` publication and subscribe through the SDK. Keep the RLS policy and wait until the database subscription is ready before relying on incoming events. See [Postgres Changes](https://supabase.com/docs/guides/realtime/postgres-changes).

The official self-hosted MCP tools cover database migrations, SQL, schema inspection, type generation, documentation search, and security and performance advisors. Configure Auth users, Storage buckets, SMTP, login redirect URLs, and other application settings through their own UI or APIs. The package does not enable the analytics service required by `query_logs`.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| Lares cannot connect to MCP | Use the protected Studio HTTPS URL ending in `/api/mcp`, with `Streamable HTTP` and `{}` headers. The separate API gateway's `/mcp` route is blocked |
| An external MCP client receives a login redirect | This connection uses Olares app identity from Lares on the same instance. Do not make Studio public or copy these empty-header settings into an unauthenticated external client |
| Lares asks before running SQL | The writable MCP connection treats `execute_sql` as a write-capable tool. Review each operation, or use a separate `?read_only=true` connection for inspection |
| `query_logs` reports that logs are disabled | Log analytics is not enabled in this package. Use the relevant container's logs in Control Hub |
| A query fails with `password authentication failed for user "supabase_read_only_user"` | Have the administrator synchronize that database account's password with the instance's Studio database configuration. Changing the MCP headers does not repair database credentials |
| Login succeeds but REST reports `role "" does not exist` | Inspect the affected user's `auth.users.role` in SQL Editor. A regular app user needs `authenticated`. Correct an empty role, then sign in again or refresh the session to obtain a new token |
| An SDK query returns no rows | Check the signed-in user, row ownership, and RLS policy before changing credentials |
| n8n or Flowise cannot authenticate | Use the Supabase API URL and the legacy `service_role` key |
| Metabase cannot connect | Check the exported port, host reachability, database password, and `postgres.supabase` username |
| Flowise reports missing WebSocket support | Apply the version-specific compatibility step above |
| Flowise embeddings return a single-text error | Set embedding **Batch Size** to `1` |

For backup planning, include both PostgreSQL data and uploaded storage objects.

## Learn more

- [Manage Olares with Lares](lares.md): Select a model and start a conversation.
- [Automate workflows with n8n](n8n.md): Set up n8n and build automation workflows.
- [Use Olares Router as your AI gateway](olares-router.md): Manage model routes and credentials.
- [Enable the self-hosted Supabase MCP server](https://supabase.com/docs/guides/self-hosting/enable-mcp): Understand the upstream server's network access requirements.
