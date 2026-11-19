---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/perplexica
outline: [2, 3]
description: Run Vane on Olares for private AI search. Connect chat models through Router and use built-in embeddings with SearXNG to get answers with source citations.
head:
  - - meta
    - name: keywords
      content: Olares, Vane, Perplexica, perplexity alternative, self-hosted perplexity, AI search, SearXNG, vane on olares
app_version: "1.1.35"
doc_version: "1.4"
doc_updated: "2026-11-18"
---
# Self-host a private AI search engine with Vane

<VersionRouteSelect />

Vane (previously Perplexica) is an open-source AI-powered answering engine. It combines web search with local or cloud LLMs to deliver cited, conversational answers while keeping your queries private.

This guide uses Qwen3.8-27B (llama.cpp) through Olares Router for answers and SearXNG for web search. Vane connects directly to SearXNG to retrieve sources.

## Learning objectives

In this guide, you will learn how to:

- Connect Vane to SearXNG in Vane's settings or through an application environment variable in Olares Settings.
- Select a Router chat model and a built-in embedding model in Vane.
- Ask a question and review the cited sources.

## Prerequisites

Before you begin, you need:

- [SearXNG](searxng.md) installed from Market and running.
- The following models:

  | Model type | Model | How to get it |
  | :--- | :--- | :--- |
  | Chat | Qwen3.8-27B (llama.cpp) | Install from Market |
  | Embedding | all-MiniLM-L6-v2 | Built into Vane |

## Install Vane

1. Open Market and search for "Vane".
   ![Vane](/images/manual/use-cases/vane.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure models

Vane uses Router for chat and its built-in all-MiniLM-L6-v2 model for embeddings.

1. Open Router from Launchpad. On **Default models**, set Qwen3.8-27B (llama.cpp) as the default chat model. Wait until the `default-chat` row on the same page shows **Callable**.
2. <!--@include: ../reusables/ai-service-connections.md#router-openai-base-url-->

3. Open Vane from Launchpad. Use the setup wizard on first launch, or open the bottom-left **Settings** > **Models** for an existing setup.
4. Click **Add Connection**, select **OpenAI**, and enter the following values:

   | Field | Value |
   | :--- | :--- |
   | **Connection Name** | `Olares Router` |
   | **API Key** | `olares` |
   | **Base URL** | The Router address from step 2, including `/v1` |

5. Click **Add Connection**. Under this connection's **CHAT MODELS**, click **Add**.
6. Set **Model name** to `Router chat` and **Model key** to `default-chat`, then click **Add Model**.
7. In the setup wizard, click **Next** to open **Select models**. In **Settings** > **Models**, use the **Select models** section on the same page.
8. Under **Select Chat Model**, choose **Olares Router - Router chat**. Under **Select Embedding Model**, choose the built-in **Transformers - all-MiniLM-L6-v2** model.
9. In the setup wizard, click **Finish**.

If the newly added model does not appear in **Select Chat Model**, refresh the page and reopen **Settings** > **Models** to select it.

To use another chat model, select it on Router's **Default models** page. Vane follows this selection through `default-chat`, using the same Base URL. To change the connection or built-in embedding model, open Vane's **Settings** > **Models**.

## Connect SearXNG

Get the SearXNG address, then choose one of the two configuration methods below. Use Vane's settings for an existing setup. The environment variable can initialize the address before first launch.

1. Open Olares **Settings** and go to **Applications** > **SearXNG** > **Entrances** > **SearXNG**.
2. Copy the **Endpoint** URL. Use this URL as the instance address without adding `/search` or query parameters.

### Option 1: Configure in Vane

1. Open Vane and go to the bottom-left **Settings** > **Search**.
2. Paste the endpoint URL into **SearXNG URL**.
3. Click outside the field to save it.

### Option 2: Configure through an environment variable

1. Open Olares **Settings** and go to **Applications** > **Vane** > **Manage environment variables**.
2. Click the edit icon beside `SEARXNG_API_URL`, paste the endpoint URL, and click **Confirm**.
3. Click **Apply**. Vane restarts automatically. Wait for it to return to the running state before opening it.
4. Open Vane's **Settings** > **Search** and check that **SearXNG URL** shows the address you entered.

:::info Existing search settings take precedence
Vane uses `SEARXNG_API_URL` only when its saved search address is empty. If an address is already saved, use Option 1 to change it. You do not need to configure both methods.
:::

For search through Router, see [Research the web with Lares](lares-web-research.md). Vane's **SearXNG URL** uses the SearXNG entrance URL.

## Start asking questions

1. On Vane's home page, open the source selector beside the message box and enable **Web**.
2. Enter a question, for example:

   ```text
   What is Olares? Find its official website and documentation, and explain
   what it does with links to your sources.
   ```

3. Submit the question and wait for the answer. Vane lists the retrieved pages under **Sources** and adds citation links to the answer.
4. Open a citation and compare the answer with the source page. Use the follow-up box to ask about a detail you want to explore.

   ![Vane answer with sources and citation links](/images/manual/use-cases/vane-example-question.png#bordered)

## Learn more

- [Olares Router](olares-router.md): Manage the models used by Vane.
- [Manage application environment variables](../manual/olares/settings/manage-app-env.md): Edit and apply app configuration in Olares Settings.
- [Vane on GitHub](https://github.com/ItzCrazyKns/Vane): Upstream project README, architecture notes, and community Discord.
