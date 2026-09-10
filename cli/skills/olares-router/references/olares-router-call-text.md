# Calling a model — text and the web

> **Prerequisite:** [the calling contract](olares-router-calling.md) — the credential, what `--model` takes, and how to read a refusal — holds for every verb here.

```
olares-cli router call chat "summarise this" --system "be terse"
cat notes.md | olares-cli router call chat --no-stream --quiet
olares-cli router call chat "what is in this picture" --image shot.png
olares-cli router call responses "summarise this" --model openai/gpt-4o
olares-cli router call count-tokens "how big is this turn" --model default-chat
olares-cli router call embed "text" --dimensions 512
olares-cli router call rerank "who wrote it" --document "…" --document "…"
olares-cli router call search "olares release notes" --limit 5
olares-cli router call scrape https://example.com/post
```

## Chat, embeddings, reranking

`chat` streams by default and prints a model and token line after the answer; `--quiet` prints only the answer, `--no-stream` waits for the whole thing. A prompt comes from the arguments or from standard input. A model running on this Olares serves a fixed number of requests at once and queues the rest, so `chat` allows ten minutes by default, says on stderr that it is waiting, and takes `--timeout` for a different budget; giving up stops the waiting and not the work, and the engine finishes an abandoned completion anyway. `--image` attaches a local file, which requires a model whose row declares `supports_vision`.

`embed` prints a summary of each vector in table form and the whole vector in JSON, and `--per-line` turns piped text into one input per line rather than a single input. `--image` embeds a picture on that same endpoint, which needs a CLIP model — one declaring `supports_embedding_image_input` — and carries one picture per call rather than a picture and text together, since the point of the shared space is comparing the vector with the text vectors already stored. `rerank` takes a query and a repeatable `--document`, or the documents one per line on standard input, and prints them in the order the model put them.

## Responses, and counting before you send

`responses` sends one request to the Responses endpoint and prints the answer the way `chat` does. It exists so that a model configured with `--mode responses` can be checked at all: that mode is served on a different endpoint, so calling such a model with `chat` fails in a way that says nothing about the model. It is deliberately only that — one request, no conversation carried across calls, nothing stored. `--stream` prints the answer as it is written: Router serves the Responses stream over a WebSocket rather than as chunked HTTP, so the flag opens a socket, sends the one request and closes. Reasoning goes to stderr and text to stdout, and `-o json` prints nothing until the end, because a document assembled from deltas is not the one Router sends.

`count-tokens` is the other verb that sends no turn: it asks how large a turn would be and sends nothing. It is mounted above the quota line and writes no usage row, so it is free and repeatable — which makes it the way to decide whether a file fits before paying to find out that it does not. Where the number comes from depends on the model and the answer does not say which: a provider with an official counter is asked and answers exactly, and everything else — a local model, a provider without one, a provider that failed or was slow — is a local estimate close enough to budget with. `--file` counts an Anthropic-shaped request that already exists rather than text.

## The web

`search` and `scrape` reach a provider that has one of those two modes; most do not, so both are commonly refused for want of a category rather than for anything wrong with the request.

`translate` is on the same tree but is [its own read](olares-router-call-translate.md): it takes a file and a turn model rather than a prompt.
