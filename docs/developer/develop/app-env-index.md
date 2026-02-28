---
outline: [2, 3]
description: Learn how variables are injected during Olares app deployment, including  declarative environment variables (.Values.olaresEnv) and system-injected runtime Helm values (.Values.*).
---

# Environment variables overview

Olares apps use App Service to inject runtime context and configuration into the app's `values.yaml` (Helm values). In Helm templates, you can reference these values via `.Values.*`.

:::info
In this document, "variables" mainly refer to Helm values. They are not automatically passed into container environment variables. If you need them inside containers, explicitly map them to `env:` in your templates.
:::

## Injection channels

Olares injects variables through two channels:

1. **Declarative environment variables**

    - Developers declare variables under `envs` in `OlaresManifest.yaml`.  
    - During deployment, App Service injects the resolved values into `.Values.olaresEnv` in `values.yaml`.  
    - Template reference: <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>

2. **System-injected runtime variables**  

   - Injected by the system at deployment time into the root of `.Values` or specific subtrees in `values.yaml` (for example, `.Values.postgres.*`).  
   - You don't need to declare the "variable itself" in the manifest (however, some variables are injected only after you declare the relevant dependency, such as middleware).  
   - Template reference: <code v-pre>{{ .Values.* }}</code> (for example, <code v-pre>{{ .Values.postgres.host }}</code>)

## Comparison by category

| Category | Source | How to declare |
| -- | -- | -- |
| **Custom environment variables** | App-specific configuration | Defined directly in `envs` |
| **Mapped system/user environment variables** | System variable pool / user variable pool | Mapped via `envs.valueFrom` |
| **Predefined runtime variables** | User/system/hardware/storage/dependencies/middleware | Injected by the system (some require dependency declarations) |

:::info Info
System environment variables and user environment variables are "variable pools". An app must map them to its own `envName` via `envs.valueFrom` before it can use them under `.Values.olaresEnv`.
:::

## Next Steps

- Configure and manage environment variables: [Custom environment variables](app-env-vars.md)
- Reference variable pools:
    - [System environment variables](app-sys-env-vars.md)
    - [User environment variables](app-user-env-vars.md)
- Use injected values: [Predefined runtime variables](runtime-values.md)