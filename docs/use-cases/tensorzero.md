---
outline: [2, 3]
description: Set up TensorZero on Olares to connect your apps to your AI models, monitor their performance, and manage your setup in one place.
head:
  - - meta
    - name: keywords
      content: Olares, TensorZero, LLMOps, AI gateway, observability, evaluation, MCP, Ollama, self-hosted
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-05-09"
---

# Use TensorZero as an AI model gateway and observability platform

TensorZero is an all-in-one platform to manage, connect, and monitor your AI models. It acts as a central gateway that connects your client applications to your local AI models. It records every chat and request so you can track performance, and it helps you test different setups to get the best results.

## Learning objectives

In this guide, you will learn how to:
- Install TensorZero on Olares.
- Understand how TensorZero manages AI connections.
- Connect a chat model and a function.
- (Optional) Connect an embedding model.
- Test your setup using the built-in Playground.
- Connect other apps to TensorZero.
- Enable your AI agent to read performance data via the built-in MCP server.

## Prerequisites

- Ensure [Ollama is installed](ollama.md) with at least one chat model downloaded (e.g., `qwen3.5:9b`) and one embedding model downloaded (e.g., `nomic-embed-text`).
- Ensure your client applications, such as OpenCode and AgentZero, are already installed and fully functional. This guide covers only the specific settings required to connect them to TensorZero.

## Install TensorZero

