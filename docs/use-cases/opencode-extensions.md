---
outline: [2, 3]
description: Extend OpenCode on Olares with skills and plugins. Use pre-installed skills for package management and web preview, or add community plugins for extra functionality.
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, skills, plugins, extensions, AI coding agent, self-hosted
---

# Extend OpenCode with skills and plugins

OpenCode supports two types of extensions:

- **Skills**: Markdown instruction files that teach OpenCode how to handle domain-specific tasks. OpenCode loads them automatically based on context.
- **Plugins**: JavaScript or TypeScript modules that add runtime capabilities. They run at startup and can hook into OpenCode's execution pipeline.

Both can be scoped globally or to individual projects.

## Skills

OpenCode loads relevant skills automatically during a conversation, so you rarely need to manage them yourself.

### Pre-installed skills

OpenCode on Olares ships with two skills:

| Skill | Description |
|:------|:------------|
| `system-admin` | System package management via `pkg-install` |
| `web-preview` | Dev server preview through a built-in reverse proxy |

#### system-admin

Ask OpenCode to install or remove a system package in the chat. For example, type "Install ffmpeg" or "Remove the curl package", and the `system-admin` skill runs the appropriate `pkg-install` command.

If the skill doesn't activate, load it manually with `/skill load system-admin`.

For the full command reference, see [Manage packages](opencode-packages.md).

#### web-preview

The `web-preview` skill starts a dev server inside the container and exposes it through a built-in reverse proxy.

Describe what you want in the chat:

```text
Start the web project in this folder on port 5544
```

OpenCode starts the server, confirms it's running, and returns a preview URL:

```text
https://<your-OpenCode-domain>/__preview/<port>/
```

The domain is the same one shown in your browser address bar when you access OpenCode.

<!-- ![Web preview in browser](/images/manual/use-cases/opencode-web-preview.png#bordered) -->

If the skill doesn't activate, load it manually with `/skill load web-preview`.

### Manage skills

List available skills or load one manually:

```text
/skill list
/skill load <skill-name>
```

<!-- ![Skill list output](/images/manual/use-cases/opencode-skill-list.png#bordered) -->

Skill files are Markdown files stored in the following locations:

| Scope | Path in Olares Files |
|:------|:-----|
| Global (all projects) | `Application/Data/opencode/.config/opencode/skills/` |
| Project-level | `Home/Code/<project>/.opencode/skills/` |

## Plugins

Plugins are npm packages or local scripts that extend OpenCode at runtime.

### Install as npm packages

Declare plugins in the OpenCode config file:

| Scope | Config file in Files |
|:------|:-----|
| Global | `Application/Data/opencode/.config/opencode/opencode.json` |
| Project-level | `opencode.json` at the project root |

Example:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": [
    "opencode-helicone-session",
    "opencode-wakatime"
  ]
}
```

OpenCode resolves and installs declared packages on startup and caches them in `~/.cache/opencode/node_modules/`.

### Install as local files

Place `.js` or `.ts` files in a plugin directory. OpenCode loads them automatically on startup.

| Scope | Path in Olares Files |
|:------|:-----|
| Global plugins | `Application/Data/opencode/.config/opencode/plugins/` |
| Project-level plugins | `Home/Code/<project>/.opencode/plugins/` |

### Popular community plugins

| Plugin | Description |
|:-------|:------------|
| `opencode-helicone-session` | Session tracking and analytics via Helicone |
| `opencode-wakatime` | Coding activity tracking via WakaTime |

## Learn more

- [Manage packages](opencode-packages.md)
- [OpenCode plugin documentation](https://opencode.ai/docs/plugins/)
- [OpenCode official documentation](https://opencode.ai/docs)
