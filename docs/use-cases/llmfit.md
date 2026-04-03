---
outline: [2, 3]
description: Use LLMFit on Olares to find the best LLM models for your hardware. It benchmarks your system and scores models on quality, speed, compatibility, and context length.
head:
  - - meta
    - name: keywords
      content: Olares, LLMFit, LLM benchmark, hardware detection, GPU, model recommendation, self-hosted, AI
app_version: "1.0.3"
doc_version: "1.0"
doc_updated: "2026-04-02"
---

# Find the best LLM models for your hardware with LLMFit

LLMFit automatically detects your system's RAM, CPU, and GPU, then recommends LLM models that run well on your hardware. It scores each model across four dimensions: quality, speed, compatibility, and context length, so you can quickly see which models will perform best on your setup.

:::warning Known limitations

The current version of LLMFit has these limitations:
* **Random node allocation**: In a multi-node environment, the specific node that the LLMFit container is assigned to is random.
* **Single GPU detection**: LLMFit can only detect and evaluate the GPU assigned to it on the specific node where it resides.
* **Multi-GPU display limitations**: If your node has multiple GPUs, the dashboard displays only one of them (which one is unpredictable). The API detects multiple cards correctly, but their VRAM is not aggregated or summed together in the interface.
:::

## Install LLMFit

1. Open Market and search for "LLMFit".
   ![Install LLMFit](/images/manual/use-cases/llmfit.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to complete.

## Use LLMFit

Open LLMFit from the Launchpad. The dashboard displays:

- **System Summary**: The current node and its hardware details such as CPU, RAM, and GPU.
- **Model Fit Explorer**: A list of LLM models with estimated TPS (tokens per second) and scores for quality, speed, fit, and context.

Use these scores to decide which models to download and run on your Olares device.

![LLMFit dashboard](/images/manual/use-cases/llmfit-dashboard.png#bordered)

## FAQ

### How do I use the LLMFit TUI?

LLMFit uses its built-in web dashboard as the primary interface for simplicity. The dashboard provides the same functionality as the TUI.

If you prefer the terminal-based TUI, open Control Hub, navigate to the LLMFit container terminal, and then run the following command:

```bash
llmfit
```

![LLMFit container terminal](/images/manual/use-cases/llmfit-terminal.png#bordered)

## Learn more

- [Download and run local AI models via Ollama](ollama.md)
- [Set up Open WebUI with Ollama](openwebui-ollama.md)
