---
outline: [2, 4] 
description: Learn how to install LobeHub (LobeChat) on Olares and integrate it with Ollama to build and enhance your local custom AI assistants.
head:
  - - meta
    - name: keywords
      content: Olares, LobeHub agent, LobeChat assistant, AI agent, AI agent team
---

# Build your local AI agent with LobeHub (LobeChat)

LobeHub (previously LobeChat) is an open‑source framework for building secure, local AI chat experiences. It supports file handling, knowledge bases, and multimodal inputs, and it supports Ollama to run and switch local LLMs.

Olares streamlines and simplifies the deployment of both, allowing you to skip complex manual environment configurations.

This guide covers the installation, configuration, and practical usage of these tools to create your personalized AI agents.

:::tip About the product name
LobeHub is the official platform name, but the application is currently listed as "LobeChat" in the Olares Market. We use both names in this guide to match exactly what you will see on your screen. The Market will be updated to reflect the new LobeHub branding in the future release.
:::

## Learning objectives

- Configure LobeHub to communicate with your local Ollama instance.
- Create specialized agents tailored to specific tasks and equip them with specific skills.
<!--- Create an agent group to enable multiple agents to collaborate on complex workflows.-->

## Prerequisites

- Ollama is installed and running in your Olares environment.
- The models you want to use are downloaded and run using Ollama. This tutorial uses `llama3.1:8b` and `qwen2.5`. For more information, see [Download and run local AI models via Ollama](ollama.md).

## Install LobeHub

