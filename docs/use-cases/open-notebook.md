---
outline: [2, 3]
description: Install and use Open Notebook on Olares to collect sources, generate AI insights, chat with your knowledge base, create notes, and generate podcasts from your research materials.
head:
  - - meta
    - name: keywords
      content: Olares, Open Notebook, AI notebook, research assistant, sources, notes, RAG, knowledge base, podcast, transformations
app_version: "1.0.4"
doc_version: "1.0"
doc_updated: "2026-04-30"
---

# Build a research notebook with Open Notebook

Open Notebook is an AI-powered research workspace for collecting source materials, generating structured insights, chatting with your knowledge base, and turning research into editable notes or podcast episodes.

This guide walks you through your first complete workflow in Open Notebook. To make the workflow easier to follow, it uses an AI research project as an example. You can use the same workflow for papers, courses, meeting notes, market research, product research, or any other topic.

## Learning objectives

In this guide, you will learn how to:

- Install Open Notebook on Olares.
- Connect AI models for chat, summaries, retrieval, and podcast generation.
- Create a research notebook.
- Add and process research sources.
- Review AI-generated insights.
- Chat with your research materials.
- Save useful AI responses as editable notes.
- Generate a podcast episode from selected sources and notes.

## Prerequisites

Before you begin, make sure you have:

- Access to at least one AI model provider, such as Ollama, OpenAI Compatible, OpenAI, Google AI, or another supported provider.
- Access to an embedding model for vector search and retrieval.
- Optional: access to a speech-to-text model if you want to process audio or video sources.
- Optional: access to a text-to-speech model if you want to generate podcasts.

:::info Recommended local AI services
For local AI workflows on Olares, you can use [Ollama](ollama.md) for language and embedding models, and [Speaches](speaches.md) for speech-to-text and text-to-speech.

Open Notebook can also use other supported cloud or OpenAI-compatible providers.
:::

## How Open Notebook works

Open Notebook organizes your work around four main content types:

| Content type | Description |
| :-- | :-- |
| **Notebook** | A workspace for one research topic or project. |
| **Source** | Original material added to Open Notebook, such as a file, web page, audio,<br> video, or pasted text. |
| **Insight** | AI-generated output created from a source by a transformation, such as a <br>summary or key takeaways. |
| **Note** | Editable knowledge saved inside a notebook. A note can be written manually,<br> saved from an AI response, or created from an insight. |

In this guide, you will create a notebook for a sample AI research project, add sources about AI, generate summaries and insights, chat with the materials, save useful outputs as notes, and turn selected materials into a podcast.

## Install Open Notebook

