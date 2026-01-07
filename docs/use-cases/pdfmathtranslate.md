---
outline: [2, 4]
description: Learn how to install and configure PDFMathTranslate on Olares. This tutorial guides you through using the local AI model Ollama to translate scientific PDFs, preserving original layouts and mathematical formulas while ensuring total data privacy.
keywords: Olares, PDFMathTranslate, Document Translation, Local AI, Ollama, Privacy, Scientific Papers
---

# Translate scientific PDFs while preserving layout

Reading academic papers in foreign languages often relies on translation tools, but standard options frequently break complex layouts or scramble mathematical formulas. PDFMathTranslate solves this pain point by using AI to translate full scientific documents while keeping the original formatting, charts, and equations intact. When running on Olares, you can choose to process these translations using local AI models. This transforms your device into a private translation server, ensuring your research data never leaves your hardware.

## Learning objectives

By the end of this tutorial, you are able to:
- Install PDFMathTranslate from the Olares Market.
- Configure the local AI model for private translation.
- Translate a scientific PDF and manage the output files.
- Troubleshoot common configuration and interface issues.

## Preparation

To use local AI model to run translations for privacy, you must have an AI model engine installed and ready. This tutorial uses Ollama as the example. For more information, see [Download and run local AI models via Ollama](ollama.md).

## Install PDFMathTranslate

1. Open the Olares Market and search for `PDFMathTranslate`.
2. Click **Get**, and then click **Install**.
   
    ![Install PDFMathTranslate](../public/images/manual/use-cases/install-pdfmathtranslate.png#bordered)

3. When the installation finishes, click **Open**. The PDFMathTranslate workspace is displayed.

    ![Open PDFMathTranslate](../public/images/manual/use-cases/open-pdfmathtranslate.png#bordered)

## Translate

Follow these steps to translate your PDF documents.

### Upload your PDF document

In the **File** area, select your input **Type**:
- If you selected **File**, drag and drop your PDF document into the upload area, or click the upload area to browse your local storage. 
    
    When the document is uploaded, a preview of it appears in the **Preview** pane on the right.

- If you selected **Link**, enter the link address of the PDF document.

    The **Preview** pane remains blank during this step. The document content only appears in this area after the translation is completed.

### Configure the translation model

You must select an engine to perform the translation. You have two options:

#### Option A: Cloud services

* From the **Service** list, select **Google** or **Bing**. These services are free to use but require an Internet connection and process data externally. Because this tutorial focuses on the privacy-preserving local AI capabilities and the configuration steps required to set up the Local AI option, cloud service options are not covered in further detail.

#### Option B: Local AI

To use your local hardware via **Ollama** (recommended for privacy), follow these specific configuration steps:

1. From the **Service** list, select **Ollama**.
2. Enter the Ollama host URL. To obtain the host address:

    a. Go to Olares **Settings** > **Application** > **Ollama**.
    
    b. In the **Entrances** section, click **Ollama API**.

    c. Click **Set up endpoint**, and then copy the endpoint address by clicking <i class="material-symbols-outlined">content_copy</i>.

    ![Obtain Ollama host address](../public/images/manual/use-cases/copy-localhost-address.png#bordered)   

3. Enter the name of the model you have downloaded, and you must specify the version tag if required. For example, `gemma3:4b`.

    ![Open PDFMathTranslate](../public/images/manual/use-cases/local-model-setup.png#bordered)

### Select languages and scope

1. Select the source and target languages:

    a. **Translate from** indicates the original document's language.
    
    b. **Translate to** indicates the language you want to read.
    
    :::info
    The app does not auto-detect the languages. You must select them manually. Supported languages include English, Simplified Chinese, Traditional Chinese, French, German, Japanese, Korean, Russian, Spanish, and Italian.
    :::

2. Specify which pages to translate:
    * **All**: Translates the entire document.
    * **First**: Translates only the cover page.
    * **First 5 pages:** Translates the first five pages.
    * **Others** Allows you to specify page ranges.

    ![Set translation scope in PDFMathTranslate](../public/images/manual/use-cases/set-translation-scope.png#bordered)

3. Click **Translate**. The translation starts immediately. 

    :::tipLimitation
    Currently, the **Cancel** button is unavailable during translation. Please wait for the process to complete or error out.

### Download your files

When the translation is completed, the translated file is displayed in the **Preview** pane, and the application generates three files: the original source file, the translated file, and a bilingual version. Download the files in two ways.

![Access files translated by PDFMathTranslate](../public/images/manual/use-cases/access-translated-files.png#bordered)

#### Option A: In pdfmathtranslate workspace

    a. On the left navigation pane, in the **Translated** section, click the download button within the app interface.

    ![Download files translated in PDFMathTranslate](../public/images/manual/use-cases/download-translated-files.png#bordered)

#### Option B: From the Files on Olares system

    a. Go to Olares **Files** > **Data** > **pdfmathtranslate** > **pdfmathtranslate**.
    
    b. Double-click a file, and then click 

    ![Download files translated from Olares Files](../public/images/manual/use-cases/download-in-files.png#bordered)

## Troubleshooting

#### Translation progress bar not showing

If you refresh the page during a translation or after an error, the progress bar might disappear. This is a display glitch; wait for the process to finish or check your output folder.

![Normal translation progress bar](../public/images/manual/use-cases/translation-progress-normal.png#bordered)

#### File overwriting

If you translate the same file multiple times, Olares overwrites the previous version in the output folder. It only saves the most recent file. Rename your source files if you need to keep multiple versions.

#### Application freezing

If the translation stalls for a long time or reports an error, the background process might be stuck. The best fix is to uninstall and re-install the application.

#### Clean uninstall

To completely remove the application, simply clicking "Uninstall" is not enough. You must also manually delete the **pdfmathtranslate** folder located in **Files** to ensure all configuration data is removed.
