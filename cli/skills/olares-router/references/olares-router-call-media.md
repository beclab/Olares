# Calling a model — pictures, clips, tracks and meshes

> **Prerequisite:** [the calling contract](olares-router-calling.md) — the credential, what `--model` takes, and how to read a refusal — holds for every verb here. `music` and `3d` require `--model`; Router keeps no default category for either.

```
olares-cli router call image "a red bicycle" --out bike.png
olares-cli router call video "a bicycle rolling downhill" --out clip.mp4
olares-cli router call music "a slow waltz" --model FlowStudio/<workflow> --out waltz.mp3
olares-cli router call music format "a slow waltz" --model FlowStudio/<workflow> --lyrics "…" --vocal-language en
olares-cli router call music draft "a defiant closing credits song" --model FlowStudio/<workflow>
olares-cli router call music alignment <generation-id>
olares-cli router call music cancel <generation-id>
olares-cli router call 3d --model FlowStudio/<workflow> --image lantern.png --out lantern.glb
```

## Work that outlives the request

`image`, `video`, `music` and `3d` are the four verbs whose work outlives the request. All four submit, wait, and write the result to `--out`; `--no-wait` prints the generation id instead, `--id <id>` collects that generation later, and `--output-id` picks one when a generation produced several. The wait defaults follow how long the work takes — five minutes for an image, twenty for a video, ten for a track, fifteen for a mesh — and a `--timeout` only stops the waiting: the provider carries on and the id stays collectable. Router holds the bytes itself, so a result does not depend on a vendor link staying alive.

The generation is anchored on `(user, key)`, so a `--no-wait` submission and the `--id` that collects it have to run under the same credential — see [the calling contract](olares-router-calling.md#the-identity-is-also-the-anchor).

## One contract, four shapes

Each verb offers only the fields its own family can express — an image has no `--fps`, a mesh has no `--lyrics` — so a field that could only be refused is not a flag there at all; all four share `--negative`, `--seed` and `--provider-option k=v`, the last carrying a vendor knob this contract has no field for. `--size` and `--aspect-ratio` describe the same shape, so giving both is refused before the request. `3d` is the one family that needs no words at all: most 3D workflows work from a picture, so `--image lantern.png` is a complete request, and a local file becomes a data URL while a data URL or a link is sent as written.

Underneath, image and video ride the OpenAI-shaped routes they shipped with, `3d` rides Router's unified one, and music has a surface of its own; the fields mean the same thing on all three. It shows in one place only: an image provider that keeps no generations to poll answers inline, and `image` handles that as well as the polled kind, which is why it was not moved onto the unified route.

## Music has four verbs the other families do not

A music model does not sing the caption and lyrics it is given: a language model in front of the audio one rewrites them first, and until these routes existed the only way to see that rewrite was to wait for the track and listen.

`music format` runs that pass alone and reports both versions — what was sent, and what would be performed — with warnings naming what it had to change. `music draft` runs it from one sentence and writes the caption, the lyrics and a tempo. Both are text-speed and priced as text; neither produces audio. `music cancel <id>` stops a running track and settles what it used, which is a bound on the cost rather than a refund — music is the only family whose upstream can be stopped mid-run. `music alignment <id>` reads a finished track's lyric timeline — each line with the second it starts and ends, which a player needs to highlight the current line and to seek by tapping one, and which is derivable from neither the lyrics nor the audio because the model decides how the words fall across the sections; Router answers it from the provider and upstream id already sealed onto the generation, so nothing is regenerated and no other model is consulted, and a track whose application serves no timeline says `lyrics_alignment_unavailable` rather than inventing one.

Both text passes return two lyric fields on purpose: `effective_lyrics` is what the song says, for a person to read, and `conditioning_lyrics` is the same words spelled for the model, with phoneme hints and section markers — a player shows one and the model sings the other. `music` itself also takes `--title`, and `--repaint` with `--audio` regenerates part of a recording you already have. A prompt whose first word is `format`, `draft`, `alignment` or `cancel` needs quoting.
