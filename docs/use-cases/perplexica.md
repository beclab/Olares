---
description: Explore Vane for advanced AI-driven search and Q&A with Olares. Learn how to set up Vane with ease.
---
# Set up a privacy-focused AI search engine with Vane

**Vane** (formerly **Perplexica**) is an open-source AI-powered search engine that provides intelligent search capabilities while maintaining user privacy. The project was renamed to **Vane**; behavior and Olares integration are unchanged. As an alternative to Perplexity AI, Vane combines advanced machine learning with comprehensive web search to deliver accurate, source-cited answers to your queries.

This guide will use Ollama as the model provider and SearXNG as the search provider.

## Prerequisites
Before you begin, make sure:
- Ollama installed and running in your Olares environment.
- At least one model installed using Ollama.

## Install SearXNG
SearXNG serves as the privacy-focused meta-search engine backend for Vane. It:
* Aggregates results from multiple search engines
* Removes tracking and preserves your privacy
* Provides clean, unbiased search results for the AI model to process

This integration enables Vane to function as a complete search solution while maintaining the security of your sensitive information.

1. In **Market**, search for "SearXNG".
   ![SearXNG](/images/manual/use-cases/perplexica-searxng.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Install Vane
With the search backend now running, you can install the main Vane app.
1. Open **Market**, and search for "Vane".
   ![Vane](/images/manual/use-cases/perplexica.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure Vane
1. Launch Vane, and click <i class="material-symbols-outlined">settings</i> in the bottom left corner to open the **Settings** page.
2. Under **Model Settings**, set **Ollama** as the **Chat Model Provider** and the **Embedding Model Provider**.
3. Select your preferred model for the **Chat Model** and **Embedding Model**.
   ![Vane configurations](/images/manual/use-cases/perplexica-configurations1.png#bordered)

4. Adjust any other settings as needed.

The changes will be automatically applied.

## Start asking questions
Once the setup is complete, go back to the main page. Try searching for a topic you're interested in to test your new, private search environment.
![Vane example](/images/manual/use-cases/perplexica-example-question1.png#bordered)

