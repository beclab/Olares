---
outline: [2, 3]
description: Learn to deploy Duix.Avatar on Olares, from model training to video synthesis, to create text-driven digital avatar videos.
---

# Create a digital avatar with Duix.Avatar

Duix.Avatar (formerly HeyGem) is an open-source AI toolkit for generating digital avatars, specializing in offline video creation and digital cloning.

This guide walks you through deploying and using Duix.Avatar on Olares, covering the complete process from model training to video synthesis to generate a text-driven digital avatar video.

## Learning objectives

In this guide, you will learn how to:
- Prepare and process video and audio assets for digital avatar cloning.
- Use Hoppscotch on Olares to call the Duix.Avatar API collection to train a model, synthesize audio, and create a video.

## Prerequisites
Before you begin, ensure the following:
- Olares 1.11 or later.
- Olares running on a machine equipped with an NVIDIA GPU.

## Install Duix.Avatar
1. In **Market**, search for "Duix.Avatar".
   ![Duix.Avatar](/images/manual/use-cases/duix-avatar.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Install Hoppscotch
In addition to Duix.Avatar, you also need Hoppscotch, an open-source API development environment to interact with the Duix.Avatar service.
1. In **Market**, search for "Hoppscotch".
   ![Hoppscotch](/images/manual/use-cases/hoppscotch.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Prepare media files
Generating a digital avatar requires a source video to use as a template for the face and voice. You will need a 10-20 second video clip of a person speaking clearly while facing the camera.

You must then separate this source video into two files: a silent video and an audio-only file. This guide uses `ffmpeg` for this step.

:::info Ensure ffmpeg is installed
To follow this guide using the `ffmpeg` command, ensure it is installed on your local computer. See https://www.ffmpeg.org/download.html.
:::
1. Open your terminal, `cd` into the folder containing your video, and run the following command:
   ```
   # Replace input.mp4 with your actual filename
    ffmpeg -i input.mp4 -c:v copy -an output_video.mp4 -c:a pcm_s16le -f wav output_audio.wav
   ```
   This creates two new files in the same folder:
   - `output_video.mp4` (silent video)
   - `output_audio.wav` (audio)
2. The Duix.Avatar service reads files from specific directories. Upload the two files you just generated to their designated locations in the Olares **Files** app.
   1. Upload `output_audio.wav` to:
   ```plain
   /Data/heygem/voice/data/ 
   ```
   ![Upload source audio](/images/manual/use-cases/duix-avatar-upload-source-audio.png#bordered)

   2. Upload `output_video.mp4` to:
   ```plain
   /Data/heygem/face2face-data/temp/
   ```
   ![Upload source video](/images/manual/use-cases/duix-avatar-upload-source-video.png#bordered)
## Import the API collection to Hoppscotch
A pre-configured Hoppscotch collection is available to simplify the API calls.
1. Run the following command in your terminal to download the API collection file:
    ```bash
    curl -o duix.avatar.json https://cdn.olares.com/app/demos/en/duix/duix.avatar.json
    ```
2. Open the Hoppscotch app in Olares.
3. In the collections panel on the right, click **Import** > **Import from Hoppscotch**, and select the `duix.avatar.json` file you just downloaded.
   ![Import from Hoppscotch](/images/manual/use-cases/duix-avatar-import-from-hoppscotch.png#bordered)

After importing, you will see a new collection named `duix.avatar` containing four pre-configured requests.
   ![Check collection](/images/manual/use-cases/duix-avatar-check-collection.png#bordered)

## Train data via API
Now you will call the four APIs in sequence to generate the digital avatar.
:::tip
The Duix.Avatar API address is tied to your Olares ID. In all of the following API requests, you must replace `<OLARES_ID_PREFIX>` in the URL with your own Olares ID prefix. For example, if your Olares access URL is `https://app.alice123.olares.com`, your prefix is `alice123`.
:::
 
### Step 1: Model training
This step preprocesses your uploaded audio, extracting features to prepare for voice cloning.

1. In Hoppscotch, expand the `duix.avatar` collection and select `1. Model Training Request`.
2. Modify the request URL, replacing `<OLARES_ID_PREFIX>` with your Olares ID's prefix.
   :::info
   The request body is pre-set to point to the `output_audio.wav` file you uploaded, so you don't need to change it.
   :::
3. Click **Send** to begin pre-training.
   A successful request returns a JSON response. Copy the values for `reference_audio_text` and `asr_format_audio_url` for later use.
   ![Pretrain](/images/manual/use-cases/duix-avatar-pretrain.png#bordered)

### Step 2: Audio synthesis
This step uses the voice model you trained in Step 1 to synthesize new audio from a text prompt.
1. Click **2. Audio Synthesis Request**. 
2. Modify the Olares ID in the request URL. 
3. In the request body, modify the following fields:
   * `text`: Enter the text you want the digital avatar to speak.
   * `reference_audio`: Paste the `asr_format_audio_url` value from Step 1.
   * `reference_text`: Paste the `reference_audio_text` value from Step 1.
   * Other parameters can be left as their defaults.
   ![Edit audio parameters](/images/manual/use-cases/duix-avatar-edit-audio-parameters.png#bordered)

4. Click **Send** to synthesize the audio. A successful request will return an audio file.

5. In the response area, click <span class="material-symbols-outlined">more_vert</span> to download the audio in MP3 format.
   ![Generate audio file](/images/manual/use-cases/duix-avatar-generate-audio-file.png#bordered)

6. Rename the downloaded file to `new.mp3`. In the same folder, convert it to `.wav` with `ffmpeg`:
    ```bashß
   ffmpeg -i new.mp3 new.wav
   ```
7. Upload the new `new.wav` file to:
   ```plain
   /Data/heygem/face2face-data/temp/
    ``` 
   ![Upload audio](/images/manual/use-cases/duix-avatar-upload-audio.png#bordered)

### Step 3: Video synthesis
Now you will merge your new synthesized audio (`new.wav`) with your original silent video (`output_video.mp4`) to create the final avatar.

1. Click **3. Video Synthesis Request**.
2. Modify the Olares ID in the request URL.
3. In the request body, change the `code` field to a new, unique task identifier. You will use this ID to check the synthesis progress.
   :::info
   The `audio_url` and `video_url` in the request body are pre-set to `new.wav` and `output_video.mp4`, which match the files you uploaded. They do not need to be changed.
   :::
4. Confirm the settings and click **Send**. A successful response will return `"success": true`, indicating the task has been submitted.
   ![Submit task](/images/manual/use-cases/duix-avatar-submit-task.png#bordered)

### Step 4: Query video synthesis progress
Video synthesis is a time-consuming task. Use this to query its processing status.
1. Click **4. Video Query Request**.
2. Modify the Olares ID in the request URL.
3. In the **Params** section, change the `code` value to the unique identifier you set in Step 3. 
4. Click **Send** to check the current progress.
5. Repeat this query until the `progress` field in the response reaches `100`, which indicates the video synthesis is complete.
   ![Task completed](/images/manual/use-cases/duix-avatar-task-completed.png#bordered)
   :::tip
   The time required for video synthesis depends on your GPU performance and video length. It may take several minutes or longer.
   :::
6. When successful, the `result` field in the response will contain the output video's filename. You can find the final generated video in the Olares Files app at:
    ```plain
   /Data/heygem/face2face-data/temp/
    ```
   ![Check video in Files](/images/manual/use-cases/duix-avatar-check-video-in-files.png#bordered)

## FAQ
### Progress is stuck or synthesis fails
If the progress query stalls for a long time or an API returns an error, go to Control Hub, find the container named `heygemgenvideo`, and check its logs for detailed error messages.
![Check Duix.Avatar in Control Hub](/images/manual/use-cases/duix-avatar-check-in-controlhub.png#bordered)

### API request fails
Confirm the following:
- You have correctly replaced the default Olares ID (`<OLARES_ID_PREFIX>`) with your own ID in the request URL.
- All media files (`output_audio.wav`, `output_video.mp4`, `new.wav`) are uploaded to the correct directories with the exact filenames.

### Media is updated, but the old video is still generated
Ensure you are using a new, unique `code` parameter for the video synthesis. The system caches results, so reusing a `code` will return the previously cached video.

