---
outline: [2, 3]
description: Learn how to link Olares applications to create workflows.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, app communication, connect apps, endpoint, internal access
---

# Connect applications <Badge type="tip" text="15 min" />

On Olares, applications can work together to create powerful workflows. A common example is connecting an AI service application to a separate chat interface application. By linking these applications, you can leverage the processing power of one app within the user-friendly interface of another.

This guide explains the core concepts of application communication and walks you through a real-world scenario by connecting Ollama to LobeChat.

## Learning objectives

- Distinguish between service apps and interface apps.
- Locate and configure application endpoints.
- Understand authentication levels.
- Connect an AI service to a chat client.

## Before you begin

To successfully connect applications, it is helpful to understand three key concepts.

### Application types

- **Service apps**: These run in the background and provide resources like AI models. They usually do not have a direct user interface. For example, Ollama provides AI computing power but has no graphical interface itself.
- **Interface apps**: These provide the visual interface you interact with. They often need to connect to a service app to function. For example, LobeChat provides the chat window but relies on a service like Ollama to generate text.

### Endpoints

An Endpoint is the URL address that defines where a service can be found on your Olares system. 

- **Standard format**: `https://{route-ID}.{username}.olares.com`.
- **Shared format**: `http://{route-ID}.shared.olares.com`. This format is used for specific backend services.

### Authentication levels

Olares provides three security levels for application access.
- **Public**: This level is open to anyone on the internet. It is generally not recommended for private services.
- **Private**: This level requires authentication for access. It provides services exclusively for individual users, families, or teams.
- **Internal**: This level allows access to other apps on your Olares system or via LarePass VPN without requiring authentication. It is the primary setting used for connecting applications together securely within your local network.

## Prerequisites

- Ollama is installed and running with at least one language model downloaded.
- LobeChat is installed and running.
- Administrative access to Olares Settings.

## How it works

Connecting any two applications generally follows two steps:

1. **Prepare the service app**: In Olares Settings, set the **Authentication level** of the service app to **Internal**, and copy the endpoint URL.
2. **Configure the interface app**: In the interface app, paste that URL into the connection settings.

    :::tip Configuration notes
    - **Fake API keys**: Some applications like third-party ChatGPT clients are hard-coded to require an API key, even if your local service like Ollama does not need one. In this case, enter any random text such as `ollama` to bypass the mandatory field check.
    - **Model name matching**: The interface app cannot guess which AI model you have installed. Ensure the model name configured in your interface app matches exactly what you have installed inside your service app.
    :::

## Scenario: Connect LobeChat to Ollama

In this scenario, we will configure the chat interface LobeChat to use your local AI service Ollama.

### Step 1. Get the service address

First, find the connection address for Ollama and make it accessible to LobeChat.

1. Open Olares **Settings**.
2. Go to **Application** > **Ollama**.
3. In **Entrances**, click **Ollama API**.
4. Change the **Authentication level** to **Internal**.
5. Click **Set up endpoint**, and then copy the endpoint address displayed. 

   ![Obtain Ollama host address from Olares Settings](/images/manual/use-cases/obtain-ollama-hosturl.png#bordered){width=60%} 

### Step 2: Configure the client

Next, tell LobeChat where to find the Ollama service.

1. Open LobeChat.
2. Go to **Settings** > **Language Model**.
3. In the **Ollama** section, find the **Interface proxy address** field, and then enter the Ollama endpoint address you just copied. 

   ![Interface proxy address connection](/images/manual/use-cases/lobechat-connection-setting.png#bordered)

4. Click **Check** under the Ollama model list to verify the connection. A **Check Passed** message indicates that the connection is successful. LobeChat is now powered by your local AI. 

## Troubleshoot connection

If the connection fails using the steps above, try the following adjustments:

1. **Change protocol**: Change `https://` to `http://`. For example,
    - Change `https://92d76a130.alice123.olares.com` 
    - To this:  `http://92d76a130.alice123.olares.com`
2. **Try alternative endpoints**: If multiple addresses are listed in the Olares Settings, try a different one. For example,
    - Try `https://92d76a131.alice123.olares.com` 
    - Or `http://d276a13c0.shared.olares.com`.
3. **Add suffix `/v1`**: Append `/v1` to the end of the URL. For example,
    - Change `https://92d76a131.alice123.olares.com`
    - To this: `https://92d76a131.alice123.olares.com/v1`
4. **Add suffix `/api`**: Append `/api` to the end of the URL. For example,
    - Change `https://92d76a131.alice123.olares.com`
    - To this: `https://92d76a131.alice123.olares.com/api`

## Resources

- [Applications](../developer/concepts/application.md)
- [Network](../developer/concepts/network.md)
- [Manage application entrances](../manual/olares/settings/manage-entrance.md)
