# Agent-facing CLI onboarding: documentation benchmark

This note compares the official documentation of four agent-facing CLIs and applies the patterns to the Olares CLI documentation. It focuses on information architecture rather than product capability.

Research date: 2026-10-20

## Executive conclusion

The consistent pattern is to give new users one uninterrupted success path:

1. Install the CLI.
2. Authenticate during setup or on first launch.
3. Run one useful task.

Products still keep detailed installation, authentication, skills, extensions, and configuration pages, but those pages support the quickstart instead of replacing it. A user should not have to navigate between several pages before the first successful task.

Olares should therefore merge the basic content from `cli-install.md`, `cli-log-in.md`, and `cli-agent-skills.md` into one **Use Olares CLI with AI agents** page. The merged page should cover installation, login, verification, and the first AI-agent task. Detailed skill behavior and the complete Agent Skills catalog can remain as reference material, but using the skills must not be a separate onboarding step.

## Product comparisons

| Product | Core onboarding | Where authentication appears | Where skills/extensions appear | Information-architecture lesson |
| --- | --- | --- | --- | --- |
| OpenAI Codex CLI | One CLI page explicitly promises to install Codex, sign in, and run the first task. Its numbered quickstart goes from installation to first-run sign-in to a sample prompt. | In the same three-step quickstart, with a link to detailed authentication options. | The CLI overview introduces skills/plugins as a later capability and links to a separate skills/plugins guide. | Keep the first successful task on one page; link to extension reference after the user is running. |
| Anthropic Claude Code | One quickstart covers installation, first launch and browser login, the first question, the first code change, and a Git task. | Login is part of “Start your first session,” not a separate prerequisite page. | The quickstart only surfaces skills as something available from `/` and links to the separate “Extend Claude Code” area for details. | Treat authentication as a step in the first session. Mention skills during use, but keep authoring and architecture out of the happy path. |
| Google Gemini CLI | A single “Get started” page is titled around installing, authenticating, configuring, and using the CLI. It includes the install command, interactive sign-in, and a sample task. | In the same page, with a link to a dedicated authentication guide for alternative account and headless cases. | Agent Skills and Extensions each have separate tutorials and reference pages under “Use Gemini CLI” and “Features.” | A unified get-started page and separate deep guides can coexist. The detailed pages are escape hatches, not sequential gates. |
| GitHub Copilot CLI | One quickstart combines installation, first launch, `/login`, folder trust, and a first project-overview prompt. | In the quickstart; GitHub also provides dedicated installation and authentication pages under a setup hub. | Skills live under customization as a separate how-to and in the command reference. | Duplicate the minimum required auth instructions in the quickstart; reserve the separate auth and skills pages for variants and reference. |

### OpenAI Codex CLI

