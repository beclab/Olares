---
outline: [2, 3]
description: Troubleshoot a local AI model that never becomes ready or an inference engine that does not start in Olares Model Console.
head:
  - - meta
    - name: keywords
      content: Olares, Model Console, model not ready, engine not running, Engine Base, AI model troubleshooting
---

# Model or engine is not ready

Use this guide when a model instance opens in Model Console but **Model** does not become **Ready**, or **Engine** does not become **Running**.

## Identify which stage is blocked

Open the model instance from the Launchpad and check **Service status** on the **Status** tab.

| What you see | What it means |
|---|---|
| **Model** is not **Ready** | The model files are still downloading or the download and verification process could not finish. |
| **Model** is **Ready**, but **Engine** is not **Running** | The files are available, but the inference service could not start with the current configuration or resources. |
| The model app is **Stopped** in Market or Settings | Olares could not reserve the resources required to launch the app, or the app was stopped manually. |

## If the model is not ready

1. If the model is still downloading, leave the model instance running. Large model files can take time to download.
2. Confirm that Olares can reach the model source and has enough free disk space. If storage is low, see [Free up disk space](../free-up-disk-space.md).
3. For an instance created from an Engine Base app, compare its **MODEL_SOURCE** and **MODEL_NAME** values with the format required by that engine. See [Configure engine environment variables](/use-cases/llm-base-apps.md#configure-engine-environment-variables).
4. If you correct an environment variable in **Settings** > **Applications** > **[model app]** > **Manage environment variables**, save the change and click **Apply**. Wait for Model Console to report the new status.

Do not delete partially downloaded model files while the instance is running. This can leave the app and its stored files out of sync.

## If the engine is not running

1. Wait until **Model** shows **Ready**. The engine cannot serve a model whose files are not ready.
2. Review the instance configuration:
   - Confirm that **MODEL_NAME** matches the configured model.
   - Check that **ENGINE_ARGS** use the syntax required by the selected engine.
   - Check that the engine's `REQUIRED_GPU_MEMORY` value does not exceed the available resource.
3. If the app is **Stopped**, resume it from **Market** > **My Olares** or **Settings** > **Applications**.
4. If Olares reports insufficient GPU memory, follow [GPU app remains stopped after installation or resume](./ts-vram-shortage.md).
5. Return to Model Console. Continue only after **Model** shows **Ready** and **Engine** shows **Running**.

## If the status does not change

Do not delete the app namespace, Application resource, or shared engine service. These actions can affect other model instances and make recovery harder.

Record the following information:

- Olares version and model app version.
- Engine Base type and whether the app is a pre-built model or a cloned instance.
- The exact **Model** and **Engine** status.
- The model source and engine arguments, with tokens or credentials removed.
- The time the failure started and any change made immediately before it.

Then download the application operation log from **Market** > **My Olares** > **Logs**. If more detail is requested, follow [Collect diagnostic information](../collect-diagnostic-information.md) and share the archive through a private support channel.

## Related guides

- [Run local LLMs with Ollama, vLLM, llama.cpp, and SGLang](/use-cases/llm-base-apps.md)
- [Connect an AI app to a model service](../best-practices/connect-ai-apps.md)
- [Memory is insufficient or not freed](./ts-free-memory.md)

