---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/perplexica
outline: [2, 3]
description: Self-host Vane (formerly Perplexica) on Olares as a private Perplexity alternative. Get cited answers using models through Olares Router and SearXNG search.
head:
  - - meta
    - name: keywords
      content: Olares, Vane, Perplexica, perplexity alternative, self-hosted perplexity, AI search, SearXNG, vane on olares
app_version: "1.12.0"
doc_version: "1.2"
doc_updated: "2026-09-23"
---
# Self-host a private AI search engine with Vane

<VersionRouteSelect />

Vane (previously Perplexica) is an open-source AI-powered answering engine. It combines web search with local or cloud LLMs to deliver cited, conversational answers while keeping your queries private.

This guide uses Qwen3.8-27B (llama.cpp) through Olares Router and SearXNG as the search backend.

## Prerequisites

Before you begin, you need:

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- Install Qwen3.8-27B (llama.cpp) from Market. Vane can use its built-in embedding models.

## Install SearXNG

SearXNG is a privacy-focused meta-search engine that aggregates results from multiple search engines without tracking users. Vane uses it to fetch clean, unbiased results for the AI model to process.

1. Open Market and search for "SearXNG".
   ![SearXNG](/images/manual/use-cases/perplexica-searxng.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Install Vane

1. Open Market and search for "Vane".
   ![Vane](/images/manual/use-cases/vane.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure Vane

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

1. Open Vane. In the setup wizard or the model section of **Settings**, add an **OpenAI** provider.
2. Set **Base URL** to the Router URL, including `/v1`, and enter `olares` for **API Key**.

   <!--
   TODO: 素材清单 17，待补 Router 截图：vane-provider-router.png；替换下方旧图后再取消注释。
   ![Manage connections](/images/manual/use-cases/vane-manage-connections.png#bordered)
   -->

3. Add a chat model manually under this provider. Use a display name such as `Router chat` and set the model ID to `default-chat`.
4. Select this chat model and one of Vane's built-in embedding models, then finish setup.

   <!--
   TODO: 素材清单 18，待补 Router 截图：vane-models-router.png；替换下方旧图后再取消注释。
   ![Configure models](/images/manual/use-cases/vane-configure-models.png#bordered)
   -->


To change models or connections later, open **Settings** using the icon in the bottom-left corner.

## Start asking questions

Try a search to test your new private search environment.
![Vane example](/images/manual/use-cases/vane-example-question.png#bordered)

## Learn more

- [Olares Router](olares-router.md): Manage the models used by Vane.
- [Vane on GitHub](https://github.com/ItzCrazyKns/Vane): Upstream project README, architecture notes, and community Discord.
