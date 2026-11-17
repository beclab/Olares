---
outline: [2, 3]
description: Use Lares and the official Supabase MCP server to prepare a private app backend on Olares, then connect your application and a reporting tool.
head:
  - - meta
    - name: keywords
      content: Olares, Supabase, Lares, MCP, PostgreSQL, RLS, Metabase, Code Server
app_version: "0.0.15"
doc_version: "2.1"
doc_updated: "2026-11-17"
---

# Build an app backend with Supabase and Lares

Supabase combines PostgreSQL, user authentication, Edge Functions, and an API layer that applications connect to directly. Its Studio interface lets you manage these services on your own Olares device.

Lares can use Supabase's official Model Context Protocol (MCP) server to create tables and configure row-level security (RLS) from a conversation. The example in this guide is a home renovation notes application, where each user can access only their own notes. A separate table of fictional project costs provides data for reporting in Metabase.

## Learning objectives

In this guide, you will learn how to:

- Connect Lares to the official Supabase MCP server.
- Ask Lares to prepare an app's tables and per-user access policy.
- Connect an application with ordinary user credentials and check its data access.
- Connect a reporting tool to the same database with a direct connection.

## Prerequisites

- Lares with a working model selected. See [Get started with Lares](lares.md#get-started).
- Code Server installed from Market for the notes application.
- Metabase installed from Market for reporting.

## Install Supabase

1. Open Market and search for "Supabase".
2. Click **Get**, then **Install**, and wait for installation to complete.

## Connect Lares to the Supabase MCP server

The MCP connection lets Lares manage the database through Supabase Studio.

:::info Connection authentication
Lares connects through Olares authentication on the same instance, so you do not need to provide an additional access token or API key. Keep the **Supabase Studio** entrance set to **Internal** in Olares Settings.
:::

1. Open Olares Settings and go to **Applications** > **Supabase**. Under **Entrances**, click **Supabase Studio**.
2. Copy the **Endpoint** address, and append `/api/mcp` to the end. For example:

   ```text
   https://304244530.laresprime.olares.com/api/mcp
   ```

3. Open Lares, go to **Settings** > **MCP**, and click **Add server**.
4. Specify the following settings:

   | Setting | Value |
   | --- | --- |
   | Server name | `supabase_official` |
   | Transport | `Streamable HTTP` |
   | MCP URL | The copied endpoint address ending in `/api/mcp` |

5. Click **Save**. Lares connects and loads the server's tools.
6. Start a new conversation in Lares to check the MCP connection with the following prompt:

   ```text
   Use supabase_official to list the tables in the public schema and return
   this Supabase installation's API URL. Use the MCP tools directly and do not
   modify data.
   ```

   Lares should return the Supabase API URL and a list of tables in the `public` schema. If you have not created any tables yet, the list will be empty.

## Create the notes table and access policy with Lares

Each note stores its owner's Supabase Auth user ID in `owner_id`. The access policy compares this value with the signed-in user's ID, allowing that user to read and change only their own notes.

:::warning Database administration access
The MCP server can change database tables and access policies. Review the proposed SQL and approve only the intended changes.
:::

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

2. Review the proposed SQL. When Lares asks for approval, approve only the intended changes and wait for it to finish.
3. Open Supabase Studio, and go to **Table Editor**. Find `user_notes` and check the `id`, `owner_id`, and `note` columns and the **RLS policy** indicator showing one policy.

4. Ask Lares to inspect `user_notes` and summarize its RLS policy:

   ```text
   Use supabase_official to inspect public.user_notes. Confirm that RLS
   is enabled and summarize the policy's name, the role it applies to, and
   both the USING read condition and the WITH CHECK write check.
   ```

   The summary should show RLS enabled and one `own_notes` policy for the `authenticated` role, with both conditions comparing `auth.uid()` with `owner_id`.

::: details Alternative: create the notes table in SQL Editor
If you have not created `user_notes` with Lares, open Supabase Studio's **SQL Editor**, paste the following query, and click **Run**. Then check the table and policy as described above.

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

## Connect the notes application to Supabase

### Create the notes application with Lares

Use Lares to create a small command-line notes application, then run it in the Code Server terminal to sign in, add notes, and list your notes. The notes are stored in Supabase's `user_notes` table.

1. Ask Lares to generate the program and its configuration files:

   ```text
   Help me connect a private notes app to Supabase. Use supabase_official
   to inspect user_notes and its RLS policy, and get the API URL and anon key.
   Create notes.mjs using @supabase/supabase-js v2 and Node.js 22.
   Read SUPABASE_URL, SUPABASE_ANON_KEY, NOTES_EMAIL, and NOTES_PASSWORD
   from process.env. I will load .env with node --env-file=.env.
   Sign in with signInWithPassword and support these commands:
   - whoami: show the signed-in user's email and ID, never session tokens
   - add "text": insert into user_notes with owner_id set to that user's ID,
     then show the saved note and its ID
   - list: select id and note from user_notes, relying on RLS for isolation
   Report SDK errors and exit after each command.
   Save notes.mjs, .env, .env.example, and .gitignore in your workspace.
   Fill .env with the API URL and anon key. Leave NOTES_EMAIL and
   NOTES_PASSWORD empty, and ignore .env and node_modules/ in Git.
   Use placeholders in .env.example. Keep passwords out of the conversation
   and source code. Tell me the exact file location. Do not create Auth
   users, change the schema or policies, or run the app.
   Stop after preparing the files.
   ```

   Lares prepares four files for the same application:

   | File | Purpose |
   | --- | --- |
   | `notes.mjs` | The program you run. It signs in to Supabase, adds notes, and lists the signed-in user's notes. |
   | `.env` | The configuration used when running the program. Lares fills in the API URL and `anon` key. You will add your Supabase Auth user's email and password. |
   | `.env.example` | A configuration template with placeholders, for reference or setting up another copy of the application. |
   | `.gitignore` | Tells Git to exclude `.env` and the `node_modules` dependency folder from version control. |

2. In Olares Files, open **Data** > **lares** > **data** > **workspace**, the default Lares workspace. If Lares saved the files in a subfolder, open that folder. Copy `notes.mjs`, `.env`, `.env.example`, and `.gitignore`.
3. In Olares Files, go to **Home** > **Code**, create a folder named `notes-app`, and paste all four files into it, including the hidden files. Code Server can access this folder.

### Create a Supabase Auth user

Create an application account in Supabase Auth for the script to sign in with. This account is separate from your Olares account.

1. Open Supabase Studio and go to **Authentication** > **Users** > **Add user** > **Create new user**.
2. Enter an email such as `demo@example.com` and a password of your own. Keep **Auto confirm user?** enabled, then click **Create user**. Auto confirmation lets you sign in for this example without receiving a confirmation email.

### Run the notes application in Code Server

1. Open Code Server from the Launchpad. Use **File** > **Open Folder** to open `/home/coder/notes-app`, which corresponds to **Home** > **Code** > **notes-app** in Olares Files.
2. If prompted, confirm that you trust the folder. Check that `notes.mjs` and `.env` appear in the Explorer.
3. Open **Terminal** > **New Terminal**. This example uses Node.js 22. Check `node -v`. If Node.js is missing or older than version 22, install it with nvm:

   ```bash
   touch "$HOME/.bashrc"
   curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.8/install.sh | bash
   \. "$HOME/.nvm/nvm.sh"
   nvm install 22
   ```

   Check `node -v` and `npm -v` before continuing. If a new terminal cannot find `node`, run `. "$HOME/.nvm/nvm.sh"` to load nvm in that terminal.
4. In the Code Server terminal, switch to the project folder and install the Supabase JavaScript SDK:

   ```bash
   cd /home/coder/notes-app
   npm install @supabase/supabase-js@2.118.0
   ```

5. Open `.env` in Code Server. Keep the `SUPABASE_URL` and `SUPABASE_ANON_KEY` values that Lares filled in. The URL must be the Supabase API entrance, without `/rest/v1` or `/api/mcp`. Fill in the user credentials and save the file:

   ```dotenv
   NOTES_EMAIL="demo@example.com"
   NOTES_PASSWORD="your-user-password"
   ```

   Replace these example values with the Supabase Auth user you created. Keep the password in `.env`, not in the Lares conversation or `notes.mjs`.

   ::: details Alternative: copy the connection details from Supabase Studio and Olares Settings
   In Olares Settings, go to **Applications** > **Supabase**. Under **Entrances**, open **Supabase API** and copy its **Endpoint** into `SUPABASE_URL` in `.env`.

   In Supabase Studio, open **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**, and copy the `anon` key into `SUPABASE_ANON_KEY`. Do not use `service_role`: it bypasses RLS and allows access to other users’ notes.
   :::

6. In the same terminal, run the following commands to sign in, add a note, and read it back. The `--env-file=.env` option loads the connection settings and user credentials from `.env`:

   ```bash
   node --env-file=.env notes.mjs whoami
   node --env-file=.env notes.mjs add "Ask the Lakeview contractor to confirm the delivery window."
   node --env-file=.env notes.mjs list
   ```

   The terminal shows the signed-in user, the new note's ID, and the saved note. Each `add` command creates a new row.

### Check that users can only access their own notes

1. In Supabase Studio, repeat [Create a Supabase Auth user](#create-a-supabase-auth-user) with a different email, such as `demo2@example.com`, and a separate password.
2. In Code Server, replace only `NOTES_EMAIL` and `NOTES_PASSWORD` in `.env` with the second user's credentials, then save the file. Keep the same API URL and `anon` key.
3. Run the following commands:

   ```bash
   node --env-file=.env notes.mjs whoami
   node --env-file=.env notes.mjs list
   ```

   The signed-in user should now be the second user. This new user's list should be empty, even though the first user's note is still in the database.
4. Add a note for the second user, then list their notes:

   ```bash
   node --env-file=.env notes.mjs add "Compare the Riverside flooring quotes."
   node --env-file=.env notes.mjs list
   ```

   The list should contain the Riverside note and exclude the first user's Lakeview note.
5. Restore the first user's email and password in `.env`, save the file, and run `node --env-file=.env notes.mjs list` again. You should see their Lakeview note without the Riverside note.

## Report on project costs with Metabase

Metabase connects directly to PostgreSQL to query project costs. Lares creates the sample data through Supabase MCP, and you configure the database connection and run the report queries in Metabase.

### Complete the initial Metabase setup

If this is your first time opening Metabase, complete its setup wizard before adding the Supabase connection:

1. Open Metabase and click **Let's get started**.
2. Enter your name, email, team name, and a password for your Metabase administrator account. Click **Next** and answer the usage question.
3. At **Add your data**, click **Continue with sample data**. You will add Supabase from the administration page below.
4. Choose your usage data preference, click **Finish**, then click **Take me to Metabase**.

### Collect the database connection details

For Metabase, collect the Supabase database host, exported session port, and database password. This connection uses the database username `postgres.supabase`.

#### Get the database host and port

1. In Olares **Settings** > **Applications** > **Supabase**, note the **Exported port** under **Permissions** > **Exported ports** > `postgres-session`.

2. On your computer, [install Olares CLI](../developer/cli-install.md#install-olares-cli-first) and [log in to your Olares instance](../developer/cli-install.md#log-in-to-olares) if you have not already done so. In your computer's terminal, run:

   ```bash
   olares-cli settings apps get supabase
   ```

3. Copy the top-level **URL** from the command output as the database host, without `https://` or a path.
4. In the **Ports** list, check that the **EXPOSE** value for `postgres-session` matches the exported port you noted in step 1. Use this port in Metabase.

#### Get the database password

1. In Control Hub, go to **Browse** > **supabase-{user-name}** > **Deployments** > **db**, replacing `{user-name}` with your Olares username. Open its Pod, then open the **db** container's terminal.
2. Run the following commands to display the database password:

   ```sh
   . /runtime/keys.env
   printf '%s\n' "$POSTGRES_PASSWORD"
   ```

3. Keep this terminal available. You will copy the password into Metabase when adding the connection. Do not paste it into the Lares conversation.

### Create the project costs table with Lares

1. In Lares, send the following prompt to create `project_costs` with three fictional renovation costs. The `amount` values in this example are in US dollars.

   ```text
   Use supabase_official to inspect the public schema and prepare a table.
   project_costs: id integer primary key, project text not null, amount integer not null.
   Enable RLS on the table. Do not add anonymous access policies.
   Insert these fictional rows:
   1, Lakeview, 100
   2, Riverside, 200
   3, Maple, 300
   Use a migration for the schema and execute_sql for the sample rows.
   Show the proposed changes first. Preserve existing data and avoid duplicate rows.
   ```

2. Review the proposed changes and confirm them in Lares.
3. When a tool call shows **Waiting for approval**, inspect its SQL and click **Allow once** for the intended operation. Repeat as needed for the migration, sample inserts, and verification queries.
4. Wait for Lares to finish, then open **Table Editor** in Supabase Studio and select `project_costs`. Check that it contains the three rows from the prompt.

::: details Alternative: prepare the table in SQL Editor
If you have not created `project_costs` with Lares, run the following query in Supabase Studio's **SQL Editor**.

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

### Connect Metabase to Supabase

1. In Metabase, open the **Settings** menu in the top-right corner and select **Admin**. Go to **Databases** > **Add database**.
2. Choose PostgreSQL, and enter:

   | Field | Value |
   | --- | --- |
   | Display name | `Home project costs` |
   | Host | The top-level URL collected in step 3 of [Get the database host and port](#get-the-database-host-and-port), without `https://` |
   | Port | The **Exported port** collected in step 1 of [Get the database host and port](#get-the-database-host-and-port) |
   | Database name | `postgres` |
   | Username | `postgres.supabase` |
   | Password | The `POSTGRES_PASSWORD` from [Get the database password](#get-the-database-password) |
   | Use a secure connection (SSL) | Off for this package's LAN connection |

   ::: warning Database administrator access
   This connection gives Metabase administrator-level access to Supabase, including access to `project_costs` without an RLS read policy. Only use these credentials in a Metabase instance you control.
   :::

3. Click **Save**.

4. On the **Home project costs** database page, check that **Connection and sync** shows **Connected**.

### Query project costs in Metabase

1. Open the **Settings** menu in the top-right corner and select **Main app**.
2. In the top right, click **New** > **SQL query**. Check the database selector above the SQL editor and select `Home project costs` if it is not already selected.
3. Paste the following query into the SQL editor, then click the run button to list the sample project costs:

   ```sql
   select project, amount
   from public.project_costs
   order by id;
   ```

   The query returns a table with the `project` and `amount` columns and three rows: `Lakeview` 100, `Riverside` 200, and `Maple` 300.

4. Replace the SQL in the editor with the following query, then click the run button again to calculate the total:

   ```sql
   select count(*) as records, sum(amount) as total_usd
   from public.project_costs;
   ```

   The query returns one row showing 3 records and a total of 600.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| Lares cannot connect to MCP | Use the Supabase Studio HTTPS URL ending in `/api/mcp` with `Streamable HTTP` transport. Keep the **Supabase Studio** entrance set to **Internal**. The Supabase API entrance's `/mcp` route is blocked |
| Code Server cannot find `notes.mjs` or `.env` | Check that both files are in **Home** > **Code** > **notes-app** in Olares Files, then run the commands from `/home/coder/notes-app` in Code Server |
| `node -v` reports a missing `libatomic.so.1` | In the Code Server terminal, run `sudo apt-get install -y libatomic1`, then check `node -v` again |
| Login succeeds but REST reports `role "" does not exist` | Inspect the affected user's `auth.users.role` in SQL Editor. A regular app user needs `authenticated`. Correct an empty role, then sign in again or refresh the session to obtain a new token |
| An SDK query returns no rows | Check the signed-in user, row ownership, and RLS policy before changing credentials |
| Metabase cannot connect | Check the exported port, host reachability, database password, and `postgres.supabase` username |

## Learn more

- [Manage Olares with Lares](lares.md): Select a model and start a conversation.
- [Enable the self-hosted Supabase MCP server](https://supabase.com/docs/guides/self-hosting/enable-mcp): Understand the upstream server's network access requirements.
