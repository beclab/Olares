# Market Manifest localization

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) and finalize the English source with [Manifest copy](olares-publish-manifest-copy.md) before translating.

## Source and locale set

Translate from the approved English Manifest, never from another translation. Do not add product facts or fix the same English problem differently in each language.

Before choosing files, inspect the repository's instructions, `spec.locale`, locale directories, filenames, generators, and the relationship between the root and English Manifests. Follow the existing layout. Do not create an undeclared locale unless the user asks for it. Report missing or extra locale files.

If the root and English copy differ, resolve the source of truth before writing translations.

`metadata.title`, `metadata.description`, and `spec.fullDescription` must be present and non-empty. `spec.upgradeDescription` is conditional:

- English absent: omit it in targets.
- English empty: prefer removing the empty field; do not create new target content.
- English non-empty: every declared target must contain a non-empty equivalent translation.
- Target non-empty while English is absent: report source drift; do not treat the target as an independent factual source.

## Translation contract

Write natural text in each target language. Preserve the source's facts, limits, warnings, required actions, section order, links, lists, and emphasis. Do not copy English sentence structure when it sounds awkward. Do not add marketing language, examples, warnings, or product facts.

Before translating, check that the approved English is understandable without specialist background and unambiguous about the subject, input, action, result, limits, and scope. Undefined task labels, compressed noun stacks, slash-separated fragments, vague pronouns, or modifiers with more than one possible attachment are source defects, not translation choices. Pause and revise or clarify the English source first; do not invent a different interpretation in each locale.

Technical accuracy alone is not enough. Each translation must preserve the plain-language explanation of what the app does. Keep a specialist term when it is useful for recognition, but do not let it replace that explanation. For example, translate both the term “speaker diarization” and the accompanying meaning that the app separates a recording by speaker and marks when each person speaks.

Keep these exact unless the publisher or repository defines an approved localized form:

- third-party product names
- `Olares`, `Olares OS`, `Olares ID`, `Olares Space`, `LarePass`, `Vault`, `Profile`, `Studio`, and `Wise`
- inline code and fenced-code content
- URLs and Markdown link targets
- versions, image names, paths, commands, ports, protocols, environment variables, configuration keys, and placeholders
- fixed technical identifiers such as `CPU`, `GPU`, `API`, `HTTP`, `JSON`, and `YAML`

Localize visible Olares built-in app names when the approved locale terminology does so; `Desktop`, `Market`, `Files`, `Settings`, `Control Hub`, and `Dashboard` are not automatically protected brand terms. Localize visible bold section labels, but keep their Markdown structure. Do not introduce ATX headings, tables, HTML, or fenced code that the approved source does not contain.

`metadata.description` remains concise and has no final punctuation in every locale.

For `upgradeDescription`, keep the most important summary within the first three lines in every locale. Do not let translated headings, links, or background detail push the key change below the collapsed release-history preview.

## Locale guidance

### `zh-CN`

Use concise Simplified Chinese. Drop unnecessary pronouns when the subject is clear. Add half-width spaces between Chinese and adjacent Latin product or technical terms. Do not alter code, paths, versions, or URLs.

### `ja-JP`

Use natural, neutral Japanese. Do not insert spaces between Japanese and adjacent Latin text or numbers: `Olaresで実行`, `GPUを使用`, `v1.2に更新`. Keep spaces that belong to an official name or protected syntax such as code, commands, paths, and URLs.

### `fr-FR`, `de-DE`, `it-IT`, and `es-ES`

Use standard regional software terms and a neutral tone. Rewrite sentence structure when needed, but do not remove or strengthen facts. Follow local punctuation and spacing without changing protected content.

## Review and handoff

Compare every locale directly with English. Check the facts, numbers, versions, warnings, actions, Markdown, protected text, terminology, punctuation, and fluency. Confirm that a non-specialist target-language reader can still understand the app's input and result without knowing the English technical term. Pay extra attention to punctuation next to bare URLs and spaces around Latin text in Chinese and Japanese.

Before writing, show the source and locale-to-path mapping when they are not obvious. After writing, list the files changed, checks run, open questions, and any table, HTML, code block, or unusually long field that still needs a human review.
