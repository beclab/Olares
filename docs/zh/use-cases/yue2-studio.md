---
outline: [2, 3]
description: 在 Olares 上选择 YuE2 Studio 模型套装，创作并编辑歌曲，更换歌词或风格，续写旋律，根据录音制作翻唱，并通过 Lares 生成音乐。
head:
  - - meta
    - name: keywords
      content: Olares, YuE2 Studio, AI 音乐, Q8, BF16, 量化, 钢琴卷帘, ABC, MIDI, 翻唱, 纯音乐, Lares, MCP, HTTP API
app_version: "0.0.7"
doc_version: "2.2"
doc_updated: "2026-11-06"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/yue2-studio.md)为准。
:::

# 用 YuE2 Studio 创作并编辑歌曲

YuE2 Studio 是一款本地 AI 音乐应用，可根据音乐风格和歌词生成带人声与伴奏的歌曲。它还会生成可编辑的乐谱，方便你先修改曲谱，再生成新的演唱或演奏版本。

本教程使用一首名为“Morning Light”的英文原声流行短歌作为示例。你将生成原曲、升高人声旋律的音高，并加快歌曲速度。也可以用自己的歌词和音乐风格完成相同操作。

## 学习目标

通过本教程，你将学会：

- 根据音乐创作任务选择模型套装和 GPU 配额。
- 在生成音频前，先谱曲并试听乐谱。
- 编辑音符、和弦、段落、调性和速度。
- 保留曲谱，更换歌词、改变风格或制作器乐版本。
- 续写一段短旋律，并根据录音制作翻唱。
- 选择生成参数，并与音乐编辑器交换乐谱。
- 通过 MCP 连接 Lares，在对话中生成歌曲。

## 前提条件

- 已在 amd64 架构的 Olares 设备上安装 YuE2 Studio。本教程适用于测试市场中的 Studio 3.8.5 预览版。
- 具有 sm_120 计算能力的 NVIDIA GeForce RTX 50 系列 GPU，以及兼容 CUDA 13 的驱动。示例使用 RTX 5090 Laptop GPU 完成。
- 已在 Olares 系统环境中配置用于下载模型的 Hugging Face 服务地址和令牌。

应用会读取 `OLARES_SYSTEM_HUGGINGFACE_SERVICE` 和 `OLARES_SYSTEM_HUGGINGFACE_TOKEN`。在 Olares 中打开 **Settings** > **Advanced** > **System environment variables** 可查看这些值。如需修改，请参阅[设置系统环境变量](../manual/manage-system-env.md)。这些设置也会影响其他使用它们的应用。

