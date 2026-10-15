---
outline: [2, 3]
description: Use Local Deep Research on Olares with AI Router and SearXNG to research a question, inspect cited sources, and save a report as Markdown or PDF.
head:
  - - meta
    - name: keywords
      content: "Olares, Local Deep Research, AI research, AI Router, SearXNG, citations, research reports, local models"
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-15"
---

# Research a topic with cited sources using Local Deep Research

Local Deep Research is an open-source AI research assistant that combines web search with a language model. It develops follow-up questions, gathers sources, and turns its findings into a report with citations. On Olares, it uses Router for model access and SearXNG for web search, while account databases and research history remain in application storage.

The example in this guide explains how heat pumps work. It uses Quick Summary with one search iteration, so you can check the complete workflow before starting a longer investigation.

## Learning objectives

In this guide, you will learn how to:

- Configure model access and web search.
- Create an application account and run a research question.
- Check citations, reopen saved research, and export a report.
- Validate the setup and identify model or search failures.

## Prerequisites

Before you begin, you need:

- An Olares device running Olares 1.12.7 or later, with sufficient memory and disk space.
- Access to [Router](olares-router.md) for local model connections.
- SearXNG installed from Market.
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

Local Deep Research searches the web through SearXNG. The installation form in the next step requires the SearXNG URL.

1. Open Olares **Settings**, go to **Applications** > **SearXNG** > **Entrances** > **SearXNG**, and then copy the **Endpoint** URL.
2. Open SearXNG from the Launchpad, search for `how heat pumps work`, and then ensure that the page contains actual results.

## Install Local Deep Research

1. Open Market and search for "Local Deep Research".
2. Click **Get**, then click **Install**.
3. In **SEARXNG_URL**, enter the SearXNG URL you copied, and click **Confirm**.
4. Wait for the installation to complete.

## Create a user account

1. Open Local Deep Research from the Launchpad.
2. On the login page, click **Register here**.
3. Enter a username and password, then confirm the password.
4. Read the notice about encrypted data, select the acknowledgment checkbox, and click **Create Encrypted Account**. You log in automatically.

:::warning Keep your password safe
Your password protects the encrypted account database. There is no password recovery for that database. Store the password securely before adding research or documents you need to keep.
:::

## Use Local Deep Research

### Choose a research mode

A research mode controls how thoroughly Local Deep Research investigates a question. Start with Quick Summary. Confirm that the selected model and search engine work before increasing the scope.


| Mode | Use it for |
|:---|:---|
| **Quick Summary** | An initial answer and a short setup check |
| **Detailed Report** | A longer investigation organized into multiple sections |
| **Chat Mode** | Follow-up questions in an experimental conversational workflow |

### Run your first research question

1. On the **New Research** page, enter the following message in **Research Query**:

   ```text
   How do heat pumps work? Give a short explanation with source citations.
   ```

2. Select **Quick Summary**.
3. Locate the **Advanced Options** section, and configure the following settings:

   | Setting | Value |
   |:---|:---|
   | **Model Provider** | OpenAI-Compatible Endpoint |
   | **Language Model** | `default-chat` |
   | **Search Engine** | SearXNG |
   | **Search Iterations** | `1` |
   | **Questions Per Iteration** | `1` |
   | **Search Strategy** | Source-Based |

   :::info
   Model and endpoint fields supplied through application environment settings can appear locked here. This is expected. Change those values in Olares Settings instead.
   :::

