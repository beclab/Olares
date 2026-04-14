---
outline: [2, 3]
description: Set up Context7 on Olares to provide AI coding assistants with up-to-date library documentation via MCP. Connect it to Olares-hosted agents or external tools like Cursor.
head:
  - - meta
    - name: keywords
      content: Olares, Context7, MCP, Model Context Protocol, AI coding assistant, documentation, Cursor, Agent Zero, LibreChat, OpenCode
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-02"
---

# Connect AI coding assistants to up-to-date docs with Context7

Context7 is a Model Context Protocol (MCP) server that provides AI coding assistants with real-time, accurate library documentation. Instead of relying on outdated training data, your AI tools can pull the latest docs for any library on demand.

On Olares, you can connect Context7 to Olares-hosted AI agents like Agent Zero, LibreChat, and OpenCode, or to external coding assistants like Cursor and Claude Desktop.

## Learning objectives

In this guide, you will learn how to:
- Look up library documentation using the Context7 Terminal.
- (Optional) Register an API key for higher rate limits.
- Connect Context7 to Olares-hosted AI agents.
- Connect Context7 to external coding assistants like Cursor.

## Install Context7

1. Open Market and search for "Context7".
   ![Install Context7](/images/manual/use-cases/context7.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to complete.

## Look up documentation with Context7 Terminal

Context7 Terminal allows you to manually search for libraries and retrieve specific documentation. Use it to test queries, verify that Context7 is working, or debug a library’s availability before integrating with an AI agent.

:::tip
In practice, you'll typically let AI agents call Context7 automatically via MCP. The Context7 Terminal is mainly useful for manual lookups and debugging.
:::

### Search for a library

Find the correct library ID that you’ll use to fetch documentation.

1. Open Context7 Terminal from the Launchpad. 
2. Use `ctx7 library <library-name> "your-query"` to find a library by name, where `"your-query"` describes what you want to do.

    ```bash
    # Example: Search for React-related libraries using "hooks usage"
    ctx7 library react "hooks usage"

    # Example: Search for Next.js using "routing"
    ctx7 library nextjs "routing"

    # Example: Search for Express using "middleware"
    ctx7 library express "middleware"
    ```

    The output returns matching library IDs (e.g., `/reactjs/react.dev`) that you can use in the next step.
  
    ```bash
    Title: React
    Context7-compatible library ID: /reactjs/react.dev
    Description: React.dev is the official documentation website for React, a JavaScript library for building user interfaces, providing guides, API references, and tutorials.
    Code Snippets: 2781
    Source Reputation: High
    Benchmark Score: 85.1
    ``` 

    When multiple results are returned, the best match is usually the one with the closest name, highest snippet count, and strongest reputation.

### Fetch documentation

Retrieve detailed, up‑to‑date content for a specific library using its ID.

Use `ctx7 docs "library-id" "your-query"`, where `"your-query"` describes what you want to know in natural language.

  ```bash
  # Example: Fetch React 19 docs about the 'use' hook for async data
  ctx7 docs "/reactjs/react.dev" "How to use the 'use' hook for async data in React 19?"

  # Example: Ask how to filter documents in React
  ctx7 docs "/reactjs/react.dev" "find and filter documents"
  ```

For more information, see the [Context7 CLI reference](https://context7.com/docs/clients/cli).

## (Optional) Register and configure an API key

Context7 supports both anonymous and authenticated usage:

- **Anonymous mode**: Works out of the box with no registration. Suitable for occasional queries.
- **Authenticated mode**: Recommended if you make more than 50 daily queries, need lower latency, or want access to private knowledge bases.

To configure an API key:

1. Go to the [Context7 dashboard](https://context7.com/dashboard).
2. In the **API Keys** section, click **Create API Key**.

  ![Context7 dashboard create api key](/images/manual/use-cases/context7-new-api-key.png#bordered){width=70%} 
3. Optionally enter a name for the API key, click **Create API Key**, and then copy the key which starts with `ctx7sk`.

  :::tip Save the key now
  Save the key now. You will not see it again.
  :::
4. Click **Done**.
5. In Olares, open Settings, and then go to **Applications** > **Context7** > **Manage environment variables**.

  ![Context7 API key configuration](/images/manual/use-cases/context7-api-key.png#bordered){width=70%}

6. Click <i class="material-symbols-outlined">edit_square</i>, paste your API key, and then click **Confirm**.
7. Click **Apply**. The Context7 app restarts automatically.
8. To verify that the API key is correctly configured, run a query in Context7 Terminal, and then check the Context7 dashboard.

   ![Context7 API call records](/images/manual/use-cases/context7-api-records.png#bordered){width=90%}

    The **REQUESTS** number (i.e., the API call count) indicates that Context7 on Olares is using the key for authenticated requests.

## Connect Context7 to Olares apps

Before proceeding, ensure that you have configured the necessary settings within each agent app, such as the model provider, model name, and base URL.

### Obtain MCP endpoint

To use Context7 with Olares-hosted AI agents, you need to obtain the MCP endpoint URL first, and then configure Context7 in your preferred agent app.

1. Open Settings, and then go to **Applications** > **Context7** > **Context7 MCP**
2. Copy the endpoint URL. For example, `https://f86d25051.olaresdemo.olares.com`.

    ![Context7 MCP endpoint](/images/manual/use-cases/context7-mcp-endpoint.png#bordered){width=70%}

### Agent Zero

Add Context7 as an MCP server in Agent Zero, then configure your model provider to enable the agent to call the documentation tools.

1. Open Agent Zero, go to **Settings** > **MCP/A2A**, and then click **Open**.

    ![Agent Zero MCP settings](/images/manual/use-cases/context7-agent-zero-settings.png#bordered){width=70%}

2. In the **MCP Servers Configuration** window, add a new MCP server with the following configuration. Replace `<your-context7-endpoint>` with your Context7 MCP endpoint:

    ```json
    {
      "mcpServers": {
        "context7": {
          "type": "streamable-http",
          "url": "<your-context7-endpoint>/mcp"
        }
      }
    }
    ```

    For example,

    ```json
    {
      "mcpServers": {
        "context7": {
          "type": "streamable-http",
          "url": "https://f86d25051.olaresdemo.olares.com/mcp"
        }
      }
    }
    ```    

3. Click **Apply now**, and then close the window.
4. Click **Save**.
5. Start a conversation. For example,

    ```text
    Use Context7 to look up the latest React 19 documentation.
    How do I fetch data with the new use hook inside a Server Component? 
    Show me a code example.
    ```

    When Context7 is called successfully, you'll see the communication process with MCP and the solution proposed for your question.

    ![Agent Zero Context7 success](/images/manual/use-cases/context7-agent-zero-success.png#bordered)

### LibreChat

Enable the Context7 MCP server in LibreChat and select it from the chat input to start using live documentation.

1. Open LibreChat, and then click the **MCP Settings** icon on the right sidebar.

    ![LibreChat MCP settings](/images/manual/use-cases/context7-librechat-settings.png#bordered)

2. Click <i class="material-symbols-outlined">add_2</i> next to **Filter MCP servers by name**.
3. In the **Add MCP Server** window:

    a. **Name**: Enter a server name such as `context7`.

    b. **MCP Server URL**: Enter the URL in the format of `<your-context7-endpoint>/mcp`. Replace `<your-context7-endpoint>` with your Context7 MCP endpoint.
  
    c. **I trust this application**: Select this option.
  
    d. Click **Create**. You get the `MCP server created successfully` message.

    ![LibreChat MCP configuration](/images/manual/use-cases/context7-librechat-config.png#bordered){width=70%}

4. On the right sidebar, find the newly added MCP server **context7**, and then click the **Connect** icon.

    ![LibreChat MCP connect](/images/manual/use-cases/context7-librechat-connect.png#bordered){width=50%}

5. In the chat window, ensure **context7** is selected from the **MCP Servers** list.

    ![LibreChat select Context7](/images/manual/use-cases/context7-librechat-select.png#bordered)

6. Ask a question. For example,

    ```text
    Use the context7 to look up React 19 documentation. Then show me 
    how to fetch data with the use hook inside a Server Component,
    including handling loading and error states.
    ```

    You will see from the responses that the info was sent to Context7, which indicates the integration is working.

    ![LibreChat Context7 success](/images/manual/use-cases/context7-librechat-success.png#bordered)

### OpenCode

Create a configuration file to register Context7 as a remote MCP server, then restart the OpenCode container to load it.

1. Open OpenCode, and then click <i class="material-symbols-outlined">terminal</i> in the upper-right corner.
2. Enter the following command to create the configuration file. Replace `<your-context7-endpoint>` with your Context7 MCP endpoint.

    ```bash
    cat << 'EOF' > /home/opencode/.config/opencode/opencode.json
    {
      "$schema": "https://opencode.ai/config.json",
      "mcp": {
        "context7": {
          "type": "remote",
          "url": "<your-context7-endpoint>/mcp",
          "enabled": true
        }
      }
    }
    EOF
    ```

    ![OpenCode config](/images/manual/use-cases/context7-opencode-config.png#bordered)

3. Open Control Hub, find the **opencode** deployment, and then click **Restart**.

    ![OpenCode restart](/images/manual/use-cases/context7-opencode-restart.png#bordered)

4. Wait for the status indicator to turn green.
5. Reopen OpenCode.
6. Click the **Status** icon in the upper-right corner, click the **MCP** tab, and then verify that **context7** is enabled.

    ![OpenCode MCP status](/images/manual/use-cases/context7-opencode-mcp.png#bordered)

7. In the chat, ask a question. For example,

    ```text
    Use Context7 to fetch the latest React 19 documentation. Then write a Server
    Component that fetches and displays user data using the use hook. Include 
    proper loading and error handling.
    ```

    You will see the process of calling context7 to produce the response, which indicates the integration is working.

    ![OpenCode MCP success](/images/manual/use-cases/context7-opencode-mcp-success.png#bordered){width=70%}

## Connect external clients via MCP

You can also connect Context7 to coding assistants running on your computer, such as Cursor or Claude Desktop. This section uses Cursor as an example.

1. Open the LarePass desktop client on your computer, and then enable **VPN connection** to connect to Olares.
   ![Enable LarePass VPN on desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

    :::tip On the same local network?
    If your computer and Olares are on the same LAN, you can skip VPN and use the `.local` domain instead. Replace `https://f86d25051.{username}.olares.com` with `http://f86d25051.{username}.olares.local` in the config in Step 3. For details, see [Use `.local` domain](../manual/best-practices/local-access.md#method-2-use-local-domain).
    :::

2. Open Cursor, and then go to **Settings** > **Tools & MCP** > **Add custom MCP**.

  ![Cursor settings](/images/manual/use-cases/context7-cursor-settings.png#bordered){width=70%}

3. Enter the following configurations in the `mcp.json` file. Replace `<your-context7-endpoint>` with your Context7 MCP endpoint.

    ```json
    {
      "mcpServers": {
        "context7": {
          "url": "<your-context7-endpoint>/mcp"
        }
      }
    }
    ```
4. Save the changes. Now Context7 is enabled.

  ![Context7 enabled in Cursor](/images/manual/use-cases/cursor-context7-enabled.png#bordered){width=50%}

5. Ask in the chat. For example,

    ```text
    Use the Context7 MCP server to look up the latest React 19 documentation.
    Then show me how to use the use hook to fetch async data inside a Server
    Component, including handling loading and error states.
    ```

    When Context7 is called successfully, you'll see the tool-use notifications in Cursor's response.

  ![Context7 called in Cursor success](/images/manual/use-cases/cursor-context7-success.png#bordered){width=50%}

:::tip Other MCP clients
The same approach works for Claude Desktop and other MCP-compatible tools. Use the configuration format and replace `<your-context7-endpoint>` with your Context7 MCP endpoint.
```json
{
  "mcpServers": {
    "context7": {
      "url": "<your-context7-endpoint>/mcp"
    }
  }
}
```
:::

## Manage Context7 skills

Context7 supports installable skills that extend its capabilities. You can search, install, and manage skills using the Context7 Terminal.

### Search and install skills

Discover available skills from the marketplace and install them to extend Context7’s capabilities (e.g., PDF processing, testing frameworks).

```bash
# Search for skills by keyword
ctx7 skills search pdf
ctx7 skills search "react testing"

# Browse all skills in a repository
ctx7 skills info /anthropics/skills

# Install a specific skill
ctx7 skills install /anthropics/skills pdf

# Get skill suggestions based on your project
ctx7 skills suggest
```

### List and remove skills

View installed skills and remove those you no longer need.

```bash
# List installed skills
ctx7 skills list

# List skills for a specific client
ctx7 skills list --claude
ctx7 skills list --cursor

# Remove a skill
ctx7 skills remove pdf
```
For more information, see the [Context7 Skills Marketplace](https://context7.com/skills).

## FAQs

### Can I call the Olares-hosted Context7 API from external programs?

No. The Context7 instance on Olares serves as an MCP server for AI assistants, not as a general-purpose API endpoint. If you need programmatic API access, use the official Context7 API at [context7.com](https://context7.com).

### Why is my AI still hallucinating (giving wrong or outdated answers)?

Your AI might not realize it has access to Context7. Unless you explicitly ask it to use Context7, it might fall back on its own training data, which could be outdated.

To fix this issue, add a phrase like “Use Context7” to your question. For example, ask "Use Context7 to find how to use the `use` hook in React 19" instead of "How do I use the `use` hook in React 19".

## Learn more

- [Context7 documentation](https://context7.com/docs)
- [Download and run local AI models via Ollama](ollama.md)
