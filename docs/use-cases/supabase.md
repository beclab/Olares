---
outline: [2, 3]
description: Keep application data on Olares with Supabase. Connect n8n workflows, Metabase reports, Flowise document search, and a project in Code Server to the same self-hosted backend.
head:
  - - meta
    - name: keywords
      content: Olares, Supabase, PostgreSQL, n8n, Metabase, Flowise, Code Server, self-hosted backend, RAG
app_version: "0.0.13"
doc_version: "1.1"
doc_updated: "2026-10-30"
---

# Connect your apps to a Supabase backend on Olares

Supabase combines a PostgreSQL database with authentication, file storage, realtime updates, and Edge Functions. Studio provides a browser interface for managing your data and services.

On Olares, these services run on your own device. Your automation workflows, reports, document search, and custom applications can share a backend without keeping a separate copy of the data in each service. AI requests follow the models you select in Router. Choose local models to keep that processing on your device too.

The examples use fictional home renovation projects. n8n records a delivery task, Metabase summarizes costs, Flowise answers questions about project notes, and Code Server connects a small application with user-level access controls.

## Learning objectives

- Choose the correct Supabase address and credential for each app.
- Record tasks with n8n and analyze costs with Metabase.
- Build document retrieval in Flowise with models from Router.
- Connect a JavaScript application using ordinary user credentials and row-level security.

## Prerequisites

This guide connects several applications to one Supabase project. Each connection works on its own, so install only the apps you plan to follow:

- n8n for recording tasks
- Metabase for cost reports
- Flowise for document search
- Code Server for the custom application

Install these apps from Market. For n8n's initial account setup, see [Automate workflows with n8n](n8n.md#set-up-n8n).

## Install Supabase

1. Open Market and search for "Supabase".
2. Click **Get**, then **Install**, and wait for installation to complete. The app appears on the Launchpad as **Supabase Studio**.

## Collect connection details

### Get connection addresses

1. Open Olares Settings, go to **Applications** > **Supabase**, and then copy the following values needed by the corresponding app:

   | Value | Where to find it | Used by |
   | --- | --- | --- |
   | Supabase API URL | **Supabase API** > **Endpoint** address, without `/rest/v1` | n8n, Flowise, Code Server |
   | Database port | **Permissions** > **Exported ports** > `postgres-session` > **Exported port** | Metabase |
   | Supabase Studio URL | **Supabase Studio** > **Endpoint** address, can differ from the Supabase API URL | Your browser |
