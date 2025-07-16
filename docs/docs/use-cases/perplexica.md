---
description: Explore Perplexica for advanced AI-driven data analysis and visualization with Olares. Learn how to set up Perplexica with ease.
---
# Perplexica

Perplexica is an open-source AI-powered search engine that provides intelligent search capabilities while maintaining user privacy. As an alternative to Perplexity AI, it combines advanced machine learning with comprehensive web search functionality to deliver accurate, source-cited answers to your queries.

## The backend: SearSNG
SearXNG serves as the privacy-focused meta-search engine backend for Perplexica. It:
* Aggregates results from multiple search engines
* Removes tracking and preserves your privacy
* Provides clean, unbiased search results for the AI model to process

This integration enables Perplexica to function as a complete search solution while maintaining the security of your sensitive information.

## Before you begin
Before getting started, ensure you have:
- Ollama installed and running in your Olares environment
- [Enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.
- Open WebUI installed with your preferred language models downloaded
  :::tip
  For optimal performance, consider using lightweight yet powerful models like `gemma2`, which offer a good balance between speed and capability.
  :::
## Set up Perplexica

1. The admin installs SearXNG from Market.
    
   :::info
   Starting from Olares 1.11.6, if "SearXNG For Cluster" or the "SearXNG" client was previously installed, uninstall them before proceeding.
   :::
    
2. Install Perplexica from Market.
3. Launch Perplexica, and click <i class="material-symbols-outlined">settings</i> in the bottom left corner to open the settings window.
4. Configure your search environment with the following settings (using `gemma2` as an example):
   - **Chat model Provider**: `Ollama`
   - **Chat Model**: `gemma2:latest`
   - **Embedding model Provider**: `Ollama`
   - **Embedding Model**: `gemma2:latest`

   ![Perplexica configurations](/images/manual/use-cases/perplexica-configurations.png#bordered){width=50%}
5. Click the confirmation button to save your configuration and return to the search interface.

Your setup is complete. Try searching for a topic you're interested in to test your new search environment.
![Perplexica example](/images/manual/use-cases/perplexica-example-question.png#bordered)

