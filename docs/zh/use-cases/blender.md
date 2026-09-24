---
outline: [2, 3]
description: 使用 Lares 在 Olares 上搭建 Blender 场景，渲染太阳系插图，并将项目与图片保存到文件管理器。
head:
  - - meta
    - name: keywords
      content: Olares, Blender, Lares, MCP, 3D 建模, 3D 场景, 渲染, 动画
app_version: "0.1.27"
doc_version: "1.0"
doc_updated: "2026-09-24"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/blender.md)为准。
:::

# 使用 Lares 创建 Blender 场景

Blender 是一套用于建模、动画和渲染的 3D 创作工具。在 Olares 上，你既可以在浏览器中使用 Blender 桌面界面，也可以通过模型上下文协议（MCP），让 Lares 直接操作同一个场景。

本示例将分步搭建一个太阳系场景，并把可继续编辑的项目文件和 PNG 图片保存到文件管理器。完成静态场景后，还可以把它制作成循环动画。

## 前提条件

- Olares 1.12.7 或更高版本。
- 已在同一台 Olares 上安装并配置 [Lares](lares.md)，且所选模型支持工具调用。

## 安装 Blender

1. 打开 Market，搜索“Blender”。
   <!-- ![Market 中的 Blender](/images/manual/use-cases/blender.png#bordered) -->

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

如果安装时需要选择计算模式，请根据硬件和任务选择：

| 模式 | 适用场景 |
|:---|:---|
| Intel | 使用可用的 Intel 核显完成本例中的视口渲染和桌面串流。尤其适合独立显卡正被本地模型占用的情况。 |
| CPU | 没有合适的 GPU 时选择。由于桌面串流需要软件编码，查看桌面时会占用更多 CPU。 |
| NVIDIA | 使用分配给 Blender 的 NVIDIA GPU 完成负载更高的渲染任务。如果本地模型也在高负载使用同一块显卡，两边速度都可能变慢。 |

三种模式的 MCP 配置相同。本示例不要求使用 NVIDIA GPU。

## 启动 Blender

1. 从启动台打开 Blender，等待默认场景出现。
2. 保持 Blender 应用运行。浏览器标签页可以关闭，Lares 不需要它一直处于打开状态。

浏览器中显示的是串流桌面。若想以更低延迟查看场景变化，请先在电脑上启用 LarePass VPN，再打开 Blender。

:::tip Olares 和电脑在同一局域网？
也可以使用 Olares 的 `.local` 地址。地址格式和 Windows 配置方法请参阅[通过本地地址访问 Olares 服务](../manual/best-practices/local-access.md)。这只会改变查看桌面的方式，Lares 仍使用内部 MCP 入口。
:::

## 在 Lares 中配置 Blender

1. 打开 Olares **Settings**，前往 **Applications** > **Blender** > **Entrances**。
2. 选择 **Blender MCP**，复制其中的 **Endpoint** URL。
3. 打开 Lares，前往 **Settings** > **MCP**，按以下信息添加服务器：

   | 设置 | 值 |
   |:---|:---|
   | Server name | `blender` |
   | Transport | `Streamable HTTP` |
   | MCP URL | 粘贴刚才复制的 Endpoint。如果结尾没有 `/mcp`，请手动补上。 |
   | Headers | `{}` |

4. 保存并启用服务器。
5. 新建一个对话，发送以下提示词：

   ```text
   使用 Blender MCP 检查当前场景。告诉我 Blender 版本、场景名称、
   对象名称和对象总数。不要修改任何内容。
   ```

Lares 应调用 Blender 工具并返回场景内容。新建的默认场景通常包含 `Camera`、`Cube` 和 `Light`。

请选择 **Blender MCP** 入口，不要选择 **Blender** 桌面入口。本应用包的内部 MCP 连接不需要应用令牌，因此 Headers 保持为空。对于本例这种在同一设备上运行的工作流，请勿将该入口公开到外部网络。

## 创建太阳系场景

继续前请先保存需要保留的工作。Lares 会直接修改浏览器中显示的同一个场景，第一个提示词会清空当前场景。

请在同一个对话中逐条发送以下提示词。等待当前步骤完成后，再发送下一条。示例中的尺寸和速度用于视觉呈现，并非按真实比例模拟太阳系。

### 设置太阳和相机

1. 在文件管理器中创建 **Home** > **blender-renders** 文件夹。如果该文件夹中已有需要保留的内容，请新建一个文件夹，并在后续所有提示词中使用新的名称。
2. 发送以下提示词：

   ```text
   使用 Blender MCP 创建一个新的太阳系场景。清空当前场景。
   在中心放置一个半径为 1.5、带橙色辉光的太阳。
   在太阳处添加能量为 26000 的点光源，并关闭阴影。
   把 42 mm 相机放在 (14, -36, 13.5)，对准场景中心。

   使用 Eevee（BLENDER_EEVEE）、48 个视口采样，图片尺寸设为
   1600x900。将 PNG 渲染到
   /config/Home/blender-renders/lares-solar-system.png。

   对于当前 Olares 中的 Blender 应用包，请在 VIEW_3D temp_override
   环境中使用 bpy.ops.render.opengl(write_still=True, animation=False)，
   并设置 CAMERA 视图和 RENDERED 着色。此 MCP 工作流不要使用
   bpy.ops.render.render()。仅使用 bpy、bmesh 和 mathutils。
   在一次 execute_blender_code 调用中完成本步骤，并返回图片保存路径。
   ```

3. 在文件管理器中打开 **Home** > **blender-renders**，确认 `lares-solar-system.png` 已生成，再继续下一步。

提示词的最后一段告诉 Lares 如何使用当前应用包支持的方式完成渲染，你不需要自己编写或运行 Python。随着场景逐步完善，后续提示词会覆盖这张示例 PNG。

### 添加行星和轨道

发送：

```text
继续编辑当前 Blender 场景，不要清空。按水星到海王星的顺序添加八颗行星，
颜色应能体现各自的典型特征。把它们放在 XY 平面的圆形轨道上，
并为每颗行星设置不同的角度。

依次使用以下轨道半径/行星半径：
3.0/0.30、4.2/0.48、5.7/0.54、7.1/0.38、
9.3/1.45、11.4/1.15、13.0/0.80、14.3/0.78。

为每条轨道绘制一条细而暗的蓝色圆环，管状半径设为 0.02。
保留太阳和相机。使用相同的视口渲染方法和路径再次渲染 PNG。
在一次 execute_blender_code 调用中完成本步骤。
```

查看更新后的图片。如果行星重叠过多，先让 Lares 调整它们在各自轨道上的位置，再添加细节。

### 添加表面细节

发送：

```text
继续编辑当前场景。为木星和土星添加宽阔的横向云带，
在土星周围添加三道薄而扁平的行星环。
为地球添加蓝色海洋和绿色陆地，并在旁边添加一颗小型月球。

使用程序化材质，不要下载纹理。保持相机和轨道不变。
使用相同的视口渲染方法再次渲染 PNG。
```

### 完善并保存场景

发送：

```text
完成当前太阳系场景。在 (-24, -30, 18) 添加能量为 6000 的冷色补光，
再在 (20, 26, 10) 添加能量为 3000 的区域光，并让两盏灯都对准场景中心。
添加深色星空背景。

确保八颗行星都清晰可见。使用相同的视口渲染方法再次渲染 PNG，
然后把可编辑的项目保存到
/config/Home/blender-renders/lares-solar-system.blend。
告诉我可以在 Olares 文件管理器的什么位置找到这两个文件。
```

![Blender 渲染的太阳系场景](/images/manual/use-cases/blender-solar-system.png#bordered)

上图由 Olares 上的 Blender 渲染。实际结果可能因所用模型和后续调整而不同。

## 查看和使用文件

打开文件管理器，前往 **Home** > **blender-renders**：

| 文件 | 用途 |
|:---|:---|
| `lares-solar-system.png` | 预览、下载或分享渲染图片。 |
| `lares-solar-system.blend` | 在 Blender 中重新打开项目并继续编辑。 |

Blender 中的 `/config/Home` 对应文件管理器中的 **Home**。Lares 和 Blender 使用不同的工作目录，因此请让 Blender 把需要共享的文件保存到 `/config/Home`，而不是 `/tmp` 或 Lares 自己的工作区。

## 制作行星动画

静态场景达到预期效果后，在同一个对话中继续操作。本例将生成 72 帧、每秒 24 帧的图片序列，对应一段三秒循环动画。

1. 渲染前，先让 Lares 设置运动轨迹：

   ```text
   为当前太阳系制作动画，不要清空场景或移动相机。
   使用第 1-72 帧，帧率为 24 fps。所有行星从上方看都应按恒定速度
   逆时针公转。

   在一个循环中，水星、金星、地球、火星、木星、土星、天王星和海王星
   分别公转 8、6、5、4、3、2、2 和 1 圈。
   轨道环、太阳和相机保持不动。每颗行星都沿各自轨道运动，
   土星环始终跟随土星。

   为每颗行星在原点创建一个 Empty 对象，用以下线性帧表达式驱动公转：
   start_angle + turns * 2 * pi * (frame - 1) / 72。
   不要使用带缓动的关键帧。第 73 帧应与第 1 帧完全一致，
   但只导出第 1-72 帧。渲染前先返回每颗行星的公转圈数。
   ```

2. 检查 Lares 返回的圈数，然后发送：

   ```text
   以 960x540 分辨率、24 个视口采样渲染第 1-72 帧，
   输出为质量 92 的 JPEG 图片。继续使用 VIEW_3D 相机视图和
   RENDERED 着色环境，并调用 bpy.ops.render.opengl(animation=True)。
   将输出文件名前缀设为
   /config/Home/blender-renders/lares-solar-system-frames/f。
   输出图片序列，不要使用 FFMPEG。保存更新后的 .blend 项目。
   完成后返回输出文件夹路径和已保存的帧数。
   ```

3. 在文件管理器中打开 **Home** > **blender-renders** > **lares-solar-system-frames**，确认其中有 72 张图片。

:::details 将图片序列转换为 MP4
本指南所用的 Blender 版本不包含 FFmpeg 视频输出功能。请把图片序列文件夹下载到已安装 FFmpeg 的电脑上，在该文件夹中打开终端并运行：

```bash
ffmpeg -framerate 24 -start_number 1 -i f%04d.jpg -c:v libx264 -pix_fmt yuv420p -crf 20 solar-system.mp4
```

该命令会在图片序列旁生成 `solar-system.mp4`。再次导出时，请使用新的文件名。
:::

## 常见问题

### 为什么 Lares 找不到 Blender 工具？

确认 Blender 正在运行，且 Lares 中已启用对应服务器。Transport 应设置为 **Streamable HTTP**，URL 应来自 **Blender MCP**，并以 `/mcp` 结尾。如果仍无法连接，请从启动台打开一次 Blender，然后重试只读的场景检查。

### 浏览器标签页需要一直打开吗？

不需要。保持 Blender 应用运行即可。仅在需要查看或手动编辑场景时打开浏览器标签页。如果串流桌面响应缓慢，请使用 LarePass VPN 或前文介绍的本地地址。

### 为什么项目已经保存，却没有生成 PNG？

在本应用包的 MCP 工作流中，通过 `bpy.ops.render.render()` 进行常规渲染时，可能完成操作但没有写入图片。请让 Lares 使用第一个提示词中的视口渲染方法，然后再次查看文件管理器。此解决方法仅适用于本指南所用的应用包版本。

### 为什么 Lares 需要很长时间才能完成？

每次只让 Lares 修改一部分场景。例如，把表面细节拆成云带、土星环和地球三个请求。等待当前工具调用完成后，再发送下一项修改。

## 了解更多

- [Lares](lares.md)：配置本示例使用的 AI 助手。
- [通过本地地址访问 Olares 服务](../manual/best-practices/local-access.md)：从电脑连接 Blender 串流桌面。
