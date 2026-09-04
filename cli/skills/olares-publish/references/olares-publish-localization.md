# Market Manifest localization

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) and finalize the English source with [Manifest copy](olares-publish-manifest-copy.md) before translating.

## Source and locale set

Translate from the approved English Manifest, not from another translation. Do not research new product facts during localization or repair a source defect differently in each language. When the root Manifest and English locale Manifest both contain copy, compare them first; unresolved drift blocks repository writes.

In a repository, discover the app boundary, instructions, generators, locale directory convention, manifest filename, root-to-English relationship, `spec.locale`, and existing locale directories before choosing paths. Follow the established mapping. Do not create an undeclared locale unless the user explicitly requests it, and report any mismatch between `spec.locale` and actual locale files.

`metadata.title`, `metadata.description`, and `spec.fullDescription` must be present and non-empty. `spec.upgradeDescription` is conditional:

- English absent: omit it in targets.
- English empty: prefer removing the empty field; do not create new target content.
- English non-empty: every declared target must contain a non-empty equivalent translation.
- Target non-empty while English is absent: report source drift; do not treat the target as an independent factual source.

## Translation contract

Translate meaning rather than English sentence shape. Preserve facts, scope, conditions, warnings, modality, section count, order, emphasis, links, lists, and blank-line grouping. Do not add marketing language, cultural examples, warnings, or product facts.

Keep these exact unless the publisher or repository defines an approved localized form:

- third-party product names
- `Olares`, `Olares OS`, `Olares ID`, `Olares Space`, `LarePass`, `Vault`, `Profile`, `Studio`, and `Wise`
- inline code and fenced-code content
- URLs and Markdown link targets
- versions, image names, paths, commands, ports, protocols, environment variables, configuration keys, and placeholders
- fixed technical identifiers such as `CPU`, `GPU`, `API`, `HTTP`, `JSON`, and `YAML`

Localize visible Olares built-in app names when the approved locale terminology does so; `Desktop`, `Market`, `Files`, `Settings`, `Control Hub`, and `Dashboard` are not automatically protected brand terms. Localize visible bold section labels, but keep their Markdown structure. Do not introduce ATX headings, tables, HTML, or fenced code that the approved source does not contain.

`metadata.description` remains concise and has no final punctuation in every locale.

## Locale guidance

### `zh-CN`

Use concise Simplified Chinese. Avoid mechanical pronouns when the subject is clear. Add half-width spaces between Chinese and adjacent Latin product or technical terms, while leaving code, paths, versions, and URLs unchanged.

### `ja-JP`

Use natural neutral Japanese and avoid literal syntax. Do not insert spaces between Japanese text and adjacent Latin letters, product names, acronyms, or numbers: `Olaresで実行`, `GPUを使用`, `v1.2に更新`. Keep a space only when it is part of an official name or required inside code, commands, paths, URLs, or other protected syntax. Kanji, hiragana, and katakana do not require word spaces merely because the scripts differ.

### `fr-FR`, `de-DE`, `it-IT`, and `es-ES`

Use standard regional software terminology and neutral documentation prose. Restructure sentences naturally when needed, but do not drop or strengthen facts. Apply visible punctuation and spacing conventions without changing protected content.

## Review and handoff

Compare every locale directly with English. Check field presence, factual scope, numbers, versions, warnings, required actions, Markdown structure, protected tokens, terminology, punctuation, and natural language. Pay particular attention to punctuation touching bare URLs and to spaces around Latin content in Chinese and Japanese.

Before writing, report the discovered source and locale-to-path mapping when it is not already explicit. After writing, report files changed, validation performed, unresolved source or terminology questions, and any rendering-sensitive table, HTML, fenced code, or unusually long field that still needs manual review.
