---
outline: deep
description: Run Open Notebook on Olares to collect research sources, generate AI insights, chat with your knowledge base, take notes, and create podcasts.
head:
  - - meta
    - name: keywords
      content: Olares, Open Notebook, AI notebook, research assistant, sources, notes, RAG, knowledge base, podcast, transformations
app_version: "1.0.4"
doc_version: "2.0"
doc_updated: "2026-07-22"
---

# Build a research notebook with Open Notebook

Open Notebook is an AI-powered research workspace for collecting source materials, generating structured insights, chatting with your knowledge base, and turning research into editable notes or podcast episodes.

This guide walks you through your first complete Open Notebook workflow using an AI research project as an example. You can apply the same workflow to papers, courses, meeting notes, market research, product research, or other topics.

## Learning objectives

In this guide, you will learn how to:

- Install Open Notebook on Olares.
- Set up AI models for chat, summaries, retrieval, and podcast generation.
- Create a research notebook.
- Add and process research sources.
- Review AI-generated insights.
- Chat with your research materials.
- Save useful AI responses as editable notes.
- Generate a podcast episode from selected sources and notes.

## Prerequisites

Before you begin, make sure you have access to the models you want to use.

- Required: At least one language model and an embedding model for vector search.
- Required for podcast generation: A text-to-speech (TTS) model.
- Optional: A speech-to-text (STT) model for processing audio or video sources.

:::info Recommended local AI services
For local AI workflows on Olares, you can use local language or embedding model services, and [Speaches](speaches.md) for speech-to-text and text-to-speech.
:::

## How Open Notebook works

Open Notebook organizes your work around four main content types:

| Content type | Description |
| :-- | :-- |
| **Notebook** | A workspace for one research topic or project. |
| **Source** | Original material added to Open Notebook, such as a file, web page, audio,<br> video, or pasted text. |
| **Insight** | AI-generated output created from a source by a transformation, such as a <br>summary or key takeaways. |
| **Note** | Editable knowledge saved inside a notebook. A note can be written manually,<br> saved from an AI response, or created from an insight. |

In this guide, you will create a sample AI research notebook, add sources, generate insights, chat with the materials, save notes, and create a podcast.

## Install Open Notebook

