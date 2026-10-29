---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/cognee
outline: [2, 3]
description: Build and explore document knowledge with Cognee on Olares
head:
  - - meta
    - name: keywords
      content: "Olares, Cognee, self-hosted, knowledge graph, documents, search"
app_version: "0.0.8"
doc_version: "1.0"
doc_updated: "2026-10-29"
---

# Connect project knowledge with Cognee

Cognee organizes text and documents into connected knowledge that you can search and explore. On Olares, the web interface, uploaded content and knowledge stores run on your own device. Language tasks use the default chat model configured in AI Router.

## Learning objectives

- Add a small document to a dataset.
- Process its contents into searchable knowledge.
- Ask a question and inspect the graph or source material.

## Prerequisites

- Olares 1.12.7 or later.
- A browser connected to your [Olares private network](../manual/larepass/private-network.md).
- AI Router with a working **default-chat** route.
- Internet access for the first download of Cognee's local embedding model.

This package uses a pinned upstream development image. Start with copies of your documents and keep the originals.

## Install Cognee

1. Open Market and install **Cognee**. Set an initial password with at least 12 characters when prompted.
2. Open Cognee from Launchpad.
3. Sign in as `default_user@example.com` using the password you set during installation.

A small embedding model runs on CPU inside Cognee. Its first download can delay the first processing job. Later runs reuse the saved model cache.

## Add a project document

Open **Brain**, select **New brain**, and name it `community_garden`. Use letters, numbers and underscores in the name. Select the new brain, choose **Paste text**, and add a few related facts. For example:

```text
The community garden opens on Saturday mornings.
Maya coordinates the volunteers.
The tool shed is beside the north entrance.
Volunteers meet Maya at the tool shed at 9:00.
```

Select **Add** to import the text. Wait until the item shows **Ready** before asking questions. Keep related documents in the same brain so you can search them together.

![A document dataset in Cognee](/images/manual/use-cases/cognee-dataset.png#bordered)

## Search and explore

Open **Search**, select the dataset and ask **Where should volunteers meet, and who coordinates them?**

Read the answer, then open **Mindmap** to inspect the relevant graph or source material. Generated relationships can vary with the model. Compare them with the original text before relying on them.

![Connected project knowledge in Cognee](/images/manual/use-cases/cognee-graph.png#bordered)

## Keep your knowledge

Uploaded content, graph files, vector indexes and model cache persist in private application data. Relational records use Olares PostgreSQL. Back up both stores when moving or restoring this instance.

For a slow first job, check whether the embedding model is still downloading before submitting the same document again. If language generation fails, check the **default-chat** route in AI Router.

## Learn more

- [Cognee documentation](https://docs.cognee.ai/)
- [Olares Router](./olares-router.md)
