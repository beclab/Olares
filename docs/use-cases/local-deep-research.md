---
outline: [2, 3]
description: Use Local Deep Research on Olares with AI Router and SearXNG to research a question, inspect cited sources, and save a report as Markdown or PDF.
head:
  - - meta
    - name: keywords
      content: "Olares, Local Deep Research, AI research, AI Router, SearXNG, citations, research reports, local models"
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-16"
---

# Research a topic with cited sources using Local Deep Research

Local Deep Research is an open-source AI research assistant that combines web search with a language model. Give it a question, and it develops follow-up questions, gathers sources, and turns its findings into a report with citations.

On Olares, it uses Router for model access and SearXNG for web search. This guide researches how heat pumps work using Quick Summary with one search iteration, so you can see the complete workflow before starting a longer investigation.

## Learning objectives

In this guide, you will learn how to:

- Prepare the model and search connections.
- Create a user account and run a research question.
- Check citations and export a report.

## Prerequisites

Before you begin, you need:

- An Olares device running Olares 1.12.7 or later, with sufficient memory and disk space.
- Access to [Router](olares-router.md) for local model connections.
- [SearXNG](searxng.md) installed from Market and running.
- The following model:

  | Model type | Model | How to get it |
  | :--- | :--- | :--- |
  | Chat | Qwen3.8-27B (llama.cpp) | Install from Market |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Check model availability

Local Deep Research calls the chat model through Router's `default-chat` route.

1. Open Router from Launchpad and go to **Default models**.
2. Confirm Qwen3.8-27B (llama.cpp) is selected as the default chat model and shows **Callable**.

## Prepare the SearXNG connection

Local Deep Research searches the web through SearXNG. You need the SearXNG URL to install Local Deep Research.

Open Olares **Settings**, go to **Applications** > **SearXNG** > **Entrances** > **SearXNG**, and then copy the **Endpoint** URL.

## Install Local Deep Research

1. Open Market and search for "Local Deep Research".
2. Click **Get**, then click **Install**.
3. When prompted, enter the SearXNG URL you copied and click **Confirm**.
4. Wait for the installation to complete.

## Create a user account

1. Open Local Deep Research from the Launchpad.
2. On the login page, click **Register here**.
3. Enter a username and password, then confirm the password.
4. Read the notice about encrypted data, select the acknowledgment checkbox, and click **Create Encrypted Account**. You log in automatically.

:::warning Keep your password safe
Your password protects the encrypted account database. There is no password recovery for that database. Store the password securely before adding research or documents you need to keep.
:::

## Run your first research question

1. On the **New Research** page, enter the following message in **Research Query**:

   ```text
   How do heat pumps work? Give a short explanation with source citations.
   ```

2. Select **Quick Summary**.

   :::tip Research modes
   **Quick Summary** returns an initial answer and works as a short setup check. **Detailed Report** runs a longer investigation organized into sections. **Chat Mode** is an experimental conversational workflow for follow-up questions. Start with Quick Summary and confirm that the selected model and search engine work before increasing the scope.
   :::

3. Locate the **Advanced Options** section, and configure the following settings:

   | Setting | Value |
   |:---|:---|
   | <nobr>**Model Provider**</nobr> | OpenAI-Compatible Endpoint |
   | <nobr>**Language Model**</nobr> | `default-chat` |
   | <nobr>**Search Engine**</nobr> | SearXNG |
   | <nobr>**Search Iterations**</nobr> | `1` |
   | <nobr>**Questions Per Iteration**</nobr> | `1` |
   | <nobr>**Search Strategy**</nobr> | Source-Based |

   :::info
   You cannot edit the model or endpoint fields here, because they are set by the application's environment variables. To change them, go to **Settings** > **Applications** > **Local Deep Research** > **Manage environment variables**.
   :::

4. Click **Start Research**. You can track the progress of the research while the model generates questions, searches for sources, and prepares the report.
5. When processing finishes, click **View Results**.

## Review and export the report

1. Read the report and confirm that it answers your question.
2. Open a citation and check that the source page supports the report's claim.
3. To save a copy, click **Export** and choose a format such as **Markdown**.
4. To save a document for reading or sharing, click **Download PDF**.

   ![Research report with citations and source list](/images/manual/use-cases/local-deep-research-report.png#bordered)

## (Optional) Use a specific model

By default, Local Deep Research calls the model through Router's `default-chat` route. To use a specific Router model instead:

1. Go to the **LLM** page in Router, find the model, ensure it shows **Callable**, and then click the **View connection example** icon.
2. In **How to call this model**, copy the complete identifier, including its provider prefix. For example, `Olares/unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL`.
3. Open Olares **Settings** > **Applications** > **Local Deep Research** > **Manage environment variables**, and set `LDR_LLM_MODEL` to the copied identifier.
4. Click **Apply**. The application restarts.
5. Open Local Deep Research and sign in again.

## FAQs

### Why do I see a `401`, invalid API key, or other model authentication error?

Check the provider credentials in Router. Leaving the application's key empty is correct for the default Olares Router connection. Select another working Router model when needed.

### Why is the model not found?

Copy the complete callable identifier from Router. Do not assume a raw local model name or another instance's identifier is valid.

### Why does the search return no sources?

Check `SEARXNG_URL` and try the question in SearXNG itself. Inspect search-engine errors and retry after a delay if an upstream engine is rate-limited.

### Why does the report contain an error even when progress shows completion?

Read the report and research log. Resolve the model or search error and run the question again. A completed progress indicator is not proof of a successful answer.

### Why are the model fields locked in the research form?

Edit the application environment variables in Olares Settings, then wait for the settings to apply.

### Why am I asked to log in after a restart?

Sign in with your user account. A new login prompt does not by itself mean saved research was deleted.

## Learn more

- [Use Olares Router as your AI gateway](olares-router.md): Understand model identifiers, access, and provider configuration.
- [Run local LLMs on Olares](llm-base-apps.md): Set up a local model for Router.
