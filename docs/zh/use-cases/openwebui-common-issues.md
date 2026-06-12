---
outline: [2, 3]
description: Open WebUI 在 Olares 上的常见问题与解决方法。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, 常见问题, 故障排除, 模型下载
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-14"
---

# 常见问题

使用本页面识别并解决 Open WebUI 在 Olares 上的常见问题。

## 模型应用卡在 "Waiting for Ollama" 或 "Needs attention"

如果模型应用在这些状态停留超过几分钟：

1. 前往**设置** > **GPU**。
2. 检查你的 GPU 模式：
   - 如果你正在使用**显存分片**，需确保模型应用已关联到 GPU，并分配了足够的 VRAM。
   - 如果你正在使用**应用独占**，需确保独占应用设置为你的模型应用。
3. 从 Launchpad 重启模型应用，然后再次检查状态。

## 下载进度消失

通过下拉菜单下载模型时，进度条有时可能会在下载完成前消失。

要继续下载：
1. 再次点击模型选择器。
2. 输入完全相同的模型名称。
3. 选择 **Pull from Ollama.com**。下载会从上次中断的位置继续。

## 麦克风出现 "Permission denied" 错误

尝试使用听写按钮或 Voice Mode 时，你可能会收到以下错误消息：
- `Permission denied when accessing microphone: NotAllowedError: Permission denied`
- `Permission denied when accessing media devices`

Olares 桌面会在嵌入式框架（iframe）中显示应用。出于严格的安全和隐私考量，现代浏览器会阻止嵌入式框架访问麦克风等敏感硬件，即使你已经在系统设置中授予浏览器相应权限。

要绕过此安全限制并使用麦克风：
1. 在 Olares 桌面的 Open WebUI 窗口右上角，选择 <i class="material-symbols-outlined">open_in_new</i>，在新的浏览器标签页中打开。
2. 在新的浏览器标签页中，点击聊天界面的麦克风图标。
3. 当浏览器提示时，允许麦克风访问。
