---
outline: [2, 3]
description: Learn how to enable web search in OpenClaw using Brave Search to give your AI agent access to real-time internet information.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, OpenClaw web search
---

# Optional: Enable web search

By default, OpenClaw answers questions only based on its training data, which means it doesn't know about current events or real-time news. To give your agent access to the live internet, you can enable the web search tool.

OpenClaw officially recommends Brave Search. It uses an independent web index optimized for AI retrieval, ensuring your agent finds accurate information.

## Prerequisites

A Brave Search API key is required to complete this setup. You can obtain a free API key from the [Brave Search API](https://brave.com/search/api/). The free tier of the "Data for Search" plan is usually sufficient for personal use.

## Enable Brave Search

1. Open the OpenClaw CLI.
2. Run the following command to start the web configuration wizard:

    ```bash
    openclaw configure --section web
    ```
3. Configure the basic settings as follows:

    | Settings | Option |
    |:-------|:-----|
    | Where will the Gateway run | Local (this machine) |
    | Enable web_search (Brave Search) | Yes |
    | Brave Search API key | Your `BraveSearchAPIkey` |
    | Enable web_fetch (keyless HTTP fetch) | Yes |

4. Finalize the configuration in Control UI.

    :::tip
    Whihe the CLI wizard sets up the API key, the Control UI allows you to customize specific parameters such as timeouts and limits.
    :::

    a. Return to the **Control UI** > **Config** > **Raw** tab. 

    b. Find the `tools` section and update as follows. Replace `{Your-Brave-Search-API-Key}` with your actual key.

    ```json
    "tools": {
        "web": {
        "search": {
            "enabled": true,
            "provider": "brave",
            "apiKey": "{Your-Brave-Search-API-Key}",
            "maxResults": 10,
            "timeoutSeconds": 30
        },
        "fetch": {
            "enabled": true,
            "timeoutSeconds": 30
        }
        }
    },
    ```

5. Click **Save** in the upper-right corner. The system validates the configuration and restarts automatically to apply the changes.
6. Open Discord and ask your agent a question that requires real-time internet data to verify the search tool is working.