---
outline: [2, 4]
description: Learn how to install LobeChat on Olares and integrate it with Ollama to build and enhance your local custom AI assistants.
---

# Build your local AI assistant with LobeChat

LobeChat is an open-source, modern AI chat framework that supports file uploads, knowledge bases, and multi-modal interactions, ensuring a secure local chat experience. 

Ollama is a lightweight platform for running open-source AI large language models (LLMs) locally, including Quen 2.5, LIama 3.1, DeepSeek V2, and more. 

LobeChat supports integration with Ollama, which allows you to use the LLMs provided by Ollama to enhance your chat applications within LobeChat easily.

This guide covers the installation, configuration, and practical usage of these tools to create your personalized AI assistant.

## Learning objectives

By the end of this guide, you will be able to:
- Understand the relationship between LobeChat and Ollama.
- Install LobeChat and Ollama directly from the Olares Market.
- Configure LobeChat to communicate with your local Ollama instance.
- Use LobeChat for specific scenarios such as content writing and coding.

## Before you begin

### Prerequisites

- Hardware: An Olares-compatible device with sufficient RAM and storage to run local models.
- System: Olares OS is installed and running.
- Network: An active internet connection required for installation and operation.

### Concept note

By combining LobeChat's intuitive frontend interface with Ollama's backend capabilities, you can turn your Olares device into a powerful, private AI workstation. 

- LobeChat: The frontend user interface that allows you to interact with various language models via a chat window.
- Ollama: The backend engine that runs the language models locally, providing computation power and API interfaces.
- Olares: The operating system (OS) that streamlines and simplifies the deployment of both, allowing you to skip complex manual environment configurations.

## Install LobeChat

Install LobeChat and related dependencies:
1. From the Olares Market, find **LobeChat**.
2. Click **Get**, and then click **Install**.
3. The system automatically detects and prompts you to install necessary dependencies if they are not already installed, such as Ollama. Allow these to install, and then wait for the installation to finish.

## Configure the connection

