---
outline: [2, 3]
description: Run Goose as your local AI agent on Olares. Chat through Goose Terminal in your browser or connect Goose Desktop on your computer.
head:
  - - meta
    - name: keywords
      content: "Olares, Goose, AI agent, AI Router, Goose Desktop, terminal, file automation"
app_version: "0.0.13"
doc_version: "1.0"
doc_updated: "2026-10-21"
---

# Run Goose as your local AI agent on Olares

Goose is an open-source, general-purpose AI agent for research, writing, coding, automation, and more. On Olares, Goose runs on your Olares device with local models through Router, and you interact with it through Goose Terminal in your browser or Goose Desktop on your computer.

The app uses Router's default chat model by default. This example creates a project brief in Goose's persistent workspace, so you can complete a file task without first granting access to your personal directories.

## Learning objectives

In this guide, you will learn how to:

- Run Goose in Goose Terminal.
- Connect Goose Desktop to the Goose service on Olares.
- Enable access to existing files when needed.

## Prerequisites

Before you begin, you need:

- Olares 1.12.7 or later.
- Access to [Router](olares-router.md) for local model connections.
- The following model installed from Market:

  | Model type | Model | How to get it |
  | :--- | :--- | :--- |
  | Chat | Qwen3.8-27B (llama.cpp) | Install from Market |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install Goose

1. Open Market and search for "Goose".
2. Click **Get**, then **Install**, and wait for installation to complete.

## Use Goose Terminal

Goose Terminal opens in your browser and connects to the Goose service directly. No client installation or connection setup is needed.

### Run a single request

Use `goose run -t` to send a single prompt without starting an interactive conversation.

1. Open Goose Terminal from Launchpad.
2. Run `goose run -t` followed by your prompt, replacing `Your_Prompt` with your question or task:

   ```bash
   goose run -t "Your_Prompt"
   ```

   For example, ask a question:

   ```bash
   goose run -t "What is the difference between Goose Terminal and Goose Desktop? Answer in three sentences."
   ```

