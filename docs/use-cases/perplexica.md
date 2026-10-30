---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/perplexica
outline: [2, 3]
description: Self-host Vane (formerly Perplexica) on Olares as a private Perplexity alternative. Get cited answers using models through Olares Router and SearXNG search.
head:
  - - meta
    - name: keywords
      content: Olares, Vane, Perplexica, perplexity alternative, self-hosted perplexity, AI search, SearXNG, vane on olares
app_version: "1.1.35"
doc_version: "1.3"
doc_updated: "2026-10-30"
---
# Self-host a private AI search engine with Vane

<VersionRouteSelect />

Vane (previously Perplexica) is an open-source AI-powered answering engine. It combines web search with local or cloud LLMs to deliver cited, conversational answers while keeping your queries private.

This guide uses Qwen3.8-27B (llama.cpp) through Olares Router for answers and SearXNG for web search. Vane connects directly to SearXNG to retrieve sources.

## Learning objectives

In this guide, you will learn how to:

- Connect Vane to SearXNG using an application environment variable in Olares Settings.
- Select a Router chat model and a built-in embedding model in Vane.
- Ask a question and review the cited sources.

## Prerequisites

Before you begin, you need:

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- Qwen3.8-27B (llama.cpp) installed from Market.
- SearXNG installed from Market and running.

Vane includes embedding models. You do not need to install EmbeddingGemma for this workflow.

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install Vane

1. Open Market and search for "Vane".
   ![Vane](/images/manual/use-cases/vane.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Connect SearXNG

Configure the search connection in Olares Settings before opening Vane for the first time.

1. Open Olares **Settings** and go to **Applications** > **SearXNG** > **Entrances** > **SearXNG**.
2. Copy the **Endpoint** URL. Use this URL as the instance address without adding `/search` or query parameters.
3. Go to **Applications** > **Vane** > **Manage environment variables**.
4. Click the edit icon beside `SEARXNG_API_URL`, paste the endpoint URL, and click **Confirm**.
5. Click **Apply**. Vane restarts automatically. Wait for it to return to the running state before opening it.

:::info Updating an existing search connection
Vane uses `SEARXNG_API_URL` to initialize an empty search address. An address already saved in Vane takes precedence over the environment variable. To change an existing connection, open Vane's bottom-left **Settings** > **Search**, update **SearXNG URL**, and click outside the field to save it.
:::

## Configure models

Configure the chat connection and embedding model in Vane's own settings.

### Get the Router connection

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

Before continuing, confirm that the chat model shows **Callable** on Router's **LLM** page.

### Select models in Vane

1. Open Vane from Launchpad. Use the setup wizard on first launch, or open the bottom-left **Settings** > **Models** for an existing setup.
2. Click **Add Connection**, select **OpenAI**, and enter the following values:

   | Field | Value |
   | :--- | :--- |
   | **Connection Name** | `Olares Router` |
   | **API Key** | `olares` |
   | **Base URL** | The Router URL you copied, including `/v1` |

3. Click **Add Connection**. Under this connection's **CHAT MODELS**, click **Add**.
4. Set **Model name** to `Router chat` and **Model key** to `default-chat`, then click **Add Model**.
5. In the setup wizard, click **Next** to open **Select models**. In **Settings** > **Models**, use the **Select models** section on the same page.
6. Under **Select Chat Model**, choose **Olares Router - Router chat**. Under **Select Embedding Model**, choose the built-in **Transformers - all-MiniLM-L6-v2** model.
7. In the setup wizard, click **Finish**.

To change models or connections later, open Vane's **Settings** > **Models**.

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
