---
outline: [2, 4]
description: Learn how to install LobeChat on Olares and integrate it with Ollama to build and enhance your local custom AI assistants.
---

# Build your local AI assistant with LobeChat

LobeChat is an open-source, modern AI chat framework that supports file uploads, knowledge bases, and multi-modal interactions, ensuring a secure local chat experience. It also supports integration with Ollama, which allows you to use the large language models (LLMs) provided by Ollama to enhance your chat applications within LobeChat easily. Olares streamlines and simplifies the deployment of both, allowing you to skip complex manual environment configurations.

This guide covers the installation, configuration, and practical usage of these tools to create your personalized AI assistants.

## Learning objectives

By the end of this guide, you are able to:
- Install LobeChat directly from the Olares Market.
- Configure LobeChat to communicate with your local Ollama instance.
- Use LobeChat for specific scenarios such as content writing and coding.

## Before you begin

Ensure that Ollama is installed and running. For more information, see [Download and run local AI models via Ollama](ollama.md).

## Install LobeChat

Install LobeChat and related dependencies.

1. From the Olares Market, search for "LobeChat".

   ![Search for LobeChat from Market](/images/manual/use-cases/find-lobechat.png#bordered)

2. Click **Get**, and then click **Install**.
3. The system automatically detects and prompts you to install necessary dependencies if they are not already installed, such as Ollama. Allow these to install, and then wait for the installation to finish.

## Configure the connection

After the installation is completed, you must connect LobeChat to Ollama to make the chat interface work.

1. Open **LobeChat** from Launchpad.
2. Click the LobeChat icon in the upper-left corner, and then go to **Settings** > **Language Model**.

   ![LobeChat settings menu](/images/manual/use-cases/lobechat-settings-menu.png#bordered)

3. In the **Ollama** section, find the **Interface proxy address** field, and then enter your local Ollama address. 

   ![Interface proxy address connection](/images/manual/use-cases/lobechat-connection-setting.png#bordered)

4. (Optional) To obtain your local Ollama host address, go to Olares **Settings** > **Application** > **Ollama** > **Ollama API** > **Set up endpoint**, and then copy the endpoint address.

   ![Obtain Ollama host address from Olares Settings](/images/manual/use-cases/obtain-ollama-hosturl.png#bordered){width=60%} 

5. Click **Check** under the Ollama model list to verify the connection. A **Check Passed** message indicates that the proxy address is correct.

   ![Ollama connection successful](/images/manual/use-cases/ollama-model-checkpass.png#bordered)

## Install language models

After connecting to Ollama, LobeChat lists supported models, including models that are not installed yet. When you choose a model, LobeChat checks local storage. If the model is not installed, LobeChat prompts you to download it. After the download finishes, the model is ready to use.

You can install these models directly through the LobeChat user interface (UI) or via the Ollama command line (CLI).

### Install via LobeChat UI

When you are chatting with a selected language model, if the language model is not installed, you are prompted right in the chat to download and install it.

![Install language model via LobeChat UI](/images/manual/use-cases/download-in-lobechat.png#bordered)

### Install via Ollama CLI

1. Check the [Ollama Library](https://ollama.com/library) to explore available models.
2. To download a model, use the following command:
```bash
ollama pull [model]
```
For more information, see [Download and run local AI models via Ollama](ollama.md).

## Use scenarios

LobeChat allows you to create specialized assistants to handle specific tasks by leveraging various language models and combining them with functional plug-ins.

- Flexible model switching: You can switch language models instantly within the same chat to achieve the best results. For example, if you are not satisfied with a response, you can select a different model from the list to leverage their unique strengths.
- Plug-in extensions: You can also install plug-ins to extend and enhance the capabilities of your assistant.

   :::info
   To install plug-ins, ensure that you select a model compatible with Function Calling. Look for <i class="material-symbols-outlined">brick</i> next to the model name, which indicates the model supports function calls.
   :::

### Polish content and visualize ideas

Create a specialized assistant to help you refine text and generate images based on descriptions.

#### 1. Add a new assistant

From the left navigation pane, click **New Assistant**. A default conversational agent is ready for customization.

#### 2. Configure the assistant

To help you get started, this guide demonstrates only some typical configurations in LobeChat.

1. Click **Open Chat Settings**.

   ![Open Chat Settings](/images/manual/use-cases/open-chat-settings.png#bordered)

2. Customize assistant identity.

   a. On the **Assistant Info** tab, set the avatar, name, and description. For example, name it **Writing Bot**.
   
   b. Click **Update Assistant Information**.

   ![LobeChat session settings](/images/manual/use-cases/lobechat-session-settings.png#bordered)   

3. Define assistant role.

   a. Click the **Role Configuration** tab.
   
   b. Click **Edit**.
   
   c. Enter your prompt for this specific role to define its behavior. For example,

      ```
      You are a creative editor. When I provide text, review it for clarity 
      and tone. When I describe a scene, use the drawing plug-in to generate 
      an image based on my description.
      ``` 

   d. Click **OK**.

   e. Close the **Session Settings** page. You return to the chat window.

#### 3. Select the language model and plug-in

1. In the basic interaction area, click <i class="material-symbols-outlined">neurology</i> to select a language model. For example, select **Qwen 2.5 7B**, because:

   - It excels at various NLP tasks such as contextual understanding and content writing.
   - It is compatible with functional calling, so you can install LobeChat plug-ins for enhanced capabilities.

2. Install LobeChat plug-ins as needed to enhance your assistant's abilities. For example, install a plug-in for image creation.

   a. Hover over the plug-in icon and click **Plugin Store**.

      ![Install LobeChat plug-in](/images/manual/use-cases/lobechat-plugin-install.png#bordered)

   b. On the **LobeHub Plugins** tab, search for `image`, select **Pollinate drawing** for example, and then click **Install**.

      ![Install LobeChat plug-in Pollinate Drawing](/images/manual/use-cases/install-pollinate-drawing.png#bordered)

#### 4. Interact with the assistant

1. Enter and send your draft content to get a refined version.
2. Hover over the plug-in icon to ensure that **Pollinate drawing** is enabled, and then ask the assistant to create a cover image for the content. It uses the enabled plug-in to generate an image.

   ![LobeChat plug-in enabled](/images/manual/use-cases/lobechat-plugin-enable.png#bordered)

3. Brainstorm and iterate with the language model to get your ideal content textually and visually.

#### 5. Pin the assistant

If you are satisfied with the performance of the assistant and want to access it quickly later on, hover over it in the sidebar, click <i class="material-symbols-outlined">more_vert</i>, and then click **Pin** to keep it accessible at the top of your list.

![Pin LobeChat assistant](../public/images/manual/use-cases/pin-writing-bot.png#bordered)

### Coding assistant

Create a specialized assistant to help you write efficient code and act as a dedicated pair programmer.

#### 1. Add a new assistant

From the left navigation pane, click **New Assistant**. A default conversational agent is ready for customization.

#### 2. Configure the assistant

To help you get started, this guide demonstrates only some typical configurations in LobeChat.

1. Click **Open Chat Settings**.

   ![Open Chat Settings](/images/manual/use-cases/open-chat-settings.png#bordered)

2. Customize assistant identity.

   a. On the **Assistant Info** tab, set the avatar, name, and description. For example, name it **Dev Bot**.
   
   b. Click **Update Assistant Information**.

   ![LobeChat session settings](/images/manual/use-cases/lobechat-session-settings.png#bordered)    

3. Define assistant role.

   a. Click the **Role Configuration** tab.
   
   b. Click **Edit**.
   
   c. Enter your prompt for this specific role to define its behavior. For example,

      ```
      You are an expert developer. When I describe a task or requirement, 
      generate clean, efficient, and well-commented code to solve it.
      ``` 
   d. Click **OK**. You return to the chat window.

#### 3. Interact with the assistant

1. In the basic interaction area, select a language model. For example, select **deepseek-coder-v2** which excels at coding use cases such as code generation and long text understanding.
2. Describe a data generation task and send to the chat. For example,
   
   ```
   Write a Python script to generate a CSV file named employees.csv with
   20 rows of mock data. Columns should include: ID, Name, Department, 
   and Salary. Use the random library to generate varied data.
   ```
3. The assistant processes your request and generates a standalone Python script with explanation.

   ```
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
4. Run the generated code to verify.

   a. Copy the generated Python code block and save it as `generate_data.py`.

   b. Open the Terminal, navigate to the folder, and run
`python3 generate_data.py`.

   c. Check your current folder. You should see a new file named `employees.csv`. Open it to verify the generated mock data.

      ![Dev bot result verification](/images/manual/use-cases/dev-bot-result.png#bordered)  

#### 4. Pin the assistant

If you are satisfied with the assistant's performance, hover over it in the sidebar, click <i class="material-symbols-outlined">more_vert</i>, and then click **Pin** to keep it accessible at the top of your list.

![Pin dev bot](/images/manual/use-cases/pin-dev-bot.png#bordered)

### Real-time news analyst

Build an assistant that keeps you updated with the latest technology trends. By using the Website Crawler plug-in, this assistant can read live news sites and provide instant summaries of what's happening right now.

#### 1. Add a new assistant

From the left navigation pane, click **New Assistant**. A default conversational agent is ready for customization.

#### 2. Configure the assistant

To help you get started, this guide demonstrates only some typical configurations in LobeChat.

1. Click **Open Chat Settings**.

   ![Open Chat Settings](/images/manual/use-cases/open-chat-settings.png#bordered)

2. Customize assistant identity.

   a. On the **Assistant Info** tab, set the avatar, name, and description. For example, name it **Daily Tech Digest**.
   
   b. Click **Update Assistant Information**.

   ![LobeChat session settings](/images/manual/use-cases/lobechat-session-settings.png#bordered)   

3. Define assistant role.

   a. Click the **Role Configuration** tab.
   
   b. Click **Edit**.
   
   c. Enter your prompt for this specific role to define its behavior. For example,

      ```
      You are a tech news reporter. When I send you a news site URL, 
      read the headlines and summarize the latest top five stories for me.
      Limit the list to five.
      ``` 

   d. Click **OK**.

   e. Close the **Session Settings** page. You return to the chat window.

#### 3. Select the language model and plug-in

1. In the basic interaction area, select a language model. For example, **Qwen 2.5 7B** , because:

   - It excels at various NLP tasks such as contextual understanding and content writing.
   - It is compatible with functional calling, so you can install LobeChat plug-ins for enhanced capabilities.

   ![Select language model](/images/manual/use-cases/select-qwen.png#bordered) 

3. Install a web-access plug-in to allow the assistant to access live web pages and analyze real-time content from any URLs you provide. For example, **Website Crawler**. 

   :::tip How Website Crawler works (Real-time vs. Offline)
   Standard local AI models are offline and rely on pre-trained data from the past. The Website Crawler plug-in, specifically the getWebsiteContent function, acts as a bridge to the live internet. When you provide a URL, the plug-in instantly accesses the web page in real time via an API, fetches the current content, and feeds it to the AI. This ensures that the AI model is accessing the latest live web content rather than using the old memory.
   :::

   a. Hover over the plug-in icon and click **Plugin Store**.

   ![Install LobeChat plug-in](/images/manual/use-cases/select-plugin.png#bordered)

   b. On the **LobeHub Plugins** tab, search for `website`, select **Website Crawler** for example, and then click **Install**.

   ![Install LobeChat plug-in Website Crawler](/images/manual/use-cases/install-website-crawler.png#bordered)

#### 4. Interact with the assistant

1. In the basic interaction area, hover over the plug-in icon to ensure that the **Website Crawler** plug-in is enabled.

   ![LobeChat crawler plug-in enabled](/images/manual/use-cases/website-crawler-enabled.png#bordered)

2. Send the URL address to the chat. For example, `https://github.com/trending`.
3. Paste and send the URL to the chat. The assistant lists specific news stories with summaries.

#### 5. Pin the assistant

If you find this useful for daily updates, hover over it in the sidebar, click <i class="material-symbols-outlined">more_vert</i>, and then click **Pin** to keep it accessible at the top of your list.

![Pin chat assistant to list for easy access](/images/manual/use-cases/pin-daily-tech-digest.png#bordered)
