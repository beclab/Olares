---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/dify
description: Self-host Dify on Olares to build local AI apps and assistants. Deploy Dify, connect models through Router, and add a personal knowledge base for private RAG workflows.
head:
  - - meta
    - name: keywords
      content: Olares, Dify, dify self hosted, dify vs n8n, dify ollama, dify on olares
---
# Customize your local AI assistant using Dify

<VersionRouteSelect />

Dify is an AI application development platform. It's one of the key open-source projects that Olares integrates to help you build and manage AI applications while maintaining full data ownership. Additionally, you can integrate your personal knowledge base documents into Dify for more personalized interactions.

## Before you begin

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- Install Qwen3.8-27B (llama.cpp) from Market.

## Install Dify
:::info
Starting from Olares 1.11.6, if "Dify For Cluster" or "Dify" was previously installed, uninstall them before proceeding.
:::

1. Install "Dify Shared" from Olares Market. 
2. Launch Dify from your desktop. Please ensure the admin has already installed Dify Shared.

## Create an AI assistant app

1. Open Dify, navigate to the **Studio** tab, and select **Create from Blank** to create an app for the AI assistant. Here, we created an agent named "Ashia".
   ![Create App](/images/manual/use-cases/dify-create-app.png#bordered)

2. Click **Go to settings** on the right to access the model provider configuration page. You can choose between remote models or locally hosted models. 
   ![App initial age](/images/manual/use-cases/dify-app-init.png#bordered)

## Add a chat model through Router

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

1. In Dify, open **Settings** > **Model Provider**, then install and select the **OpenAI-API-compatible** provider plugin.
2. Add a model with **Model Name** set to `default-chat` and the model type set to chat or **LLM**.
3. Set **Base URL** to the URL copied from Router, including `/v1`. Leave the API key empty if allowed, or enter `olares` if required.
4. Save the configuration.

## Configure Ashia

1. Navigate to Dify's **Studio** tab and enter Ashia.  
2. From the model list on the right, select the `default-chat` model you just configured.


   <!--
   TODO: 素材清单 05，待补 Router 截图：dify-model-router.png；替换下方旧图后再取消注释。
   ![Select model](/images/manual/use-cases/dify-select-model.png#bordered)
   -->

3. Click **Publish**. Now you can chat with Ashia in the **Debug & Preview** window. 

   ![Chat](/images/manual/use-cases/dify-chat-with-ashia.png#bordered)

## Set up local knowledge base
1. In Dify, navigate to the **Knowledge** tab.
2. Locate your default knowledge base. It will be named after your Olares ID and monitors the `/Documents` folder in Files.
   ![Default KB](/images/manual/use-cases/dify-default-knowledge-base.png#bordered)
3. Enter `/Documents` and add documents to the knowledge base.
   ![Default KB](/images/manual/use-cases/dify-add-kb-file.png#bordered)
4. In Ashia's orchestration page, click **<i class="material-symbols-outlined">add</i>Add** to add context support for Ashia.
    ![Add KB](/images/manual/use-cases/dify-add-knowledge-base.png#bordered){width=70%}
5. Click **Publish**. Now try asking a domain-specific question with the help of the knowledge base.
    ![Add KB](/images/manual/use-cases/dify-chat-kb.png#bordered)
