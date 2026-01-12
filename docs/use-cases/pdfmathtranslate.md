---
outline: [2, 4]
description: Learn how to install and configure PDFMathTranslate on Olares. This tutorial guides you through using the local AI model Ollama to translate scientific PDFs, preserving original layouts and mathematical formulas.
---

# Translate scientific PDFs while preserving layout

Reading academic papers in foreign languages often relies on translation tools, but standard options frequently break complex layouts or scramble mathematical formulas. PDFMathTranslate solves this pain point by using AI to translate full scientific documents while keeping the original formatting, charts, and equations intact. When running on Olares, you can choose to process these translations using local AI models. This transforms your device into a private translation server, ensuring your research data never leaves your hardware.

## Learning objectives

By the end of this tutorial, you are able to:
- Install PDFMathTranslate from the Olares Market.
- Configure the local AI model for private translation.
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
Ensure that the PDF file is a standard PDF document that is not password-protected or corrupted. Invalid PDFs will fail.
:::

In the **File** area, select your input **Type**:
- If you select **File**, drag and drop your PDF document into the upload area, or click the area to browse your local storage. 
    
    When the document is uploaded, a preview of it appears in the **Preview** pane on the right.

- If you select **Link**, enter the link address of the PDF document.

    The **Preview** pane remains blank during this step. The document content only appears in this area after the translation is completed.

### Configure the translation service

Select the service you want to use for the translation. You can choose between an external cloud provider or a private local AI service.

#### Cloud services

* From the **Service** list, select **Google** or **Bing**. These services are free to use but require an Internet connection and process data externally. Because this tutorial focuses on the privacy-preserving local AI capabilities and the configuration steps required to set up the local AI option, cloud service options are not covered in further detail.

#### Local AI services

To use the local Ollama service, configure the following settings:

1. From the **Service** list, select **Ollama**.
2. Enter the Ollama host address. 
3. (Optional) To obtain the Ollama host address:

    a. Go to Olares **Settings** > **Application** > **Ollama**.
    
    b. In the **Entrances** section, click **Ollama API**.

    c. Click **Set up endpoint**, and then copy the endpoint address by clicking <i class="material-symbols-outlined">content_copy</i>.

    ![Obtain Ollama host address](/images/manual/use-cases/copy-localhost-address.png#bordered){width=60%}

3. Enter the name of the model you have downloaded, and you must specify the version tag if required. For example, "gemma3:4b".

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
    * **First 5 pages:**: Translates the first five pages.
    * **Others**: Translates a custom range of pages.

    ![Set translation scope in PDFMathTranslate](/images/manual/use-cases/set-translation-scope.png#bordered)

3. Click **Translate**. The translation starts immediately. 

    :::info
    The **Cancel** function is currently unavailable during active translation processing. According to best practices, do not click this button; otherwise, the translation progress might report an error.

### Download your files

When the translation is completed, the translated file is displayed in the **Preview** pane, and the application generates three files:

- Original source file
- Translated file
- Bilingual version


Download the files in two ways.

![Access files translated by PDFMathTranslate](/images/manual/use-cases/access-translated-files.png#bordered)

#### In PDFMathTranslate workspace

On the left side of the pdfmathtranslate workspace, in the **Translated** section, click the download button next to the file.

![Download files translated in PDFMathTranslate](/images/manual/use-cases/download-translated-files.png#bordered)

#### From Olares Files app

a. Go to Olares **Files** > **Data** > **pdfmathtranslate** > **pdfmathtranslate**.

b. Double-click a file, and then click the download icon in the upper-right corner.

![Download files translated from Olares Files](/images/manual/use-cases/download-in-files.png#bordered)

## Troubleshooting

#### Translation progress bar not shown

If you refresh the page while a translation is running, the progress bar might disappear from the screen. This is a display issue only, and the translation is still processing in the background. Please wait for the task to complete or check your output folder for the new file.

#### File overwriting

If you translate the same file name multiple times, the system replaces the previous version with the new one. It does not create numbered copies such as `file_1.pdf`. To keep multiple versions, rename your source file before translating it again.

#### Application unresponsive

If the translation takes significantly longer than usual or the application stops responding, the background process might have stalled. To resolve this issue, uninstall and then reinstall pdfmathtranslate.

#### Uninstall PDFMathTranslate

Using the standard **Uninstall** button removes the PDFMathTranslate application but keeps your settings. To completely remove the application and its data:
1. Uninstall the app from My Olares. For more information, see [Uninstall applications](../../docs/manual/olares/market/market.md).
2. Open the Files app, go to **Application** > **Data**, and then delete the **pdfmathtranslate** folder.
