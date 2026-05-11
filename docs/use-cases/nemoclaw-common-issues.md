---
outline: [2, 3]
description: Common issues and workarounds for NemoClaw on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, common issues, troubleshooting
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-05-11"
---

# Common issues

This page lists common issues for NemoClaw on Olares and their workarounds.

## Discord channel lookup fails

When configuring Discord in the NemoClaw CLI sandbox, you might see the following error:

```text
Channel lookup failed; keeping entries as typed. TypeError: fetch failed
```

To work around this, restart the NemoClaw container from Control Hub and try the Discord configuration again.

## Missing default workspace files

NemoClaw might fail to create the default workspace files during installation. As a temporary workaround, manually create the required files by referring to the [OpenClaw default agent documentation](https://docs.openclaw.ai/reference/AGENTS.default) and the [official templates](https://github.com/openclaw/openclaw/tree/main/docs/reference/templates).

## Olares CLI login and skills don't persist across restarts

NemoClaw doesn't persist your Olares CLI login or installed ClawHub skills across restarts. After restarting NemoClaw, log in to Olares CLI again and reinstall the Olares skills. For details, see [Manage Olares with Olares CLI](nemoclaw-olares-cli.md).
