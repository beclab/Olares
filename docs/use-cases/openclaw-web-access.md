---
outline: [2, 3]
description: 
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, OpenClaw web search
---

# Optional: Enable web search

By default, OpenClaw answers questions only based on its training data, which means it doesn't know about current events or real-time news. To give your agent access to the live internet, you can enable the web search tool.

## Enable web search

OpenClaw officially recommends Brave Search. It uses an independent web index optimized for AI retrieval, ensuring your agent finds accurate information.

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
    | Enable web_fetch (keyless HTTP <br>fetch) | Yes |

4. Finalize the configuration in Control UI.

    The CLI wizard sets up the API key, but you can customize specific tool parameters such as timeouts and limits in the Control UI.

    a. Return to the **Control UI** > **Config** > **Raw** tab. 

    b. Find the `tools` section and update as follows: 

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

5. Now you can ask the agent in Discord to answer questions that require real-time internet data.