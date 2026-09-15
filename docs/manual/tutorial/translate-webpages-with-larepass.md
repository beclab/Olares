---
outline: [2, 3]
description: Translate webpages privately with the LarePass Chrome extension and the Hy-MT2 model running on your Olares.
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, webpage translation, immersive translation, Hy-MT2, local translation, private translation
---

# Translate webpages privately with LarePass <Badge type="tip" text="^ 1.12.7" />

Cloud-based translation extensions process webpage text on the provider's servers. LarePass takes a self-hosted approach: it sends the text through Router to **Hy-MT2** running on your Olares, rather than to a third-party translation provider. You retain more control over how your content is processed.

You also get the convenience of an integrated browser translator:

- **Private by design:** Router routes webpage text from LarePass to the model running on your Olares, so it is not processed by a third-party translation provider.
- **Lightweight and multilingual:** Hy-MT2 supports translation across 33 languages. Its compact 1.8B size balances translation quality with the resources required to run it locally.
- **Automatic integration:** Router uses your Olares identity to authenticate requests and automatically detects and configures the installed Hy-MT2 model. You do not need to configure model endpoints or API keys.

## Prerequisites

- [Install the LarePass browser extension](../install-larepass-browser-extension.md).
- [Import your Olares account](../larepass/manage-accounts.md#chrome-extension).
- Ensure **Router** is installed on your Olares.

## Get started with local translation

1. In [Olares Market](../olares/market/market.md#install-models), search for and install **Hy-MT2-1.8B**. Wait for the installation to finish so Router can automatically detect and configure the model.
2. Open a webpage and click the LarePass icon in the Chrome toolbar. Select <i class="material-symbols-outlined">translate</i>, choose the source and target languages, and select **Olares Hy-MT2** as the **Translation provider**. Then click **Translate this page**.

To restore the original text, click **Show original**. If the webpage was already open when you installed the extension or model, refresh it before translating.

## Customize your experience

On the Translate page, you can adjust:

- **Automatic translation**: Follow the global setting, always translate the current website, or never translate it.
- **Show translation only**: Hide the original text after translation.
- **Show original on hover**: Available after you turn on **Show translation only**. Point to a translated paragraph to temporarily view the original text.

To change global defaults and translation styles, click <i class="material-symbols-outlined">settings</i>.

## Troubleshooting

- **Provider unavailable:** Router might still be detecting and configuring Hy-MT2. Wait for it to finish, then refresh the webpage.
- **Page cannot be translated:** Chrome extensions cannot translate internal pages such as `chrome://extensions/`, or pages that do not expose regular webpage text. Try a standard article or blog.
