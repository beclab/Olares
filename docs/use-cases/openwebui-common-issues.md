---
outline: [2, 3]
description: Common issues and solutions for Open WebUI on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, common issues, troubleshooting, model download
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-13"
---

# Common issues

This page lists common issues for Open WebUI on Olares and their workarounds.

## Model app is stuck at "Waiting for Ollama" or "Needs attention"

If the model app stays in these states for more than a few minutes, the cause is likely GPU allocation.

#### Cause

The model app requires GPU resources to load the model, but it might not be correctly linked to the GPU or might not have enough VRAM allocated.

#### Solution

1. Go to **Settings** > **GPU**.
2. Check your GPU mode:
   - If you are using **Memory slicing**, make sure the model app is linked to the GPU and has enough VRAM allocated.
   - If you are using **App exclusive**, make sure the exclusive app is set to your model app.
3. Restart the model app from Launchpad and check the status again.

## Download progress disappears

When downloading a model via the dropdown menu, the progress bar might sometimes disappear before completion.

#### Cause

The download UI can lose connection to the background download process, especially during large file transfers.

#### Solution

To resume the download:
1. Click the model selector again.
2. Enter the exact same model name.
3. Select **Pull from Ollama.com**. The download will resume from where it left off.