1. From the Olares Market, search for "LobeChat".

   ![Search for LobeChat from Market](/images/manual/use-cases/find-lobechat1.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Sign in to LobeHub

1. Open **LobeChat** from the Launchpad.
2. Enter your email address, and then follow the prompts on the page to create a LobeHub account and sign in.

   ![LobeHub home page](/images/manual/use-cases/lobehub-start.png#bordered)

## Configure the connection

Connect LobeHub to Ollama to make the chat interface work.

1. From the left sidebar, go to **Settings** > **AI Service Provider** > **Ollama**.

      ![Configure Ollama in LobeHub](/images/manual/use-cases/lobehub-config-ollama.png#bordered)

2. In the **Interface proxy address** field, enter your local Ollama address.

   :::tip Obtain local Ollama host address
   To obtain your local Ollama host address, go to Olares **Settings** > **Applications** > **Ollama**, click **Ollama API** under **Entraces** or **Shared entrances**, and then copy the endpoint address.

   ![Obtain Ollama host address from Olares Settings](/images/manual/use-cases/obtain-ollama-hosturl1.png#bordered){width=60%}
   :::

3. Disable the **Use Client Request Mode** option.

   :::tip
   When you are running local models, do not enable the **Use Client Request Mode** option.
   :::
4. In the **Model List** section, click **Fetch models** to pull the list of supported models, and then click <i class="material-symbols-outlined">toggle_off</i> to enable the models you want to use.

   ![Fetch model list and enable models](/images/manual/use-cases/lobehub-fetch-enable-model.png#bordered){width=85%} 

5. In the **Connectivity Check** section, select the model you just enabled from the list, and then click **Check** to verify the connection. If the model is large, it might take a little longer to load.

   ![Connectivity check](/images/manual/use-cases/lobehub-connectivity-check.png#bordered){width=85%} 

   The button changes to **Check Passed**, indicating that the proxy address is correct. 

   ![Connectivity check success](/images/manual/use-cases/lobehub-checkpass.png#bordered){width=85%}    

6. Click the home icon at the upper-left corner to return to the LobeHub home page.

   ![Return to home page](/images/manual/use-cases/lobehub-return-home.png#bordered){width=45%} 

## Use Lobe AI

Lobe AI is the official default agent from LobeHub. It is designed to help you accomplish a wide range of tasks without the need for complex setup, such as software development, learning support, creative writing, data analysis, and daily personal tasks.

If Lobe AI does not meet your specific workflow needs, you can build your own specialized agents. For more information, see [Create an agent](#create-an-agent).

1. From the left sidebar, click **Lobe AI**.
   
   ![Click Lobe AI](/images/manual/use-cases/lobe-ai.png#bordered){width=85%} 

2. In the chat window, click the model selector and select a local language model.
3. Chat as you would with any standard conversational AI.

## Create an agent

Create your own specialized agents by using the conversational Agent Builder or by manually configuring the settings from scratch.

LobeHub allows you to create specialized assistants to handle specific tasks by leveraging various language models and combining them with skills.
- **Flexible model switching**: You can switch language models instantly within the same chat to achieve the best results. For example, if you are not satisfied with a response, you can select a different model from the list to leverage their unique strengths.
- **Skill extensions**: You can also install additional skills to extend and enhance the capabilities of your agent.
   To install skills, ensure that you select a model compatible with Function Calling. Look for <i class="material-symbols-outlined">brick</i> next to the model name, which indicates the model supports function calls.

### Create using Agent Builder

Agent Builder is LobeHub's built-in assistant that helps you create specialized agents through conversations. Describe your needs, and it will automatically generate a complete agent configurations, including role settings, system prompts, and skills.

1. On the home page, click **Create Agent** under the chat box.

   ![Create Agent button](/images/manual/use-cases/lobehub-create-agent.png#bordered){width=85%} 

2. In the chat box, describe the specific task you want the agent to handle. For example,

   ```
   I need an agent to review my daily work items and summarize them.
   The summary should focus on the overall purpose of the tasks and
   highlight specific action items.
   ```
3. Select the language model. For example, `llama3.1:8b`.
4. Press **Enter**. The profile page of the new agent opens, and you can see the Agent Builder starts configuring your agent automatically.

   ![Agent builder](/images/manual/use-cases/lobehub-agent-builder.png#bordered){width=85%} 

5. Use the chat interface on the lower right to interact with the Agent Builder. As you provide more details or refine your requirements, the Agent Builder automatically drafts and updates accordingly. 
6. After the creation is completed, click **Start Conversation** to use the agent.
7. Provide your text in the chat, and then you can get the refined results. For example, 
   ```
   - fix bug 405 on login
   - discuss with design on new dashboard
   - answer customer question about billing in email.
   - review pr112, ddl 11:00 am tmrw
   ```
   You get the output:

   ![Sample output by agent builder](/images/manual/use-cases/agent-builder-example.png#bordered){width=85%}    

8. If you are satisfied with the agent's performance, pin it for quick access:

   a. Return to the home page.
   
   b. Hover over the agent from the left sidebar, click <i class="material-symbols-outlined">more_horiz</i>, and then click **Pin**.

### Create a custom agent

If you have specific requirements and prefer to configure the agent entirely manually, create a custom agent.

Custom agents offer the highest level of personalization. You can set the agent's avatar, name, AI model, skills, and prompt to create a unique AI agent.

1. On the home page, click the robot icon in the upper left corner, and then select **Create Agent**.

   ![Create custom agent](/images/manual/use-cases/lobehub-create-custom-agent.png#bordered){width=40%} 

   The **Agent Profile** page opens.

   ![Custom agent profile](/images/manual/use-cases/lobehub-custom-agent-profile.png#bordered){width=85%} 

2. Click the default robot avatar to select a new icon for your agent.
3. Enter the agent name. For example, `SEO Copywriter`.
4. Select the language model. For example, `qwen 2.5`.
5. Click **+ Add Skill** to equip the agent with additional tools. For example, select **Web Browsing** for gathering SEO data.
6. Define role and behavior by filling out the structured markdown template to define exactly how the agent operates. For example,

   ```
   #### Goal
   Write SEO-optimized blog posts based on the user-provided topic.
   #### Skills
   - Keyword research, deployment, and and density optimization
   - Engaging headline generation
   - Markdown formatting
   #### Workflow
   1. Ask the user for a topic.
   2. Suggest target keywords, an H1 title, and an optimal meta description.
   3. Generate a structured outline designed for google's featured snippets.
   4. Generate a structured outline for approval.
   5. Write the full blog post once the outline is approved.
   #### Constraints
   - Use simple language and avoid technical jargon.
   - Focus on user values instead of listing product features.
   - Avoid using passive voice.
   - Target users with the second person "you"
   ```
7. Click **Start Conversation** to use it. For example, type the following request:

   ```
   I want to rank for "local AI alternatives"
   ```
8. Review the proposal and output, and then iterate with it until you are satisfied with the results.

   ![Custom agent result sample](/images/manual/use-cases/lobehub-seo-sample.png#bordered){width=85%} 

9. If you are satisfied with the agent's performance, pin it for quick access:

   a. Return to the home page.
   
   b. Hover over the agent from the left sidebar, click <i class="material-symbols-outlined">more_horiz</i>, and then click **Pin**.

<!--
## Manage agents

When you have many assistants and group chats, organizing them into groups is the most intuitive way to manage them. It keeps your assistant list clean and makes switching between them easier.

### Pin agents

Pin frequently used assistants to the top of the agent list for quicker access. 
1. On the LobeHub home page, find the assistant in the **Agent** section on the left sidebar.
2. Point to it, click <i class="material-symbols-outlined">more_horiz</i>, and then click **Pin**. The pinned assistants will stay at the top of the list for easy access.

### Categorize agents

create categories to group different agents for

1. On the LobeHub home page, point to **Agent** from the left sidebar, click <i class="material-symbols-outlined">more_horiz</i>, and then click **Add New Category**. A **New Category** section is created under **Agent**.
   ![Add New Category menu](/images/manual/use-cases/lobehub-new-category.png#bordered){width=45%} 

2. Point to **New Category**, click <i class="material-symbols-outlined">more_horiz</i>, and then click **Rename Category**. 

### Move to a group

If you have multiple groups, go to the assistant list or group menu and select "Manage Groups" to easily rename or reorder them.

## Create an agent team

For complex workflows, a single agent might not be enough. LobeHub allows you to create an agent team, where multiple specialized agents collaborate as members, execute tasks in parallel, and iterate on each other's work.

1. On the home page, click **Create Group** under the chat box.

   ![Create Group button](/images/manual/use-cases/lobehub-create-group.png#bordered){width=85%} 

2. In the chat box, describe the specific task you want the agent team to handle. For example,

   ```
   I need a team to research trending AI tech news and write a daily 
   newsletter. One agent should gather the facts, and another should
   format them into an engaging email draft.
   ```
3. Select the language model, and then press **Enter**.

   ![Create Group chat box](/images/manual/use-cases/lobehub-create-group-start.png#bordered){width=85%} 

   The **Group Profile** opens with a **Supervisor** created by default. Every agent team chat includes a built-in moderator responsible for: Understanding your needs and assigning discussion tasks, Coordinating the speaking order of assistants, Summarizing the discussion and extracting key conclusions, and Keeping the conversation organized and on-topic.
   
   Meanwhile, the Lobe AI starts designing the team automatically and lists the steps to complete the task.

   ![Agent group builder](/images/manual/use-cases/lobehub-agent-group-builder.png#bordered){width=85%} 

4. Communicate with Lobe AI to complete the steps:
   - Provide detailed for group settings and agent configurations.
   - Approve the requests to create individudal agent members.
   - Clarity your requirements when necessary.

   When the creation of the team agents is completed, the agents are displayed in Members on the left sidebar.

    ![Agent team member created](/images/manual/use-cases/agent-group-member-created.png#bordered){width=85%}

5. Click **Group Profile** and check the configurations of each agent on its tab. Make adjustments as needed. For example,
 
   - Group Settings:
      - Group name: AI Tech News Research & Newsletter Team
      - Group objectives or work modes: I need a team to research trending AI tech news and write a daily newsletter. This will be used as the shared prompt for team agents.

   - Configure the Supervisor, including the avatar, name, model, skill, and supervisor information to enable more precise workflow coordination.
      - Name: Supervisor
      - Model: Qwen2.5 7B
      - Skill: Web browsing
      - Description: I need a team to research trending AI tech news and write a daily newsletter. This will be used as the shared prompt for team agents.

6. Click **Start Conversation** to use it. For example, type `crawl this webpage https://news.ycombinator.com/ and draft a short, engaging newsletter for the latest three AI news`, and then 

   ![Agent team work result sample](/images/manual/use-cases/lobehub-team-result.png#bordered){width=85%} 

## Manage agent teams

### Add or remove members
 
1. In the team chat, from the left sidebar, point to **Memebers**, and then click the **Add Member** icon to bring additional assistants into the group chat.
2. From the left sidebar, point to an existing member, and then click the **Remove Member** icon to delete the member from the team chat.

### Delete agent teams

1. On the LobeHub home page, point to the target agent team, click <i class="material-symbols-outlined">more_horiz</i>, and then click **Delete**.
-->

## FAQ

### Why did the connection check fail when I connected to Ollama?

If you encounter the `Error requesting Ollama service` error, troubleshoot as follows and retry:

   ![Connectivity error](/images/manual/use-cases/lobehub-connection-error.png#bordered){width=85%} 
1. Ensure the specific model you are using is downloaded using Ollama.
2. Ensure the **Use Client Request Mode** option on the Ollama settings page is disabled.

   ![Disable the use client request mode option](/images/manual/use-cases/lobehub-disable-client-request-mode.png#bordered){width=85%} 

### Error: Model requires more system memory than is available

If you encounter errors similar te:

```json
{
  "error": {
    "message": "model requires more system memory (4.5 GiB) than is available (3.8 GiB)",
    "name": "ResponseError",
    "status_code": 500
  },
  "provider": "ollama"
}
```