1. Open Market and search for "TensorZero".

    ![Search for TensorZero from Market](/images/manual/use-cases/tensorzero.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Understand the configuration requirements

TensorZero does not provide a graphical interface for configuring models. You manage all settings by editing its configuration file in Files.

Before you edit the file, review the following rules to avoid errors:
- **Strict permission**: TensorZero rejects direct requests to raw model names like `gpt-4o` and `qwen3.5`. You must define an alias for every model you want to use. Do not use dots or colons in your alias names. For example, use `qwen3_5_9b`, not `qwen3.5:9b`.
- **Exact naming**: When you connect other apps to TensorZero, you must prepend your model aliases with specific prefixes, such as `tensorzero::model_name::<alias>` and `tensorzero::function_name::<alias>`.

    :::tip
    For applications that use the LiteLLM framework, such as the embedding feature in AgentZero, you must include the `openai/` prefix in the embedding model name. For example, use `openai/tensorzero::embedding_model_name::nomic_embed`.
    :::

- **Formatting rules**: The configuration file uses the TOML text format. You must maintain at least one empty line between different sections, for example, between `[models]` and `[functions]`. If you remove the empty lines, the application fails to start.

## Configure a chat model and function

To make TensorZero work, you need two things: a model to act as the AI engine, and a function to serve as the access point your apps use to communicate with that engine.

You define the model to tell TensorZero where the AI is, and then you link it to a function to handle the requests. 

This example connects a local Ollama model.

1. Open Settings, go to **Applications** > **Ollama** > **Shared entrances** > **Ollama API**, and then copy the endpoint URL. For example, `http://d54536a50.shared.olares.com`.
2. Open Files, and then go to **Data** > **tensorzero** > **config**.
3. Right-click `tensorzero.toml`, and then rename it to `tensorzero.toml.txt`.
4. Double-click `tensorzero.toml.txt`, and then click <i class="material-symbols-outlined">edit_square</i>.
5. In the editor, add the following snippet:

    - Replace `api_base` with your copied Ollama endpoint URL and append `/v1`.
    - Replace `model_name` with the exact name of the model you downloaded in Ollama.

    This configuration registers your Ollama model under the alias `qwen3_5_9b`, and creates a client-facing function `general_chat` that routes incoming app requests to that model.

    ```bash
    # models
    [models.qwen3_5_9b]
    routing = ["ollama"]
    [models.qwen3_5_9b.providers.ollama]
    type = "openai"
    api_base = "<ollama-shared-entrance>/v1"
    model_name = "qwen3.5:9b"
    api_key_location = "none"

    # functions
    [functions.general_chat]
    type = "chat"
    [functions.general_chat.variants.my_default_variant]
    type = "chat_completion"
    model = "qwen3_5_9b"
    ```

    ![Connect to Ollama](/images/manual/use-cases/tensorzero-config-ollama.png#bordered)

6. Click <i class="material-symbols-outlined">save</i>, and then close the file.
7. Rename `tensorzero.toml.txt` back to `tensorzero.toml`.
8. Open Control Hub, go to **Browse** > **tensorzero-{username}** > **Deployments** > **tensorzero**, and then click **Restart** for the application to apply the new settings.

    ![TensorZero pod restart](/images/manual/use-cases/tensorzero-pod-restart.png#bordered)

## (Optional) Configure an embedding model

Many apps require embedding models to search through documents or build memory features. TensorZero treats embedding models  separately from chat models. You must define a dedicated embedding model. Do not use a chat function for memory tasks.

1. Add the following snippet in `tensorzero.toml` to define an embedding model. Replace `model_name` with the exact name of the embedding model you downloaded in Ollama.

    This configuration registers your Ollama embedding model under the alias `nomic_embed`.

    ```bash
    # embedding_models
    [embedding_models.nomic_embed]
    routing = ["ollama"]
    [embedding_models.nomic_embed.providers.ollama]
    type = "openai"
    api_base = "<ollama-shared-entrance>/v1"
    model_name = "nomic-embed-text"
    api_key_location = "none"
    ```

    ![Connect to embedding model](/images/manual/use-cases/tensorzero-config-embedding.png#bordered)    

2. Restart the **tensorzero** container in Control Hub for the application to apply the new settings.

## Verify the connection

Use the built-in Playground to test that your function works correctly with your Ollama model.

The Playground requires at least one test case, called a Datapoint, to display the chat interface. If you do not have one, you must create it manually.

1. Open TensorZero from the Launchpad.
2. Select **Datasets** from the left sidebar.
3. Click **New Datapoint**, and then configure the test case details.

    For example, to create a basic geography test:
    - **Dataset**: Specify a name to create a new collection of test cases. For example, `Baseline tests`.
    - **Function**: Select the function you configured earlier. For example, `general_chat`.
    - **Input**: Select **+ User Message**, click **+ Text**, and then enter a test prompt. For example, `What is the capital of Spain?`.
    - **Output**: Select **+ Text**, and then enter the exact answer you expect the model to generate. For example, `Madrid`.
    - (Optional) **Tags** and **Metadata**: Enter labels to help identify this test case later. For example, add a tag with **Key** set to `type` and **Value** set to `QA`.

    ![Create a new datapoint](/images/manual/use-cases/tensorzero-new-datapoint.png#bordered)      

4. Click **Create Datapoint**.
5. Select **Playground** from the left sidebar.
6. Select your function, the dataset you just created, and your variant. The chat interface appears. If you receive a normal reply, the setup is successful.

    ![Verify connection](/images/manual/use-cases/tensorzero-playground.png#bordered)   

## Obtain the TensorZero endpoint

To connect other apps to TensorZero, get its entrance address.

1. Open Settings, and then go to **Applications** > **TensorZero** > **Entrances** > **TensorZero**.

    ![TensorZero endpoint addres](/images/manual/use-cases/tensorzero-endpoint.png#bordered){width=70%} 

2. Copy the endpoint URL. For example, `https://ea581361.laresprime.olares.com`. For OpenAI‑compatible clients, you must append `/openai/v1` to this URL.

## Route models to client applications

Configure your third-party applications to use TensorZero.

### Determine your model name string

Construct the correct model name using the following prefixes based on the resource you want to call:

| Resource type | TOML definition | Required string format | Example |
| :--- | :--- | :--- | :--- |
| **Function** | `[functions.<alias>]` | `tensorzero::function_name::<alias>` | `tensorzero::function_name::general_chat` |
| **Model** | `[models.<alias>]` | `tensorzero::model_name::<alias>` | `tensorzero::model_name::qwen3_5_9b` |
| **Embedding** | `[embedding_models.<alias>]` | `tensorzero::embedding_model_name::<alias>` | `tensorzero::embedding_model_name::nomic_embed` |

:::tip
- Do not use dots or colons in your alias names. For example, use `qwen3_5_9b`, not `qwen3.5:9b`.
- If the model name does not work, prepend `openai/` to satisfy the LiteLLM framework and try again. For example, use `openai/tensorzero::embedding_model_name::nomic_embed`.
:::

### Connect your clients

<Tabs>
<template #OpenCode>

1. In OpenCode, click <i class="material-symbols-outlined">settings</i> in the bottom-left corner.
   ![Open OpenCode settings](/images/manual/use-cases/opencode-settings.png#bordered)

2. Select **Providers**, then scroll down and select **Connect** next to **Custom provider**.
   ![Select custom provider](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. Enter the following details, and then click **Submit**.
   - **Provider ID**: A unique identifier for the model provider. For example, `olares-ollama-tensorzero`.
   - **Display name**: The name shown in the provider list. For example, `Olares TensorZero`.
    - **Base URL**: Enter your TensorZero endpoint URL ending with `/openai/v1`. For example, `https://ea581361.laresprime.olares.com/openai/v1`.
    - **API key**: Enter any text. This field cannot be empty.
    - **Models**:
        - **Model ID**: Enter the exact function string, `tensorzero::function_name::general_chat`.
        - **Display Name**: Enter a descriptive name for easy identification in the interface, such as `TensorZero Qwen`.

    ![TensorZero config in OpenCode](/images/manual/use-cases/tensorzero-opencode.png#bordered){width=70%}     

4. Refresh OpenCode, go to **Settings** > **Models**, and then locate your custom provider **Olares TensorZero**.
5. Verify the model you added is enabled.

    ![TensorZero enabled in OpenCode](/images/manual/use-cases/tensorzero-opencode-enable.png#bordered)   

6. Start a new session, and select the TensorZero-managed model to begin a chat.

    ![TensorZero chat in OpenCode](/images/manual/use-cases/tensorzero-opencode-chat.png#bordered)

7. Open TensorZero, and check the observability data. For example, on the **Inferences** page, each request you send appears as a log entry, which confirms that TensorZero routes the traffic successfully.

    ![TensorZero Inferences page](/images/manual/use-cases/tensorzero-inferences.png#bordered)

8. Select an entry to view the details.

    ![TensorZero Inferences entry details](/images/manual/use-cases/tensorzero-inferences-details.png#bordered)
</template>
<template #AgentZero>

1. Open Agent Zero, and then go to **Settings** > **Agent Settings**.
2. Click **Chat Model**, configure as follows, and then click **Save**.

    - **Chat model provider**: Select **Other OpenAI compatible**.
    - **Chat model name**: Enter `tensorzero::function_name::general_chat`.
    - **Chat model API base URL**: Enter your TensorZero endpoint URL ending with `/openai/v1`. For example, `https://ea581361.laresprime.olares.com/openai/v1`.
    - **API key**: Enter any text. This field cannot be empty.

    ![TensorZero config in AgentZero](/images/manual/use-cases/tensorzero-agentzero.png#bordered)

3. Click **Embedding Model**, configure as follows, and then click **Save**.

    - **Embedding model provider**: Select **Other OpenAI compatible**.
    - **Embedding model name**: Enter `openai/tensorzero::embedding_model_name::nomic_embed`.

        :::tip
        For applications that use the LiteLLM framework, such as the embedding feature in AgentZero, you must include the `openai/` prefix in the model name.
        :::

    - **API key**: Enter any text. This field cannot be empty.
    - **Embedding model API base URL**: Enter your TensorZero endpoint URL ending with `/openai/v1`. For example, `https://ea581361.laresprime.olares.com/openai/v1`.

    ![TensorZero config in AgentZero, embedding model config](/images/manual/use-cases/tensorzero-agentzero-embed.png#bordered)    

4. Start a new chat. 

    ![TensorZero chat in AgentZero](/images/manual/use-cases/tensorzero-agentzero-chat.png#bordered)

5. To test the embedding model's memory effect, tell the agent a specific fact to remember, and then ask it to recall the fact.

    ![TensorZero memory verify in AgentZero](/images/manual/use-cases/tensorzero-agentzero-memory.png#bordered)

6. Open TensorZero, and check the observability data. For example, on the **Inferences** page, each request you send appears as a log entry, which confirms that TensorZero routes the traffic successfully. 

    ![TensorZero inferences page](/images/manual/use-cases/tensorzero-agentzero-inferences.png#bordered)

7. Select an entry to view the details.

    ![TensorZero chat in AgentZero inference details](/images/manual/use-cases/tensorzero-agentzero-inference-details.png#bordered)

</template>
</Tabs>

## Access the built-in MCP server

TensorZero includes a built-in Model Context Protocol (MCP) server located at the `/mcp` endpoint. This feature allows your AI agent to look up the performance data.

For example, you can ask your agent to retrieve the average response time for `general_chat` today, and the agent will use the MCP connection to read the logs and report the data back to you.

The following example demonstrates how to configure OpenCode to access this MCP tool.

1. Open Files, and then go to **Data** > **opencode** > **.config** > **opencode**.
2. Double-click `opencode.json`, and then click <i class="material-symbols-outlined">edit_square</i>.
3. Add the following MCP configuration block. Ensure you replace `<tensorzero-endpoint>` with your actual TensorZero endpoint URL.

    ```json
    {
    "mcp": {
        "tensorzero": {
        "type": "remote",
        "url": "<tensorzero-endpoint>/mcp",
        "enabled": true
        }
    }
    }
    ```
4. Click <i class="material-symbols-outlined">save</i>.
5. Restart the OpenCode app to apply the changes. In the upper-right corner, on the **MCP** tab, verify that **tensorzero** shows up as enabled.

    ![TensorZero MCP enabled in OpenCode](/images/manual/use-cases/tensorzero-mcp.png#bordered){width=50%}

6. Instruct your AI agent directly in the chat to explicitly use the tool. For example, enter `Use the TensorZero MCP tool to analyze the latest inference logs`.

    ![OpenCode MCP use](/images/manual/use-cases/tensorzero-mcp-opencode.png#bordered)

## FAQs

### What is the difference between connecting to a model versus a function?

In TensorZero, both models and functions allow your applications to communicate with the AI, but connecting your applications to a function is highly recommended.

- **Model (`tensorzero::model_name::...`)**: This represents the raw AI engine. While you can connect your client applications directly to a model, doing so bypasses TensorZero's advanced monitoring features.
- **Function (`tensorzero::function_name::...`)**: This represents the specific task your application is performing, such as `coding_assistant` or `text_summarizer`. Connecting through a function can use TensorZero's detailed observability and statistical tracking. It also allows you to link multiple different functions to the same underlying model, helping you track and optimize each specific task separately.

### Error: model field must start with `tensorzero::function_name::...`

**Why it happens**: You entered a raw model name like `qwen3.5:9b` or an incorrect format in your client’s model field.

**How to fix**: Always use one of these three exact formats, depending on what you want to connect:

| You want to call | Format | Example |
| :--- | :--- | :--- |
| A function | `tensorzero::function_name::<alias>` | `tensorzero::function_name::general_chat` |
| A model directly | `tensorzero::model_name::<alias>` | `tensorzero::model_name::qwen3_5_9b` |
| An embedding model | `tensorzero::embedding_model_name::<alias>` | `tensorzero::embedding_model_name::nomic_embed` |

### Error: `litellm.BadRequestError: LLM Provider NOT provided`

**Why it happens**: This error occurs in applications that rely on the LiteLLM framework, such as the embedding feature in AgentZero. These specific applications do not automatically recognize the standard TensorZero model string. They require an explicit provider prefix to understand how to format the connection.

**How to fix**:
Review the error message details to identify exactly which model is failing. Open your application settings and add `openai/` to the very beginning of that model name.

For example, if the error mentions `model=tensorzero::embedding_model_name::nomic_embed`, you must change your embedding model name or ID to `openai/tensorzero::embedding_model_name::nomic_embed`. Save your settings and try the request again.

### TensorZero fails to start after I edit the configuration file

**Why it happens**: The TOML format is broken, which is usually caused by a missing empty line between sections or an invalid character in an alias name.

**How to fix**:
1. Open Control Hub, go to **tensorzero-{username}** > **Deployments** > **tensorzero** > **Pods**, and then click the tensorzero pod.
2. In the **Containers** section, locate **gateway**, and then click <i class="material-symbols-outlined">article</i> next to it.

    ![Container logs](/images/manual/use-cases/tensorzero-container-logs.png#bordered)

3. Look for the following common errors:

    - `Failed to parse tensorzero.toml`: Syntax error. Ensure you have exactly one empty line between every section block (`# models`, `# functions`, `# embedding_models`). If you deleted the empty lines when pasting the code, the application will fail to start.
    - `unknown field`: Incorrect setting name, such as dots or colons in aliases. Use underscores, like `qwen3_5_9b`, not `qwen3.5:9b`.
    - `provider...not found`: The provider name in your `routing = ["name"]` line does not match the block defined immediately below it `[models.alias.providers.name]`. For example, if you write `routing = ["ollama"]`, you must have a matching `[models.xxx.providers.ollama]` block.

4. After fixing the syntax, restart the TensorZero pod.

### My configuration changes not showing up in TensorZero UI

**Why it happens**: The UI cache, the gateway did not reload the configuration, or the restart failed.

**How to fix**:

Try the following methods:
- Hard refresh your browser by pressing Ctrl+Shift+R or Cmd+Shift+R to clear the browser cache.
- Check the `gateway` container logs for `Starting gateway server...`. If you see migration messages, wait another 30 seconds.
- Restart the TensorZero pod. 

### Some pages mentioned in the official docs (Autopilot, Config Editor) are missing

**Why it happens**: Those are advanced components that are not included in the default Olares deployment. Olares provides the core gateway, UI, and observability stack.

**How to fix**: If you need those features, refer to the [TensorZero official documentation](https://www.tensorzero.com/docs) for self‑hosting the additional services.

## Learn more

- [Set up Bifrost as an AI model gateway](bifrost.md)
- [Use LiteLLM as a unified AI model gateway](litellm.md)
