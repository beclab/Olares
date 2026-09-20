# Calling a model — translation

> **Prerequisite:** [the calling contract](olares-router-calling.md) — the credential, what `--model` takes, and how to read a refusal — holds for every verb here.

```
olares-cli router call translate "hello" --to zh
olares-cli router call translate --detect "¿dónde está?"
olares-cli router call translate --languages
olares-cli router call translate --transcript meeting.json --to en
```

`translate` translates, `--detect` identifies a language instead, and `--languages` lists the pairs the configured model serves. `--to` is required for a translation and `--from` is optional, since detection is the default.

## Translating a conversation, not a line

`--transcript <file>` is the fourth route and the one that is not a text at a time: it translates a conversation as a stretch, so the model reads the turns around each line, and answers one result per turn under the ids that were sent — never shorter, so nothing shifts onto the wrong turn. Use it for anything a diarizer or a meeting tool produced.

`--per-line` on the same file translates each turn with nothing around it, which is how a two-word reply comes back as "yes" when it meant "right". The difference is the whole reason this route exists, so reach for `--per-line` only when the lines really are unrelated.

The file is a `segments` array, or an object carrying one alongside `context`, `background` and `glossary`. 64 turns is the ceiling, and splitting means carrying the tail of each part into the next one's `context`. `--background` and repeatable `--term source=target` add the same two things from the command line.

A turn the model failed on comes back carrying its reason, in place, so the rest of the page is still usable.
