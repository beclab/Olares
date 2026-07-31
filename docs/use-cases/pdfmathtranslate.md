---
outline: [2, 4]
title: Translate scientific PDFs in PDFMathTranslate
description: Run PDFMathTranslate on Olares with a local OpenAI-compatible model to translate scientific PDFs while preserving page layouts, formulas, and other technical content.
head:
  - - meta
    - name: keywords
      content: Olares, PDFMathTranslate, translate PDF, scientific PDF translation, self-hosted PDF translator, PDFMathTranslate on Olares
---

# Translate scientific PDFs while preserving layout

PDFMathTranslate is an application designed to translate scientific PDF documents while retaining the original layout and mathematical formulas.

This tutorial uses a local OpenAI-compatible model to translate PDFs on Olares.

## Learning objectives

By the end of this tutorial, you are able to:
- Configure the local AI model for PDF translation.
- Translate a scientific PDF and manage the output files.

## Prerequisites

Before you begin, you need the following model:

| Used for | Model | How to get it |
| :--- | :--- | :--- |
| Translation | Qwen3.6-27B (llama.cpp) | Install from Market |

## Install PDFMathTranslate

1. Open the Olares Market and search for "PDFMathTranslate".
2. Click **Get**, and then click **Install**.
   
    ![Install PDFMathTranslate](/images/manual/use-cases/install-pdfmathtranslate.png#bordered)

3. When the installation finishes, click **Open**. The PDFMathTranslate workspace is displayed.

    ![Open PDFMathTranslate](/images/manual/use-cases/open-pdfmathtranslate.png#bordered)

## Translate

### Upload your PDF document

:::warning PDF format requirements
Ensure that the PDF file is a standard PDF document that is not password-protected or corrupted. The application cannot process invalid PDF files.
:::

In the **File** area, select your input **Type**:
- If you select **File**, drag and drop your PDF document into the upload area, or click the area to browse your local storage. 
    
    When the document is uploaded, a preview of it appears in the **Preview** pane on the right.

- If you select **Link**, enter the link address of the PDF document.

    The **Preview** pane remains blank during this step. The document content only appears in this area after the translation is completed.

### Get model connection details

PDFMathTranslate connects to Qwen3.6-27B through its OpenAI-compatible API.

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

For Qwen3.6-27B (llama.cpp):

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

Use the **OpenAI-Compatible** API format. The model name is `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`.

### Configure the translation service

1. In PDFMathTranslate, select **OpenAI** from the **Service** list.
2. In the **OPENAI_BASE_URL** field, enter the Base URL copied from Model Console.
3. In the **OPENAI_API_KEY** field, enter any non-empty value, such as `none`. The local model endpoint does not require an API key.
4. In the **OPENAI_MODEL** field, enter the copied Model name.

### Select languages and scope

1. Select the source and target languages:

    a. **Translate from** indicates the original document's language.
    
    b. **Translate to** indicates the language you want to read.
    
    :::info
    PDFMathTranslate does not auto-detect the languages. You must select them manually. Supported languages include English, Simplified Chinese, Traditional Chinese, French, German, Japanese, Korean, Russian, Spanish, and Italian.
    :::

2. Specify which pages to translate:
    * **All**: Translates the entire document.
    * **First**: Translates only the first page.
    * **First 5 pages**: Translates the first five pages.
    * **Others**: Translates a custom range of pages.

    ![Set translation scope in PDFMathTranslate](/images/manual/use-cases/set-translation-scope.png#bordered)

3. Click **Translate**. The translation starts immediately. 

    :::warning
    Do not click the **Cancel** button during translation. This might cause the process to report an error.
    :::

### Download your files

When the translation is completed, the translated file is displayed in the **Preview** pane, and the application generates three files:

- Original source file
- Translated file
- Bilingual version

You can find these files in the Files app. To download them to your local computer, download them directly from the PDFMathTranslate workspace, or use the Files app.

#### In PDFMathTranslate workspace

On the left side of the pdfmathtranslate workspace, in the **Translated** section, click the download button next to the file.

![Download files translated in PDFMathTranslate](/images/manual/use-cases/download-translated-files.png#bordered)

#### From Olares Files app

1. Open the Files app, and then go to **Data** > **pdfmathtranslate** > **pdfmathtranslate**.

    ![Access files translated by PDFMathTranslate](/images/manual/use-cases/access-translated-files.png#bordered)

2. Double-click a file, and then click the download icon in the upper-right corner.

    ![Download files translated from Olares Files](/images/manual/use-cases/download-in-files.png#bordered)

## FAQ

### Why did the translation progress bar disappear?

If you refresh the page while a translation is running, the progress bar might disappear from the screen. This is a display issue only, and the translation is still processing in the background. 

Wait for it to finish or check the output folder.

### Will the app keep multiple versions of the same file?

If you translate the same file name multiple times, the system replaces the previous version with the new one. It does not create numbered copies such as `file_1.pdf`. To keep multiple versions, rename your source file before translating it again.

### What should I do if the app becomes unresponsive or takes unusually long?

If the translation takes significantly longer than usual or the application stops responding, the background process might have stalled. To resolve this issue, uninstall and then re-install pdfmathtranslate.

### How do I perform a clean uninstallation of PDFMathTranslate?

To completely remove the application and its data:
1. Uninstall the app from Market or Desktop. For more information, see [Uninstall applications](../manual/olares/market/market.md#uninstall-applications).
2. Open the Files app, go to **Application** > **Data**, and then delete the `pdfmathtranslate` folder.