2. Open Control Hub, go to **Terminal** > **Olares**, and then run the following command:

    ```bash
    olares-cli settings apps get supabase
    ```

    If the command fails with `Error: no Olares profile is configured`, [log in to Olares](https://www.olares.com/docs/developer/cli-install#log-in-to-olares) first and run it again.

    ::: details Example output
    ```text
    root@olares:/# olares-cli settings apps get supabase
    Name:               supabase
    Title:              Supabase
    State:              running
    Owner:              laresprime
    Namespace:          supabase-laresprime
    Deployment:         -
    URL:                30424453.laresprime.olares.com
    System App:         no
    Cluster-Scoped:     no
    Mobile Supported:   no
    Required GPU:       -

    Ports:
      NAME                  HOST       PORT  EXPOSE  PROTOCOL
      postgres-session      supavisor  5432  45432   tcp
      postgres-transaction  supavisor  6543  46543   tcp

    Entrances:
      NAME         TITLE            STATE    AUTH LEVEL  INVISIBLE  URL
      supabase     Supabase Studio  running  internal    no         304244530.laresprime.olares.com
      supabaseapi  Supabase API     running  internal    yes        304244531.laresprime.olares.com
    ```
    :::
3.  Use the top-level **URL** as the database host for Metabase, without `https://` or a path. In this example, it is `30424453.laresprime.olares.com`.

### Get connection credentials

1. Open Supabase Studio, and go to **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**.
2. Copy the `anon` key. The Code Server application uses it.
3. Click **Reveal** and copy the `service_role` key. The n8n and Flowise connections use it.

   :::warning
   The `service_role` key bypasses row-level security and grants full database access. Treat it like a password and store it only in a trusted backend app.
   :::
4. Open Control Hub, and go to **Browse** > **supabase-{user-name}** > **Secrets** > **supabase-secrets**.
5. In the **Data** section, click the <i class="material-symbols-outlined">visibility</i> icon to reveal the decoded values, and then copy the value for `POSTGRES_PASSWORD`. Metabase uses it.

## Prepare data for n8n and Metabase

This section creates the tables the n8n and Metabase examples use, with fictional home renovation projects as sample data. `project_tasks` records tasks for the n8n connection, and `project_costs` stores costs for Metabase reports.

1. Open Supabase Studio, and go to **SQL Editor**.
2. Paste the following SQL into the query editor, and then click **Run**.

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
3. Open **Table Editor**, and check that `project_tasks` and `project_costs` are created, both with the RLS badge next to their names. 

    `project_costs` shows the three sample rows, and `project_tasks` is empty until the n8n example inserts the first task.

:::info
RLS is enabled on both tables, but the n8n, Metabase, and Flowise connections use trusted credentials or a direct database login that bypass it. The Code Server example relies on an RLS policy for per-user access, which shows RLS in action.
:::

## Connect other apps

The examples below connect four types of applications to the Supabase backend. Each uses a different address and credential from [Collect connection details](#collect-connection-details).

### Record a delivery task with n8n

n8n runs automated workflows between apps. This example uses a manual trigger to insert a delivery task into `project_tasks` through the Supabase node.

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

8. Click **Execute step**.
9. Go to Supabase Studio's **Table Editor**, find `project_tasks`, and check for the new added task.

### Summarize project costs with Metabase

Metabase builds reports and dashboards from a database. This example connects it to Supabase and summarizes the sample project costs.

1. Open Metabase, and go to **Admin** > **Databases** > **Add database**.
2. Choose PostgreSQL, and enter:

   | Field | Value |
   | --- | --- |
   | Display name | `Home project costs` |
   | Host | The top-level URL collected in step 3 of [Get connection addresses](#get-connection-addresses), without `https://` |
   | Port | The **Exported port** next to `postgres-session` |
   | Database name | `postgres` |
   | Username | `postgres.supabase`, the package's fixed pooler username |
   | Password | The decoded `POSTGRES_PASSWORD` collected in step 5 of [Get connection credentials](#get-connection-credentials) |
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

#### Prepare Flowise and Router

1. In Router, open the **Default models** page, and assign **Qwen3.8-27B (llama.cpp)** as the chat default and **Jina CLIP v2** as the embedding default. The `default-chat` and `default-embedding` route names then point to these models.
2. Copy the OpenAI-compatible **Base URL** for **Apps in Olares** from Router's connection example, including `/v1`. For other connection locations and key requirements, see [Connect AI apps](https://www.olares.com/docs/manual/best-practices/connect-ai-apps).
3. In Control Hub, go to **Browse** > **flowise-{user-name}** > **Deployments** > **flowise**, open the flowise container's **Terminal**, and run the following script:

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

   This keeps the bundled model options and adds the two routes. If you already maintain a custom catalog, add the routes to that file instead.
4. In **ConfigMaps > flowise-config**, add the following entry under `data`, and restart Flowise:

   ```yaml
   MODEL_LIST_CONFIG_JSON: "/home/node/.flowise/olares-models.json"
   ```

   A direct ConfigMap edit can be replaced during an app upgrade.

:::info
Jina CLIP v2 outputs 1024-dimensional vectors. Keep the same embedding model for document ingestion and retrieval. A different dimension requires a different vector column and a fresh index.
:::

:::details WebSocket compatibility for Flowise image 3.1.4
This image reports Flowise 3.1.2 and uses Node.js 20. Its Supabase dependency can fail with `Node.js 20 detected without native WebSocket support`. Apply this workaround to that version before continuing:

1. On your computer, create `supabase-websocket.cjs` containing:

   ```js
   globalThis.WebSocket ||= require('/usr/local/lib/node_modules/flowise/node_modules/ws');
   ```

2. Upload it with Files to `Data/flowise/`. The file is mounted in Flowise at `/home/node/.flowise/supabase-websocket.cjs`.
3. In Control Hub, open the Flowise namespace, select **ConfigMaps > flowise-config**, and add the following entry under `data`:

   ```yaml
   NODE_OPTIONS: "--require=/home/node/.flowise/supabase-websocket.cjs"
   ```

   Append the option if `NODE_OPTIONS` already has a value.
4. Restart the Flowise workload. See [Configure variables in Control Hub](https://www.olares.com/docs/manual/olares/settings/manage-app-env#configure-variables-in-control-hub).

Direct Control Hub edits can be replaced during an app upgrade. Keep the file and reapply the setting, or use an app package that includes the compatibility fix.
:::

#### Create a document store

1. Open Supabase Studio, and go to **SQL Editor**.
2. Run the following query:

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
   | Plain Text | A project note in **Text**. Set **Additional Metadata** to `{"project":"Lakeview"}` |
   | OpenAI Embeddings | **Model Name** `default-embedding`, Router **Base Path**, **Batch Size** `1`, **Encoding Format** `float` |
   | Supabase | The Supabase API URL in **Supabase Project URL**, a credential containing `service_role`, **Table Name** `project_documents`, **Query Name** `match_project_documents` |
   | ChatOpenAI | **Model Name** `default-chat`, Router **Base Path** |
   | Retrieval QA Chain | Connect the chat model and the Supabase retriever |

   For the first note, use:

   ```text
   Lakeview cabinet delivery is on 10 November 2026.
   Confirmation reference: LAKEVIEW-6728.
   ```
2. Connect Plain Text's **Document** and OpenAI Embeddings to the Supabase node. Connect **Supabase Retriever** and ChatOpenAI to Retrieval QA Chain.
3. Save the flow, and run the vector upsert action before asking a question.
4. Ask “What is the Lakeview cabinet delivery date and confirmation reference?”, and compare the answer with the source note.
5. Add another Plain Text node for Riverside with its own note and `{"project":"Riverside"}` metadata. Set **Supabase Metadata Filter** to that JSON object to limit retrieval to Riverside.

For model credential fields, apps inside Olares do not need a Router API key. Use `olares` as a placeholder when the field requires a value.

The single-item embedding batch is required by the Jina CLIP v2 service used here. Change batch size only when the selected embedding service supports it.

### Build a small application in Code Server

Code Server runs a code editor in the browser. This example builds a small Node.js application that signs in as a regular user and reads only its own notes, so you can see row-level security in action.

1. Open Code Server from the Launchpad, and open the built-in terminal with **Terminal** > **New Terminal**. Code Server is only the code editor and does not include a Node.js runtime. Install Node.js 22 or later with nvm:

   ```bash
   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.8/install.sh | bash
   \. "$HOME/.nvm/nvm.sh"
   nvm install 22
   ```

   If `node -v` reports `libatomic.so.1: cannot open shared object file`, run `sudo apt-get install -y libatomic1` and check `node -v` again.
2. Open Supabase Studio, go to **SQL Editor**, and run the following query:

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
3. In Supabase Studio, open **Authentication** > **Users** > **Add user** > **Create new user**. Enter an email and password, keep **Auto confirm user?** enabled for this local example, and create the user.

   Use any email address and password you choose, such as `demo@example.com`. This creates an application user in Supabase Auth, not an Olares login or a database account. The script below signs in with this email and password, and the RLS policy limits each signed-in user to their own notes.
   
4. In the Code Server project terminal, install the SDK:

   ```bash
   npm install @supabase/supabase-js@2.118.0
   ```
5. In Code Server, create a file named `.env` in the project directory with the following content:

   ```dotenv
   SUPABASE_URL=https://YOUR_SUPABASE_API_URL
   SUPABASE_ANON_KEY=YOUR_LEGACY_ANON_KEY
   SUPABASE_USER_EMAIL=YOUR_USER_EMAIL
   SUPABASE_USER_PASSWORD=YOUR_USER_PASSWORD
   ```

   Replace the placeholders with the Supabase API URL from [Get connection addresses](#get-connection-addresses), the legacy `anon` key from [Get connection credentials](#get-connection-credentials), and the email and password of the user you created.
6. Add `.env` to `.gitignore`, and create `notes.mjs`:

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

   const { error: insertError } = await client.from('user_notes').insert({
     owner_id: session.user.id,
     note: 'Ask the Lakeview contractor to confirm the delivery window.'
   });
   if (insertError) throw insertError;

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
7. Run `node --env-file=.env notes.mjs`. The terminal displays the user's notes and the response from the included `hello` Edge Function. A different signed-in user cannot read these notes under the policy above. Do not replace the `anon` key with `service_role` to bypass a policy error.

## Extend your application

Use **Storage** in Studio to create a private bucket for receipts or project files. Define storage policies for the intended users before accessing it with the SDK. Uploading, downloading, and short-lived signed URLs use the same API address. See [Storage access control](https://supabase.com/docs/guides/storage/security/access-control).

For live updates, add the intended table to the `supabase_realtime` publication and subscribe with the SDK. Keep RLS policies in place. Wait for the database subscription to be ready before relying on the first event, especially just after adding a table. See [Postgres Changes](https://supabase.com/docs/guides/realtime/postgres-changes).

Authentication, storage, realtime updates, and functions let your application grow beyond a shared database. External sign-in providers and email delivery need their own configuration. A local, auto-confirmed user does not configure those services.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| n8n or Flowise cannot authenticate | Use the **Supabase API** URL and the legacy `service_role` key |
| Metabase cannot connect | Check the app's mapped ports, reachable host, database password, and the `postgres.supabase` username |
| Metabase connects but cannot see data | Check the selected schema and the database account's permissions |
| Flowise reports missing WebSocket support | Apply the version-specific compatibility step above |
| Flowise embeddings return a single-text error | Set embedding **Batch Size** to `1` |
| An SDK query returns no rows | Check the signed-in user and RLS policy before changing credentials |

For backup planning, include both PostgreSQL data and uploaded storage objects. A Studio session or a successful application query is not a database backup.

## Learn more

- [Automate workflows with n8n](n8n.md): Set up n8n and build automation workflows on Olares.
- [Use Olares Router as your AI gateway](olares-router.md): Manage model routes and credentials in Router.
