---
outline: [2, 4] 
title: Build a local AI agent with LobeHub
description: Install LobeHub on Olares and connect it to local models to build self-hosted AI assistants with knowledge bases, skills, and multimodal input.
head:
  - - meta
    - name: keywords
      content: Olares, LobeHub, LobeChat, self-hosted lobechat, AI agent, lobechat on olares
app_version: "1.0.14"
doc_version: "2.0"
doc_updated: "2026-07-29"      
---

# Build your local AI agent with LobeHub

LobeHub (previously known as LobeChat) is an open-source platform for building secure, self-hosted AI agents and chat experiences. It connects to your local models, supports file handling and knowledge bases, and allows you to create specialized agents with custom skills.

This guide covers the installation, configuration, and practical usage of LobeHub to create your personalized AI agents.

:::tip About the product name
LobeHub is the official platform name, but the application is currently listed as "LobeChat" in the Olares Market. We use both names in this guide to match exactly what you will see on your screen. The Market will be updated to reflect the new LobeHub branding in the future release.
:::

## Learning objectives

- Install LobeHub on Olares and connect it to your local model.
- Chat with Lobe AI for everyday tasks.
- Create specialized agents using the Agent Builder or custom settings.
<!--- Create an agent group to enable multiple agents to collaborate on complex workflows.-->

## Prerequisites

Before you begin, you need the following model:
| Model type | Model | How to get it |
| :--- | :--- | :--- |
| Chat | Qwen3.6-27B (llama.cpp) | Install from Market |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install LobeHub

