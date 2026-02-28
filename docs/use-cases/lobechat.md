---
outline: [2, 4] 
description: Learn how to install LobeChat on Olares and integrate it with Ollama to build and enhance your local custom AI assistants.
head:
  - - meta
    - name: keywords
      content: Olares, LobeHub agent, LobeChat assistant, AI agent, AI agent team
---

# Build your local AI agent with LobeHub

LobeHub (previously LobeChat) is an open‑source framework for building secure, local AI chat experiences. It supports file handling, knowledge bases, and multimodal inputs, and it supports Ollama to run and switch local LLMs.

Olares streamlines and simplifies the deployment of both, allowing you to skip complex manual environment configurations.

This guide covers the installation, configuration, and practical usage of these tools to create your personalized AI agents.

## Learning objectives

- Configure LobeHub to communicate with your local Ollama instance.
- Create specialized agents tailored to specific tasks and equip them with specific Skills.
- Create an agent group to enable multiple agents to collaborate on complex workflows.
- Use LobeHub for practical scenarios such as content writing and coding.

## Prerequisites

Before you begin, make sure:

- Ollama is installed and running in your Olares environment.
- The models you want to use are downloaded using Ollama. This tutorial uses `llama3.1:8b` and `qwen2.5`. For more information, see [Download a model](ollama.md#download-a-model).

## Install LobeHub

1. From the Olares Market, search for "LobeChat".

   ![Search for LobeChat from Market](/images/manual/use-cases/find-lobechat1.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Sign in to LobeHub

1. Open **LobeChat** from Launchpad.
2. In the **Sign up or log in to your LobeHub account** field, enter your email address, and then click <i class="material-symbols-outlined">arrow_forward_ios</i>.
3. If you have not signed up before, you are prompted to create an account using this email. And then you are signed in to LobeHub directly. Click let's get started. Set up your user name, your interested areas, language preference, default model used by the agent (you can change later). Then you are ready to go.

      ![LobeHub welcome interface](/images/manual/use-cases/lobehub-start.png#bordered)

## Configure the connection

Connect LobeHub to Ollama to make the chat interface work.

1. From the left sidebar, click **AI Service Provider**, and then select **Ollama**.

      ![Configure Ollama in LobeHub](/images/manual/use-cases/lobehub-config-ollama.png#bordered)

2. In the **Interface proxy address** field, enter your local Ollama address.

   :::tip Obtain your local Ollama host address
   To obtain your local Ollama host address, go to Olares **Settings** > **Applications** > **Ollama**, click **Ollama API** under **Entraces** or **Shared entrances**, and then copy the endpoint address.

   ![Obtain Ollama host address from Olares Settings](/images/manual/use-cases/obtain-ollama-hosturl1.png#bordered){width=60%} 

3. Disable the **Use Client Request Mode** option.

   :::tip
   When you are running local models, do not enable the **Use Client Request Mode** option.
   :::
4. In **Model List**, click **Fetch models** to pull the list of installed models, and then enable the models you want to use.

   ![Fetch model list and enable models](/images/manual/use-cases/lobehub-fetch-enable-model.png#bordered){width=85%} 

5. In the **Connectivity Check** section, select the model you just enabled, and then click **Check** to verify the connection. When the model is large, it might take a while to load.

   ![Connectivity check](/images/manual/use-cases/lobehub-connectivity-check.png#bordered){width=85%} 

   The button turning to **Check Passed** indicates that the proxy address is correct. 

   ![Connectivity check success](/images/manual/use-cases/lobehub-checkpass.png#bordered){width=85%}    

6. Click the home icon at the upper-left corner to return to the main page.

   ![Return to homepage](/images/manual/use-cases/lobehub-return-home.png#bordered){width=45%} 

## Use Lobe AI

1. From the left sidebar, click **Lobe AI**.
   
   ![Click Lobe AI](/images/manual/use-cases/lobe-ai.png#bordered){width=85%} 

## Create an agent 

### Create an agent using Agent Builder

If you prefer not to configure everything manually, you can use Agent Builder. This is LobeHub's built-in assistant that helps you create specialized agents through conversations. Simply describe your needs, and it will automatically generate a complete agent configuration, including role settings, system prompts, and skill setups.

1. On the homepage, click **Create Agent** under the chat box.

   ![Create Agent button](/images/manual/use-cases/lobehub-create-agent.png#bordered){width=85%} 

2. In the chat box, describe the specific task you want the agent to handle. For exmaple,

   ```
   I need an agent to review my daily work items and summarize them.
   The summary should focus on the overall purpose of the tasks and
   highlight specific action items.
   ```
3. Select the language model. For exmaple, `llama3.1:8b`.
4. Press **Enter**. The profile page of the new agent opens, and you can see the Agent Builder immediately starts configuring your agent automatically.

   ![Agent builder](/images/manual/use-cases/lobehub-agent-builder.png#bordered){width=85%} 

5. Use the chat interface on the lower right to interact with the Agent Builder. As you provide more details or refine your requirements, the Agent Builder automatically drafts and updates. 
6. After the creation is completed, click **Start Conversation** to use the agent.
7. Provide your text in the chat, and then you can get the refined results.

   ![Start using agent](/images/manual/use-cases/lobehub-agent-use1.png#bordered){width=85%} 

### Create a custom agent

If you have specific needs, you can create a custom assistant tailored to your personal requirements.

Custom assistants offer the highest level of personalization. You can set the assistant’s avatar, name, prompts, preferred AI model, and plugins to create a truly unique AI assistant. All customization can be done manually on the assistant profile page.

1. On the homepage, click the robot icon and then select **Create Agent**.

   ![Create custom agent](/images/manual/use-cases/lobehub-create-custom-agent.png#bordered){width=40%} 

   The **Agent Profile** page opens.

   ![Custom agent profile](/images/manual/use-cases/lobehub-custom-agent-profile.png#bordered){width=85%} 

2. Enter the agent name, select the language model,

## Manage agents

When you have many assistants and group chats, organizing them into groups is the most intuitive way to manage them. It keeps your assistant list clean and makes switching between them easier.

### Create a group

On the LobeHub homepage, open the assistant list menu and select "Add New Group" to create a new group.

### Move to a group

If you have multiple groups, go to the assistant list or group menu and select "Manage Groups" to easily rename or reorder them.

### Pin freqnetly used agents

You can pin frequently used assistants to the top of the list for quicker access. Select the assistant and choose "Pin" from the menu. Pinned assistants will stay at the top of the list for easy access.

## Create an agent team

For complex workflows, a single agent might not be enough. LobeHub allows you to create an Agent Group (or Agent Team), where multiple specialized agents collaborate, execute tasks in parallel, and iterate on each other's work.

<!--1. On the homepage, click the robot icon and then select **Create Group**.

   ![Create agent group](/images/manual/use-cases/lobehub-create-agent-group.png#bordered){width=40%} 

   The Group Profile page opens.

2. In the **Agent Builder** panel, describe the specific task you want the agent team to handle in the chat box. For exmaple,

   ```
   I need a team to research trending tech news and write a daily newsletter.
   One agent should gather the facts, and another should format them into an
   engaging email draft.
   ```
3. Select the language model, and then press **Enter**.

   ![Agent group builder](/images/manual/use-cases/lobehub-group-builder.png#bordered){width=40%} 

   The profile page of the new agent group opens, and you can see the Agent Builder immediately starts designing your agent team automatically.-->

1. On the homepage, click **Create Group** under the chat box.

   ![Create Group button](/images/manual/use-cases/lobehub-create-group.png#bordered){width=85%} 

2. In the chat box, describe the specific task you want the agent team to handle. For exmaple,

   ```
   I need a team to research trending tech news and write a daily newsletter.
   One agent should gather the facts, and another should format them into an
   engaging email draft.
   ```
3. Select the language model, and then press **Enter**.

   ![Create Group chat box](/images/manual/use-cases/lobehub-create-group-start.png#bordered){width=85%} 

   The profile page of the new agent group opens, and you can see the Lobe AI immediately starts designing your agent team automatically.

   ![Agent group builder](/images/manual/use-cases/lobehub-agent-group-builder.png#bordered){width=85%} 

4. Communicate with Lobe AI to clarity your requirements, provide details to it for next step, confirm operations, or approve actions.

   For example, you agree with the Lobe AI's design to create a new agent for tech news gathering, and then you provide the following details to it:

   ```
   Agent name: TechNewsGatherer
   Description: Responsible for collecting and summarizing the latest tech news from give URL.
   Tools: Web-crawler to fetch relevant articles from related tech websites.
   System Role: Information catching.
   ```
   
   It will send a create agent request for you to approve:

   ![Building agent group](/images/manual/use-cases/creating-agent-team.png#bordered){width=35%} 

   Once approved, the new agent is created and added to the agent team automatically.

    ![Agent team member created](/images/manual/use-cases/agent-group-member-created.png#bordered){width=85%} 

5. From the left sidebar, point to **Memebers**, and then click the **Add Member** icon to bring additional assistants into the group chat.
6. From the left sidebar, point to an existing member, and then click the **Remove Member** icon to delete a member from the group chat.
7. Configure the group setting, such as group name, description on group objectives or work modes.
8. When all set, click **Start Conversation** to use it.



<!--## Create an assistant

LobeHub allows you to create specialized assistants to handle specific tasks by leveraging various language models and combining them with functional plug-ins.

- **Flexible model switching**: You can switch language models instantly within the same chat to achieve the best results. For example, if you are not satisfied with a response, you can select a different model from the list to leverage their unique strengths.
- **Plug-in extensions**: You can also install plug-ins to extend and enhance the capabilities of your assistant.

   :::info
   To install plug-ins, ensure that you select a model compatible with Function Calling. Look for <i class="material-symbols-outlined">brick</i> next to the model name, which indicates the model supports function calls.
   :::

The following steps outline the standard workflow for creating and configuring any assistant in LobeChat. You can apply this procedure using specific settings provided in the [use scenarios](#use-scenarios) section.

1. Create a new assistant:
   - From the left navigation pane, click **Lobe AI**. 
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
3. Paste and send the URL to the chat. The assistant lists specific news stories with summaries.-->

## FAQ

### Why did the connection check fail when I connected to Ollama?

1. Ensure the **Use Client Request Mode** option on the Ollama settings page is disabled.
2. Ensure the model you are using is downloaded using Ollama.

   ![Connectivity error](/images/manual/use-cases/lobehub-connection-error.png#bordered){width=85%} 
