---
description: Explore Perplexica for advanced AI-driven data analysis and visualization with Olares. Learn how to set up Perplexica with ease.
---
# Set up a privacy-focused AI search engine with Perplexica

Perplexica is an open-source AI-powered search engine that provides intelligent search capabilities while maintaining user privacy. As an alternative to Perplexity AI, it combines advanced machine learning with comprehensive web search functionality to deliver accurate, source-cited answers to your queries.

This guide will use Ollama as the model provider and SearXNG as the search provider.

## Prerequisites
Before you begin, make sure:
- Ollama installed and running in your Olares environment.
- You have downloaded at least one model using Ollama.

## Install SearXNG
SearXNG serves as the privacy-focused meta-search engine backend for Perplexica. It:
* Aggregates results from multiple search engines
* Removes tracking and preserves your privacy
* Provides clean, unbiased search results for the AI model to process

This integration enables Perplexica to function as a complete search solution while maintaining the security of your sensitive information.

1. In **Market**, search for "SearXNG".
   ![SearXNG](/images/manual/use-cases/perplexica-searxng.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Install Perplexica
With the search backend now running, you can install the main Perplexica app.
1. Open **Market**, and search for "Perplexica".
   ![Perplexica](/images/manual/use-cases/perplexica.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure Perplexica
1. Launch Perplexica, and click <i class="material-symbols-outlined">settings</i> in the bottom left corner to open the **Settings** page.
2. Under **Model Settings**, set **Ollama** as the **Chat Model Provider** and the **Embedding Model Provider**.
3. Select your preferred model for the **Chat Model** and **Embedding Model**.
   ![Perplexica configurations](/images/manual/use-cases/perplexica-configurations1.png#bordered)

4. Make any other changes as needed.

The changes will be automatically applied.

## Start asking questions
Once the setup is complete, go back to the main page. Try searching for a topic you're interested in to test your new, private search environment.
![Perplexica example](/images/manual/use-cases/perplexica-example-question1.png#bordered)

