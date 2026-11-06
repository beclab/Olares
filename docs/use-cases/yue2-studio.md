---
outline: [2, 3]
description: Choose a YuE2 Studio model set on Olares, create and edit a song, change its lyrics or style, continue a melody, make a cover from a recording, and generate music through Lares.
head:
  - - meta
    - name: keywords
      content: Olares, YuE2 Studio, AI music, Q8, BF16, quantization, piano roll, ABC, MIDI, covers, instrumental music, Lares, MCP, HTTP API
app_version: "0.0.7"
doc_version: "2.2"
doc_updated: "2026-11-06"
---

# Create and edit a song with YuE2 Studio

YuE2 Studio is a local AI music app that turns a musical style and lyrics into songs with vocals and accompaniment. It also produces an editable score, so you can change the composition before generating another performance.

This guide uses a short English acoustic pop song called "Morning Light". You will create the original, raise its vocal melody, and increase its tempo. You can follow the same workflow with your own lyrics and musical style.

## Learning objectives

In this guide, you will learn how to:

- Choose a model set and GPU allocation for your music workflow.
- Compose and preview a score before generating audio.
- Edit notes, chords, sections, key and tempo.
- Reuse a composition with new lyrics, another style or instrumental parts.
- Continue a short melody and create a cover from a recording.
- Choose generation settings and exchange scores with a music editor.
- Connect Lares through MCP and create a song from a conversation.

## Prerequisites

- YuE2 Studio installed on an amd64 Olares device. This guide targets the test-market preview of Studio 3.8.5.
- An NVIDIA GeForce RTX 50 series GPU with compute capability sm_120 and a CUDA 13 compatible driver. The example was prepared on an RTX 5090 Laptop GPU.
- A Hugging Face service address and token configured in the Olares system environment for model downloads.

The app reads `OLARES_SYSTEM_HUGGINGFACE_SERVICE` and `OLARES_SYSTEM_HUGGINGFACE_TOKEN`. To check these values, open Olares **Settings** > **Advanced** > **System environment variables**. For changes, follow [Set system environment variables](../manual/manage-system-env.md). These settings also affect other apps that use them.

