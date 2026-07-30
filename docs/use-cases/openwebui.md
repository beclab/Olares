---
outline: [2, 3]
description: Self-host Open WebUI on Olares for private, local AI chat. Connect it to local models and keep conversations on your device.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, self-hosted AI platform, local LLM, open webui on olares
app_version: "1.0.38"
doc_version: "3.0"
doc_updated: "2026-07-30"
---

# Set up Open WebUI for local AI chat

Open WebUI is a self-hosted chat interface that lets you interact with local models on your Olares device.

This guide walks you through installing Open WebUI, connecting it to a local model, and starting your first conversation.

## Learning objectives

In this guide, you will learn how to:

- Install Open WebUI on Olares.
- Create an admin account.
- Connect Open WebUI to a local model.
- Start a chat session with your configured model.

## Prerequisites

Before you begin, you need:

- An Olares device with sufficient disk space and memory
- The following model:
   | Model type | Model | How to get it |
   | :--- | :--- | :--- |
   | Chat | Qwen3.6-27B (llama.cpp) | Install from Market |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install Open WebUI

1. Open Market and search for "Open WebUI".

   ![Open WebUI](/images/one/open-webui.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Create an admin account

The first time you launch Open WebUI, you need to create a local administrator account to manage your models and settings.

1. Open Open WebUI from the Launchpad.
2. On the welcome page, click **Get started**.
3. Enter your name, email, and password to create the admin account.

   ![Create account](/images/one/open-webui-create-account.png#bordered)

## Get model connection details

To connect Open WebUI to a model, you first need to collect the model's connection information from its Model Console.

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

## Configure the connection in Open WebUI

With the connection details ready, add the model as an OpenAI-compatible provider in Open WebUI.

1. In Open WebUI, click your profile icon and select **Admin Panel**.
2. Select the **Settings** tab, and then choose **Connections** from the left sidebar.
3. To the right of **Manage OpenAI API Connections**, click <span class="material-symbols-outlined">add</span> to add a new connection.
4. In the **API Base URL** field, enter the **Base URL** you copied from the Model Console. For example, `https://e46e044d.laresprime.olares.com/v1`.

   ![Connection established](/images/manual/use-cases/open-webui-connection-established1.png#bordered)

5. Click **Save**. Open WebUI verifies the connection automatically.

   When you see the "OpenAI API settings updated" message, the connection is established.

## Start chatting

Once you connect a model, you are ready to use the chat interface.

1. In the chat area, select your configured model.

   ![Select model](/images/manual/use-cases/open-webui-chat.png#bordered)

2. Enter your prompt in the text box, and then press **Enter** to start your conversation.

   ![Chat with LLM](/images/manual/use-cases/open-webui-chat-result.png#bordered)

## Learn more

- [Set up multi-user access](openwebui-multiuser.md): Share Open WebUI with other users on your Olares device.
- [Configure audio](openwebui-audio.md): Enable speech-to-text and text-to-speech.
- [Enable web search](openwebui-search.md): Add web search capabilities to your chats.
- [Use knowledge base](openwebui-knowledge.md): Upload documents and create a knowledge base for RAG.