4. Click **Start Research**. Track the progress page while the model generates questions, searches for sources, and prepares the report.
5. When processing finishes, click **View Results**.

   ![Research in progress with the View Results button](/images/manual/use-cases/local-deep-research-query-1.png#bordered)

One iteration can search both your original question and a generated follow-up question. Increasing iterations expands the investigation and adds model calls and search requests.

### Review and export the report

1. Read the report and check that it contains an answer, not an error message or troubleshooting report.
2. Open a citation and compare the claim with the source. Check the publication date, publisher, and whether the cited page supports the statement.
3. Click **Back to History**, then reopen the research to check that it was saved.
4. To save a text copy, click **Export** > **Markdown**.
5. To save a document for reading or sharing, click **Download PDF**. Open the downloaded file and check the report and references.

   ![Research report with citations and source list](/images/manual/use-cases/local-deep-research-report.png#bordered)

A citation makes a claim traceable. It does not establish that the claim is correct. Review the underlying sources before using the report for a decision.

## Validate your setup

### Check the core workflow

Use the heat-pump question above for these checks. The checks below cover the Quick Summary workflow. They do not validate every research mode or document-indexing option. Judge the actual report content, rather than the progress percentage alone.

| Test | Action | Expected result |
|:---|:---|:---|
| Account access | Sign out, then sign in with your application account | The account opens successfully |
| Model and search connection | Run the one-iteration Quick Summary | A substantive answer with citations is produced, and no authentication error appears |
| Source review | Open citations for several key claims | Sources are accessible and support the cited claims |
| Saved history | Reload the report, then reopen it from **History** | The query, report, and references are retained |
| Markdown export | Download and open the Markdown file | The file contains the report and source references |
| PDF export | Download and open the PDF | The document opens and its report and references are readable |

In a reference run with Qwen3.6-27B MTP (llama.cpp) through Router, the heat-pump example completed in about 130 seconds and saved eight sources. History, Markdown export, and a two-page PDF were checked successfully. Completion time, source count, and document length vary with the model, search results, and settings. These figures are observations, not pass criteria. Source review remains a separate check.

### Try other research questions

These are additional examples for evaluating your own setup. They are not benchmark results or guaranteed answers.

| Type | Example question | What to check |
|:---|:---|:---|
| Technology comparison | Compare PostgreSQL and MySQL for a transactional web application. Discuss concurrency, indexing, and replication, using official documentation where possible. | Balanced comparison, identifiable versions, and claims supported by current documentation |
| Historical explanation | What factors contributed to the fall of the Western Roman Empire? Separate scholarly explanations from disputed claims and cite sources. | Multiple causes, uncertainty acknowledged, and sources support the interpretation |
| Scientific explanation | Why do lithium-ion batteries lose capacity over time? Distinguish calendar aging from cycle aging and cite authoritative sources. | Clear distinction between mechanisms, suitable technical sources, and no unsupported numerical claims |

To request Chinese output, add `Answer in Chinese and retain source citations` to your question. Compare the sources as well as the fluency of the answer.

## Change the model connection

To choose a specific Router model, open that model in Router and copy its model identifier from **How to call this model**. Use the complete Router identifier, including its provider prefix. A raw model name displayed by the model application's Launchpad page might omit this prefix.

Open **Settings** > **Applications** > **Local Deep Research**, find the application environment variables, and set `LDR_LLM_MODEL` to the copied identifier. For example, a Router instance serving the model used in the reference run exposes:

```text
Olares/unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL
```

Use this example only when that model is available in your Router. Keep the endpoint override and API key empty to continue calling through Router. Wait for the application to finish applying the settings, then sign in again if prompted.

For a different OpenAI-compatible service, set its model name, API base URL including `/v1`, and API key if required. A provider's API key authenticates model requests. It does not replace your Local Deep Research account password.

## Troubleshoot common issues

| Symptom | What to check |
|:---|:---|
| `401`, invalid API key, or a model authentication error | Check the provider credentials in Router. Leaving the application's key empty is correct for the default Olares Router connection. Select another working Router model when needed. |
| Model not found | Copy the complete callable identifier from Router. Do not assume a raw local model name or another instance's identifier is valid. |
| Search returns no sources | Check `SEARXNG_URL` and try the question in SearXNG itself. Inspect search-engine errors and retry after a delay if an upstream engine is rate-limited. |
| Progress shows completion but the report contains an error | Read the report and research log. Resolve the model or search error and run the question again. A completed progress indicator is not proof of a successful answer. |
| Model fields are locked in the research form | Edit the application environment variables in Olares Settings, then wait for the settings to apply. |
| You are asked to log in after a restart | Sign in with the application account. A new login prompt does not by itself mean saved research was deleted. |

Web research sends your question to the selected search services. A cloud model also receives the content used for inference. Using a local model keeps model inference on Olares, but does not make public web search offline.

## Learn more

- [Use Olares Router as your AI gateway](olares-router.md): Understand model identifiers, access, and provider configuration.
- [Run local LLMs on Olares](llm-base-apps.md): Set up a local model for Router.
- [Shared applications](../manual/olares/market/shared-apps.md): Understand shared application access on Olares.
