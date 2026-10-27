---
outline: [2, 3]
description: Common issues and solutions for Open WebUI on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, common issues, troubleshooting, model download
app_version: "1.0.38"
doc_version: "1.1"
doc_updated: "2026-07-31"
---

# Open WebUI common issues

Use this page to identify and resolve common issues with Open WebUI on Olares.

## Model app is stuck at "Waiting for Ollama" or "Needs attention"

If the model app stays in these states for more than a few minutes:

1. Go to **Settings** > **Accelerator**.
2. Check your GPU mode:
   - If you are using **Memory slicing**, make sure the model app is linked to the GPU and has enough VRAM allocated.
   - If you are using **Exclusive**, make sure the model app has full GPU access.
3. Restart the model app from Launchpad and check the status again.

## Microphone "Permission denied" error

When attempting to use the dictate button or Voice Mode, you might receive the following error messages:
- `Permission denied when accessing microphone: NotAllowedError: Permission denied`
- `Permission denied when accessing media devices`

The Olares desktop displays applications inside embedded frames (iframes). For strict security and privacy reasons, modern web browsers prevent embedded frames from accessing sensitive hardware like your microphone, even if you already granted the browser permission in your system settings.

To bypass this security restriction and use your microphone:
1. In the top-right corner of the Open WebUI window on the Olares desktop, select <i class="material-symbols-outlined">open_in_new</i> to open it in a new browser tab.
2. In the new browser tab, select the microphone icon in the chat interface.
3. When the browser prompts you, allow microphone access.

## Open WebUI does not search the web

If the AI response does not include web search results, work through the following checks:

1. **Search is enabled for the conversation.**  
   In the chat area, click the **Integrations** icon, and make sure **Web Search** is enabled.

2. **The embedding model is configured and reachable.**  
   Go to **Admin Panel** > **Settings** > **Tools** > **Documents**,  and verify the embedding model settings. If the embedding model is missing or unreachable, Open WebUI cannot process web search results.

3. **SearXNG returns results for your query.**  
   Open SearXNG from the Launchpad and run the same search. If SearXNG returns no results, check that its search engines are enabled and working in **Preferences** > **ENGINES**.

4. **The web loader is being blocked.**  
   If you need full-page content and the default web loader fails because of anti-scraping measures, go to **Admin Panel** > **Settings** > **Tools** > **Web Search**, and then enable **Bypass Web Loader**. This uses search result summaries instead of fetching full pages.

   :::tip
   For reliable full-page retrieval, install and configure Firecrawl as the web loader. See [Use Firecrawl as a web page loader](firecrawl.md#configure-open-webui).
   :::
