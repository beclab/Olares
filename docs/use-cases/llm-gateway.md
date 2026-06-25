---
outline: [2, 3]
description: Learn how to install and configure LLM Gateway on your Olares device to unify and manage your local and cloud AI models.
head:
  - - meta
    - name: keywords
      content: Olares, LLM Gateway, AI models, provider, OpenAI compatible
app_version: "v1.0.4"
doc_version: "1.0"
doc_updated: "2026-06-25"
---

# Manage local and cloud AI models with LLM Gateway

LLM Gateway gives your Olares device a single control panel to connect, manage, and route requests to both cloud-based AI services (such as OpenAI or Anthropic) and locally hosted models.

## Learning objectives

In this guide, you will learn how to:
- Install LLM Gateway from the Market.
- Add cloud and local AI model providers.
- Configure system and personal default models.
- Issue API keys and manage user quotas.
- Manage client applications.
- Monitor model usage and spend logs.

## Prerequisites

- You have admin privileges to install the app and access its management features.
- You have the API keys for any external cloud providers you want to connect.

## Install LLM Gateway

1. Open Market and search for "LLM Gateway".
   <!-- ![LLM Gateway](/images/manual/use-cases/llm-gateway.png#bordered) -->
2. Click **Get**, then **Install**, and wait for installation to complete.

Click **Open** to launch the app. You're logged in automatically.

## Add a model provider

A provider represents your connection to an AI service. Add cloud providers manually or install local AI models directly from the Olares ecosystem.

<tabs>
<template #Local-Model>

Install and connect local models running directly on your Olares device.

1. Open LLM Gateway.
2. Go to **Providers**.
3. Click **INSTALL FROM MARKET**.
4. Choose a local model application from the catalog.
5. Click **Install**.

:::info
Local models do not require external credentials. The Gateway configures the connection automatically when the installation completes.
:::
</template>
<template #Cloud-Provider>

Add external services like OpenAI, Anthropic, or DeepSeek.

1. Open LLM Gateway.
2. Go to **Providers** in the left sidebar.
3. Click **CREATE PROVIDER (MANUAL)**.
4. Choose your preferred provider from the catalog, or select a custom OpenAI-compatible endpoint.
5. Enter your provider credentials (such as your API Key).
6. Click **CREATE**.

    <!--![Add a manual cloud provider](/images/manual/use-cases/llmgateway-add-manual-provider.png#bordered){width=70%}-->

:::tip
For your security, the application masks your credentials after you save them. To update an existing key, enter the new key directly. Leaving the field blank removes the stored credential entirely.
:::
</template>
</tabs>

## Manage models and defaults

After adding a provider, manage the specific models available to your users and set fallbacks for routing requests.

### Enable or disable models

1. Go to **Models**.
2. Locate the model you want to manage.
3. Toggle the status switch to enable or disable the model.

:::info
Disabling a model prevents users from routing new requests to it, but it does not delete the model configuration from your system.
:::

### Set default models

Configure which model the system uses when an application or user does not specify one.

1. Open **Settings**.
2. Go to the **My default models** section.
3. Specify your preferred model for each category (e.g., Chat, Embedding, Text-to-Speech).
4. Click **SAVE**.

:::tip
Administrators set system-wide defaults in the **Models** section. Your personal user settings always override the system defaults.
:::

## Generate and manage API keys

Users require an API key to authenticate their requests through the LLM Gateway.

1. Go to **API Keys**.
2. Click **ISSUE NEW KEY**.
3. Select the target user as the owner of the key, and then click **NEXT**.
4. Enter a descriptive name for the key, and then click **NEXT**.
5. (Optional) Configure a qupta for this key. Skip to inherit the default quota. Click **NEXT**.
6. Review the specified info, and then click **ISSUE**.
7. Copy the generated API key.

:::warning
The application displays the API key in plain text only once. Ensure you copy and store it securely before you close the window.
:::

## Monitor usage and manage quotas

Track how your resources are used and enforce limits to control costs and system load.

### View spend logs

1. Go to **Spend Logs**.
2. Specify your filter criteria, such as the date range, specific user, or provider ID.
3. Review the detailed ledger of all requests, token usage, and associated costs.

### Configure user quotas

Administrators enforce specific request and cost limits for individual users.

1. Go to **Users**.
2. Expand a user's row to view the details.
3. Click **ADD QUOTA**.
4. Specify the following settings:

    - **Dimension**: Select the limit type — **max_rpm** (requests per minute, sliding 60-second window), **max_tpm** (tokens per minute), or **max_budget_usd** (spend in US dollars).
    - **Limit**: Enter the maximum value allowed for the selected dimension.
    - **Soft threshold (%)**: Optional. Enter the percentage of the limit at which the system warns you before enforcing the hard limit.
5. Click **SAVE**.

    <!--![Configure user quotas inline](/images/manual/use-cases/llmgateway-add-user-quota.png#bordered){width=70%}-->

## Manage client apps

Olares applications automatically connect to the LLM Gateway using their internal application IDs. The Gateway registers these applications automatically upon their first request; you do not need to create them manually.

Administrators monitor these applications, adjust their usage limits, or permanently revoke their access.

1. Open LLM Gateway.
2. Go to **Client Apps**.
3. Select an application from the list to view its details.
4. Choose one of the following management actions:
   * **Configure quotas:** Set specific request or cost limits for the application.
   * **View spend:** Monitor the application's token usage and costs.
   * **Archive:** Revoke the application's access to the Gateway.

:::warning
Archiving a client application is a permanent action. If you archive an application, it permanently loses access to the Gateway and you cannot restore it.
:::

## FAQs

## Learn more

- [Use LiteLLM as a unified AI model gateway](litellm.md): Unify multiple AI model providers behind a single OpenAI-compatible API.
- [Set up Bifrost as an AI model gateway](bifrost.md): Aggregate Ollama and single-model apps behind one endpoint.