3. Goose displays the answer and returns to the terminal prompt.

   To ask follow-up questions in an ongoing conversation, see [Start a conversation](#start-a-conversation).

### Start a conversation

Use an interactive session when you want a back-and-forth conversation.

1. Open Goose Terminal from Launchpad.
2. Start an interactive session in the workspace:

   ```bash
   cd /config/workspace
   goose session
   ```

3. The first time you run this command, Goose asks whether to share anonymous usage data. Choose **Yes** or **No** to continue.
4. Send the following prompt to create a sample project brief:

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

5. Follow the terminal prompts to approve tools when requested.
6. Verify the result in the same terminal:

   ```bash
   ls -l /config/workspace/goose-demo
   cat /config/workspace/goose-demo/brief.md
   ```

   The brief should contain the project goal, both deliverables, the target date, and the unresolved ownership question. Wording can vary with the model. The facts should remain unchanged.

7. Press `Ctrl+C` to exit the session. If Goose asks you to press it again, press `Ctrl+C` a second time. This ends the terminal interaction.

:::info File location
Goose stores the files in the `/config/workspace` directory on your Olares device. Files there persist across normal app restarts. Keep a separate backup for documents you need to retain.
:::

### Find and resume a session

Sessions are saved on the Goose service. You can list them and resume any session later, in Goose Terminal or Goose Desktop.

- To list saved sessions, run `goose session list`.
- To resume a session, run `goose session --resume --session-id SESSION_ID` with the ID shown in the list.

   :::tip
   Two active sessions can overwrite each other's changes when they edit the same file. Exit one session before editing the same file in another.
   :::

## Use Goose Desktop

Goose Desktop runs on your computer and connects to the Goose service on Olares through the Goose ACP entrance.

### Prepare connection details

1. Open Goose Terminal from Launchpad and run:

   ```bash
   printenv GOOSE_SERVER__SECRET_KEY
   ```

2. Copy the returned value and save it for the Goose Desktop setup later.

   This is the connection secret between Goose Desktop and the Goose service, not an AI model API key. Keep it private.

3. Open Olares **Settings**, go to **Applications** > **Goose** > **Entrances** > **Goose ACP**, and then copy the **Endpoint** URL. For example, `https://615b32570.primelares.olares.com`.
4. In Router, go to the **LLM** page, find the model assigned as `default-chat`, and click **View connection example** on the right.
5. Select how your computer connects:
   - If your computer is on the same local network as your Olares device, select **Device in LAN** and copy the Router URL.
   - If your computer connects over the internet, select **Remote** and copy the Router URL.

### Connect to the Goose service

:::tip
If your computer is not on the same local network as your Olares device, turn on LarePass VPN before connecting.
:::

1. Install Goose Desktop on your computer using the [official installation guide](https://goose-docs.ai/docs/getting-started/installation/).
2. Open Goose Desktop and select **Connect to a Provider** > **Add a custom provider** > **Configure manually**.
3. In the **Add Custom Provider** window, configure the following settings:

   | Field | Value |
   |:---|:---|
   | <nobr>**Provider Type**</nobr> | Select **OpenAI Compatible**. |
   | <nobr>**Display Name**</nobr> | Enter a name for easy identification, such as `olares-router`. |
   | <nobr>**API URL**</nobr> | Enter the Router URL you copied earlier. |
   | <nobr>**Available Models**</nobr> | Enter `default-chat`. |

4. Click **Create Provider**.
5. Select the privacy option as needed, and click **Get Started**.
6. In Goose Desktop, select **Settings** from the left sidebar, and go to **Agent**.
7. Locate **External Backend (ACP)**, toggle on **Use external backend**, and configure the following settings:

   | Setting | Value |
   |:---|:---|
   | <nobr>**Backend Base URL**</nobr> | The **Endpoint** URL of the Goose ACP entrance you copied in **Prepare connection details**. Do not append `/acp` or `/v1`. |
   | <nobr>**Remote Working Directory**</nobr> (optional) | `/config/workspace`<br><br>**Note**: Although Goose Desktop marks this field as optional, you must set it here. The agent needs a server-side directory to work in. |
   | <nobr>**Secret Key**</nobr> | The connection secret you saved in **Prepare connection details**. |

8. Restart Goose Desktop.
9. Open a new chat. Under the chat box, confirm the model is `default-chat`, then select a working directory.
10. Send the following message and confirm that Goose replies:

      ```text
      Reply with exactly: Goose is ready.
      ```

### Continue a Terminal session

The session you created in Goose Terminal also appears in Goose Desktop as a **CLI Session** entry under **CHATS** in the left sidebar. Open it and you can continue the conversation where you left off, with the same workspace and files.

## Configure access and models

### Work with Home or External files

Home and External directory access are optional. To use existing files:

1. Open Olares **Settings**, and go to **Applications** > **Goose** > **Manage environment variables**.
2. Set the access variable to `true` for the directory you need:

   | Variable | Default | Directory available to Goose when enabled |
   |:---|:---|:---|
   | <nobr>`ALLOW_HOME_DIR_ACCESS`</nobr> | `false` | Home at `/config/Home` |
   | <nobr>`ALLOW_EXTERNAL_DIR_ACCESS`</nobr> | `false` | External at `/config/External` |

3. Click **Apply**, and wait for Goose to return to **Running**.
4. Reconnect Goose Desktop or reopen Goose Terminal.
5. Give Goose the exact path of the files to process.

   :::warning
   The agent can modify files in the enabled directory. Start with copies of your documents and review the output before applying it to originals.
   :::

   :::info Scope of access
   Enabling a variable exposes the whole directory to the agent, not only a single file mentioned in a prompt. To revoke access, set the variable back to `false` and apply the changes.
   :::

### Use another model provider

The default model connection is Router. To use a third-party provider, edit the app's environment variables in Olares Settings. These settings control the server used by both Goose Desktop and Goose Terminal.

For an OpenAI-compatible provider, configure:

| Variable | Value |
|:---|:---|
| <nobr>`GOOSE_PROVIDER`</nobr> | `openai` |
| <nobr>`GOOSE_MODEL`</nobr> | The exact model ID supplied by your provider |
| <nobr>`OPENAI_HOST`</nobr> | The provider's server root URL, such as `https://api.openai.com`, without `/v1` |
| <nobr>`OPENAI_BASE_PATH`</nobr> | `v1/chat/completions`, or the chat completions path specified by your provider |
| <nobr>`OPENAI_API_KEY`</nobr> | Your provider's API key |

Save the changes, wait for Goose to return to **Running**, and start a new chat. For Anthropic, set `GOOSE_PROVIDER` to `anthropic`, enter its model ID in `GOOSE_MODEL`, and set `ANTHROPIC_API_KEY`.

To return to Router, set `GOOSE_PROVIDER` to `custom_olares_router` and `GOOSE_MODEL` to `default-chat`. Remove third-party credentials you no longer need. Changing provider settings in Goose Desktop does not replace environment variables configured on the Olares server.

## FAQs

### Why does Goose Desktop fail to connect?

Confirm that Goose is running, and Goose Desktop uses the Goose ACP URL copied from Olares Settings. Use the connection secret from Goose Terminal. A secret rejection indicates that the value in Goose Desktop does not match the server value.

### Why does a model request return 503?

Check that Router's default chat model is ready. If the error says the model is already serving its configured number of requests, wait for the current request to finish before retrying. Avoid sending another request from Goose Desktop or Goose Terminal at the same time.

### The model returned an empty response. Please resend your message to continue.

External Backend (ACP) is not enabled or configured. In Goose Desktop, open **Settings** > **Agent** and check the External Backend (ACP) settings.

### Why can Goose not find my files?

Use a path on Olares, such as `/config/workspace/goose-demo/brief.md`. For Home or External files, enable the corresponding directory access first and use `/config/Home` or `/config/External`. A path on your computer is not a server path.

## Learn more

- [Olares Router](olares-router.md): Set up and manage models used by Goose.
- [Goose installation guide](https://goose-docs.ai/docs/getting-started/installation/): Install Goose Desktop for your operating system.