After the installation is completed, you must connect LobeChat to Ollama to make the chat interface work:
1. From the Olares launchpad, click **LobeChat** to open the application.
2. In LobeChat, go to **Settings** > **Language Model**.
3. In the **Ollama** section, find the **Interface proxy address** field, and then enter your local Ollama address https://39975b9a1.*UserID*.olares.com. 

   :::info
   Replace *UserID* with the Olares Admin's local name. For example, `https://39975b9a1.alexmiles.olares.com`.
   :::

   ![Interface proxy address connection](../public/images/manual/use-cases/lobechat-connection-setting.png#bordered)

5. Click **Check** to verify the connection. A **Check Passed** message indicates that the proxy address is correct.

## Install language models

You can install language models directly through the LobeChat user interface (UI) or via the Ollama command line (CLI).

### Method A: Install via LobeChat UI

1. In the Chat window, click <i class="material-symbols-outlined">neurology</i> to select the target language model.
2. Type and send a message in the chat.
3. If the language model is not installed, you are prompted right in the chat to download and install it.

    ![Install language model via LobeChat UI](../public/images/manual/use-cases/download-in-lobechat.png#bordered)

4. When the installation is completed, you can chat with the newly installed language model.

### Method B: Install via Ollama CLI

1. Visit the Ollama Library to find the target model. For example, gemma2.
2. Select the parameters suitable for your hardware specifications.
3. Copy the run command. For example, `ollama run gemma2`.
4. Open the Ollama terminal from the Olares launchpad.
5. Paste the command and press Enter.
6. To verify the installation, run the `ollama list` command.

For more information, see [Download and run local AI models via Ollama](ollama.md).

## How to use

LobeChat allows you to create specialized assistants to handle specific tasks by combining various language models with functional plugins.

- Flexible model switching: You can switch language models instantly within the same chat to achieve the best results. For example, if you are not satisfied with a response, you can select a different model from the list to leverage their unique strengths.
- Plugin extensions: You can also install plugins to extend and enhance the capabilities of your assistant.

   :::info
   To install plugins, ensure that you select a model compatible with Function Calling. Look for <i class="material-symbols-outlined">brick</i> next to the model name, which indicates the model supports function calls.
   :::

### Scenario 1: Polish content and visualize ideas

Create a specialized assistant to help you refine text and generate images based on descriptions.

#### 1. Add a new assistant

From the left navigation pane, click **New Assistant**. A default conversational agent is ready for customization.

#### 2. Configure the assistant

To help you get started, this guide demonstrates only some typical configurations in LobeChat.

1. Click **Open Chat Settings**.

   ![Open Chat Settings](../public/images/manual/use-cases/open-chat-settings.png#bordered)

2. Customize assitant identity:

   a. On the **Assistant Info** tab, set the avatar, name, and description. For example, name it **Writing bot**.
   
   b. Click **Update Assistant Information**.

   ![LobeChat session settings](../public/images/manual/use-cases/lobechat-session-settings.png#bordered)   

3. Define assistant role:

   a. Click the **Role Configuration** tab.
   
   b. Click **Edit**.
   
   c. Enter your prompt for this specific role to define its behavior. For example,

      ```
      You are a creative editor. When I provide text, review it for clarity 
      and tone. When I describe a scene, use the drawing plugin to generate 
      an image based on my description.
      ``` 

   d. Click **OK**.

#### 3. Select the language model and plugin

1. Return to the chat window.
   
2. In the basic interaction area, click <i class="material-symbols-outlined">neurology</i> to select a language model. For example, select **Qwen 2.5 7B** , because:

   - It excels at various NLP tasks such as contextual understanding and content writing.
   - It is compatible with functional calling, so I can install LobeChat plugin for enhanced capabilities.

3. Install LobeChat plugins as needed to enhance your assistant's abilities. For example, install the **Pollinate drawing** plugin for image creation.

   ![Install LobeChat plugin](../public/images/manual/use-cases/lobechat-plugin-install.png#bordered)

#### 4.Interact with the assistant

1. Enter and send your draft content to get a refined version.
2. Ensure that the **Pollinate drawing** plugin is enabled, and then ask the assistant to create a cover image for the content. It will use the enabled plugin to generate an image.

   ![LobeChat plugin enabled](../public/images/manual/use-cases/lobechat-plugin-enable-2.png#bordered)

3. Iterate to get your ideal content textually and visually.

#### 5. Pin the assistant

If you are satisfied with the performance of the assistant and want to access it quickly later on, click **Pin** to keep it at the top of your assistant list.

![Pin LobeChat assistant](../public/images/manual/use-cases/lobechat-pin.png#bordered)
<!--this senario pending the text to speech plugin work
### Scenario 2: Training and eduction

#### Task: Language learning

#### Selected model: Llama 3.1

LIAMA (Language Intelligence Model for AI Applications) is a large language model that specializes in multilingual support, natural language understanding, domain-specific knowledge, content creation, summarization, and translation.

#### Procedure

1. From the left navigation pane, click **New Assistant**.
2. From the top language model list, switch the active model to **Llama 3.1**.
2. In the **Role Setting** area, click <i class="material-symbols-outlined">edit_square</i> to enter your prompt for this specific role. For example, 

   c. Customize your own opening message or question displayed when the conversation starts by going to **Settings > Default Assistant > Opening Settings**. You can introduce this chat assistant's features or facilitate your conversations by providing guiding questions. 
   d. Go back to the chat and engage in the conversation to practice French syntax and vocabulary without language barriers.
-->

### Scenario 2: Research technical documentations and generate code in one workflow

Boost your programming efficiency by turning your assistant into a research-capable pair programmer. 

#### 1. Add a new assistant

From the left navigation pane, click **New Assistant**. A default conversational agent is ready for customization.

#### 2. Configure the assistant

To help you get started, this guide demonstrates only some typical configurations in LobeChat.

1. Click **Open Chat Settings**.

   ![Open Chat Settings](../public/images/manual/use-cases/open-chat-settings.png#bordered)

2. Customize assitant identity:

   a. On the **Assistant Info** tab, set the avatar, name, and description. For example, name it **Dev bot**.
   
   b. Click **Update Assistant Information**.

   ![LobeChat session settings](../public/images/manual/use-cases/lobechat-session-settings.png#bordered)    

3. Define assistant role:

   a. Click the **Role Configuration** tab.
   
   b. Click **Edit**.
   
   c. Enter your prompt for this specific role to define its behavior. For example,

      ```
      You are a senior developer. When I send a URL, use the crawler to read it. 
      Summarize technical details and generate code based on my requests.
      ``` 
   d. Click **OK**.

#### 3. Select the language model and plugin

1. Return to the chat window.
   
2. In the basic interaction area, click <i class="material-symbols-outlined">neurology</i> to select a language model. 

   For example, select **CodeQwen 1.5 7B** which excels at coding use cases such as code generation and long text understanding.

3. Install a web-access plugin to allow the assistant to access live web pages and analyze real-time content from any URLs you provide. For example, **Website Crawler**. 

   :::info
   Standard local AI models are offline and rely on pre-trained data from the past. The website crawler plugin acts as a bridge to the live internet. When you provide a URL, the plugin instantly accesses the web page in real time via an API, fetches the current content, and feeds it to the AI. This ensures that the AI model is accessing the live web content rather than using the old memory. 
   :::

#### 4. Interact with the assistant

1. Ensure that the **Website Crawler** plugin is enabled.

   ![LobeChat crawler plugin enabled](../public/images/manual/use-cases/lobe-chat-plugin-enable-2.png#bordered)


2. Paste a URL to a technical documentation page and ask the assistant to analyze this page and summarize the API endpoints. For example,
   
   ```
   Analyze this page and summarize the API endpoints:
   https://jsonplaceholder.typicode.com/guide/.
   ```

3. Based on the summary, ask the assistant to complete your requirement. For example,

   ```
   Write a Python script to fetch data from the /users endpoint and 
   save it as a CSV.
   ```

   The assistant reads the live documentation and generates accurate, copy-paste-ready code.

4. Run the generated code.

   a. Copy the Python code block from LobeChat and save it as a `fetch_data.py` file.

   b. Open the terminal and run the script `python3 fetch_users.py`.

   c. Check your current folder to confirm that a new .csv file containing the data has been created.

### Scenario 3: Tech news analyst

Build an assistant that keeps you updated with the latest technology trends. By using the Website Crawler plugin, this assistant can read live news sites and provide instant summaries of what's happening right now.

#### 1. Add a new assistant

From the left navigation pane, click **New Assistant**. A default conversational agent is ready for customization.

#### 2. Configure the assistant

To help you get started, this guide demonstrates only some typical configurations in LobeChat.

1. Click **Open Chat Settings**.

   ![Open Chat Settings](../public/images/manual/use-cases/open-chat-settings.png#bordered)

2. Customize assitant identity:

   a. On the **Assistant Info** tab, set the avatar, name, and description. For example, name it **Daily Tech Digest**.
   
   b. Click **Update Assistant Information**.

   ![LobeChat session settings](../public/images/manual/use-cases/lobechat-session-settings.png#bordered)   

3. Define assistant role:

   a. Click the **Role Configuration** tab.
   
   b. Click **Edit**.
   
   c. Enter your prompt for this specific role to define its behavior. For example,

      ```
      You are a tech news reporter. When I send you a news site URL, 
      read the headlines and summarize the latest top five stories for me.
      Limit the list to five.
      ``` 

   d. Click **OK**.

#### 3. Select the language model and plugin

1. Return to the chat window.
   
2. In the basic interaction area, click <i class="material-symbols-outlined">neurology</i> to select a language model. For example, select **Qwen 2.5 7B** , because:

   - It excels at various NLP tasks such as contextual understanding and content writing.
   - It is compatible with functional calling, so I can install LobeChat plugin for enhanced capabilities.

3. Install a web-access plugin to allow the assistant to access live web pages and analyze real-time content from any URLs you provide. For example, **Website Crawler**. 

   :::noteHow Website Crawler works (Real-time vs. Offline):
   Standard local AI models are offline and rely on pre-trained data from the past. The Website Crawler plugin, specifically the getWebsiteContent function, acts as a bridge to the live internet. When you provide a URL, the plugin instantly accesses the web page in real time via an API, fetches the current content, and feeds it to the AI. This ensures that the AI model is accessing the latest live web content rather than using the old memory.
   :::

   ![Install LobeChat plugin](../public/images/manual/use-cases/lobechat-plugin-install.png#bordered)

#### 4. Interact with the assistant

1. In the basic interaction area, check that the **Website Crawler** plugin is enabled.
2. Send the URL address to the chat. 
3. Paste and send the URL to the chat. The assistant will list five specific news stories with summaries.

#### 5. Pin the assistant

If you find this useful for daily updates, Pin it to your sidebar for quick access later.
