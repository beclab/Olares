---
outline: [2, 4]
description: Learn how to install and configure PDFMathTranslate on Olares. This tutorial guides you through using the local AI model Ollama to translate scientific PDFs, preserving original layouts and mathematical formulas.
---

# Translate scientific PDFs while preserving layout

PDFMathTranslate is an application designed to translate scientific PDF documents while retaining the original layout and mathematical formulas.

This tutorial provides instructions on how to install and use PDFMathTranslate on Olares, using the local AI model Ollama for translation.

## Learning objectives

By the end of this tutorial, you are able to:
- Configure the local AI model for PDF translation.
- Translate a scientific PDF and manage the output files.

## Before you begin

To ensure privacy by using a local AI model for translation, you must have an AI model engine installed and ready. This tutorial uses Ollama as the example. For installation instructions, see [Download and run local AI models via Ollama](ollama.md).

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

### Configure the translation service

To use the local AI service Ollama for translation, configure the following settings:

1. From the **Service** list, select **Ollama**.
2. (Optional) To obtain the Ollama host address:

    a. Go to Olares **Settings** > **Application** > **Ollama**.
    
    b. In the **Entrances** section, click **Ollama API**.

    c. Click **Set up endpoint**, and then click <i class="material-symbols-outlined">content_copy</i> to copy the endpoint address.

    ![Obtain Ollama host address](/images/manual/use-cases/copy-localhost-address.png#bordered){width=60%}

3. In the **OLLAMA_HOST** field, enter the Ollama host address. 
4. In the **OLLAMA_MODEL** field, enter the exact identifier of the Ollama model you have downloaded. For example, `gemma3:4b`.

    ![Open PDFMathTranslate](/images/manual/use-cases/local-model-setup.png#bordered)

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
    During translation, do not click the **Cancel** button; otherwise, the process might report an error.
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
1. Uninstall the app from My Olares. For more information, see [Uninstall applications](../../docs/manual/olares/market/market.md).
2. Open the Files app, go to **Application** > **Data**, and then delete the `pdfmathtranslate` folder.
