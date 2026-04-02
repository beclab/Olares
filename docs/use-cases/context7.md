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

Context7 is an MCP (Model Context Protocol) server that provides AI coding assistants with real-time, accurate library documentation. Instead of relying on outdated training data, your AI tools can pull the latest docs for any library on demand.

On Olares, you can connect Context7 to Olares-hosted AI agents like Agent Zero, LibreChat, and OpenCode, or to external coding assistants like Cursor and Claude Desktop.

## Learning objectives

In this guide, you will learn how to:
- Look up library documentation using the Context7 CLI.
- Connect Context7 to Olares-hosted AI agents.
- Connect Context7 to external coding assistants like Cursor.
- (Optional) Register an API key for higher rate limits.

## Install Context7

1. Open Market and search for "Context7".
   <!-- ![Install Context7](/images/manual/use-cases/context7.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Look up documentation with the CLI

:::tip
In practice, you'll typically let AI agents call Context7 automatically via MCP. The CLI is mainly useful for manual lookups and debugging.
:::

### Search for a library

Use `ctx7 library` to find a library by name:
```bash
# Search for React-related libraries
ctx7 library react "hooks usage"

# Search for Next.js
ctx7 library nextjs "routing"

# Search for Express
ctx7 library express "middleware"
```

<!-- ![Context7 library search](/images/manual/use-cases/context7-library-search.png#bordered) -->

The output returns matching library IDs (e.g., `/reactjs/react.dev`) that you can use in the next step.

### Fetch documentation

Use `ctx7 docs` with a library ID to retrieve specific documentation:

```bash
# Fetch React 19 docs about the 'use' hook
ctx7 docs "/reactjs/react.dev" "How to use the 'use' hook for async data in React 19?"

# Fetch general docs
ctx7 docs "/reactjs/react.dev" "find and filter documents"
```

For more CLI commands, see the [Context7 CLI reference](https://context7.com/docs/clients/cli).

## (Optional) Register and configure an API key

Context7 supports both anonymous and authenticated usage:

- **Anonymous mode**: Works out of the box with no registration. Suitable for occasional queries.
- **Authenticated mode**: Recommended if you make more than 50 daily queries, need lower latency, or want access to private knowledge bases.

:::info
Context7 on Olares does not support OAuth login. Use an API key instead.
:::

To configure an API key:

1. Go to the [Context7 dashboard](https://context7.com/dashboard) and create an API key. Copy the key (it starts with `ctx7sk`).
2. In Olares, navigate to **Settings** > **Applications** > **Context7** > **Manage environment variables**.
3. Paste your API key and click **Confirm**, then click **Apply**.

    <!-- ![Context7 API key configuration](/images/manual/use-cases/context7-api-key.png#bordered) -->

4. The app restarts automatically. You can check the restart status in Market.

    <!-- ![Context7 restarting](/images/manual/use-cases/context7-restarting.png#bordered) -->

5. To verify, run a CLI query and check the [Context7 dashboard](https://context7.com/dashboard) for API call records.

    <!-- ![Context7 API call records](/images/manual/use-cases/context7-api-records.png#bordered) -->

## Connect Context7 to Olares apps

To use Context7 with Olares-hosted AI agents, you need the MCP endpoint URL:

1. Navigate to **Settings** > **Applications** > **Context7**.
2. Under **Context7 MCP** > **Endpoint**, copy the endpoint URL.

<!-- ![Context7 MCP endpoint](/images/manual/use-cases/context7-mcp-endpoint.png#bordered) -->

Then configure Context7 in your preferred agent app.

### Agent Zero

1. Open Agent Zero and go to **Settings** > **MCP/A2A**.

    <!-- ![Agent Zero MCP settings](/images/manual/use-cases/context7-agent-zero-settings.png#bordered) -->

2. Add a new MCP server with the following configuration. Replace the URL with your Context7 MCP endpoint:

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

3. Click **Apply Now**, then **Save**.
4. Go to **Settings** > **Agent settings** and configure your model provider, model name, and URL for each model role (Chat Model, Utility Model, Web Browser Model).

    <!-- ![Agent Zero model settings](/images/manual/use-cases/context7-agent-zero-models.png#bordered) -->

5. Start a conversation. When Context7 is called successfully, you'll see a notification in the response.

    <!-- ![Agent Zero Context7 success](/images/manual/use-cases/context7-agent-zero-success.png#bordered) -->

### LibreChat

1. Open LibreChat and go to the MCP settings in the sidebar.

    <!-- ![LibreChat MCP settings](/images/manual/use-cases/context7-librechat-settings.png#bordered) -->

2. Enter a server name (e.g., `context7`) and paste your Context7 MCP endpoint URL. Check **Trust this app** and click **Update**.

    <!-- ![LibreChat MCP configuration](/images/manual/use-cases/context7-librechat-config.png#bordered) -->

3. In the chat input, select Context7 from the MCP server list.

    <!-- ![LibreChat select Context7](/images/manual/use-cases/context7-librechat-select.png#bordered) -->

4. Ask a question. If the response shows that the message was sent to Context7, the integration is working.

    <!-- ![LibreChat Context7 success](/images/manual/use-cases/context7-librechat-success.png#bordered) -->

### OpenCode

1. Open the OpenCode terminal (top-right corner) and create the configuration file:

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

    Replace `<your-context7-endpoint>` with your Context7 MCP endpoint URL.

2. Open Control Hub, navigate to the OpenCode app, find the backend container under **Deployments**, and click **Restart**.
3. Wait for the status indicator to turn green.

    <!-- ![OpenCode restart](/images/manual/use-cases/context7-opencode-restart.png#bordered) -->

4. Reopen OpenCode. Click **Status** > **MCP** (top-right corner) to verify that Context7 appears in the MCP plugin list.

    <!-- ![OpenCode MCP status](/images/manual/use-cases/context7-opencode-mcp.png#bordered) -->

:::tip Verify the connection
If asking "check Context7 MCP servers" in the chat doesn't produce a response, try alternative commands such as `restart MCP servers`, `mcp list`, or `/mcp list`.
:::

## Connect external clients via MCP

You can also connect Context7 to coding assistants running on your computer, such as Cursor or Claude Desktop. This section uses Cursor as an example.

<tabs>
<template #Use-.local-domain-(LAN)>

If your computer is on the same local network as your Olares device, you can connect directly using the `.local` domain.

:::info Windows users
On Windows, multi-level `.local` domains require additional setup. Try one of these:
- **Import hosts in LarePass**: Open the LarePass desktop app and use the built-in option to import Olares hosts to your system.
- **Use the single-level domain**: Change `https://806ba3e40.{username}.olares.com` to `http://806ba3e40-{username}-olares.local`.

For details, see [Access Olares services locally](../manual/best-practices/local-access.md).
:::

1. Get your Context7 MCP endpoint from **Settings** > **Applications** > **Context7** > **Context7 MCP** > **Endpoint**.
2. For the MCP server URL, use your endpoint with the `.local` domain and `http`. For example, if your endpoint is:
    ```plain
    https://f86d25051.username.olares.com
    ```
    Change it to:
    ```plain
    http://f86d25051.username.olares.local
    ```
3. In Cursor, go to **Settings** > **Tools & MCP** > **New MCP Server**.
4. Enter the following configuration:
    ```json
    {
      "mcpServers": {
        "context7": {
          "url": "http://f86d25051.username.olares.local/mcp"
        }
      }
    }
    ```
5. Save the configuration. When Context7 is called successfully, you'll see a tool-use notification in Cursor's response.

</template>
<template #Use-.com-domain-(VPN)>

If your computer is on a different network, you need to update the access policy and enable LarePass VPN.

1. Update Context7's access policy to enable direct access from external apps:

    a. Navigate to **Settings** > **Applications** > **Context7**.

    b. Set **Authentication level** to `Internal`.

    <!-- ![Context7 authentication level](/images/manual/use-cases/context7-auth-level.png#bordered) -->

2. Enable LarePass VPN on your computer.

    ![Enable LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered){width=70%}

3. Get your Context7 MCP endpoint from **Settings** > **Applications** > **Context7** > **Context7 MCP** > **Endpoint**.
4. In Cursor, go to **Settings** > **Tools & MCP** > **New MCP Server**.
5. Enter the following configuration, replacing the URL with your endpoint:
    ```json
    {
      "mcpServers": {
        "context7": {
          "url": "<your-context7-endpoint>/mcp"
        }
      }
    }
    ```
6. Save the configuration. When Context7 is called successfully, you'll see a tool-use notification in Cursor's response.

    <!-- ![Cursor Context7 success](/images/manual/use-cases/context7-cursor-success.png#bordered) -->

</template>
</tabs>

:::tip Other MCP clients
The same approach works for Claude Desktop and other MCP-compatible tools. Use this configuration format:
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

Context7 supports installable skills that extend its capabilities. You can search, install, and manage skills using the CLI.

### Search and install skills

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

```bash
# List installed skills
ctx7 skills list

# List skills for a specific client
ctx7 skills list --claude
ctx7 skills list --cursor

# Remove a skill
ctx7 skills remove pdf
```

For more skills, visit the [Context7 Skills Marketplace](https://context7.com/skills).

## FAQ

### Can I call the Olares-hosted Context7 API from external programs?

No. The Context7 instance on Olares serves as an MCP server for AI assistants, not as a general-purpose API endpoint. If you need programmatic API access, use the official Context7 API at [context7.com](https://context7.com).

## Learn more

- [Context7 documentation](https://context7.com/docs): Official docs for Context7, including CLI reference, MCP setup, and skills.
- [Download and run local AI models via Ollama](ollama.md): Set up a local model backend for your AI agents.
- [Set up Open WebUI with Ollama](openwebui-ollama.md): Add a chat interface for local models.