Olares 会在启动应用时分配 GPU 资源。如果分配提示导致应用无法启动，请按照[恢复应用并分配资源](../manual/olares/settings/gpu-resource.md#恢复应用并分配资源)中的步骤操作。

## 准备音乐引擎

四套预设使用同一个 YuE2 主模型，主要区别是保存权重时采用的精度。较低精度可减小模型文件，也可能引入更多量化误差。

| 模型套装 | 主模型精度 | 设置中显示的套装大小（约） | 适用场景 |
|:---|:---|:---|:---|
| Light | `Q5_K_M` | 4.0 GB | 需要最小的模型套装，并接受更强的权重压缩。 |
| Balanced | `Q6_K` | 4.3 GB | 希望在 Light 与 Quality 之间折中。 |
| Quality | `Q8_0` | 5.1 GB | 日常生成和编辑从这一档开始。本 Olares 应用包以此作为推荐基线。 |
| Full native | `BF16` | 9.9 GB | 希望与原始主模型权重对比，并能分配更多内存和显存。 |

四套模型的音频 VAE 都使用 `F32`。Light、Balanced 和 Quality 使用 `Q8_0` 版本的 SheetSage2 与 MERT 处理录音转谱，Full native 使用它们的 `F32` 版本。每套模型还包含解码器配套权重，因此套装大小不只是主模型文件的大小。

对于本教程使用的 RTX 5090 Laptop GPU，建议从 Quality Q8 开始。Full native 可用于对比，但下载量和内存需求更大。更高的权重精度不保证每个提示词都得到更好的结果，较小的文件也不保证在每张 GPU 上都生成得更快。

1. 从 Launchpad 打开 YuE2 Studio。
2. 点击左下角的账户名打开 **Settings**，然后选择 **Models**。
3. 选择一套音乐引擎模型。可以先使用 **Quality - Q8_0 backbone, near lossless**。**Full native - BF16 backbone, original weights** 会下载原始主模型权重，占用更多存储空间。
4. 下载所选模型，等待状态显示为 **installed**。只需准备一套模型。
5. 点击 **Done**，返回 **Create**，等待显示 **Engine ready** 后再开始创作。

![YuE2 Studio 设置中的音乐引擎模型套装](/images/manual/use-cases/yue2-studio-model-sets.png#bordered)

模型权重与应用分开下载，重启应用后仍会保留。设置中显示的模型套装大小是文件大小，并非生成时的显存峰值。

Olares 应用包为单首 Q8 歌曲声明了 5 GiB 最低显存请求。时间分片模式不会将它作为固定显存上限。在一次 32 步、约 130 秒的 Q8 生成中，观测到的整卡显存峰值约为 4.48 GiB。一次 BF16 生成达到约 5.72 GiB。这些样本可辅助选择配额，但更长的素材、更多变体和录音转谱都可能改变资源占用。本环境尚未测量 Q5、Q6 的资源占用。

切换到 BF16 或增加并发前，先分配更多显存。在 Memory slicing 模式下，可按照[恢复应用并分配资源](../manual/olares/settings/gpu-resource.md#恢复应用并分配资源)中的步骤更改配额。切换已安装的模型套装会重启音乐引擎，请先完成当前任务。

## 生成原曲

### 填写风格与歌词

1. 在 **Create** 中选择 **Studio**。
2. 将标题设为 `Morning Light Original`。
3. 在 **Style** 中输入以下内容：

    ```text
    English acoustic pop, warm female vocals, piano, acoustic guitar, gentle drums, hopeful, clean production, 100 BPM
    ```

4. 在 **Lyrics** 中输入以下原创示例歌词。每个段落标签单独占一行，段落之间留一个空行。

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

### 谱曲并生成音频

1. 在 **Score (ABC)** 下保持选中 **Melody + chords**，点击 **Compose the score**。
2. 等待乐谱出现。点击 **Listen**，用简单音色试听音符，先检查旋律，再生成演唱版本。
3. 本例使用以下设置：

    | 设置 | 示例值 | 用途 |
    |:---|:---|:---|
    | **Max duration** | `2:00` | 为短歌留出完整结束的时间。这是时长上限，并非目标时长。 |
    | **Flow-matching steps** | `32` | 设置音频合成步数。 |
    | **Solver** | **Midpoint (reference)** | 使用参考渲染方法。 |
    | **Songs per request** | `1` | 创建一份曲谱。 |
    | **Variations** | `1` | 生成一个演唱或演奏版本。 |

4. 为了方便比较，关闭 **Randomize seed**，将 **Composition seed** 设为 `20261106`，**Sound seed** 设为 `42`。
5. 点击 **Create**。等待歌曲出现在 **Library** 中，然后播放试听。

![YuE2 Studio 创作界面中的英文歌曲](/images/manual/use-cases/yue2-studio-create.png#bordered)

生成的歌曲会保留乐谱和生成设置，之后可以继续基于它制作新版本。

谱曲模式有三种：

| 谱曲模式 | 提供的内容 | 适用场景 |
|:---|:---|:---|
| **Melody + chords** | 带和弦符号的旋律 | 创作需要进一步编辑旋律与和声的歌曲。 |
| **Melody** | 只有旋律的乐谱 | 翻唱时保留曲调，让新风格决定和声。 |
| **No score** | 不使用符号乐谱，直接生成音频 | 不以乐谱编辑为目标，探索更自由的结果。 |

## 编辑乐谱并生成新版本

### 修改旋律与速度

1. 在 **Library** 中找到 `Morning Light Original`，点击信息图标打开 **Song Details**。
2. 在乐谱区域点击 **Edit**，打开钢琴卷帘编辑器。
3. 选择 **Vocal** 轨道。依次选择 **Edit** > **Select All**，然后选择 **Edit** > **Transpose**。输入 `2`，将这条轨道升高两个半音，并确认修改。
4. 将底部栏的 **BPM** 改为 `120`，加快歌曲速度。
5. 点击右上角的 **Apply**，将编辑后的乐谱带回创作表单。

![乐谱编辑器中的人声音符、和弦与段落栏，以及速度控件](/images/manual/use-cases/yue2-studio-edit-score.png#bordered)

如果只想做局部调整，可以上下拖动单个音符来改变音高，水平移动来改变起始时间，或拖动右边缘来改变音符长度。点击音符上方栏中的和弦或段落标签，即可编辑相应内容。

:::warning 刷新前保存乐谱
刷新页面会清空创作表单中尚未保存的 ABC 乐谱。请使用 **Save the score as .abc** 保存文件，或先生成歌曲，将乐谱存入 Library。
:::

### 生成编辑后的版本

1. 将标题改为 `Morning Light Variation`，以便与原曲区分。
2. 将 **Style** 中的 `100 BPM` 改为 `120 BPM`，与编辑后的乐谱保持一致。
3. 启用乐谱下方的 **Keep the edited score for new words**。这样在修改歌词或风格时仍会保留曲谱。第一次对比时，保持示例歌词不变。
4. 保持其他生成设置不变，然后点击 **Create**。
5. 在 Library 中分别播放原曲和新版本，对比人声旋律与节奏快慢。

![原曲和编辑后的版本分别保存在曲库中](/images/manual/use-cases/yue2-studio-compare-versions.png#bordered)

**Apply** 用于更新表单中的乐谱。**Create** 用于生成新的音频。原曲仍会保留在 Library 中。

钢琴卷帘中的移调操作只影响所选轨道。如果希望旋律与和声一起改变调性，请使用下方的整曲移调方法。

## 细调曲谱

### 绘制、移动和拉长音符

1. 打开已保存歌曲的 **Song Details**，点击乐谱旁的 **Edit**。
2. 选择 **Vocal** 编辑人声旋律，或选择 **Ins** 编辑器乐声部。
3. 选中铅笔工具后，在网格空白处点击添加音符。可以先在已有乐句中加入一个短音。
4. 上下拖动音符可改变音高，水平拖动可改变起始时间。
5. 拖动右边缘可改变音符长度。需要对齐固定拍点细分时，使用网格吸附设置。
6. 在编辑器中试听乐句，然后点击 **Apply**。使用新标题并点击 **Create**，听取生成后的效果。

横向位置表示时间，纵向位置表示音高，音符宽度表示时值。水平滚动可查看后面的小节，垂直滚动可将较高或较低的旋律移入视野。

### 编辑和弦与段落

1. 在乐谱编辑器中找到音符上方的 **Chords** 栏。
2. 点击和弦标签并输入新和弦。例如，在本例的 F 大调歌曲中，可以尝试将一个和弦改为 `Gm7`。
3. 在 **Sections** 栏中修改段落名，或添加 `Interlude` 等段落标记。
4. 试听乐谱，然后点击 **Apply** 并生成新版本。

让段落名与歌词标签保持一致。如果加入器乐间奏，可以在歌词中使用对应的 `[Instrumental Break]` 标签。生成后试听段落衔接，单靠段落标记不能决定编曲的具体听感。

### 改变调性或人声音区

整首歌移调的方法如下：

1. 点击 **Use this score** 载入已保存的乐谱。
2. 展开 **Advanced**，找到 **Transpose score**。
3. 输入半音偏移量，例如 `2` 表示整体升高，`-2` 表示整体降低。
4. 使用新标题并点击 **Create**，新版本会根据移调后的乐谱生成。
5. 继续其他编辑前，将 **Transpose score** 重置为 `0`，除非还要对当前版本再次移调。

如果只想调整人声八度，使用乐谱旁的 **Voice −8va** 或 **Voice +8va**。它们将人声声部移动一个八度，伴奏和和弦符号保持原样。生成前先试听，并查看是否出现人声音域提示。八度调整改变的是乐谱中的音区，声音特征仍通过 **Style** 描述。

## 复用或续写歌曲

### 保留曲谱，更换歌词

1. 打开原曲的三点菜单，选择 **Sing again with new words**，将乐谱和提示词带入创作表单。
2. 为新版本设置标题，例如 `Morning Light New Words`。
3. 将副歌换成另一段四行歌词。例如：

    ```text
    [Chorus]
    Follow me into the day
    Let our worries drift away
    Hear a new song in the air
    There is hope for us to share
    ```

4. 第一次尝试时保留原来的主歌，尽量让新歌词的句长和重读音节与原歌词相近。
5. 启用乐谱下方的 **Keep the edited score for new words**，然后点击 **Create**。
6. 试听是否有词语出现过早、拖得过长或发音不清。缩短或改写这些句子，再生成一个版本。

如果通过 **Use this score** 载入乐谱，需要自行填写用于演唱的歌词和风格。带标记的乐谱原本对应其他歌词或风格时，需要启用保留选项。如果应用提示编辑后的乐谱属于其他歌词或风格，先核对表单，再启用该选项，继续使用预期的曲谱。

### 改变声音或音乐风格

1. 载入原曲乐谱，第一次比较时保持歌词不变。
2. 在 **Style** 中完整描述新的演唱和编曲方式。例如：

    ```text
    English jazz ballad, warm baritone vocals, upright bass, brushed drums, jazz piano, intimate and mellow, 100 BPM
    ```

3. 如果出现 **Keep the edited score for new words**，将其启用。
4. 使用 `Morning Light Jazz Version` 等新标题，然后点击 **Create**。

保留乐谱会保留写好的曲谱，由模型生成新的演唱或演奏。请试听后判断声音和编曲是否符合描述。风格提示词不能克隆某位歌手的声音。

### 将旋律转为纯音乐

1. 载入原曲乐谱，点击谱面旁的 **Instrumental**，将人声旋律转入器乐声部。
2. 将风格改为纯音乐描述，例如：

    ```text
    Instrumental acoustic pop, piano, acoustic guitar, gentle drums, calm and hopeful, 100 BPM
    ```

3. 删除演唱文字，保留 `[Verse]`、`[Chorus]` 等段落标签。
4. 如果出现保留选项，将其启用。使用新标题，然后点击 **Create**。

此流程使用乐谱旁的 **Instrumental** 按钮。**Create** 附近独立的 **Instrumental only** 开关还可能添加 Instrumental LoRA，在缺少权重时下载该适配器。它与四套基础模型分开提供。

### 续写一段短旋律

当你已经有开头的乐句，希望 YuE2 写出后续部分时，可以使用此功能。

1. 输入完整歌曲的预期歌词和风格。
2. 在 **Score (ABC)** 中粘贴一小段开头乐谱，或使用 **Open an .abc score** 载入 ABC 文件。保留 ABC 文件头、调性、拍号和速度，以及至少一个包含人声音符的乐句。
3. 启用 **Continue the song from this score**。
4. 点击 **Continue the score**，先写出后续乐谱，再生成音频。
5. 试听完整乐谱，并按需在 **Score editor** 中修改。设置新标题后点击 **Create**。

如果想从已保存的歌曲开始，先导出它的 ABC。在文件副本中保留文件头和开头的主歌，删除后面的段落，再导入。保留两个声部的声明，以及剩余段落对应的完整小节。如果开头只有休止符或器乐前奏，需要先加入一段演唱或写好的人声旋律，才能引导这种续写流程。

### 将录音翻唱为新风格

Cover 流程先用 SheetSage2 将录音转写为旋律，再根据这份乐谱生成新的演唱。生成前可以检查并修正转录结果。

1. 在 **Create** 中选择 **Cover**，保持启用 **Melody only**。这样会保留曲调，让新风格决定和声。
2. 选择来源：
    - 使用电脑上的文件时，点击 **Choose a recording…**。
    - 使用曲库中的演示原曲时，打开 `Morning Light Original` 的三点菜单，选择 **Transcribe to score: Melody for a cover**。
3. 等待显示 **Score ready**。使用 **Listen** 试听旋律，按需在 **Score editor** 中修正音符或速度。
4. 粘贴原曲歌词，段落标签应与录音对应。第一次翻唱时保留原词。
5. 在 **Style** 中描述新编曲，例如使用前面的爵士提示词。设置 `Morning Light Jazz Cover` 等新标题。
6. 检查 **Max duration** 是否为录音留出足够的结束时间，然后点击 **Create**。

![将来源录音转为乐谱，准备制作英文爵士翻唱](/images/manual/use-cases/yue2-studio-cover-score.png#bordered)

如果应用提示缺少 SheetSage2 转录模型，请在 **Settings** > **Models** 中添加。本教程所用的预设套装已包含它。来源录音编曲较密或旋律不清楚时，应仔细检查转录结果。根据转录乐谱生成的翻唱可能与原录音有所不同。

## 选择生成参数

### 从和弦进行开始

如果想先确定和声，再让模型写旋律，可以使用 **Chord bed**。

1. 在 **Studio** 中准备风格和歌词，将谱曲模式设为 **Melody + chords**。
2. 打开乐谱下方的 **Chord bed**。
3. 设置 **Tempo, BPM**、**Key** 和 **Meter**，例如 `100`、`C` 和 `4/4`。
4. 点击 **Write the bed**。如果表单中已有乐谱，按钮会显示为 **Write another bed**。
5. 检查生成的和弦符号。底稿由和弦与休止符组成，段落依据歌词标签划分。
6. 点击 **Create**，让模型在这些和声基础上生成旋律和编曲。

写入和弦底稿会替换表单中的当前乐谱。使用前，请先导出想要保留的乐谱。

![和弦底稿的速度、调性、拍号，以及生成的和弦进行](/images/manual/use-cases/yue2-studio-chord-bed.png#bordered)

### 权衡合成步数与求解器

第一个版本保持 **32** 步和 **Midpoint (reference)**。想更快试听一个想法时，可以尝试 **16** 步并比较结果。减少合成步数可能改变声音，具体影响取决于歌曲。

另一种求解器 **AB2 (about twice as fast)** 在声音生成阶段使用较少的网络计算次数。可以在同一曲谱上尝试，并试听人声、起音和段落衔接的变化。谱曲、模型加载和其他阶段仍然需要时间，因此界面上的名称不代表整个请求的耗时都会减半。

![生成设置中的合成步数、求解器和歌曲数量](/images/manual/use-cases/yue2-studio-quality-settings.png#bordered)

### 有目的地使用种子与变体

| 控件 | 影响的内容 | 使用方法 |
|:---|:---|:---|
| **Composition seed** | 模型根据提示词写出的计划 | 比较提示词或乐谱修改时保持固定。 |
| **Sound seed** | 演唱或演奏合成时的噪声 | 改变它，尝试同一曲谱的另一种渲染。 |
| **Songs per request** | 不同的曲谱 | 每首歌都会单独谱曲，占用更多工作内存。 |
| **Variations** | 同一首曲谱的不同演唱或演奏版本 | 使用最低 Q8 配额时从 `1` 开始。 |

如果 **Songs per request** 不可用，表示本地引擎当前将歌曲数量限制为一首。修改 **Settings** > **Local engine** 中的引擎限制前，先完成当前任务。尝试更多并发任务前，先增加 GPU 配额。使用 5 GiB 的 Q8 基线时，保持一首歌曲、一个变体。

### 重新渲染已有演唱

1. 打开歌曲的三点菜单，选择 **Re-render**。
2. 在对话框中调整步数、声音种子、变体数量或输出格式。
3. 提交新的渲染任务，并与已保存的歌曲比较。

重新渲染会复用已保存的演唱计划，跳过重新谱曲。需要改变音符、歌词或风格时，使用乐谱编辑流程。相同设置有助于比较，但不保证得到逐字节完全相同的音频文件。

![重渲染对话框中的步数、变体、声音种子和输出格式](/images/manual/use-cases/yue2-studio-rerender.png#bordered)

## 保存与交换作品

### 保存音频与 ABC 乐谱

1. 打开想要保留的版本的 **Song Details**。
2. 点击 **Download Audio**，将生成的录音保存到电脑。
3. 点击乐谱区域中的 **.abc**，保存可编辑的乐谱。需要把已保存歌曲的乐谱带回创作表单时，点击 **Use this score**。

之后需要恢复文件时，打开 **Create**，点击乐谱工具栏中的 **Open an .abc score**。先试听，再填写用于生成的风格和歌词。

### 与 MIDI 编辑器交换音符

点击乐谱旁的 **MIDI**，将音符和速度导出到其他音乐编辑器。如果希望保留原始和弦符号与乐谱文本，也请同时保存 ABC 文件。

将 MIDI 文件带回歌曲创作流程的方法如下：

1. 选择 **Cover**，点击 **Choose a recording…**，选择 MIDI 文件。
2. 在 **Load MIDI** 中，按需要选择 **Melody only** 或 **With chords**。
3. 检查轨道表，并明确选择 **Voice** 和 **Instrument**。对于本例导出的文件，将 **Voice** 设为 **1. Vocal Melody**，**Instrument** 设为 **2. Ins Melody**。自动识别可能选中另一条旋律轨道。
4. 选择是否保留文件中的段落标记。接受结果前，检查时间网格、八度调整和转换提示。
5. 等待转换完成，**Use as the score** 按钮可用后点击它，并试听导入的乐谱。MIDI 轨道中的和弦或速度变化，可能需要简化后才能写入歌曲乐谱。
6. 填写歌词和风格，然后使用新标题并点击 **Create**，生成新的演唱或演奏。

![在 MIDI 导入对话框中明确指定人声和器乐轨道](/images/manual/use-cases/yue2-studio-import-midi.png#bordered)

将 MIDI 转回 ABC 时，和弦名称可能会重新推断。在将导入结果视为原和声的副本前，请先检查这些名称。

### 分享 PDF 前先检查谱面

乐谱旁的 **PDF** 按钮可导出适合打印的谱面。分享前请检查生成的页面。此预览版中发现部分降号和弦符号显示不正确，需要保留准确且可编辑的乐谱时，请使用 ABC 或 MIDI。

### 查找并比较已保存的版本

使用曲库搜索，输入 `Morning Light` 等共同标题筛选版本。为不同版本添加清楚的标题后缀，比较音频前打开 **Song Details**，核对歌词和已保存的生成参数。歌曲旁的时长表示音频长度，相对时间表示它何时加入曲库，都不是生成耗时。

## 通过 Lares 生成歌曲

你可以通过 YuE2 Studio 的 MCP 服务，让 Lares 生成歌曲，再到 Studio 中试听或编辑乐谱。此流程需要 Olares 中的 YuE2 Studio 应用包为 0.0.5 或更新版本，并且提供 **YuE2 Studio MCP & API** 入口。

在同一台 Olares 上安装 [Lares](lares.md)，并选择支持工具调用的对话模型。本例使用 Qwen3.8。启动对应的模型应用，等待它在 Lares 中可用。对话模型负责理解请求，YuE2 负责生成音乐。同时保持 YuE2 Studio 运行，安装好一套音乐模型，并等待显示 **Engine ready**。

如果两个模型使用同一张 GPU，需要为它们留出足够显存，并且一次只提交一个音乐生成请求。时间分片可以共享 GPU 计算时间，但不会自动释放已加载模型占用的显存。

### 连接 MCP 服务

令牌认证默认关闭，入口仍遵循 Olares 的内部访问策略。所有能够访问此入口的客户端都可以调用工具。如需应用令牌认证，请按照[开启令牌认证](#开启令牌认证)操作。

1. 在 Olares 中打开 **Settings** > **Applications** > **YuE2 Studio** > **Entrances**。
2. 选择 **YuE2 Studio MCP & API**，复制其 **Endpoint** 地址。如果地址末尾没有 `/mcp`，请将其补上。此内部入口可供同一台 Olares 上的 Lares 访问。
3. 在 Lares 中打开 **Settings** > **MCP**，点击 **Add server**，填写：

    | 设置 | 值 |
    |:---|:---|
    | **Server name** | `yue2studio` |
    | **Transport** | `Streamable HTTP` |
    | **MCP URL** | 第 2 步获取的、以 `/mcp` 结尾的入口地址 |
    | **Headers** | `{}` |

4. 点击 **Save**。Lares 会连接服务并载入工具。

### 开启令牌认证

应用环境变量 `MCP_AUTH_ENABLED` 同时控制 MCP 和 HTTP API 的令牌认证，默认值为 `false`。

1. 在 Olares **Settings** > **Applications** > **YuE2 Studio** 中找到 **Environment variables**，点击 **Manage environment variables**。
2. 编辑 `MCP_AUTH_ENABLED`，设为 `true`，点击 **Confirm**，再点击 **Apply**。应用会重启以应用修改，请等待恢复可用。设置界面的操作说明请参阅[管理应用环境变量（英文）](../../manual/olares/settings/manage-app-env.md)。
3. 打开 Files，找到 **Application** > **Data** > **yue2studio** > **config**。将 `mcp-token` 作为文本打开，复制完整的一行。如果无法预览文本，可以下载后用文本编辑器打开。
4. 在 Lares **Settings** > **MCP** 中编辑 `yue2studio`，将 **Headers** 替换为下方的 JSON。把 `<your-token>` 替换为复制的完整一行，并在 `Bearer` 后保留一个空格。

    ```json
    {
      "Authorization": "Bearer <your-token>"
    }
    ```

5. 点击 **Save**，后续请求需要携带令牌。

令牌由应用自动生成，正常升级后会保持不变。不要把它放进聊天提示词或分享的截图。如需再次关闭令牌认证，将 `MCP_AUTH_ENABLED` 设为 `false` 并应用修改，客户端即可将 **Headers** 留空（`{}`）。

### 在对话中生成 Morning Light

1. 在 Lares 中新建对话，选择 Qwen3.8 或另一个可用且支持工具调用的对话模型。
2. 先让 Lares 检查 Studio：

    ```text
    Use the configured yue2studio MCP server to call studio_status. Report the actual result. Do not generate music or change settings.
    ```

    确认 Lares 调用了 `studio_status` 并返回 Studio 的当前状态。如果它只描述准备如何操作，请要求它实际使用已配置的 MCP 工具。

3. 发送以下请求。你可以修改标题、风格或歌词，换成自己的歌曲构思。示例中的 30 秒、16 步和 AB2 求解器用于快速生成第一首歌。

::: details 示例：通过 Lares 生成 Morning Light
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

4. 等待 Lares 跟进任务直到完成。如果它请求运行生成工具的权限，请先审阅请求，再允许执行。提交后返回的任务 ID 表示歌曲已排队或正在生成，还需要等待完成并获得保存后的歌曲 ID。
5. 打开 YuE2 Studio，在曲库中搜索 `Morning Light` 并试听新曲。打开 **Song Details** 查看歌词和已保存的设置，再点击 **Download Audio** 保存音频。
6. 如需修改曲谱，点击 **Use this score**，按照[编辑乐谱并生成新版本](#编辑乐谱并生成新版本)继续操作。生成修改后的版本时，请使用新标题。

调用失败时，先阅读错误信息再重试。开启令牌认证后，HTTP 401 表示缺少应用令牌或令牌不正确。如果 Lares 连普通聊天消息也无法回答，请先检查它选择的对话模型和连接，再重新提交音乐请求。生成失败时，打开 YuE2 Studio，检查音乐引擎状态和可用显存。

### 使用其他客户端或 HTTP API

其他兼容 MCP 的客户端可以使用同一个 Streamable HTTP 地址。同一入口也提供应用的 HTTP API。去掉 MCP 地址末尾的 `/mcp` 即可得到基础地址。当 `MCP_AUTH_ENABLED=true` 时，每次请求都要在 Authorization 请求头中携带 Bearer 令牌。

| 任务 | HTTP 请求 |
|:---|:---|
| 提交生成任务 | `POST /v1/music/jobs` |
| 查询任务状态和已保存的歌曲 | `GET /v1/music/jobs/{job_id}` |
| 下载已保存歌曲的音频 | `GET /v1/library/media/{song_id}` |

生成任务异步执行。请保留返回的任务 ID，等待任务完成后再下载音频。工具说明与自动化示例请参阅 [YuE2 Studio MCP 参考](https://github.com/timoncool/YuE2-Studio/blob/main/docs/mcp-skill.md)。

## 关于其他 Studio 工具

上面的核心流程使用音乐引擎及其附带的转录模型。界面中的其他工具，例如助手写作、LoRA 训练、分轨、VST 处理和全乐器音频转 MIDI，可能需要额外的服务商、模型或平台组件，未纳入此 Linux 预览版的已验证范围。本教程介绍的乐谱 MIDI 导出属于核心流程。

:::info 模型使用条款
YuE2 模型权重使用 CC BY-NC 4.0 许可。在项目中使用前，请查看 [YuE2 模型说明](https://huggingface.co/m-a-p/YuE2-3B)。
:::

## 了解更多

- [Lares](lares.md)：配置助手及其对话模型。

- [管理加速资源](../manual/olares/settings/gpu-resource.md)：为 Olares 应用分配或释放 GPU 资源。
- [设置系统环境变量](../manual/manage-system-env.md)：管理共享的 Hugging Face 下载设置。