The [Codex CLI page](https://developers.openai.com/codex/cli) presents a three-step quickstart: install Codex, run it and sign in, then start the first task. The sign-in step happens the first time the user runs `codex`, and the page links out to authentication options rather than interrupting the flow with all variants.

The same page introduces skills and plugins as an additional capability after the quickstart. The separate [Skills & Plugins guide](https://learn.chatgpt.com/docs/skills-and-plugins?surface=cli) explains reusable workflows, skill selection, building, installation, and plugins. This keeps extension concepts discoverable without making them a prerequisite for first use.

### Anthropic Claude Code

The [Claude Code quickstart](https://code.claude.com/docs/en/quickstart) states that it covers installation, sign-in, and the first code change. Its sequence is install, start the first session and authenticate, ask the first question, make the first change, then use Git. The authentication flow is explained where the user first runs `claude`, which matches the user's actual sequence.

The quickstart mentions that `/` shows commands and skills, then routes advanced customization to [Extend Claude Code](https://code.claude.com/docs/en/features-overview). The extension overview treats skills, subagents, hooks, MCP, and plugins as capabilities that users add when a need appears. Skills have their own guide and are not a separate prerequisite in the onboarding journey.

### Google Gemini CLI

The [Gemini CLI get-started guide](https://geminicli.com/docs/get-started/) deliberately combines install, authentication, configuration, and use. It gives the global npm install, starts `gemini`, walks through Google sign-in, and ends with a concrete prompt. It links to a dedicated [authentication guide](https://geminicli.com/docs/get-started/authentication/) for account-specific, enterprise, API-key, and headless cases.

Gemini keeps Agent Skills and Extensions in deeper documentation. The [Agent Skills overview](https://geminicli.com/docs/cli/skills/) covers discovery and lifecycle; the [skills tutorial](https://geminicli.com/docs/cli/tutorials/skills-getting-started/) covers creation and activation; the [Extensions guide](https://geminicli.com/docs/extensions/) covers discovery, installation, and management. Its [documentation landing page](https://geminicli.com/docs/) groups the quickstart, installation, and authentication under “Get started,” while skills are grouped under use/features.

### GitHub Copilot CLI

The [Copilot CLI quickstart](https://docs.github.com/en/copilot/get-started/cli-quickstart) combines installation, first launch, `/login`, folder trust, and a first prompt. GitHub also has a [setup hub](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli) that separates installation, authentication, configuration, and troubleshooting for users who need more detail. The dedicated [installation page](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/install-copilot-cli) still includes first-launch authentication and points onward to use, so the user is not left at “the binary exists.”

Agent Skills are documented under customization in [Adding agent skills for GitHub Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills), with deeper semantics in [Comparing GitHub Copilot CLI customization features](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/comparing-cli-features) and command details in the [CLI command reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference). This separates first use from authoring, storage locations, enable/disable behavior, and reference syntax.

## What this means for Olares

### Why the three current pages should not be sequential prerequisites

The current Olares journey is spread across:

- `docs/developer/cli-install.md`
- `docs/developer/cli-log-in.md`
- `docs/developer/cli-agent-skills.md`

That split is more costly for Olares than for the products above because the recommended command, `npx @olares/cli@latest install`, installs both the CLI and its bundled Agent Skills. The product already treats them as one setup operation; the documentation should reflect that product behavior.

Login is also not a separate user goal. It is the bridge between installation and a successful command. Likewise, “use Agent Skills” is not a separate setup phase: after the installer writes the skills, the user should be able to give an agent an Olares task immediately.

### Recommended primary page

Keep the existing `/developer/cli-install` route to minimize URL churn, but retitle the page **Use Olares CLI with AI agents**. Recommended sequence:

1. **Before you begin** — supported client machines, Node.js requirement if applicable, Olares ID, and the distinction between the standalone user CLI and the host-bundled OS CLI.
2. **Install Olares CLI and Agent Skills** — use `npx @olares/cli@latest install` as the recommended path and explain in one sentence what it installs.
3. **Log in to Olares** — run `olares-cli profile login --olares-id ...`; include password and optional TOTP behavior inline.
4. **Verify the connection** — run `olares-cli profile whoami`.
5. **Run your first task with an AI agent** — check that Olares skills are visible in the agent's chat window: use `/` if the agent lists skills as slash commands, or ask it to list the available Olares skills. Then ask it to list the Home folder in natural language. The user does not need to load `olares-shared` first. Keep direct commands as verification or troubleshooting tools rather than a parallel onboarding path.
6. **Try more tasks** — a few natural-language examples covering files, Market, and dashboard/status.
7. **Keep the CLI and skills up to date** — `olares-cli update`, which updates both together.
8. **Special case: an Olares Linux host** — retain the current binary-conflict warning and separate-prefix guidance.
9. **Learn more** — link to Lares, the conceptual blog post, profile/auth reference, and the Agent Skills guide.

The entire tutorial should be framed around using Olares CLI with an AI agent. It should not split direct CLI use and AI-agent use into parallel onboarding paths.

### What may remain separate

Merging the onboarding path does not require putting every Agent Skills detail on one long page.

- **Profile management:** keep the small command table with the main tutorial so login and account switching stay in one place.
- **Agent Skills guide:** keep the complete skill catalog, storage/link behavior, manual reinstall/export, provenance, and mismatch notices on a dedicated page.
- **Host/cluster management:** keep separate because it uses a different identity and privilege model and serves a different job.

The key boundary is user intent: “install, log in, and use a skill for the first time” belongs on one page; “understand every profile command or every skill bundle” is reference material.

### Navigation and redirect recommendation

Suggested primary navigation:

- Olares CLI overview
- Use Olares CLI with AI agents
- Olares CLI Agent Skills
- Cluster management

Remove the separate login page and redirect `/developer/cli-log-in` to `/developer/cli-install#log-in-to-olares`. Keep the existing Agent Skills route, rename its title to **Olares CLI Agent Skills**, and remove first-use instructions from it.

## Recommended decision

Yes: merge installation, login, and first use of Agent Skills into one **Use Olares CLI with AI agents** page. Preserve a separate Agent Skills page only for deep reference, not as a required next step. This matches the dominant pattern in Codex CLI, Claude Code, Gemini CLI, and GitHub Copilot CLI, and it aligns especially well with Olares because the recommended installer already deploys the CLI and its matching Agent Skills together.