1. Open Market and search for "Open Notebook".

   ![Open Notebook in Market](/images/manual/use-cases/open-notebook.png#bordered){width=90%}

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, configure the required providers and models before starting your first research notebook.

## Set up AI models

Open Notebook uses AI models for summaries, chat, retrieval, and podcast generation. You only need to set them up once.

### Get provider connection details

How you get the connection details depends on whether you connect a standalone model or another Olares app.

#### Connect a standalone model

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

For each standalone model used in this guide:

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

In this case, we use `qwen3.5-9b` and `qwen3-embedding:0.6b`. Open Notebook connects to them through the **Ollama** provider, so view the **Ollama** format in each Model Console and copy the corresponding Base URL.

#### Connect an app

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

This guide uses Speaches as the TTS and STT provider:

1. Go to Olares **Settings** > **Applications** > **Speaches** > **Entrances**.
2. Select **Speaches API**, then copy the **Endpoint** URL.

### Add provider configurations

Go to **Manage** > **Models**. For each service, find the matching provider and click **Add Configuration**.

| Service | Provider | Base URL |
| :-- | :-- | :-- |
| Qwen language model | **Ollama** | Base URL from its Model Console |
| Qwen embedding model | **Ollama** | Base URL from its Model Console |
| Speaches | **OpenAI Compatible** | Speaches endpoint with `/v1` appended |

Enter a recognizable configuration name and the Base URL. If an API key is required, enter `olares`, then save the configuration.

### Add models

In each configuration, click **Models** and add the following models:

| Configuration | Type | Model ID |
| :-- | :-- | :-- |
| Qwen language model | **Language** | `qwen3.5-9b` |
| Qwen embedding model | **Embedding** | `qwen3-embedding:0.6b` |
| Speaches | **TTS** | `speaches-ai/Kokoro-82M-v1.0-ONNX` |
| Speaches | **STT** | `Systran/faster-whisper-small` |

### Assign default models

Under **Default Model Assignments**, assign the models as follows:

| Slot | Model |
| :-- | :-- |
| Chat Model | `qwen3.5-9b` |
| Embedding Model | `qwen3-embedding:0.6b` |
| Text-to-Speech Model | `speaches-ai/Kokoro-82M-v1.0-ONNX` |
| Speech-to-Text Model | `Systran/faster-whisper-small` |
| Transformation Model | `qwen3.5-9b` |
| Tools Model | `qwen3.5-9b` |
| Large Context Model | `qwen3.5-9b` |

If **Auto-assign Defaults** is available, you can use it to fill the slots automatically, then review the selections.

![Model assignments](/images/manual/use-cases/open-notebook-set-models-result1.png#bordered)

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
2. In the **Source** area, click **Add Source** > **Add Source**.
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
   | **Simple Summary** | You want a shorter summary before reading in detail. |
   | **Key Insights** | You want main takeaways, claims, or findings. |
   | **Paper Analysis** | You are processing an academic paper. |
   | **Reflection Questions** | You want questions for discussion or follow-up research. |
   | **Table of Contents** | You want to understand the structure of a long source. |

10. Keep **Enable search vector embedding** selected.
11. Click **Done**.

![Add first source](/images/manual/use-cases/open-notebook-add-first-source.png#bordered){width=70%}

Open Notebook starts processing the source. When processing finishes, you can use it for insights, chat, notes, and citations.

### Add other source types

After the first workflow succeeds, you can add other materials in a similar way.

| Source type | Supported content |
| :-- | :-- |
| **Upload file** | Documents, images, archives, and media files. <br>Audio or video files require a Speech-to-Text model. |
| **Add URL** | Web pages and other supported online content. |
| **Enter text** | Content pasted or typed directly. |

:::warning Avoid heavy processing
When using local models, process one source with one transformation first. Processing many large sources or applying multiple transformations at the same time may cause slow processing, timeouts, or failed tasks.

You can generate additional insights later from the source's **Insights** tab by using **Generate New Insight**.
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
3. Click the icon next to each source to choose how much of each source the AI can use:
      | Icon | Context level | Recommended use |
      | :-- | :--  | :-- |
      | <i class="material-symbols-outlined">news</i> | Full content | Use for the most important sources when you need detailed answers and citations. |
      | <i class="material-symbols-outlined">lightbulb_2</i> | Insights only | Use for background sources when a summary is enough. |
      | <i class="material-symbols-outlined">visibility_off</i>| Not included in chat| Use for irrelevant, sensitive, or unnecessary sources.  |
   
4. Enter your question and send it.

![Chat with AI](/images/manual/use-cases/open-notebook-chat.png#bordered){width=90%}

Open Notebook answers based on the sources included in the current chat context.

:::tip Verify citations
When an answer includes citations, click them to open the referenced source passages. Compare the answer with the original content to check whether the AI response is supported by your sources.
:::

## Create notes

Notes are editable items for summaries, outlines, questions, drafts, or conclusions.

### Save an AI answer as a note

When you receive a useful answer in chat:

1. Click the <i class="material-symbols-outlined">save</i> icon under the AI response.
2. In the **Notes** area, click the saved note with the `AI Generated` tag to review it.
  
   ![AI generated note](/images/manual/use-cases/open-notebook-ai-note.png#bordered){width=70%}

3. Update the title or content when needed, then click **Save Note**.

You can use saved notes as part of future notebook context, or include them in podcast generation.

### Create a note manually

You can also create notes manually.

1. Open your notebook.
2. Go to the **Notes** area.
3. Click **Write Note**.
4. Enter a title and write the note content. Markdown is supported.
5. Click **Create Note**.

Your note appears in the **Notes** area with a `Human` tag.

![Manually created note](/images/manual/use-cases/open-notebook-manual-note.png#bordered){width=70%}

## Generate a podcast

After you have sources, insights, and notes, you can turn your research materials into a podcast episode.

Podcast generation requires:

- A language model for outline generation.
- A language model for transcript generation.
- A text-to-speech model for audio generation.
- Processed sources or notes to use as context.

### Configure podcast profiles

Before generating a podcast, configure the required models and voices in podcast profiles.

1. Go to **Create** > **Podcasts**, then click the **Profiles** tab.
2. Open any profile marked with **Needs Configuration** or a warning icon.
3. For a **Speaker Profile**, select a voice model and enter a voice ID supported by that model.

   :::warning Use a voice ID supported by the selected TTS model
   The default voice ID in a speaker profile may not be supported by your selected TTS model. For example, `nova` is not available in `speaches-ai/Kokoro-82M-v1.0-ONNX`. Use a supported Kokoro voice ID such as `af_heart` instead.
   :::

4. For an **Episode Profile**, select the speaker profile, outline model, transcript model, language, segment count, and briefing.
5. Save your changes.

### Generate the audio

1. Click the **Episodes** tab, then click **Generate Podcast**.
2. Select the sources or notes to include.
3. Select the episode profile.
4. Set the episode name.
5. Add extra instructions if needed.
   :::tip Match the language and voice
   Some TTS voices work best with specific languages. Make sure the podcast language matches the selected voice. If you use an English voice, add extra instructions such as: `Generate the entire podcast script in ENGLISH only.`
   :::
6. Click **Generate**. 

![Generate podcast](/images/manual/use-cases/open-notebook-generate-podcast.png#bordered){width=90%}

After the episode is complete, you can:

- Play it in the browser.
- Download the audio file.
- Review the generated transcript in **Details**.

![Generated podcast](/images/manual/use-cases/open-notebook-podcast-result.png#bordered){width=90%}


## Explore more features

### Search across your knowledge base

Go to **Process** > **Ask and Search** when you want to find information across your sources and notes.

#### Ask a question

Use **Ask** when you want a synthesized answer.

1. Open **Process** > **Ask and Search**.
2. Click the **Ask** tab.
3. Enter your question.

   Example:

   ```plain
   Based on my sources, what are the main benefits and risks of using generative AI in business?
   ```

4. Click **Ask**.

Open Notebook returns a synthesized answer based on matching content from your knowledge base.

![Ask a question](/images/manual/use-cases/open-notebook-ask-result.png#bordered){width=90%}

#### Search for source fragments

Use **Search** when you want to inspect matching fragments yourself.

1. Open **Process** > **Ask and Search**.
2. Click the **Search** tab.
3. Choose a search type:
   - **Text Search**: Use this when you remember exact words or phrases.
   - **Vector Search**: Use this when you remember the meaning but not the exact wording.
4. Enter your query.

:::warning Embedding model required
Vector search requires a configured Embedding Model. The source also needs embeddings enabled during processing. If embeddings are missing or incorrectly configured, vector search may return no useful results.
:::

### Customize transformations

Transformations are reusable AI prompts that turn source content into structured insights, such as summaries, key takeaways, paper analysis, or reflection questions.

:::tip How transformations are applied
Use **Manage** > **Transformations** to view, edit, test, or create templates.

To apply a transformation to a source:
- Select a transformation when adding a source.
- For an existing source, open the source, go to the **Insights** tab, and click **Generate New Insight**.
:::

You can either edit an existing transformation or create a new one.

#### Edit an existing transformation

Use this when a built-in transformation is close to what you need.

1. Go to **Manage** > **Transformations**.
2. Find the transformation you want to adjust, then click **Edit**.
3. Modify the title, description, or prompt.
4. Click **Edit Transformation** to save your changes.

Editing a transformation only changes how it works the next time you apply it. Existing insights are not updated automatically. To get a new result, run the transformation again from the source's **Insights** tab.

#### Create a new transformation

Use this when you want a separate template for a specific analysis task.

1. Go to **Manage** > **Transformations**.
2. Click **Create New**.
3. Enter a name, title, description, and prompt.
4. Click **Create New**.

   ![New transformation](/images/manual/use-cases/open-notebook-new-trans.png#bordered){width=90%}

#### Test a transformation

Before applying a transformation to real sources, test it with a short sample.

1. In **Manage** > **Transformations**, find the transformation you want to test.
2. Click **Playground**.
3. Paste a short excerpt from a source.
4. Run the transformation and review the output.
5. If the output does not meet your expectation, click **Edit** to refine the prompt, then test again.

:::tip Test before applying to full sources
Use **Playground** to check the output format, length, and accuracy before applying a transformation to full sources.
:::

## FAQs

### Why is processing slow or unstable?

Processing can be slow or unstable for two common reasons:

- You are processing multiple sources or applying multiple transformations at the same time.
- Other GPU-intensive apps are using GPU resources needed by Open Notebook or its configured model services.

To improve performance:

1. Process one source with one transformation first.
2. Stop or pause other GPU-intensive apps. Keep only the model services assigned in **Manage** > **Models** running. See [Manage accelerator resources](/manual/olares/settings/gpu-resource.md) for details.

### Why does vector search return no useful results?

Vector search requires:

- A configured **Embedding Model**.
- Sources processed with embeddings enabled.

If vector search returns no useful results:

1. Go to **Manage** > **Models** and check the **Embedding Model** assignment.
2. Make sure the selected model is an embedding model, not a language model.
3. Check whether the source was processed with **Enable search vector embedding** selected.
4. Reprocess the source.

### Why does podcast generation fail?

Podcast generation may fail if required models are missing, selected sources are not ready, the TTS provider is unavailable, the transcript language does not match the selected voice, or the speaker voice ID is not supported by the selected TTS model.

Check the following:

- The episode profile has an outline model and a transcript model.
- The speaker profile has a valid TTS model.
- Each speaker uses a voice ID supported by the selected TTS model.
- The selected sources are processed and ready.
- The podcast language matches the selected voice.
- The TTS provider is running.

If the error message lists supported voice IDs, update the speaker profile with one of those IDs, then create a new podcast generation task.

## Learn more

- [Ollama](ollama.md): Download and run local language models on Olares.
- [Speaches](speaches.md): Set up local speech-to-text and text-to-speech services.
- [Manage accelerator resources](/manual/olares/settings/gpu-resource.md): Allocate accelerator resources for local AI apps.
