---
outline: [2, 3]
description: Configure Olares' video playback, transcoding, and hardware acceleration parameters to optimize media playback performance and compatibility.
---

# Video settings

The **Video** settings page allows you to configure Olares' video playback, transcoding, and hardware acceleration parameters to optimize media playback performance and compatibility.

## Hardware acceleration

Enable and configure hardware acceleration for video processing in Olares. Hardware acceleration can significantly reduce CPU load and speed up transcoding. Use the dropdown menu to select the type of hardware acceleration your system supports. If unsure, select "None".

**Supported accelerators**: AMD AMF, Nvidia NVENC, Intel QuickSync (QSV), Video Acceleration API (VAAPI), Rockchip MPP (RKMPP), Apple VideoToolBox, Video4Linux2 (V4L2).

## Encoding scheme

Choose the video encoding formats to be used during transcoding. When the selected format does not support hardware acceleration, the system automatically falls back to software encoding.

| Setting                          | Description                                                                       |
| :------------------------------- | :-------------------------------------------------------------------------------- |
| **Allow encoding in HEVC format**| When enabled, videos can be transcoded to HEVC (H.265) format.                    |
| **Allow encoding in AV1 format** | When enabled, videos can be transcoded to AV1 format.                             |

-----

## Transcoding settings

Provide detailed settings for the transcoding process to balance performance and playback quality.

| Setting                 | Description                                                                                                                                                                                   |
|:------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Transcoding threads** | Sets the maximum number of threads used for transcoding. <br/>Reducing threads lowers CPU usage but may slow down transcoding <br/>and affect playback smoothness. <br/>Recommended to keep this to "Auto".  |
| **Transcoding path**    | Specifies a custom path to store transcoded files. <br/>Leave blank to use the server's default path.                                                                                              |

## Audio transcoding

Configure options for multi-channel audio downmixing and encoding.

| Setting                          | Description                                                                                                                           |
| :------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------ |
| **Enable VBR audio encoding** | Enables Variable Bitrate encoding. <br/>This often provides better quality but may cause compatibility<br/> or buffering issues in a few cases. |
| **Audio boost when downmixing** | Sets the amount to boost audio volume during downmixing. A value <br/>of 1 retains the original volume.                                      |
| **Stereo downmix algorithm** | Selects the algorithm used to downmix multi-channel audio to stereo.<br>Options include: None, Dave750, NightmodeDialogue.                 |


## Encoding quality

Adjust video encoding quality to balance file size, CPU consumption, and video quality.

| Setting                | Description                                                                                                                                  |
|:-----------------------|:---------------------------------------------------------------------------------------------------------------------------------------------|
| **Encoder preset**     | Choose a faster preset value to improve performance or a slower value<br/> to improve quality.                                                    |
| **H.265 Encoding CRF** | Sets the Constant Rate Factor (CRF) for H.265 encoding. The value <br/>ranges from 0-51; a lower value means higher quality but larger file size. |
| **H.264 Encoding CRF** | Sets the CRF for H.264 encoding. The value ranges from 0-51; <br/>a lower value means higher quality but larger file size.                        |

::: tip
CRF is the default quality setting for H.264 and H.265 encoders. A reasonable range is typically between 18 and 28. We recommend using the default values for both as an initial reference.
:::

## Other

Provide more advanced settings related to transcoding and streaming to optimize the playback experience and resource usage.

| Setting                         | Description                                                                                                                                                                               |
| :------------------------------ |:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Throttle transcoding** | Pauses the transcoding or remuxing process when the progress is far <br/>ahead of the current playback position to reduce resource consumption.<br/> Disable it if playback issues occur. |
| **Delete segments** | Automatically deletes old segments after the client finishes <br/>downloading them to prevent fully transcoded files from taking<br/> up disk space. Disable it if playback issues occur.  |
| **Throttle delay** | The buffer time (in seconds) before the transcoder starts throttling. <br/>Ensure the client has a sufficient buffer. Only valid when **Throttle <br/>transcoding** is enabled.           |
| **Segment retention time** | The retention time (in seconds) for segments after the client<br/> downloads them. Only valid when **Delete segments** is enabled.                                                        |
| **Maximum muxing queue** | Sets the retention time (in seconds) for segments after the client <br/>downloads them. Only valid when **Delete segments** is enabled.                                                   |


