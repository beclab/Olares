---
outline: [2, 3]
description: Set up Vane (previously Perplexica) on Olares for private, AI-powered search and Q&A. Use local LLMs with Ollama and SearXNG as the search backend.
head:
  - - meta
    - name: keywords
      content: Olares, Vane, Perplexica, AI search, privacy, Ollama, SearXNG, self-hosted
app_version: "1.12.0"
doc_version: "1.2"
doc_updated: "2026-04-17"
---
# Set up a privacy-focused AI search engine with Vane

Vane (previously Perplexica) is an open-source AI-powered answering engine. It combines web search with local or cloud LLMs to deliver cited, conversational answers while keeping your queries private.

This guide uses Ollama as the model provider and SearXNG as the search backend.

## Prerequisites

Before you begin, make sure:
- [Ollama is installed](ollama.md) and running in your Olares environment.
- At least one chat model is installed in Ollama. An embedding model is optional, since Vane ships with built-in ones.

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

1. Launch Vane. A setup wizard opens on first launch, with Ollama and its installed models detected automatically.
   ![Manage connections](/images/manual/use-cases/vane-manage-connections.png#bordered)

2. Click **Next**.
3. Select a chat model and an embedding model, then click **Finish**.
   ![Configure models](/images/manual/use-cases/vane-configure-models.png#bordered)

   :::tip Embedding model options
   If you don't have an embedding model in Ollama, you can pick one of Vane's built-in embedding models instead.
   :::

You're taken to the main chat page. To change models or connections later, click <i class="material-symbols-outlined">settings</i> in the bottom-left corner to open the **Settings** page.

## Start asking questions

Try a search to test your new private search environment.
![Vane example](/images/manual/use-cases/vane-example-question.png#bordered)

## Learn more

- [Ollama](ollama.md): Run local LLMs on Olares as Vane's model backend.
- [Vane on GitHub](https://github.com/ItzCrazyKns/Vane): Upstream project README, architecture notes, and community Discord.
