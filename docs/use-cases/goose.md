---
outline: [2, 3]
description: Run Goose on Olares with AI Router, connect Goose Desktop, and turn project notes into a Markdown brief. Use Goose Terminal for conversations and file tasks in your browser.
head:
  - - meta
    - name: keywords
      content: "Olares, Goose, AI agent, AI Router, Goose Desktop, terminal, file automation"
app_version: "0.0.13"
doc_version: "1.0"
doc_updated: "2026-10-13"
---

# Turn project notes into a brief with Goose on Olares

Goose is an open-source AI agent that can read files, edit documents, and run commands through natural language. On Olares, the agent runs on your Olares device while you interact with it through Goose Desktop on your computer or Goose Terminal in your browser.

The app uses AI Router's default chat model out of the box. This example creates a project brief in Goose's persistent workspace, so you can complete a file task without first granting access to your personal directories.

## Learning objectives

In this guide, you will learn how to:

- Connect Goose Desktop to the Goose service on Olares.
- Create and verify a Markdown document through a conversation.
- Use Goose Terminal and enable access to existing files when needed.

## Prerequisites

- Olares 1.12.7 or later.
- [AI Router](olares-router.md) with a working default chat model that supports tool use.

## Install Goose

1. Open Market and search for "Goose".
   <!-- ![Goose in Market](/images/manual/use-cases/goose.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

Goose is already configured to use AI Router. You do not need to enter a model provider's API key for this default connection. Model availability and performance depend on the model configured in AI Router.

## Configure Goose Desktop

1. Install Goose Desktop on your computer using the [official installation guide](https://goose-docs.ai/docs/getting-started/installation/). Use version 1.54.0 to match the server covered by this guide.

2. Enable LarePass VPN on your computer to connect to Olares.

   :::tip On the same local network?
   You can also use Olares local access. See [Access Olares services locally](../manual/best-practices/local-access.md) for domain and Windows setup details. The steps below use the HTTPS entrance URL.
   :::

3. In Olares, open **Settings** > **Applications** > **Goose**. Under **Entrances**, select **Goose ACP** and copy its HTTPS URL. This entrance connects a client to the agent; it is not a browser chat page.

4. Open Goose Terminal from Launchpad and run:

   ```bash
   printenv GOOSE_SERVER__SECRET_KEY
   ```

   Copy the returned value. This is the connection secret for your Goose service, not an AI model API key. Keep it private.

5. In Goose Desktop, open **Settings** > **goose Server** and configure:

   | Setting | Value |
   |:---|:---|
   | **Use external server** | Enabled |
   | **URL** | The Goose ACP HTTPS URL copied from Olares Settings. Do not append `/acp` or `/v1`. |
   | **Secret Key** | The value returned by `printenv GOOSE_SERVER__SECRET_KEY` |
   | **Remote Working Directory** | `/config/workspace` |
   | **Certificate Fingerprint** | Leave empty for the initial connection |

6. Apply the settings and open a new chat. Restart Goose Desktop if an existing window still uses the previous connection or working directory.

7. Send the following message and confirm that Goose replies:

   ```text
   Reply with exactly: Goose is ready.
   ```

With the external server enabled, model requests and agent commands run from Olares. A file path on your computer is not automatically available to the remote agent. The connection fields follow Goose's [remote server configuration](https://goose-docs.ai/docs/guides/remote-goose-server/).

## Create a project brief

### Generate the document

1. In the connected chat, send:

   ```text
   Work only in /config/workspace/goose-demo. Create that directory if needed.
   Create notes.md with these facts:
   - Project: Customer onboarding refresh
   - Goal: Reduce the time needed to complete the first setup
   - Deliverables: A setup checklist and a troubleshooting guide
   - Target date: 2026-11-30
   - Open question: Who will own the troubleshooting guide?
   Then read notes.md and create brief.md with these sections:
   Overview, Deliverables, Target date, and Open questions.
   Preserve every fact. Do not invent an owner or a schedule.
   Finally, report the absolute paths of both files and summarize what you created.
   ```

2. Review any tool approval request before allowing Goose to write files. Wait for the task to finish.

3. Confirm that the response reports `/config/workspace/goose-demo/notes.md` and `/config/workspace/goose-demo/brief.md`.

The brief should contain the project goal, both deliverables, the target date, and the unresolved ownership question. Wording can vary with the model; the facts should remain unchanged.

### Verify the result

1. Open Goose Terminal from Launchpad and inspect the saved files:

   ```bash
   ls -l /config/workspace/goose-demo
   cat /config/workspace/goose-demo/brief.md
   ```

2. Compare the brief with the original notes. Confirm that Goose wrote an actual file rather than only displaying a proposed document in the chat.

3. Return to the same chat and ask:

   ```text
   Read /config/workspace/goose-demo/brief.md. Add a Next actions section
   proposing how to resolve the ownership question. Label your suggestions
   as proposals and preserve the existing project facts.
   ```

4. Read the file again in Goose Terminal and review the changes.

Files under `/config/workspace` persist across normal app restarts. Keep a separate backup for documents you need to retain.

## Use Goose Terminal

### Start a conversation

1. Open Goose Terminal from Launchpad.

2. Start an interactive session in the workspace:

   ```bash
   cd /config/workspace
   goose session
   ```

3. Enter your message at the prompt. You can use the same project brief task above. Follow the terminal prompts to approve tools when requested.

4. Press `Ctrl+C` to exit the session. This ends the terminal interaction; the Goose service remains available for Desktop connections.

To run a single request without an interactive session:

```bash
cd /config/workspace
goose run -t "Read goose-demo/brief.md and summarize its open questions. Do not change any files."
```

### Find and resume a session

List saved sessions:

```bash
goose session list
```

Resume a session using the ID shown in the list:

```bash
goose session --resume --session-id SESSION_ID
```

Replace `SESSION_ID` with an actual session ID. Avoid editing the same files from two active sessions at once.

## Configure access and models

### Work with Home or External files

Home and External directory access are optional. To use existing files:

1. Open **Settings** > **Applications** > **Goose** and edit the app's environment variables.

2. Enable the directory you need:

   | Variable | Default | Directory available to Goose when enabled |
   |:---|:---|:---|
   | `ALLOW_HOME_DIR_ACCESS` | `false` | Home at `/config/Home` |
   | `ALLOW_EXTERNAL_DIR_ACCESS` | `false` | External at `/config/External` |

3. Save the changes, wait for Goose to return to **Running**, and reconnect your client or reopen Goose Terminal.

4. Give Goose the exact path of the files to process. Start with copies of your documents and review the output before applying it to originals.

Enabling access exposes the selected directory to the agent, rather than only a single file mentioned in a prompt. You can disable access again through the same environment variable settings.

### Use another model provider

The default model connection is AI Router. To use a third-party provider, edit the app's environment variables in Olares Settings. These settings control the server used by both the connected Desktop client and Goose Terminal.

For an OpenAI-compatible provider, configure:

| Variable | Value |
|:---|:---|
| `GOOSE_PROVIDER` | `openai` |
| `GOOSE_MODEL` | The exact model ID supplied by your provider |
| `OPENAI_HOST` | The provider's server root URL, such as `https://api.openai.com`, without `/v1` |
| `OPENAI_BASE_PATH` | `v1/chat/completions`, or the chat completions path specified by your provider |
| `OPENAI_API_KEY` | Your provider's API key |

Save the changes, wait for Goose to return to **Running**, and start a new chat. For Anthropic, set `GOOSE_PROVIDER` to `anthropic`, enter its model ID in `GOOSE_MODEL`, and set `ANTHROPIC_API_KEY`.

To return to AI Router, set `GOOSE_PROVIDER` to `custom_olares_router` and `GOOSE_MODEL` to `default-chat`. Remove third-party credentials you no longer need. Changing provider settings in a client does not replace environment variables configured on the Olares server.

## FAQs

### Why does the client fail to connect?

Confirm that Goose is running, LarePass VPN is enabled, and the client uses the Goose ACP URL copied from Olares Settings. Use the connection secret from Goose Terminal. A secret rejection indicates that the client value does not match the server value.

### Why does a model request return 503?

Check that AI Router's default chat model is ready. If the error says the model is already serving its configured number of requests, wait for the current request to finish before retrying. Avoid sending another request from Desktop or Terminal at the same time.

### Why are conversations named New Chat?

Automatic AI session titles are disabled by default. You can rename a session manually. To enable automatic titles, set `GOOSE_DISABLE_SESSION_NAMING` to `false` in the app's environment variables. Title generation sends additional model requests, which can compete with chat requests on a model with limited concurrency.

### Why can Goose not find my files?

Use a path on Olares, such as `/config/workspace/goose-demo/brief.md`. For Home or External files, enable the corresponding directory access first and use `/config/Home` or `/config/External`. A path on the client computer is not a server path.

## Learn more

- [Olares Router](olares-router.md): Set up and manage models used by Goose.
- [Goose installation guide](https://goose-docs.ai/docs/getting-started/installation/): Install the Desktop client for your operating system.