1. From the Olares Market, search for "LobeChat".

   ![Search for LobeChat from Market](/images/manual/use-cases/find-lobechat2.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Sign in to LobeHub

1. Open **LobeChat** from the Launchpad.
2. Enter your email address, and then follow the prompts on the page to create a LobeHub account and sign in.

   ![LobeHub home page](/images/manual/use-cases/lobehub-start1.png#bordered)

## Configure the connection

Connect LobeHub to your local model to make the chat interface work.

### Get model connection details

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### Configure the connection in LobeHub

1. From the left sidebar, go to **Settings** > **AI Service Provider** > **OpenAI**.

      ![Configure model connection in LobeHub](/images/manual/use-cases/lobehub-config-model.png#bordered)

2. Configure the following settings:

   - **API Key**: Enter any placeholder text such as `local`.
   - **API Proxy URL**: Enter the **Base URL** you copied from the Model Console. For example, `https://e46e044d.laresprime.olares.com/v1`.
   - **Use Responses API Specification**: Ensure this option is disabled.
   - **Use Client Request Mode**: Ensure this option is disabled.

      :::tip
      Do not enable the **Use Client Request Mode** option when running local models. This mode is designed for remote API calls and might cause connection errors.
      :::

3. In the **Model List** section, click **Fetch models** to pull the list of supported models. The model name `unsloth/Qwen3.6-27B-GGUF:Q4_K_M` appears in the list.

   ![Fetch model list and enable models](/images/manual/use-cases/lobehub-fetch-enable-model1.png#bordered)

4. Click <i class="material-symbols-outlined">toggle_off</i> to enable it.
5. In the **Connectivity Check** section, select the model you just enabled from the list, and then click **Check** to verify the connection. If the model is large, it might take a little longer to load.

   The button changes to **Check Passed**, indicating that the connection is established. 

   ![Connectivity check success](/images/manual/use-cases/lobehub-checkpass2.png#bordered)  

6. Click the home icon at the upper-left corner to return to the LobeHub home page.

   ![Return to home page](/images/manual/use-cases/lobehub-return-home.png#bordered){width=50%} 

## Use Lobe AI

Lobe AI is the official default agent from LobeHub. It is designed to help you accomplish a wide range of tasks without the need for complex setup, such as software development, learning support, creative writing, data analysis, and daily personal tasks.

If Lobe AI does not meet your specific workflow needs, you can build your own specialized agents. For more information, see [Create an agent](#create-an-agent).

1. From the left sidebar, click **Lobe AI**.
   
   ![Click Lobe AI](/images/manual/use-cases/lobe-ai.png#bordered) 

2. Under the chat window, click the model selector and select the local model.
3. Chat as you would with any standard conversational AI.

## Create an agent

Create your own specialized agents by using the conversational Agent Builder or by manually configuring the settings from scratch.

LobeHub allows you to create specialized assistants to handle specific tasks by leveraging various language models and combining them with skills.
- **Flexible model switching**: You can switch language models instantly within the same chat to achieve the best results. For example, if you are not satisfied with a response, you can select a different model from the list to leverage their unique strengths.
- **Skill extensions**: You can also install additional skills to extend and enhance the capabilities of your agent.
   To install skills, ensure that you select a model compatible with Function Calling. Look for <i class="material-symbols-outlined">brick</i> next to the model name, which indicates the model supports function calls.

### Create using Agent Builder

Agent Builder is LobeHub's built-in assistant that helps you create specialized agents through conversations. Describe your needs, and it will automatically generate a complete agent configuration, including role settings, system prompts, and skills.

1. On the home page, click **Create Agent** under the chat box.

   ![Create Agent button](/images/manual/use-cases/lobehub-create-agent1.png#bordered)

2. In the chat box, describe the specific task you want the agent to handle. For example,

   ```text
   I need an agent to review my daily work items and summarize them.
   The summary should focus on the overall purpose of the tasks and
   highlight specific action items.
   ```
3. Select the local language model.
4. Press **Enter**. The profile page of the new agent opens, and you can see the Agent Builder starts configuring your agent automatically.

   ![Agent builder](/images/manual/use-cases/lobehub-agent-builder1.png#bordered)

5. Use the chat interface on the lower right to interact with the Agent Builder. As you provide more details or refine your requirements, the Agent Builder automatically drafts and updates accordingly. 
6. When the creation is completed, click **Start Conversation** to use the agent.
7. Provide your text in the chat, and then you can get the refined results. For example:

   ```text
   - fix bug 405 on login
   - discuss with design on new dashboard
   - answer customer question about billing in email.
   - review pr112, ddl 11:00 am tmrw
   ```
   You get the output:

   ![Sample output by agent builder](/images/manual/use-cases/agent-builder-example1.png#bordered)  

8. If you are satisfied with the agent's performance, pin it for quick access:

   a. Return to the home page.
   
   b. Hover over the agent from the left sidebar, click <i class="material-symbols-outlined">more_horiz</i>, and then click **Pin**.

### Create a custom agent

If you have specific requirements and prefer to configure the agent entirely manually, create a custom agent.

Custom agents offer the highest level of personalization. You can set the agent's avatar, name, AI model, skills, and prompt to create a unique AI agent.

1. On the home page, click the robot icon in the upper left corner, and then select **Create Agent**.

   ![Create custom agent](/images/manual/use-cases/lobehub-create-custom-agent.png#bordered){width=50%} 

   The **Agent Profile** page opens.

   ![Custom agent profile](/images/manual/use-cases/lobehub-custom-agent-profile1.png#bordered)

2. Click the default robot avatar to select a new icon for your agent.
3. Enter the agent name. For example, `SEO Copywriter`.
4. Select the local model.
5. Click **+ Add Skill** to equip the agent with additional tools. For example, select **Web Browsing** for gathering SEO data.
6. Define role and behavior by filling out the structured markdown template to define exactly how the agent operates. For example,

   ```text
   #### Goal
   Write SEO-optimized blog posts based on the user-provided topic.
   #### Skills
   - Keyword research, deployment, and density optimization
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

   ```text
   I want to rank for "local AI alternatives"
   ```
8. Review the proposal and output, and then iterate with it until you are satisfied with the results.

   ![Custom agent result sample](/images/manual/use-cases/lobehub-seo-sample1.png#bordered)

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

   ![Connectivity error](/images/manual/use-cases/lobehub-connection-error.png#bordered)
1. Check the Model Console to confirm that the Model shows **READY** and the Engine shows **RUNNING**.
2. Ensure the **Use Client Request Mode** option on the Ollama settings page is disabled.

   ![Disable the use client request mode option](/images/manual/use-cases/lobehub-disable-client-request-mode3.png#bordered)
