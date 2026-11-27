# AI service connections

<!-- #region router-openai-base-url -->
Copy the Router address from your browser. Keep only the protocol and domain, then append `/v1`. Replace the domain in this example with your own:

   ```text
   https://router.laresprime.olares.com/v1
   ```
<!-- #endregion router-openai-base-url -->

<!-- #region model-context-window -->
1. In Router, open **LLM**, find the model, and click the information icon at the right end of its row to open **Model card**.

   ![Open a model card from the Router LLM list](/images/manual/use-cases/router-model-card-entry.png#bordered)

2. Review **Context window** and **Engine args**. For the llama.cpp model shown below, `-c 104448` gives the configured context size: **104448 tokens**. The list and model card display this as **102K**. Use the exact value from **Engine args** when a client asks for a token count.

   ![Read the exact llama.cpp context size in Router Engine args](/images/manual/use-cases/router-model-card-context.png#bordered)

:::tip Match the client's context setting
Set the client's context length to the value configured in your model instance's **Engine args**, or lower. Review this value when changing the model behind `default-chat`. Each engine uses its own context parameter, such as `-c` for llama.cpp.
:::
<!-- #endregion model-context-window -->
