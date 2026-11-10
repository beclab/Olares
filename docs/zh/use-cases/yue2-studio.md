---
outline: [2, 3]
description: 在 Olares 上选择 YuE2 Studio 模型套装，创作并编辑歌曲，更换歌词或风格，续写旋律，并根据录音制作翻唱。
head:
  - - meta
    - name: keywords
      content: Olares, YuE2 Studio, AI 音乐, Q8, BF16, 量化, 钢琴卷帘, ABC, MIDI, 翻唱, 纯音乐
app_version: "0.0.7"
doc_version: "2.6"
doc_updated: "2026-11-07"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/yue2-studio.md)为准。
:::

# 用 YuE2 Studio 创作并编辑歌曲

YuE2 Studio 是一款本地 AI 音乐应用，可根据音乐风格和歌词生成带人声与伴奏的歌曲。它还会生成可编辑的乐谱，方便你先修改曲谱，再生成新的演唱或演奏版本。

本教程使用一首名为“Morning Light”的英文原声流行短歌作为示例。你将生成原曲、升高人声旋律的音高，并加快歌曲速度。也可以用自己的歌词和音乐风格完成相同操作。

## 学习目标

通过本教程，你将学会：

- 安装 YuE2 Studio，并为模型下载做好准备。
- 根据音乐创作任务选择模型套装和 GPU 配额。
- 在生成音频前，先谱曲并试听乐谱。
- 编辑音符、和弦、段落、调性和速度。
- 保留曲谱，更换歌词、改变风格或制作器乐版本。
- 续写一段短旋律，并根据录音制作翻唱。
- 选择生成参数，并与音乐编辑器交换乐谱。

## 前提条件

- 一台运行 Olares 1.12.7 或更高版本的 amd64 设备，配备 NVIDIA GPU，并有至少 5 GiB 可用显存。
- 一个 Hugging Face 账号。

## 安装 YuE2 Studio

YuE2 Studio 从 Hugging Face 下载模型套装。安装应用前，先在 Olares 中添加一个只读令牌。

