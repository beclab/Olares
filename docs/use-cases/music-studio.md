---
outline: [2, 3]
description: Create a short song with Music Studio on Olares, review the result, and download the audio for your own project.
head:
  - - meta
    - name: keywords
      content: Olares, Music Studio, ACE-Step, AI music, lyrics, instrumental, local music generation
app_version: "0.1.76"
doc_version: "1.0"
doc_updated: "2026-11-12"
---

# Create a song with Music Studio

Music Studio is an AI music creation app powered by ACE-Step XL-SFT. You can turn lyrics and a style description into a vocal track, or create instrumental music. Finished tracks stay in your Olares music library for playback and download.

This example creates a 30-second song called *Morning Light*. Replace the sample lyrics and style with material for your own video or music project.

## Learning objectives

- Create a short song using your own lyrics and musical direction.
- Use AI to draft lyrics and review its suggestions.
- Play, organize, and download your finished tracks.

## Prerequisites

- Olares 1.12.7 or later with Router installed and running.
- An amd64 device with an NVIDIA GPU and 24 GiB of GPU memory, as specified by this app package.
- Sufficient disk space for the initial model download of approximately 30 GB and your generated audio.

## Install Music Studio

1. Open Market and search for "Music Studio".
2. Click **Get**, then **Install**, and wait for installation to complete.
3. Open Music Studio from Launchpad. On first use, the model setup page downloads approximately 30 GB of required model files and displays the download progress. Wait for the download and model loading to finish. The creation page opens automatically when the service is ready.

The app includes its music model. You do not need to select a model in the creation form. Allow network access and sufficient disk space for the first download. Download and loading time depend on your connection and device. Existing cached model files can be reused on later starts.

## Create your song

### Start with your own lyrics

1. Open **Create** and enter these words in **Lyrics**:

   ```text
   [Verse]
   Morning light across the bay
   Let the shadows drift away
   [Chorus]
   Take a breath and start again
   Every road can find a friend
   ```

2. In **Music style**, enter:

   ```text
   Hopeful acoustic pop, warm piano, gentle guitar and light drums,
   clear English vocals, clean studio production
   ```

3. Set **Title** to `Morning Light` and **Language** to **English**.
4. Under **Song duration**, choose **Fixed** and set **Duration** to 30 seconds. Leave **Instrumental** unchecked for this vocal example.
5. Click **Create**. Music Studio first refines the lyrics, then displays the song under **In progress** while it generates audio. Wait for the finished track to appear under **Recent songs**.

![Prepare lyrics and musical direction in Music Studio](/images/manual/use-cases/music-studio-create.png#bordered)

Music Studio runs one creation task at a time. You can use **Cancel** on an in-progress generation if you want to stop it. Generation time varies with duration, quality settings, and model loading.

### Let AI draft a starting point

Enter a short description in **Song idea**, choose the language and duration, then click **Write with AI**. For example:

```text
A hopeful morning song with piano and a simple memorable chorus
```

Review the resulting **Lyrics**, **Music style**, and **Title** before clicking **Create**. AI suggestions can contain repetition or musical directions that do not match your intended language. Edit them directly to keep the parts you want.

If the app rejects repetitive lyrics, revise the idea and try again, or enter your own lyrics. A failed draft does not prevent you from creating a song with manually written content.

### Make an instrumental variation

For background music without vocals, select **Instrumental** and describe the instruments, mood, and arrangement in **Music style**. Set the duration and click **Create**. You can find the result later using the **Instrumental** filter in **Library**.

### Adjust the musical direction

Expand **Advanced options** when you want more control. **Quality profile** offers XL-SFT with 50 steps or 64 steps with ADG. **Production profile** selects a clean studio sound or textured ambience. **Style handling** lets you enhance the description or preserve it more strictly.

You can also specify BPM, key, time signature, vocal type, and section structure. These settings guide generation rather than guaranteeing an exact musical result. Change a small number of settings at a time so you can compare the variations.

## Review and keep your music

1. Click the play button beside a completed track. Use the bottom player to pause, move to another track, or drag the playback position.
2. Click the track to open its details and review the lyrics and music style. Vocal tracks can show synchronized lyrics when timing is available. Instrumental tracks do not need lyric timing.
3. Open **Library** to search your saved tracks. Use **Liked** or **Instrumental** to narrow the list, or change the model filter and sort order.
4. Open a track's **More actions** menu to choose **Like**, **Rename**, or **Download**. The generated audio downloads as a WAV file.

![Review a saved track and use the player](/images/manual/use-cases/music-studio-playback.png#bordered)

Choose **Reuse seed** from the same menu to copy a track's creation settings back into the form. Edit the lyrics or style and create another version. Reusing a seed does not guarantee identical output after other settings change.

Use **Delete** only for tracks you no longer need. Download a copy first if you want to keep the audio.

:::info Section editing
The current package does not provide a working section repaint workflow. Create another version with revised lyrics or musical direction instead.
:::

## Change the display language

Open **Settings** and choose **Display language**. Select **English**, **简体中文**, or **Auto** to follow your Olares language. This changes the interface language. Set the song's language separately in the creation form.

## FAQs

### Why does the app show no available music model?

After starting or resuming the app, allow the model service to finish loading. If the creation page still shows no available model, refresh it and check that Router is running. Avoid starting multiple creation tasks while a model task is active.

### Why are synchronized lyrics missing?

Lyric highlighting depends on timing returned for that vocal track. You can still read the lyrics and play or download the audio when timing is unavailable. Instrumental tracks do not have sung lyrics.

## Learn more

- [Olares Router](olares-router.md): Manage AI model availability and routing on Olares.