1. Open Market and search for "Open Notebook".

   ![Open Notebook in Market](/images/manual/use-cases/open-notebook.png#bordered){width=90%}

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, configure the required models before starting your first research notebook.

## Connect AI models

Open Notebook needs AI models to summarize sources, answer questions, retrieve source context, and generate podcasts. You only need to set this up once.

Go to **Manage** > **Models**. Setting up models has four main steps:

### Get provider endpoints

If you use local Olares apps, such as Ollama, Speaches, or Whisper-WebUI, as AI providers, copy their endpoints first.

1. Go to Olares **Settings** > **Applications** and click the app.
2. Look for **Shared entrances** and copy the endpoint.
3. If **Shared entrances** is not available for the app, copy its **API Entrance** or standard entrance instead.

:::tip Endpoint format
Use the endpoint format required by the provider:

- **Ollama**: Use the endpoint exactly as copied. Do not append `/v1`.
- **OpenAI-compatible providers**, such as Speaches or Whisper-WebUI: Append `/v1` to the endpoint.

For example, if the Speaches endpoint is `http://edd26bab0.shared.olares.com`, configure it in Open Notebook as `http://edd26bab0.shared.olares.com/v1`
:::

### Add a provider configuration

1. In Open Notebook, go to **Manage** > **Models** and find the provider you want to use.
2. Click **Add Configuration** under the provider, such as Ollama or OpenAI Compatible.
3. Enter a name for the configuration.
4. Paste the endpoint URL you copied.
5. Enter an API key if required.
6. Click **Add Configuration**.

### Add models

After adding a provider configuration, add the models that Open Notebook can use from it.

1. In the provider configuration you added, click **Models**.
2. Select the model type, such as **Language**, **Embedding**, **TTS**, or **STT**.
3. Select the available models you want to use.
4. Click **Add**.

### Set default models

Go back to **Default Model Assignments** at the top of **Manage** > **Models**. Here, you tell Open Notebook which model to use for each task. Assign the models you added to the slots you need:

| Slot | What to select  |
| :--| :-- |
| Chat Model   | A language model. |
| Embedding Model | An embedding model.  |
| Text-to-Speech Model | A TTS model, if you generate podcasts. |
| Speech-to-Text Model | An STT model, if you process audio or video. |
| Transformation Model | A language model. |
| Tools Model | A language model, if you use tool-based tasks. |
| Large Context Model | A language model suitable for long documents. |

If **Auto-assign Defaults** is available, you can use it to fill available slots automatically, then review the selections.

![Model assignments](/images/manual/use-cases/open-notebook-set-models-result.png#bordered)

## Create your first research notebook

A notebook is the workspace for one topic, project, course, or research question. In this guide, you will create a notebook for learning about generative AI.

1. Go to **Process** > **Notebooks**.
2. Click **New** > **Notebook**.
3. Enter the notebook name and description.
    :::tip Write a useful description
    The notebook description helps the AI understand the context of your project. Describe the topic, purpose, and expected use of the notebook as clearly as possible.
    :::
4. Click **Create New Notebook**.

![Create a notebook](/images/manual/use-cases/open-notebook-create-a-notebook.png#bordered){width=70%}

## Add your first sources

Sources are the original materials you want Open Notebook to process. For your first run, use a lightweight text source so you can avoid failures caused by external websites, large PDFs, or unavailable video transcripts.

This guide provides sample text about generative AI. Add it as a **Text** source. After the first workflow succeeds, you can add more text sources or try external URLs, PDFs, YouTube videos, or audio/video files.

### Add a text source

1. Open the notebook you just created.
2. Click **Add Source** > **Add Source**.
3. Click the **Enter Text** tab.
4. Paste the following content:
   ```plain
   Generative AI is a type of artificial intelligence that can create new content based on patterns learned from data. It can generate text, images, audio, video, code, and other forms of content. In everyday work, generative AI is often used to draft documents, summarize long materials, rewrite text for different audiences, brainstorm ideas, create outlines, and answer questions.

   A common example is using a language model to help write a product document. The user can provide rough notes, requirements, or meeting records, and the model can turn them into a clearer draft. The user still needs to review the result, check accuracy, and decide whether the writing fits the intended audience.

   Generative AI is useful because it can reduce the time spent on repetitive writing and analysis tasks. It can also help users explore unfamiliar topics by explaining concepts, comparing viewpoints, and suggesting follow-up questions.

   However, generative AI has limitations. It may produce inaccurate or unsupported statements. It may miss important context. It may also sound confident even when the answer is incomplete or wrong. For this reason, users should treat generative AI as an assistant rather than a final authority.

   A reliable workflow is to combine AI output with source verification. Users can collect original materials, generate summaries, ask targeted questions, save useful answers as notes, and manually review the final result before using it in real work.
   ```

5. Enter the following title:
   ```plain
   What generative AI can do
   ```
6. Click **Next**.
7. Link the source to your notebook if prompted.
8. Click **Next** to open the processing settings.
9. Under **Transformations**, select **Dense Summary** for the first run. 

   You can choose a different transformation depending on your goal:
   
   | Transformation | Use it when |
   |:---------------|:------------|
   | **Dense Summary** | You want a compact, information-rich overview of the source. <br>Recommended for the first run. |
   | **Simple Summary** | You want a shorter, easier-to-read summary before deciding<br> whether the source is worth reading in detail. |
   | **Key Insights** | You want the main takeaways, important claims, or notable <br>findings. |
   | **Paper Analysis** | You are processing an academic paper and want a structured<br> analysis of research question, method, findings, and limitations. |
   | **Reflection Questions** | You want questions for deeper thinking, discussion, or<br> follow-up research. |
   | **Table of Contents** | You want to understand the structure of a long source. |

   :::tip
   For your first run, select only one transformation. You can open the source's **Insights** tab later and click **Generate New Insight** to apply another transformation.
   :::
10. Keep **Enable search vector embedding** selected.
11. Click **Done**.

![Add first source](/images/manual/use-cases/open-notebook-add-first-source.png#bordered){width=70%}

Open Notebook starts processing the source. After processing completes, the source becomes available for insights, chat, notes, and citations.

:::info Processing time
Processing time depends on source size, selected transformations, model speed, and available hardware resources. Small web pages may finish quickly, while large files may take longer.
:::

### Add other source types

After the first workflow succeeds, you can add other materials in a similar way.

| Source type | Supported content |
| :-- | :-- |
| **Upload file** | Documents, images, archives, and media files. Audio or video files require a Speech-to-Text model. |
| **Add URL** | Web pages and other supported online content. |
| **Enter text** | Content pasted or typed directly. |

:::warning Avoid heavy processing
When using local models, especially on a single-GPU setup, avoid processing too many large sources or applying too many transformations at the same time. This may cause slow processing, timeouts, or failed tasks.

For better stability, process one source with one transformation first. You can generate additional insights later from the source's **Insights** tab by using **Generate New Insight**.
:::

## Review generated insights

Insights are AI-generated outputs created from sources. For example, **Dense Summary** helps you quickly understand what a source is about before reading it in detail.

### View an insight

1. Open a processed source.
2. Click the **Insights** tab.
3. Click **View Insight** to review the generated insight.

![Review insight](/images/manual/use-cases/open-notebook-review-insight.png#bordered){width=90%}

Use the insight to decide whether the source is useful and whether it should be included in notebook chat.

### Generate another insight

If you want to analyze the same source in another way:

1. Open the processed source.
2. Click the **Insights** tab.
3. Under **Generate New Insight**, select a transformation.

   ![New insight](/images/manual/use-cases/open-notebook-new-insight.png#bordered){width=90%}

4. Click **New**.

## Chat with your research materials

After your sources are processed, you can ask questions based on the materials in your notebook.

1. Open your notebook.
2. In **Chat with Notebook**, select the model you want to use.
3. Click the icon next to each source to change the source context level.
   
      | Icon | Context level | Recommended use |
      | :-- | :--  | :-- |
      | <i class="material-symbols-outlined">news</i> | Full content: The AI can use the<br> full source content. | Use for the most important sources when you need detailed answers and citations. |
      | <i class="material-symbols-outlined">lightbulb_2</i> | Insights only: The AI can use<br> generated summaries or insights. | Use for background sources when a summary is enough. |
      | <i class="material-symbols-outlined">visibility_off</i>| Not included in chat: The AI cannot<br> use this source.| Use for irrelevant, sensitive, or unnecessary sources.  |
   
4. Enter your question and send it.

![Chat with AI](/images/manual/use-cases/open-notebook-chat.png#bordered){width=90%}

Open Notebook answers based on the sources included in the current chat context.

:::tip Verify citations
When an answer includes citations, click them to open the referenced source passages. Compare the answer with the original content to check whether the AI response is supported by your sources.
:::

## Create notes

Notes are editable knowledge items inside a notebook. Use them to keep summaries, outlines, questions, drafts, or your own conclusions.

### Save an AI answer as a note

When you receive a useful answer in chat:

1. Click the <i class="material-symbols-outlined">save</i> icon under the AI response.
2. After the note appears in **Notes** with the `AI Generated` tag, open it.
3. Enter a note title, then click **Save Note**.

You can edit the note later, use it as part of future notebook context, or include it in podcast generation.

### Create a note manually

You can also create notes manually.

1. Open your notebook.
2. Go to the **Notes** area.
3. Click **Write Note**.
4. Enter a title and write the note content. Markdown is supported.
5. Click **Create Note**.

Your note appears in the **Notes** area with a `Human` tag.

## Generate a podcast

After you have sources, insights, and notes, you can turn your research materials into a podcast episode.

Podcast generation requires:

- A language model for outline generation.
- A language model for script generation.
- A text-to-speech model for audio generation.
- Processed sources or notes to use as context.

### Configure podcast profiles

Before generating a podcast, make sure the podcast profiles have the required models and voices configured.

1. Go to **Create** > **Podcasts**, then click the **Profiles** tab.
2. Open any profile marked with **Needs Configuration** or a warning icon.
3. For a **Speaker Profile**, select a voice model and a supported voice for each speaker.
4. For an **Episode Profile**, select the speaker profile, outline model, transcript model, language, segment count, and briefing.
5. Save your changes.

:::tip Start simple
For your first podcast, use one speaker and a short briefing. After the first episode works, you can try multi-speaker formats or more detailed instructions.
:::

### Generate the audio

1. Click **Episodes** tab and click **Generate Podcast**.
2. Select the sources or notes to include.
3. Select the episode profile.
4. Set the episode name.
5. Add extra instructions if needed.

   ![Generate podcast](/images/manual/use-cases/open-notebook-generate-podcast.png#bordered){width=90%}

6. Click **Generate**. 

After the episode is complete, you can:

- Play it in the browser.
- Download the audio file.
- Review the generated transcript, if available.

:::info Generation time
Podcast generation can take several minutes. Text-to-speech is usually the slowest stage, especially for longer episodes or multi-speaker podcasts.
:::

## Explore: Search across your knowledge base

Go to **Process** > **Ask and Search** when you want to find information across your sources and notes.

### Ask a question

Use **Ask** when you want a complete answer instead of reviewing search results manually.

1. Open **Process** > **Ask and Search**.
2. Click the **Ask** tab.
3. Enter your question.

   Example:

   ```plain
   Based on my sources, what are the main benefits and risks of using generative AI in business?
   ```

4. Click **Ask**.

### Search for source fragments

Use **Search** when you want to find and inspect matching content yourself.

1. Open **Process** > **Ask and Search**.
2. Click the **Search** tab.
3. Choose a search type:
   - **Text Search**: Use this when you remember exact words or phrases.
   - **Vector Search**: Use this when you remember the meaning but not the exact wording.
4. Enter your query.


:::warning Embedding model required
Vector search requires a configured Embedding Model. The source also needs embeddings enabled during processing. If embeddings are missing or incorrectly configured, vector search may return no useful results.
:::

## FAQs

### Why does vector search return no useful results?

Vector search requires:

- A configured **Embedding Model**.
- Sources processed with embeddings enabled.

If vector search returns no useful results:

1. Go to **Manage** > **Models** and check the **Embedding Model** assignment.
2. Make sure the selected model is an embedding model, not a chat model.
3. Check whether the source was processed with **Enable search vector embedding** selected.
4. Reprocess the source if needed.

### Why does podcast generation fail?

Podcast generation may fail if required models are missing, selected sources are not ready, the TTS model is unavailable, or the script language does not match the selected voice.

Check the following:

- The episode profile has an outline model and a transcript model.
- The speaker profile has a valid TTS model and voice.
- The selected sources are processed and ready.
- The podcast language matches the selected TTS voice.
- The TTS provider is running.

### Why is processing slow or unstable when I add multiple sources or transformations?

When using local models, especially on a single-GPU setup, processing multiple sources or applying multiple transformations at the same time can overload available resources. This may cause slow processing, duplicate generation, timeouts, or failed tasks.

For better stability, process one source with one transformation first. You can generate additional insights later from the source's **Insights** tab by using **Generate New Insight**.

## Learn more

- [Ollama](ollama.md): Download and run local language models on Olares.
- [Speaches](speaches.md): Set up local speech-to-text and text-to-speech services.
- [Manage GPU resources](/manual/olares/settings/single-gpu.md): Allocate GPU resources for local AI apps.