---
outline: [2, 3]
description: Use Lares to build a Supabase backend for a private notes application on Olares, verify per-user access, and query project costs with Metabase.
head:
  - - meta
    - name: keywords
      content: Olares, Supabase, Lares, MCP, PostgreSQL, RLS, Metabase, Code Server
app_version: "0.0.15"
doc_version: "2.1"
doc_updated: "2026-11-17"
---

# Build an app backend with Supabase and Lares

Supabase provides a PostgreSQL database, user authentication, and APIs for applications on your Olares device. With Lares connected through the Supabase Model Context Protocol (MCP) server, you can create tables and access policies through conversation.

In this tutorial, you will build and run a command-line notes application, using row-level security (RLS) to keep each user's notes private. You will also use Metabase to query sample project costs stored in a separate database table.

## Learning objectives

In this guide, you will learn how to:

- Connect Lares to the Supabase MCP server.
- Create a database table to store users’ notes and configure per-user access with RLS.
- Use Lares to generate a notes application, then run it in Code Server and sign in with a user account created in Supabase.
- Verify that each user can access only their own notes.
- Connect Metabase to Supabase and query sample project costs.

## Prerequisites

- Lares with a working model selected. See [Get started with Lares](lares.md#get-started).
- Code Server installed from Market for the notes application.
- Metabase installed from Market for reporting.

## Install Supabase

1. Open Market and search for "Supabase".
2. Click **Get**, then **Install**, and wait for installation to complete.

## Connect Lares to the Supabase MCP server

Connect Lares to Supabase so you can use database tools in a conversation.

:::info Connection authentication
Lares connects through Olares authentication on the same instance, so you do not need to provide an additional access token or API key. Keep the **Supabase Studio** entrance set to **Internal** in Olares Settings.
:::

1. Open Olares Settings, and go to **Applications** > **Supabase** > **Supabase Studio**.
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

## Build and run a notes application with Supabase

Use Lares to prepare a Supabase database table, configure access policies, and generate a notes application. Then run the application in Code Server and check that each user can access only their own notes.

### Create the notes table and access policy with Lares

With the MCP connection ready, use Lares to create a database table for the notes application and an access policy so each user can access only their own notes.

Each note stores its owner's user ID in `owner_id`. The policy compares this value with the signed-in user's ID to control access.

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

### Create the notes application with Lares

The command-line application signs users in, adds notes, and lists their notes from Supabase's `user_notes` table.

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
   Save notes.mjs, .env, .env.example, and .gitignore directly in your
   workspace root, not in a subfolder.
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

2. Open Olares Files, go to **Data** > **lares** > **data** > **workspace**, and copy `notes.mjs`, `.env`, `.env.example`, and `.gitignore`.
3. In Olares Files, go to **Home** > **Code**, create a folder named `notes-app`, and paste all four files into it. Code Server can access this folder.

### Create a user account for the notes application

Create a user account in Supabase to sign in to the notes application. This account is separate from your Olares account.

1. Open Supabase Studio and go to **Authentication** > **Users** > **Add user** > **Create new user**.
2. Enter an email such as `demo@example.com` and a password of your own. Keep **Auto confirm user?** enabled, then click **Create user**. Auto confirmation lets you sign in for this example without receiving a confirmation email.

### Run the notes application in Code Server

1. Open Code Server from the Launchpad.
2. Use **File** > **Open Folder** to open `/home/coder/notes-app`, which corresponds to **Home** > **Code** > **notes-app** in Olares Files.
3. If prompted, confirm that you trust the folder. Check that `notes.mjs` and `.env` appear in the Explorer.
4. Open **Terminal** > **New Terminal**.
5. Run `node -v` to check the installed Node.js version. This example uses Node.js 22.
6. If Node.js is missing or older than version 22, run the following four commands in order to install it with nvm:

   ```bash
   touch "$HOME/.bashrc"
   curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.8/install.sh | bash
   . "$HOME/.nvm/nvm.sh"
   nvm install 22
   ```

7. Run `node -v` and `npm -v` to confirm that both commands display version numbers before continuing. If you installed Node.js in the previous step, `node -v` should show `v22.x.x`.
8. In the Code Server terminal, switch to the project folder and install the Supabase JavaScript SDK:

   ```bash
   cd /home/coder/notes-app
   npm install @supabase/supabase-js@2.118.0
   ```

9. Open `.env` in Code Server. Leave `SUPABASE_URL` and `SUPABASE_ANON_KEY` unchanged. Set `NOTES_EMAIL` and `NOTES_PASSWORD` to the email and password of the notes application account you created, then save the file:

   ```dotenv
   NOTES_EMAIL="your-email@example.com"
   NOTES_PASSWORD="your-password"
   ```

   ::: warning Keep your password private
   Store the password in `.env`. Do not paste it into the Lares conversation or include it in `notes.mjs`.
   :::

   ::: details Alternative: copy the connection details from Supabase Studio and Olares Settings
   Open Olares Settings, and go to **Applications** > **Supabase** > **Supabase API**. Copy its **Endpoint** into `SUPABASE_URL` in `.env`, without `/rest/v1` or `/api/mcp`.

   In Supabase Studio, open **Project Settings** > **API Keys** > **Legacy anon, service_role API keys**, and copy the `anon` key into `SUPABASE_ANON_KEY`. Do not use `service_role`: it bypasses RLS and allows access to other users’ notes.
   :::

10. In the same terminal, run the following commands to sign in, add a note, and read it back. The `--env-file=.env` option loads the connection settings and user credentials from `.env`:

    ```bash
    node --env-file=.env notes.mjs whoami
    node --env-file=.env notes.mjs add "Ask the Lakeview contractor to confirm the delivery window."
    node --env-file=.env notes.mjs list
    ```

    The terminal shows the signed-in user, the new note's ID, and the saved note. Each `add` command creates a new row.

### Check that users can only access their own notes

1. In Supabase Studio, repeat [Create a user account for the notes application](#create-a-user-account-for-the-notes-application) with a different email, such as `demo2@example.com`, and a separate password.
2. In Code Server, open `.env` and update `NOTES_EMAIL` and `NOTES_PASSWORD` with the second user's email and password. Leave `SUPABASE_URL` and `SUPABASE_ANON_KEY` unchanged, then save the file.
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

## Query Supabase data with Metabase

Metabase connects directly to PostgreSQL to query project costs. Lares creates the sample data through Supabase MCP, and you configure the database connection and run the report queries in Metabase.

### Complete the initial Metabase setup

If this is your first time opening Metabase, complete its setup wizard before adding the Supabase connection:

1. Open Metabase and click **Let's get started**.
2. Enter your name, email, team name, and a password for your Metabase administrator account. Click **Next** and answer the usage question.
3. At **Add your data**, click **Continue with sample data** to continue setup without connecting Supabase yet. You will add the Supabase connection in [Connect Metabase to Supabase](#connect-metabase-to-supabase).
4. Choose your usage data preference, click **Finish**, then click **Take me to Metabase**.

### Collect the database connection details

For Metabase, collect the Supabase database host, exported session port, and database password. This connection uses the database username `postgres.supabase`.

#### Get the database host and port

1. Open Olares Settings, and go to **Applications** > **Supabase** > **Permissions** > **Exported ports** > `postgres-session`. Note down the **Exported port**.

2. On your computer, [install Olares CLI](../developer/cli-install.md#install-olares-cli-first) and [log in to your Olares instance](../developer/cli-install.md#log-in-to-olares) if you have not already done so. In your computer's terminal, run:

   ```bash
   olares-cli settings apps get supabase
   ```

3. Note down the top-level **URL** from the command output as the database host, without `https://` or a path.
4. In the **Ports** list, check that the **EXPOSE** value for `postgres-session` matches the exported port you noted down in step 1. Use this port in Metabase.

#### Get the database password

1. In Control Hub, go to **Browse** > **supabase-{user-name}** > **Deployments** > **db**, open its pod, and then open the **db** container's terminal.
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
| Lares cannot connect to MCP | Use the Supabase Studio HTTPS URL ending in `/api/mcp` with `Streamable HTTP` transport. Keep the **Supabase Studio** entrance set to **Internal**. |
| Code Server cannot find `notes.mjs` or `.env` | Check that both files are in **Home** > **Code** > **notes-app** in Olares Files, then run the commands from `/home/coder/notes-app` in Code Server |
| The terminal reports `node: command not found` | Run `. "$HOME/.nvm/nvm.sh"`, then `nvm use 22` in the same terminal. Check `node -v` before running the application again. If nvm or Node.js 22 is not installed, follow the installation steps in [Run the notes application in Code Server](#run-the-notes-application-in-code-server). |
| `node -v` reports a missing `libatomic.so.1` | In the Code Server terminal, run `sudo apt-get install -y libatomic1`, then check `node -v` again |
| The notes application returns an empty list | An empty list is expected if the signed-in user has not added any notes. If notes are missing, run `node --env-file=.env notes.mjs whoami` to check the user, then check that the notes' `owner_id` matches that user's ID and that the RLS policy matches [Create the notes table and access policy with Lares](#create-the-notes-table-and-access-policy-with-lares). |
| Metabase cannot connect | Check the database host and the exported `postgres-session` port from [Get the database host and port](#get-the-database-host-and-port). Confirm that Metabase can reach that host and port. Use `postgres.supabase` as the username and the password from [Get the database password](#get-the-database-password). |

## Learn more

- [Manage Olares with Lares](lares.md): Select a model and start a conversation.
- [Enable the self-hosted Supabase MCP server](https://supabase.com/docs/guides/self-hosting/enable-mcp): Understand the upstream server's network access requirements.