Olares assigns GPU resources when you launch the app. If launch is blocked by an allocation message, follow [Resume an app and assign resources](../manual/olares/settings/gpu-resource.md#resume-an-app-and-assign-resources).

## Prepare the music engine

The four ready-made sets use the same YuE2 backbone. Their main difference is the precision used to store its weights. Lower precision reduces the size of the model and can introduce more quantization error.

| Model set | Backbone precision | Approximate set size shown in Settings | When to choose it |
|:---|:---|:---|:---|
| Light | `Q5_K_M` | 4.0 GB | You need the smallest set and accept stronger weight compression. |
| Balanced | `Q6_K` | 4.3 GB | You want a middle option between Light and Quality. |
| Quality | `Q8_0` | 5.1 GB | Start here for everyday generation and editing. This is the recommended baseline for this Olares package. |
| Full native | `BF16` | 9.9 GB | You want to compare against the original backbone weights and can allocate more memory. |

All four sets keep the audio VAE in `F32`. Light, Balanced and Quality use `Q8_0` versions of SheetSage2 and MERT for recording transcription. Full native uses their `F32` versions. Each set also includes the decoder companion. The set size therefore includes more than the backbone file.

For the RTX 5090 Laptop GPU used in this guide, start with Quality Q8. Full native is an optional comparison, with larger downloads and memory needs. Its higher weight precision does not guarantee a better result for every prompt, and a smaller file does not guarantee faster generation on every GPU.

1. Open YuE2 Studio from Launchpad.
2. Click your account name in the lower-left corner to open **Settings**, then select **Models**.
3. Choose one music-engine model set. **Quality - Q8_0 backbone, near lossless** is a starting option. **Full native - BF16 backbone, original weights** downloads the original backbone weights and uses more storage.
4. Download the selected set and wait for it to show **installed**. You only need one set.
5. Click **Done**, return to **Create**, and wait for **Engine ready** before starting a song.

![Music-engine model sets in YuE2 Studio Settings](/images/manual/use-cases/yue2-studio-model-sets.png#bordered)

Model weights are downloaded separately from the application and remain on disk across app restarts. The model-set sizes shown in Settings are file sizes, not a measurement of peak GPU memory during generation.

The Olares package declares a 5 GiB minimum GPU memory request for a single-song Q8 workflow. Time slicing does not enforce this as a fixed VRAM cap. In one 130-second Q8 generation at 32 steps, the observed whole-GPU memory peak was about 4.48 GiB. A BF16 run reached about 5.72 GiB. These samples help choose an allocation, but longer material, more variations and transcription can change memory use. Q5 and Q6 were not profiled in this setup.

Allocate more VRAM before moving to BF16 or increasing concurrency. In Memory slicing mode, follow [Resume an app and assign resources](../manual/olares/settings/gpu-resource.md#resume-an-app-and-assign-resources) to change the quota. Switching an installed model set restarts the music engine, so finish the current job first.

## Create the original song

### Enter a style and lyrics

1. In **Create**, select **Studio**.
2. Set the title to `Morning Light Original`.
3. Enter the following in **Style**:

    ```text
    English acoustic pop, warm female vocals, piano, acoustic guitar, gentle drums, hopeful, clean production, 100 BPM
    ```

4. Enter these original example lyrics in **Lyrics**. Keep each section tag on its own line and leave a blank line between sections.

    ```text
    [Verse]
    Morning light across the floor
    Open skies beyond the door
    Every road can lead us home
    We will never walk alone

    [Chorus]
    Carry on into the light
    Keep our little dreams in sight
    Let the music lift us high
    Like a sunrise in the sky
    ```

### Compose and generate

1. Under **Score (ABC)**, keep **Melody + chords** selected and click **Compose the score**.
2. Wait for the score to appear. Click **Listen** to hear a simple preview of the notes. This preview helps you review the melody before generating the sung version.
3. Use these settings for the example:

    | Setting | Example value | Purpose |
    |:---|:---|:---|
    | **Max duration** | `2:00` | Gives the short song room to finish. This is a ceiling, not a target length. |
    | **Flow-matching steps** | `32` | Sets the number of audio synthesis steps. |
    | **Solver** | **Midpoint (reference)** | Uses the reference rendering method. |
    | **Songs per request** | `1` | Creates one composition. |
    | **Variations** | `1` | Creates one performance. |

4. For easier comparisons, turn **Randomize seed** off. Set **Composition seed** to `20261106` and **Sound seed** to `42`.
5. Click **Create**. Wait for the track to appear in **Library**, then play it.

![An English song in the YuE2 Studio creation workspace](/images/manual/use-cases/yue2-studio-create.png#bordered)

The generated track retains its score and generation settings. You can return to it to make another version.

There are three planning modes:

| Planning mode | What it supplies | Useful for |
|:---|:---|:---|
| **Melody + chords** | A melody with chord symbols | Writing a song whose melody and harmony you want to edit. |
| **Melody** | A melody-only plan | Covers where the new style should supply the harmony. |
| **No score** | Audio generation without a symbolic plan | Exploring a freer result when score editing is not your goal. |

## Edit and perform the score

### Change the melody and tempo

1. Find `Morning Light Original` in **Library** and open **Song Details** using the information icon.
2. In the score area, click **Edit** to open the piano roll.
3. Select the **Vocal** track. Choose **Edit** > **Select All**, then **Edit** > **Transpose**. Enter `2` to raise that track by two semitones and confirm the change.
4. Change **BPM** to `120` in the bottom bar for a faster version.
5. Click **Apply** in the upper-right corner. The edited score returns to the creation form.

![Vocal notes, chord and section lanes, and the tempo control in the score editor](/images/manual/use-cases/yue2-studio-edit-score.png#bordered)

To make a smaller change, drag an individual note up or down to change its pitch, move it horizontally to change its start time, or drag its right edge to change its length. Click a chord or section label in the lanes above the notes to edit it.

:::warning Save the score before refreshing
The creation form does not preserve an unsaved ABC score when you refresh the page. Use **Save the score as .abc**, or generate a track to store the score in Library.
:::

### Generate the edited version

1. Change the title to `Morning Light Variation` so you can distinguish it from the original.
2. In **Style**, change `100 BPM` to `120 BPM` to match the edited score.
3. Enable **Keep the edited score for new words** under the score. This keeps the composition when the lyrics or style change. Leave the example lyrics unchanged for this first comparison.
4. Keep the other generation settings the same, then click **Create**.
5. Play the original and the variation in Library to compare the vocal melody and pace.

![The original song and the edited variation saved as separate tracks](/images/manual/use-cases/yue2-studio-compare-versions.png#bordered)

**Apply** updates the score in the form. **Create** renders a new audio track. The original track remains in Library.

The piano-roll transpose operation affects the selected track. To move the melody and harmony together into another key, use the whole-score transpose workflow below.

## Refine the composition

### Draw, move and resize notes

1. Open the saved song's **Song Details**, then click **Edit** beside the score.
2. Select **Vocal** to edit the sung melody or **Ins** to edit the instrument line.
3. With the pencil tool selected, click an empty point in the grid to add a note. Start with a short note inside an existing phrase.
4. Drag the note vertically to change its pitch or horizontally to change when it starts.
5. Drag its right edge to change its length. Use the grid's snap setting when you want the edit to land on a regular beat division.
6. Preview the phrase in the editor, then click **Apply**. Use a new title and **Create** to hear the change in a generated performance.

Horizontal position represents time, vertical position represents pitch, and note width represents duration. Scroll horizontally to reach later bars and vertically to bring a high or low melody into view.

### Edit chords and sections

1. In the score editor, find the **Chords** lane above the notes.
2. Click a chord label and enter the replacement. For example, try changing a chord to `Gm7` in the F-major example.
3. Use the **Sections** lane to edit section names or add a section marker such as `Interlude`.
4. Preview the score, then click **Apply** and generate a new version.

Keep section names and lyrics tags consistent. If you add an instrumental break, use a corresponding tag such as `[Instrumental Break]` in the lyrics. Listen to the transition after generation, because a section marker alone does not determine how the arrangement will sound.

### Change the key or the vocal register

For a whole-song key change:

1. Load the saved score with **Use this score**.
2. Expand **Advanced** and find **Transpose score**.
3. Enter a semitone offset such as `2` to raise the whole score, or `-2` to lower it.
4. Use a new title and click **Create**. The new performance uses the transposed score.
5. Reset **Transpose score** to `0` before another edit unless you want to transpose that version again.

For an octave adjustment to the voice alone, use **Voice −8va** or **Voice +8va** beside the score. These move the vocal line by one octave and keep the accompaniment and chord symbols in place. Preview the result and check any vocal-range warning before generating. Adjusting the octave changes the written register, while the description of the voice belongs in **Style**.

## Reuse or extend a song

### Sing the same composition with new lyrics

1. Open the original song's three-dot menu and select **Sing again with new words**. This brings its score and prompt into the creation form.
2. Give the new version a title such as `Morning Light New Words`.
3. Replace the chorus with another four-line chorus. For example:

    ```text
    [Chorus]
    Follow me into the day
    Let our worries drift away
    Hear a new song in the air
    There is hope for us to share
    ```

4. Keep the original verse for this first attempt. Match the original line lengths and stressed syllables as closely as you can.
5. Enable **Keep the edited score for new words** under the score, then click **Create**.
6. Listen for words that arrive too early, stretch too far or lose clarity. Shorten or rephrase those lines and make another version.

If you load the score through **Use this score** instead, fill in the lyrics and style yourself before generating. The keep option is needed when a marked score belongs to different words or a different style. If the app reports that the edited score belongs to other words or style, review the form and enable that option to continue with the intended score.

### Change the voice or musical style

1. Load the original score and keep the lyrics unchanged for the first comparison.
2. Change **Style** to a complete description of the new performance. For example:

    ```text
    English jazz ballad, warm baritone vocals, upright bass, brushed drums, jazz piano, intimate and mellow, 100 BPM
    ```

3. Enable **Keep the edited score for new words** when it is shown.
4. Use a title such as `Morning Light Jazz Version`, then click **Create**.

Keeping the score preserves the written composition while the model generates another performance. Listen to the result before deciding whether the voice and arrangement match your description. A style prompt does not clone a particular singer.

### Turn the melody into instrumental music

1. Load the original score and click **Instrumental** beside the notation. This moves the vocal melody into the instrument line.
2. Replace the style with an instrumental description, for example:

    ```text
    Instrumental acoustic pop, piano, acoustic guitar, gentle drums, calm and hopeful, 100 BPM
    ```

3. Remove the sung words and keep the section tags, such as `[Verse]` and `[Chorus]`.
4. Enable the keep option when shown, use a new title, and click **Create**.

This workflow uses the score's **Instrumental** button. The separate **Instrumental only** switch near **Create** can also add an Instrumental LoRA and download its weights when missing. That additional adapter is separate from the four base model sets.

### Continue a short melody

Use this when you have an opening phrase and want YuE2 to write the rest of the song.

1. Enter the full intended lyrics and style.
2. In **Score (ABC)**, paste a short opening score, or load an ABC file with **Open an .abc score**. Keep the ABC header, key, meter and tempo, plus at least one phrase with vocal notes.
3. Enable **Continue the song from this score**.
4. Click **Continue the score** to compose the remainder before generating audio.
5. Preview the completed score and adjust it in **Score editor**. Give the result a new title and click **Create**.

To use a saved song as the starting point, export its ABC first. In a copy of that file, keep its header and the opening verse and remove later sections before importing it. Keep both voice declarations and the complete bars belonging to the remaining sections. An opening made only of rests or an instrumental introduction needs a sung or written vocal phrase before it can guide this continuation workflow.

### Cover a recording in a new style

The Cover workflow first transcribes a recording into a melody with SheetSage2, then generates a new performance from that score. You can inspect and correct the transcription before generating.

1. In **Create**, select **Cover** and keep **Melody only** enabled. This retains the tune and lets the new style supply the harmony.
2. Choose a source:
    - For a file on your computer, click **Choose a recording…**.
    - For the demo already in Library, open `Morning Light Original`'s three-dot menu and choose **Transcribe to score: Melody for a cover**.
3. Wait for **Score ready**. Preview the melody with **Listen** and correct notes or tempo in **Score editor** when needed.
4. Paste the source song's lyrics with section tags that match the recording. Keep the original words for the first cover.
5. Describe the new arrangement in **Style**, such as the jazz prompt above. Use a title such as `Morning Light Jazz Cover`.
6. Check **Max duration** gives the recording room to finish, then click **Create**.

![A source recording transcribed into a score for an English jazz cover](/images/manual/use-cases/yue2-studio-cover-score.png#bordered)

If the app reports that no SheetSage2 transcriber is available, add it through **Settings** > **Models**. The ready-made sets in this guide include it. Review the transcription when the source has a dense arrangement or unclear melody. A cover generated from a transcription can differ from the recording.

## Choose generation settings

### Start with a chord progression

Use **Chord bed** when you want to define harmony before the model writes the melody.

1. Prepare the style and lyrics in **Studio**, and select **Melody + chords** as the planning mode.
2. Open **Chord bed** below the score.
3. Set **Tempo, BPM**, **Key** and **Meter**, for example `100`, `C` and `4/4`.
4. Click **Write the bed**. If a score already exists, the button is **Write another bed**.
5. Review the chord symbols. The bed contains chords over rests, with sections based on the lyrics tags.
6. Use **Create** to generate a melody and arrangement over that harmony.

Writing a chord bed replaces the current score in the form. Export a score you want to keep before using it.

![Chord-bed tempo, key, meter and generated chord progressions](/images/manual/use-cases/yue2-studio-chord-bed.png#bordered)

### Balance synthesis steps and solver choice

For the first version, keep **32** flow-matching steps and **Midpoint (reference)**. To audition an idea more quickly, try **16** steps and compare the result. Fewer synthesis steps can change the sound, and their effect depends on the song.

The other solver, **AB2 (about twice as fast)**, uses fewer network evaluations in the sound stage. Try it on the same composition and listen for changes in vocals, attacks and transitions. Score composition, model loading and other stages still take time, so the label does not mean that the entire request finishes in half the time.

![Flow-matching steps, solver and song count in the generation settings](/images/manual/use-cases/yue2-studio-quality-settings.png#bordered)

### Use seeds and variations deliberately

| Control | What it changes | How to use it |
|:---|:---|:---|
| **Composition seed** | The plan written from the prompt | Keep it fixed while comparing prompt or score changes. |
| **Sound seed** | The synthesis noise for the performance | Change it to try another rendering of a composition. |
| **Songs per request** | Separate compositions | Each song writes its own music and uses more working memory. |
| **Variations** | Performances of the same composed song | Start with `1` when using the minimum Q8 allocation. |

If **Songs per request** is disabled, the local engine's current song limit is one. Finish a job before changing engine limits in **Settings** > **Local engine**, and increase the GPU allocation before trying more concurrent work. Keep one song and one variation for the 5 GiB Q8 baseline.

### Re-render an existing performance

1. Open a song's three-dot menu and choose **Re-render**.
2. Adjust the step count, sound seed, number of variations or output format in the dialog.
3. Submit the new render and compare it with the saved track.

Re-render reuses the saved performance plan and skips writing a new composition. Use the score-editing workflows when you want different notes, lyrics or style. Repeating the same settings is useful for comparison, but it is not a promise of a byte-for-byte identical audio file.

![Re-render controls for steps, variations, sound seed and output format](/images/manual/use-cases/yue2-studio-rerender.png#bordered)

## Save and exchange your work

### Save audio and ABC scores

1. Open **Song Details** for the version you want to keep.
2. Click **Download Audio** to save the generated recording to your computer.
3. Click **.abc** in the score area to save the editable score. Use **Use this score** when you want to bring a saved track's score back into the creation form.

To restore the file later, open **Create** and click **Open an .abc score** in the score toolbar. Preview it, then supply the style and lyrics for the performance you want.

### Exchange notes with a MIDI editor

Click **MIDI** beside a score to export its notes and tempo for another music editor. Keep the ABC file as well when you want to preserve the original chord symbols and score text.

To bring a MIDI file into the song workflow:

1. Select **Cover**, click **Choose a recording…**, and select the MIDI file.
2. In **Load MIDI**, choose **Melody only** or **With chords** for the score you want.
3. Check the track table and explicitly select the **Voice** and **Instrument** tracks. For the example export, choose **1. Vocal Melody** for **Voice** and **2. Ins Melody** for **Instrument**. Automatic selection can pick a different melodic track.
4. Choose whether to keep the file's section markers. Review the timing grid, octave adjustments and conversion notices before accepting them.
5. Wait for conversion to finish and **Use as the score** to become available. Click it and preview the imported notation. A MIDI track can contain chords or tempo changes that need to be simplified into the song score.
6. Supply the lyrics and style, then use a new title and **Create** to generate a performance.

![Explicit vocal and instrument track assignments in the MIDI import dialog](/images/manual/use-cases/yue2-studio-import-midi.png#bordered)

Converting MIDI back to ABC can infer chord names again. Review those names before treating the imported score as a copy of the original harmony.

### Review a PDF before sharing

The score's **PDF** button exports printable notation. Check the generated page before sharing it. Some flat chord symbols displayed incorrectly in this preview version, so use ABC or MIDI when you need to preserve an accurate editable score.

### Find and compare saved versions

Use Library search to filter by a shared title such as `Morning Light`. Give each version a descriptive suffix, then open **Song Details** to review its lyrics and saved generation settings before comparing audio. The duration beside a track is the recording's length, and the relative timestamp shows when it was added. Neither is the time spent generating it.

## Create a song with Lares

You can ask Lares to generate a song through YuE2 Studio's MCP server, then open the result in Studio to listen or edit its score. This workflow requires YuE2 Studio package 0.0.5 or later on Olares, with the **YuE2 Studio MCP & API** entrance available.

Have [Lares](lares.md) installed on the same Olares device and select a chat model that supports tool calls. This example uses Qwen3.8. Start its model app and wait for it to become available in Lares. The chat model handles your request, while YuE2 generates the music. Keep YuE2 Studio running with **Engine ready** and a model set installed.

When both models use the same GPU, leave enough memory for both and run one music request at a time. Time slicing shares GPU processing time but does not automatically release the memory held by a loaded model.

### Connect the MCP server

Token authentication is off by default. The entrance keeps Olares's internal access policy. Any client that can reach this entrance can call its tools. To require an application token, follow [Enable token authentication](#enable-token-authentication).

1. Open Olares **Settings** > **Applications** > **YuE2 Studio** > **Entrances**.
2. Select **YuE2 Studio MCP & API** and copy its **Endpoint** URL. Append `/mcp` if the URL does not already end with it. This internal entrance is accessible to Lares on the same Olares device.
3. Open Lares and go to **Settings** > **MCP**. Click **Add server** and enter:

    | Setting | Value |
    |:---|:---|
    | **Server name** | `yue2studio` |
    | **Transport** | `Streamable HTTP` |
    | **MCP URL** | The endpoint ending in `/mcp` from step 2 |
    | **Headers** | `{}` |

4. Click **Save**. Lares connects and loads the server's tools.

### Enable token authentication

The `MCP_AUTH_ENABLED` application environment variable controls token authentication for both MCP and the HTTP API. Its default is `false`.

1. In Olares **Settings** > **Applications** > **YuE2 Studio**, find **Environment variables** and click **Manage environment variables**.
2. Edit `MCP_AUTH_ENABLED`, set it to `true`, and click **Confirm**, then **Apply**. The application restarts to apply the change. Wait for it to become available again. See [Manage application environment variables](../manual/olares/settings/manage-app-env.md) for the Settings workflow.
3. Open Files and find **Application** > **Data** > **yue2studio** > **config**. Open `mcp-token` as text and copy the complete line. If text preview is unavailable, download it and open it in a text editor.
4. In Lares **Settings** > **MCP**, edit `yue2studio` and replace **Headers** with the JSON below. Replace `<your-token>` with the copied line and keep one space after `Bearer`.

    ```json
    {
      "Authorization": "Bearer <your-token>"
    }
    ```

5. Click **Save**. Subsequent requests require the token.

The token is generated automatically and retained across normal upgrades. Keep it out of chat prompts and shared screenshots. To turn token authentication off again, set `MCP_AUTH_ENABLED` to `false` and apply the change. Clients can then use empty **Headers** (`{}`).

### Generate Morning Light from a conversation

1. Start a new Lares conversation and select Qwen3.8 or another available chat model with tool support.
2. Ask Lares to check Studio first:

    ```text
    Use the configured yue2studio MCP server to call studio_status. Report the actual result. Do not generate music or change settings.
    ```

    Check that Lares calls `studio_status` and returns the current Studio state. If it only describes what it would do, ask it to use the configured MCP tool.

3. Send the following request. Change the title, style or lyrics to use your own song idea. The 30-second request, 16 steps and AB2 solver are example settings for a quick first song.

::: details Example: generate Morning Light through Lares
```text
Use the configured yue2studio MCP server to generate one song.

Title: Morning Light
Style: English acoustic pop, warm female vocals, acoustic guitar, piano and soft drums, hopeful and calm, clean studio production, 100 BPM.

Lyrics:
[Verse 1]
Morning light across the sea
Brings a little hope to me
Every road can lead us home
We will never walk alone

[Chorus]
Let the music find its way
Carry us into the day
Raise your voice and hold on tight
We are brighter than the night

Call song_create with duration_seconds=30, steps=16, solver="ab2", lm_batch_size=1, synth_batch_size=1 and output_format="flac".

Use studio_wait until the job finishes. Report the actual job ID and saved song ID. Only report success after completion.

Do not use shell commands, download models, change settings or modify existing songs. If a tool fails, report its exact error and stop.
```
:::

4. Let Lares wait for the job to finish. If it asks for permission to run a generation tool, review the request and approve it. A submitted job ID means the song is queued or running. Wait for completion and the saved song ID.
5. Open YuE2 Studio and search the library for `Morning Light`. Play the new track. Open **Song Details** to check its lyrics and saved settings, then click **Download Audio** to save it.
6. To change the composition, click **Use this score** and continue with [Edit and perform the score](#edit-and-perform-the-score). Use a new title when generating the edited version.

If a call fails, read the error before trying again. When token authentication is enabled, an HTTP 401 response means the application token is missing or incorrect. If Lares cannot answer a plain chat message, check its selected chat model and connection before retrying the music request. For a generation failure, open YuE2 Studio and check the engine state and available GPU memory.

### Use another client or the HTTP API

Other compatible MCP clients can use the same Streamable HTTP URL. The same entrance also serves the app's HTTP API. Remove the final `/mcp` from the MCP URL to get the base address. When `MCP_AUTH_ENABLED=true`, send the Bearer token in the Authorization header with each request.

| Task | HTTP request |
|:---|:---|
| Submit a generation job | `POST /v1/music/jobs` |
| Read a job's state and saved songs | `GET /v1/music/jobs/{job_id}` |
| Download a saved song's audio | `GET /v1/library/media/{song_id}` |

Generation is asynchronous. Keep the returned job ID and wait for completion before downloading the audio. For tool descriptions and automation examples, see the [YuE2 Studio MCP reference](https://github.com/timoncool/YuE2-Studio/blob/main/docs/mcp-skill.md).

## About additional Studio tools

The core workflows above use the music engine and its included transcription models. Other visible tools, including assistant writing, LoRA training, stem separation, VST processing and all-instrument audio-to-MIDI transcription, can require additional providers, models or platform components. They are outside the validated scope of this Linux preview. The MIDI score export described here is part of the core workflow.

:::info Model terms
YuE2 model weights use CC BY-NC 4.0. Check the [YuE2 model card](https://huggingface.co/m-a-p/YuE2-3B) before using the model for your project.
:::

## Learn more

- [Lares](lares.md): Configure the assistant and its chat model.

- [Manage accelerator resources](../manual/olares/settings/gpu-resource.md): Assign or release GPU resources for Olares apps.
- [Set system environment variables](../manual/manage-system-env.md): Manage the shared Hugging Face download settings.
