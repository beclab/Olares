---
outline: [2, 3]
description: Common issues and solutions for Open WebUI on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, common issues, troubleshooting, model download
app_version: "1.0.38"
doc_version: "1.1"
doc_updated: "2026-07-31"
---

# Open WebUI common issues

Use this page to identify and resolve common issues with Open WebUI on Olares.

## Model app is stuck at "Waiting for Ollama" or "Needs attention"

If the model app stays in these states for more than a few minutes:

1. Go to **Settings** > **Accelerator**.
2. Check your GPU mode:
   - If you are using **Memory slicing**, make sure the model app is linked to the GPU and has enough VRAM allocated.
   - If you are using **Exclusive**, make sure the model app has full GPU access.
3. Restart the model app from Launchpad and check the status again.

## Microphone "Permission denied" error

When attempting to use the dictate button or Voice Mode, you might receive the following error messages:
- `Permission denied when accessing microphone: NotAllowedError: Permission denied`
- `Permission denied when accessing media devices`

The Olares desktop displays applications inside embedded frames (iframes). For strict security and privacy reasons, modern web browsers prevent embedded frames from accessing sensitive hardware like your microphone, even if you already granted the browser permission in your system settings.

To bypass this security restriction and use your microphone:
1. In the top-right corner of the Open WebUI window on the Olares desktop, select <i class="material-symbols-outlined">open_in_new</i> to open it in a new browser tab.
2. In the new browser tab, select the microphone icon in the chat interface.
3. When the browser prompts you, allow microphone access.
