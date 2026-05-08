---
outline: [2, 3]
description: Connect NemoClaw to Google Workspace so your agent can read emails, manage calendars, and access files through the gog skill.
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, Google Workspace, Gmail, Google Calendar, Google Drive, gog, skills, OAuth, AI agent
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-05-08"
---

# Integrate with Google Workspace

The gog skill lets your NemoClaw agent interact with Google Workspace services such as Gmail, Google Calendar, and Google Drive. Once configured, you can ask your agent to search emails, schedule meetings, or access files in natural language.

## Prerequisites

- NemoClaw installed and running on Olares.
- A Google Workspace or personal Google account.
- Admin access to the Google Cloud Console to create an OAuth application.

## Step 1: Create a Google Cloud OAuth application

1. Go to the [Google Cloud Console](https://cloud.google.com/) and sign in with your Google account.
2. Navigate to **Console**, and create a new project or select an existing one.
3. Navigate to **APIs and services** > **Library**.
4. From the filter on the left, select **Google Workspace** and enable the Google Workspace APIs you need. For example:

   - Gmail API
   - Google Drive API
   - Google Calendar API
   - Google People API (for contacts)
   - Google Sheets API and Google Docs API

5. If this is your first time configuring OAuth, navigate to **APIs and services** > **OAuth consent screen**, and follow the on-screen instructions to set up the OAuth consent screen first.

   :::info Personal Google accounts
   If you use a personal Gmail account (not a Google Workspace organization), select **External** for **Audience**. After publishing the app, go to **Audience** > **Test users** and add your own email address. Without this step, authentication will fail for unauthorized users.
   :::

6. Go to **APIs and services** > **Credentials**, click **Create credentials**, and select **OAuth client ID**.
7. For **Application type**, select **Web application**.
8. Configure the authorized origins and redirect URIs:

   - **Authorized JavaScript origins**: Your OpenClaw Web UI URL. You can copy it directly from the address bar. For example, `https://d38aad901.laresprime.olares.com`.
   - **Authorized redirect URIs**: Your OpenClaw Web UI URL with `/oauth2/callback` appended. For example, `https://d38aad901.laresprime.olares.com/oauth2/callback`.

   ![Configure OAuth redirect URIs](/images/manual/use-cases/google-cloud-oauth-uris.png#bordered){width=70%}

9. Click **Create**, then click **Download JSON** to save the client secrets file.

## Step 2: Install the gog skill

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

3. Run the skills configuration wizard:

   ```bash
   openclaw config --section skills
   ```

4. Follow the prompts to configure your installation. Use the arrow keys to navigate and press **Enter** to confirm.

    | Settings | Option |
    |:---------|:-------|
    | Where will the Gateway run | Local (this machine) |
    | Configure skills now | Yes |
    | Install missing skill dependencies | Navigate to the skill **gog**, press **Space** to select it, then press **Enter**. |
    | Set [API_KEY] for [skill] | Select **No** for all these settings. |

## Step 3: Authenticate with Google

1. Upload the downloaded JSON file to the directory that NemoClaw can access:

   a. Open Files and navigate to `Data/nemoclaw/openclaw-config/inbox/`.

   b. Upload the JSON file.

   ![Upload client secrets](/images/manual/use-cases/nemoclaw-upload-secrets.png#bordered)

2. In the NemoClaw CLI sandbox, add the credentials file, and use **Tab** to autocomplete the name of the JSON file:

   ```bash
   gog auth credentials inbox/client_secret_....json
   ```

3. Start the authentication flow. Replace the `email`, `services`, and `redirect-host` values with your own:

   ```bash
   gog auth add your-email@example.com --services gmail,calendar,drive,contacts,sheets,docs \
        --listen-addr 0.0.0.0:8080 \
        --redirect-host <your-control-ui-domain> \
        --force-consent
   ```

   :::tip
   Do not include `https://` in the `--redirect-host` value. Use only the domain, such as `d38aad901.laresprime.olares.com`.
   :::

   For example:

   ```bash
   gog auth add laresprime@gmail.com --services calendar \
        --listen-addr 0.0.0.0:8080 \
        --redirect-host d38aad901.laresprime.olares.com \
        --force-consent
   ```

4. Click the URL that appears in the terminal and complete the Google authentication within 3 minutes.

   :::tip
   The final callback URL should look like `https://d38aad901.{username}.olares.com/oauth2/callback?state=...`. If your browser appends `/chat` to the URL due to cache, complete the authentication in an incognito window.
   :::

5. When authentication succeeds, you see a confirmation message in the terminal.

   Example output:
   ```text
   email laresprime@gmail.com
   services calendar
   client default
   ```

## Step 4: Use Google Workspace with your agent

You can now use `gog` commands in the NemoClaw CLI sandbox (`my-assistant`). For example, to create a meeting:

```bash
sandbox@my-assistant:~$ gog calendar create primary \
  --summary "Coffee with Team" \
  --from "2026-05-09T10:00:00+02:00" \
  --to "2026-05-09T11:00:00+02:00" \
  --description "Catching up at the local cafe"
```

Example output:
```text
id      5gsacgvqem82o29cct65oq2234
summary Coffee with Team
timezone        Europe/Amsterdam
event-timezone  Etc/GMT-2
start   2026-05-09T10:00:00+02:00
start-day-of-week       Saturday
start-local     2026-05-09T10:00:00+02:00
end     2026-05-09T11:00:00+02:00
end-day-of-week Saturday
end-local       2026-05-09T11:00:00+02:00
description     Catching up at the local cafe
reminders       (calendar default)
link    https://www.google.com/calendar/event?eid=...
```
Click the link to open the event in Google Calendar and verify it.

![Create calendar event with gog](/images/manual/use-cases/gog-calendar-event.png#bordered){width=70%}

You can also ask your agent in natural language through the OpenClaw Web UI or TUI. For example:

```text
Create a meeting on my Google Calendar at 12 PM today, Pacific time
```

<!-- ![Create calendar event via chat](/images/manual/use-cases/nemoclaw-google-calendar-chat.png#bordered) -->

:::tip
If the agent doesn't call `gog` after authentication, restart NemoClaw so the agent picks up the new skill.
:::

## Learn more

- [Run NemoClaw with a local LLM](nemoclaw.md): Set up NemoClaw with a local model.
- [Manage skills and plugins](openclaw-skills.md): Install and manage other OpenClaw skills.