1. 登录 Hugging Face，或注册一个账号。
2. 打开 [Access Tokens](https://huggingface.co/settings/tokens)，点击 **New token**。
3. 输入令牌名称，选择 **Read** 权限，创建并复制令牌。
4. 在 Olares 中打开 **Settings** > **Advanced** > **System environment variables**。
5. 找到 `OLARES_SYSTEM_HUGGINGFACE_TOKEN`，点击编辑图标，将令牌粘贴到 **Value**。
6. 点击 **Confirm**，再点击页面底部的 **Apply**。Olares 已在 `OLARES_SYSTEM_HUGGINGFACE_SERVICE` 中提供 Hugging Face 服务地址。
7. 打开 Market，搜索 `YuE2 Studio`。
8. 点击 **Get**，再点击 **Install**，等待安装完成。

令牌会保存为系统环境变量，并供使用 Olares Hugging Face 设置的应用共享。如需稍后更新，请参阅[设置系统环境变量](../manual/manage-system-env.md)。

## 下载模型套装

四套预设使用同一个 YuE2 主模型，主要区别是保存权重时采用的精度。较低精度可减小模型文件，也可能引入更多量化误差。

| 模型套装 | 主模型精度 | 设置中显示的套装大小（约） | 适用场景 |
|:---|:---|:---|:---|
| Light | `Q5_K_M` | 4.0 GB | 需要最小的模型套装，并接受更强的权重压缩。 |
| Balanced | `Q6_K` | 4.3 GB | 希望在 Light 与 Quality 之间折中。 |
| Quality | `Q8_0` | 5.1 GB | 日常生成和编辑从这一档开始。本 Olares 应用包以此作为推荐基线。 |
| Full native | `BF16` | 9.9 GB | 希望使用原始主模型权重，并能分配更多内存和显存。 |

日常生成和编辑建议从 Quality Q8 开始。Full native 保留主模型的原始精度，会占用更多存储空间和显存。

1. 从 Launchpad 打开 YuE2 Studio。
2. 首次打开时，YuE2 Studio 会直接提示下载模型套装。如果之前打开过应用，请点击左下角的账户名，然后进入 **Settings** > **Models**。
3. 选择一套音乐引擎模型。可以先使用 **Quality - Q8_0 backbone, near lossless**。**Full native - BF16 backbone, original weights** 会下载原始主模型权重，占用更多存储空间。
4. 下载所选模型，等待状态显示为 **installed**。
5. 点击 **Done**，返回 **Create**，等待显示 **Engine ready** 后再开始创作。

![YuE2 Studio 设置中的音乐引擎模型套装](/images/manual/use-cases/yue2-studio-model-sets.png#bordered)

下载的模型套装会在应用重启后保留。应用初始分配 5 GiB 显存。使用 Full native、生成更长的歌曲、转写录音或一次生成多个结果时，需要增加显存配额。在 Memory slicing 模式下，可按照[恢复应用并分配资源](../manual/olares/settings/gpu-resource.md#恢复应用并分配资源)中的步骤更改配额。切换模型套装会重启音乐引擎，请先完成当前歌曲。

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
    | **Max duration** | `2:00` | 允许歌曲最长生成两分钟。 |
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
| **No score** | 不使用符号乐谱，直接生成音频 | 在进入乐谱工作流前探索音频效果。 |

## 编辑乐谱并生成新版本

### 修改旋律与速度

1. 在 **Library** 中找到 `Morning Light Original`，点击信息图标打开 **Song Details**。
2. 在乐谱区域点击 **Edit**，打开钢琴卷帘编辑器。
3. 选择 **Vocal** 轨道。依次选择 **Edit** > **Select All**，然后选择 **Edit** > **Transpose**。输入 `2`，将这条轨道升高两个半音，并确认修改。
4. 将底部栏的 **BPM** 改为 `120`，加快歌曲速度。
5. 点击右上角的 **Apply**，将编辑后的乐谱带回创作表单。

![乐谱编辑器中的人声音符、和弦与段落栏，以及速度控件](/images/manual/use-cases/yue2-studio-edit-score.png#bordered)

如果只想做局部调整，可以上下拖动单个音符来改变音高，水平移动来改变起始时间，或拖动右边缘来改变音符长度。点击音符上方栏中的和弦或段落标签，即可编辑相应内容。

:::warning 刷新页面前保存乐谱
刷新页面前，点击 **Save the score as .abc**。也可以先生成歌曲，将乐谱保存到 Library。
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

让段落名与歌词标签保持一致。如果加入器乐间奏，可以在歌词中使用对应的 `[Instrumental Break]` 标签。生成新版本后试听段落衔接，再按需要调整乐谱或风格描述。

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

如果通过 **Use this score** 载入乐谱，请填写歌词和风格，再启用 **Keep the edited score for new words** 以保留载入的曲谱。

### 改变声音或音乐风格

1. 载入原曲乐谱，第一次比较时保持歌词不变。
2. 在 **Style** 中完整描述新的演唱和编曲方式。例如：

    ```text
    English jazz ballad, warm baritone vocals, upright bass, brushed drums, jazz piano, intimate and mellow, 100 BPM
    ```

3. 如果出现 **Keep the edited score for new words**，将其启用。
4. 使用 `Morning Light Jazz Version` 等新标题，然后点击 **Create**。

保留乐谱会保留写好的曲谱，由模型生成新的演唱或演奏。试听生成结果并继续调整风格描述，直到声音和编曲适合这首歌。

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

如果想从已保存的歌曲开始，先导出它的 ABC。在文件副本中保留文件头、两个声部的声明、开头的主歌和该主歌对应的完整小节，删除后面的段落，再导入。至少保留一段人声旋律来引导续写。

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

如果应用提示缺少 SheetSage2 转录模型，请在 **Settings** > **Models** 中添加。本教程使用的模型套装已包含它。来源录音编曲较密或旋律不清楚时，先检查并修正转录结果，再生成翻唱。

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

另一种求解器 **AB2 (about twice as fast)** 可以加快声音生成阶段。两种求解器的谱曲和模型加载时间相同。可以在同一曲谱上尝试 AB2，并试听人声、起音和段落衔接的变化。

![生成设置中的合成步数、求解器和歌曲数量](/images/manual/use-cases/yue2-studio-quality-settings.png#bordered)

### 有目的地使用种子与变体

| 控件 | 影响的内容 | 使用方法 |
|:---|:---|:---|
| **Composition seed** | 模型根据提示词写出的计划 | 比较提示词或乐谱修改时保持固定。 |
| **Sound seed** | 演唱或演奏合成时的噪声 | 改变它，尝试同一曲谱的另一种渲染。 |
| **Songs per request** | 不同的曲谱 | 每首歌都会单独谱曲，占用更多工作内存。 |
| **Variations** | 同一首曲谱的不同演唱或演奏版本 | 使用最低 Q8 配额时从 `1` 开始。 |

本地引擎初始限制为一次生成一首歌曲。修改 **Settings** > **Local engine** 中的引擎限制前，先完成当前歌曲。一次生成多首歌曲或多个变体前，请增加 GPU 配额。

### 重新渲染已有演唱

1. 打开歌曲的三点菜单，选择 **Re-render**。
2. 在对话框中调整步数、声音种子、变体数量或输出格式。
3. 提交新的渲染任务，并与已保存的歌曲比较。

重新渲染会复用已保存的演唱计划。调整声音种子、合成步数、变体数量或输出格式，可以生成新的演唱或演奏。需要改变音符、歌词或风格时，请使用乐谱编辑流程。

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
3. 选择 **Voice** 和 **Instrument** 轨道。对于本例导出的文件，将 **Voice** 设为 **1. Vocal Melody**，**Instrument** 设为 **2. Ins Melody**。
4. 选择是否保留文件中的段落标记。接受结果前，检查时间网格、八度调整和转换提示。
5. 等待转换完成，**Use as the score** 按钮可用后点击它，并试听导入的乐谱。MIDI 轨道中的和弦或速度变化，可能需要简化后才能写入歌曲乐谱。
6. 填写歌词和风格，然后使用新标题并点击 **Create**，生成新的演唱或演奏。

![在 MIDI 导入对话框中明确指定人声和器乐轨道](/images/manual/use-cases/yue2-studio-import-midi.png#bordered)

将 MIDI 转回 ABC 后，请检查推断出的和弦名称，并按需要在乐谱编辑器中进行调整。

### 导出可打印乐谱

点击乐谱旁的 **PDF**，导出适合打印的谱面。如果还要继续编辑音符和和弦符号，请导出 ABC 或 MIDI。

### 查找并比较已保存的版本

使用曲库搜索，输入 `Morning Light` 等共同标题筛选版本。为不同版本添加清楚的标题后缀，比较音频前打开 **Song Details**，核对歌词和已保存的生成参数。

:::info 模型使用条款
YuE2 模型权重使用 CC BY-NC 4.0 许可。在项目中使用前，请查看 [YuE2 模型说明](https://huggingface.co/m-a-p/YuE2-3B)。
:::

## 了解更多

- [通过 Lares 使用 YuE2 Studio 创作歌曲](yue2-studio-lares.md)：连接 MCP 服务，并在对话中生成音乐。
- [使用令牌保护 YuE2 Studio MCP](yue2-studio-mcp-auth.md)：要求 Lares 和其他 MCP 客户端提供 Bearer 令牌。
- [开始使用 Lares](lares.md)：准备助手和支持工具调用的对话模型。
- [管理加速资源](../manual/olares/settings/gpu-resource.md)：为 Olares 应用分配或释放 GPU 资源。
- [设置系统环境变量](../manual/manage-system-env.md)：管理共享的 Hugging Face 下载设置。
