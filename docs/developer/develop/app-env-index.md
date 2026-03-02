---
outline: [2, 3]
description: Learn how variables are injected during Olares app deployment, including declarative environment variables (.Values.olaresEnv) and system-injected runtime Helm values (.Values.*).
---

# Environment variables overview

Olares apps use app-service to inject runtime context and configuration into the app's `values.yaml`. In Helm templates, you can reference these values via `.Values.*`.

:::info Variables and Helm values
In this document, "variables" mainly refer to Helm values. They are not automatically passed into container environment variables. If you need them inside containers, explicitly map them to `env:` in your templates.
:::

## How variables are injected

Olares injects variables through two channels:

- **Declarative environment variables**: The developer declares variables under `envs` in `OlaresManifest.yaml`. At deployment, app-service resolves and injects the values into `.Values.olaresEnv` in `values.yaml`.

- **System-injected runtime variables**: Injected automatically by Olares at deployment time. No declaration is required, though some values are only available after you declare the relevant dependency, such as middleware.

## Next steps

1. [Declarative environment variables](app-env-vars.md): Field reference for the `envs` schema.
2. [System environment variables](app-sys-env-vars.md): Available system variables and how to map them.
3. [User environment variables](app-user-env-vars.md): Available user variables and how to map them.
4. [System-injected runtime values](app-sys-injected-values.md): Full reference for all system-injected values.