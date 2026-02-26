---
outline: [2, 4]
description: Learn how to install LobeChat on Olares and integrate it with Ollama to build and enhance your local custom AI assistants.
---

# Build your local AI assistant with LobeChat

LobeChat is an open‑source framework for building secure, local AI chat experiences. It supports file handling, knowledge bases, and multimodal inputs, and it supports Ollama to run and switch local LLMs.

Olares streamlines and simplifies the deployment of both, allowing you to skip complex manual environment configurations.

This guide covers the installation, configuration, and practical usage of these tools to create your personalized AI assistants.

## Learning objectives

By the end of this guide, you are able to:
- Configure LobeChat to communicate with your local Ollama instance.
- Use LobeChat for specific scenarios such as content writing and coding.

## Prerequisites

Before you begin, make sure:

- Ollama is installed and running in your Olares environment.
- At least one model is installed using Ollama. For more information, see [Ollama](ollama.md).

## Install LobeChat

1. From the Olares Market, search for "LobeChat".

   ![Search for LobeChat from Market](/images/manual/use-cases/find-lobechat.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Configure the connection

After the installation is completed, you must connect LobeChat to Ollama to make the chat interface work.

1. Open **LobeChat** from Launchpad.
2. Click the LobeChat icon in the upper-left corner, and then go to **Settings** > **Language Model**.

   ![LobeChat settings menu](/images/manual/use-cases/lobechat-settings-menu.png#bordered){width=75%} 

3. (Optional) To obtain your local Ollama host address, go to Olares **Settings** > **Application** > **Ollama** > **Ollama API** > **Set up endpoint**, and then copy the endpoint address.

   ![Obtain Ollama host address from Olares Settings](/images/manual/use-cases/obtain-ollama-hosturl.png#bordered){width=60%} 

4. In the **Ollama** section, find the **Interface proxy address** field, and then enter your local Ollama address. 

   ![Interface proxy address connection](/images/manual/use-cases/lobechat-connection-setting.png#bordered)

5. Click **Check** under the Ollama model list to verify the connection. A **Check Passed** message indicates that the proxy address is correct.

   ![Ollama connection successful](/images/manual/use-cases/ollama-model-checkpass.png#bordered)

## Install language models

By default, LobeChat lists supported models, including models that are not installed yet. 

You can click **Get Model List** to fetch available models from Ollama.

To download additional models, use one of the following options:
   - Download via LobeChat UI

      When you are chatting with a selected language model, if the language model is not installed, you are prompted right in the chat to download and install it.

      ![Install language model via LobeChat UI](/images/manual/use-cases/download-in-lobechat.png#bordered)

   - Download using the Ollama CLI. For more information, see [Download a model](ollama.md#download-a-model).

## Create an assistant

LobeChat allows you to create specialized assistants to handle specific tasks by leveraging various language models and combining them with functional plug-ins.

- **Flexible model switching**: You can switch language models instantly within the same chat to achieve the best results. For example, if you are not satisfied with a response, you can select a different model from the list to leverage their unique strengths.
- **Plug-in extensions**: You can also install plug-ins to extend and enhance the capabilities of your assistant.

   :::info
   To install plug-ins, ensure that you select a model compatible with Function Calling. Look for <i class="material-symbols-outlined">brick</i> next to the model name, which indicates the model supports function calls.
   :::

The following steps outline the standard workflow for creating and configuring any assistant in LobeChat. You can apply this procedure using specific settings provided in the [use scenarios](#use-scenarios) section.

1. Create a new assistant:
   - From the left navigation pane, click **New Assistant**. 
   - If you already have active chats, click <i class="material-symbols-outlined">add_comment</i> to create a new one.
2. Configure the assistant such as identity and role:

   a. Click **Open Chat Settings**.

      ![Open Chat Settings](/images/manual/use-cases/open-chat-settings.png#bordered)

   b. On the **Assistant Info** tab, set the avatar, name, and description, and then click **Update Assistant Information**.

      ![LobeChat session settings](/images/manual/use-cases/lobechat-session-settings.png#bordered)   

   c. On the **Role Configuration** tab, enter your prompt for this specific role to define its behavior, and then click **OK**.

   d. Close the **Session Settings** page to return to the chat window.

3. Select the language model from the basic interaction area.
   
   ![Select language model](/images/manual/use-cases/select-qwen.png#bordered) 
   
4. (Optional) Install LobeChat plug-ins to enhance the assistant's capabilities:

   a. In the basic interaction area, hover over the plug-in icon and click **Plugin Store**.

      ![Install LobeChat plug-in](/images/manual/use-cases/lobechat-plugin-install.png#bordered)

   b. On the **LobeHub Plugins** tab, search for the target plug-in, and then click **Install**.

5. Interact with the assistant.
6. (Optional) Pin for quick access:

   If you are satisfied with the assistant's performance, hover over the assistant in the sidebar, click <i class="material-symbols-outlined">more_vert</i>, and then click **Pin** to keep it accessible at the top of your list.

## Use scenarios

The following scenarios provide some practical examples for your daily tasks. Apply these specific settings during the [general creation procedure](#create-an-assistant) to build specialized assistants tailored to your workflow.

### Polish content and visualize ideas

Create a specialized assistant to help you refine text and generate images based on descriptions.

#### Configurations

- **Name**: `Writing Bot`
- **Role prompt**:

   ```
   You are a creative editor. When I provide text, review it for clarity 
   and tone. When I describe a scene, use the drawing plug-in to generate 
   an image based on my description.
   ``` 

- **Language model**: `qwen2.5:7b`

   :::info
   `qwen2.5:7b` excels at various NLP tasks such as contextual understanding and content writing. It is also compatible with functional calling, so you can install LobeChat plug-ins for enhanced capabilities.
   :::

- **LobeChat plug-in**: "Pollinate drawing", which is used to create images based on description

   ![Install LobeChat plug-in Pollinate Drawing](/images/manual/use-cases/install-pollinate-drawing.png#bordered)

#### Interaction

1. Enter and send your draft content to get a refined version.
2. Hover over the plug-in icon to ensure that **Pollinate drawing** is enabled, and then ask the assistant to create a cover image for the content.
3. Brainstorm and iterate with the language model to get your ideal content textually and visually.

### Coding assistant

Create a specialized assistant to help you write efficient code and act as a dedicated pair programmer.

#### Configurations

- **Name**: `Dev Bot`
- **Role prompt**: 
   ```
   You are an expert developer. When I describe a task or requirement, 
   generate clean, efficient, and well-commented code to solve it.
   ``` 
- **Language model**: `deepseek-coder-v2`

   :::info
   `deepseek-coder-v2` is good at coding use cases such as code generation and long text understanding.
   :::

#### Interaction

1. Describe a data generation task and send to the chat.
   ```
   Write a Python script to generate a CSV file named employees.csv with
   20 rows of mock data. Columns should include: ID, Name, Department, 
   and Salary. Use the random library to generate varied data.
   ```
2. The assistant processes your request and generates a standalone Python script with explanation.

   ```python
   import csv
   import random

   def generate_mock_data():
      departments = ['HR', 'Engineering', 'Marketing', 'Sales', 'Finance']
      filename = "employees.csv"

      print(f"Generating {filename}...")

      with open(filename, 'w', newline='', encoding='utf-8') as csvfile:
         fieldnames = ['ID', 'Name', 'Department', 'Salary']
         writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
         writer.writeheader()

         for i in range(1, 21):
               writer.writerow({
                  'ID': f'EMP{i:03d}',
                  'Name': f'Employee {i}',
                  'Department': random.choice(departments),
                  'Salary': random.randint(50000, 120000)
               })

      print(f"Successfully created {filename} with 20 records.")

   if __name__ == "__main__":
      generate_mock_data()
   ```
3. Run the generated code to verify.

   a. Copy the generated Python code block and save it as `generate_data.py`.

   b. Open the Terminal, navigate to the folder, and run the following command:

   ```python
   python3 generate_data.py
   ```

   c. Check your current folder. You should see a new file named `employees.csv`. Open it to verify the generated mock data.

      ![Dev bot result verification](/images/manual/use-cases/dev-bot-result.png#bordered)  

### Real-time news analyst

Build an assistant that keeps you updated with the latest technology trends. By using the Website Crawler plug-in, this assistant can read live news sites and provide instant summaries of what's happening right now.

#### Configurations

- **Name**: `Daily Tech Digest`
- **Role prompt**:

   ```
   You are a tech news reporter. When I send you a news site URL, 
   read the headlines and summarize the latest top five stories for me.
   Limit the list to five.
   ``` 
- **Language model**: `qwen2.5:7b`

   :::info
   `qwen2.5:7b` excels at various NLP tasks such as contextual understanding and content writing. It is also compatible with functional calling, so you can install LobeChat plug-ins for enhanced capabilities.
   :::
- **LobeChat plug-in**: "Website Crawler", which is used to access live web pages and analyze real-time content from provided URLs

   :::info How Website Crawler works (Real-time vs. Offline)
   Standard local AI models are offline and rely on pre-trained data from the past. The Website Crawler plug-in, specifically the getWebsiteContent function, acts as a bridge to the live internet.
   
   When you provide a URL, the plug-in instantly accesses the web page in real time via an API, fetches the current content, and feeds it to the AI. This ensures that the AI model is accessing the latest live web content rather than using the old memory.
   :::

   ![Install LobeChat plug-in Website Crawler](/images/manual/use-cases/install-website-crawler.png#bordered)

#### Interaction

1. In the basic interaction area, hover over the plug-in icon to ensure that the **Website Crawler** plug-in is enabled.
2. Send the URL address to the chat. For example, `https://github.com/trending`.
3. Paste and send the URL to the chat. The assistant lists specific news stories with summaries.

## FAQ

### Why did the connection check fail when I connected to Ollama?

This usually happens due to network or authentication settings. Follow these steps to resolve it:

1.  Open Settings, and ensure that the **Authentication level** for Ollama is set to **Internal**.
2.  Turn on the VPN in the LarePass desktop client, and confirm that the status changes to **Intranet** or **P2P**.
3.  Run the connection check again.

:::tip
If you are on the same local network as your device and the **Authentication level** for Ollama is set to **None**, you do not need to enable the LarePass VPN.
:::

For more information, see [Ensure network connectivity](ollama.md#ensure-network-connectivity).