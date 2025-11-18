---
outline: [2, 3]
description: Learn how to set up DeerFlow on your Olares device, complete with Ollama integration and Tavily for web research.
---

# Build a local deep research agent with DeerFlow
DeerFlow is an open-source framework that transforms a simple research topic into a comprehensive, detailed report.

This guide will walk through the process of setting up DeerFlow on your Olares device, integrating it with a local Ollama model and the Tavily search engine for web-enabled research.

## Learning objectives
In this guide, you will learn how to:
- Configure DeerFlow to communicate with a local LLM via Ollama.
- Configure the Tavily search API for web access.
- Execute deep research tasks and manage reports.

## Prerequisites
Before you begin, make sure:
- Ollama is installed and running in your Olares environment.
- At least one model installed using Ollama. See [Ollama](./ollama.md) for details.
- You have a [Tavily](https://www.tavily.com/) account (a free account is sufficient).

## Install DeerFlow
1. Open **Market**, and search for "DeerFlow".
   ![Install DeerFlow](/images/manual/use-cases/deerflow.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure DeerFlow
DeerFlow requires connection details for the LLM. You will configure this by editing the `conf.yaml` file using either the graphical interface or the command line.

### Configure DeerFlow to use Ollama
<tabs>
<template #Use-graphical-interface>

1. Open the Files app and navigate to `/Applications/Data/Deerflow/app/`.
2. Locate the `conf.yaml` file and download it to your local computer.
    ![Find `conf.yaml` in Files](/images/manual/use-cases/deerflow-conf-yaml-in-files.png#bordered)

3. Open the `conf.yaml` file with a text editor.
4. Modify the default model settings:
   ```yaml
    BASIC_MODEL:
      base_url:  # Your Ollama API endpoint (ensure /v1 suffix is included)
      model: # The model name
      api_key: # Any non-empty string
   ```
   For example:
   ```yaml
    BASIC_MODEL:
      base_url: https://39975b9a1.{YOURUSERNAME}.olares.com/v1
      model: "cogito:14b"
      api_key: ollama
   ```
5. Save the file.
6. Return to the Files app, delete the original `conf.yaml` file, and upload your modified version.

</template>

<template #Use-command-line>

You can edit the configuration file directly on the host machine via the terminal.
1. Open Control Hub and select the DeerFlow project from the sidebar.
2. Navigate to **Deployments** > **deerflow** and click the running pod.
3. Expand the **deerflow** container details to view the **Volumes** section.
   ![Locate DeerFlow's containers](/images/manual/use-cases/deerflow-locate-containers.png#bordered)

   ![Find app folder](/images/manual/use-cases/deerflow-app-volume.png#bordered)

4. Copy this path.
5. Open the Olares terminal from Control Hub, and change directory to the copied path:
   ```bash
   # Replace with your actual path
   cd /olares/rootfs/userspace/pvc-userspace-laresprime-raizlofhiszoin5c/Data/deerflow/app
   ```
6. Edit the `conf.yaml` file using a command-line text editor like `nano` or `vi`. For exmaple:
   ```Bash
   nano conf.yaml
   ```
7. Modify the default model settings:
   ```yaml
    BASIC_MODEL:
      base_url:  # Your Ollama API endpoint (ensure /v1 suffix is included)
      model: # The model name
      api_key: # Any non-empty string
   ```
   For example:
   ```yaml
    BASIC_MODEL:
      base_url: https://39975b9a1.{YOURUSERNAME}.olares.com/v1
      model: "cogito:14b"
      api_key: ollama
   ```
8. Save the changes and exit the editor.
</template>
</tabs>


### Configure DeerFlow to use Tavily
To enable web search, add your Tavily API key to the application configuration.
1. In Control Hub, select the DeerFlow project.
2. Click **Configmaps** in the resource list and select **deerflow-config**.
    ![Browse to DeerFlow's configmaps](/images/manual/use-cases/deerflow-configmap.png#bordered)

3. Click <span class="material-symbols-outlined">edit_square</span> in the top-right to open the editor.
4. Add the following key-value pairs under the `data` section:
   ```yaml
   SEARCH_API: tavily
   TAVILY_API_KEY: tvly-xxx # Your Tavily API Key
   ```
   ![Configure Tavily](/images/manual/use-cases/deerflow-configure-tavily.png#bordered)

5. Click **Confirm** to save the changes.

### Restart DeerFlow
Restart the service to apply the new model and search configurations.

1. In Control Hub, select the DeerFlow project.
2. Under **Deployments**, locate **deerflow** and click **Restart**.
   ![Restart DeerFlow](/images/manual/use-cases/deerflow-restart.png#bordered)

2. In the confirmation dialog, type `deerflow` and click **Confirm**.
3. Wait for the status icon to turn green, which indicates the service has successfully restarted.

## Run DeerFlow
### Run a deep research task
1. Open **DeerFlow** from the Olares Launchpad.
2. Click **Get Started** and enter your research topic in the prompt box.
    ![Enter research prompt](/images/manual/use-cases/deerflow-enter-prompt.png#bordered)
 
3. Click the wand icon to have DeerFlow refine your prompt for better results.
4. Enable **Investigation**.
5. Select your preferred writing style (e.g., **Popular Science**).
6. Click <span class="material-symbols-outlined">arrow_upward</span> to send the request.

DeerFlow will generate a preliminary research plan. Review and edit this plan if necessary, or allow it to proceed.
![Generate research plan](/images/manual/use-cases/deerflow-generate-research-plan.png#bordered)

Once the process is complete, a detailed analysis report will be displayed.
![View research report](/images/manual/use-cases/deerflow-generate-research-report.png#bordered)

To audit the sources and steps taken, click the **Activities** tab.
![Review research activities](/images/manual/use-cases/deerflow-review-research-activities.png#bordered)
### Edit and save the report
:::info Verify citations
AI models may occasionally generate inaccurate citations or "hallucinated" links. Manually verify important sources in the citations section.
:::

1. Click <span class="material-symbols-outlined">edit</span> in the top-right corner to enter editing mode.
2. You can adjust formatting using Markdown or select a section and ask the AI to improve or expand on it.
   ![Ask AI to edit the report](/images/manual/use-cases/deerflow-ask-ai-to-edit.png#bordered)
3. Click <span class="material-symbols-outlined">undo</span> in the top-right corner to exit editing mode.
4. Click <span class="material-symbols-outlined">download</span> to save the report to your local machine as a Markdown file.

## Add an MCP server
The Model Context Protocol (MCP) extends DeerFlow's capabilities by integrating external tools. For example, adding the Fetch server allows the agent to scrape and convert web content into Markdown for analysis.

1. Open your DeerFlow app, and click <span class="material-symbols-outlined">settings</span> to open the **Settings** dialog.
2. Select the **MCP** tab and click **Add Servers**.
3. Paste the JSON configuration for the server. The following example adds the fetch server:
   ```json
    {
      "mcpServers": {
        "fetch": {
          "command": "uvx",
          "args": ["mcp-server-fetch"]
        }
      }
    }
   ```
4. Click **Add**. The server is automatically enabled and available for research agents.
   ![Add MCP server](/images/manual/use-cases/deerflow-add-mcp-server.png#bordered)

## Turn research report to a podcast (TTS)
DeerFlow can convert reports into MP3 audio using a Text-to-Speech (TTS) service, such as Volcengine TTS. This requires adding API credentials to the application environment.

1. Obtain your **Access Token** and **App ID** from the [Volcengine](https://console.volcengine.com) console.
2. In Control Hub, select the DeerFlow project and go to **Configmaps** > **deerflow-config**.
3. Click the **Edit** icon in the top-right corner.
4. Add the following keys under the `data` section:
   ```yaml
   VOLCENGINE_TTS_ACCESS_TOKEN: # Your Access Token
   VOLCENGINE_TTS_APPID: # Your App ID
   ```
5. Click **Confirm** to save the changes.
6. Navigate to **Deployments** > **deerflow** and click **Restart**.

Once restarted, DeerFlow should detect these keys and the podcast/TTS feature will be available.

## FAQ
### DeerFlow does not generate a response
If the agent fails to start or hangs:
- **Check model compatibility:** DeerFlow does not support reasoning models (e.g., DeepSeek R1). Switch to a standard chat model and try again.
- **Check endpoint configuration:** Ensure the Ollama API endpoint in `conf.yaml` includes the `/v1` suffix.

### No web search results during the research
If the report is generic and lacks external data:
- **Check model capabilities:** The selected LLM may lack strong tool-calling capabilities. Switch to a model known for effective tool use, such as Qwen 2.5 or Llama 3.1.
- **Verify API Key:** Ensure the `TAVILY_API_KEY` in the ConfigMap is correct and the account has remaining quota.