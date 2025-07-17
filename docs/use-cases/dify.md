---
description: Guide to leveraging Dify on Olares for building AI applications. Learn how to deploy Dify and add personal knowledge base with ease.
---
# Customize your local AI assistant using Dify

Dify is an AI application development platform. It's one of the key open-source projects that Olares integrates to help you build and manage AI applications while maintaining full data ownership. Additionally, you can integrate your personal knowledge base documents into Dify for more personalized interactions.

## Before you begin
To use local AI models on Dify, ensure you have:
- [Ollama installed](ollama.md) and running in your Olares environment.
- LassPass installed and [VPN enabled](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.
- [Open WebUI installed](openwebui.md) with your preferred language models downloaded.
  :::tip
  For optimal performance, consider using lightweight yet powerful models like `gemma2` or `qwen`, which offer a good balance between speed and capability.
  :::

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

## Add Ollama as model provider

1. Navigate to **Settings** > **Application** > **Ollama** > **Entrances**, and set the authentication level for Ollama to **Internal**. This configuration allows other applications to access Ollama services within the local network without authentication. 
    
    ![Ollama entrance](/images/manual/use-cases/dify-ollama-entrance.png#bordered)

2. In Dify, navigate to **Settings** > **Model Provider**.
3. Select Ollama as the model provider, with the following configurations:
    - **Model Name**: Enter the model name. For example: `gemma2`.
    - **Base URL**: Enter Ollama's local address: `https://39975b9a1.{username}.olares.com`. Replace `{username}` with the Olares Admin's local name. For example, `https://39975b9a1.marvin123.olares.com`.

     ![Add gemma2](/images/manual/use-cases/dify-add-gemma2.png#bordered){width=70%}

      :::tip
      You can keep default values for other required fields.
      :::
4. Click **Save**.

## Configure Ashia

1. Navigate to Dify's **Studio** tab and enter Ashia.  
2. From the model list on the right, select the Gemma2 model you just configured.

   ![Select model](/images/manual/use-cases/dify-select-model.png#bordered)
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