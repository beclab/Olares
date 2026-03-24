---
outline: [2, 3]
description: 排查在 Olares 1.12.5 中，已暂停的应用在独占模式下无法移除的问题。
---

# 无法在应用独占模式移除已暂停应用

在**应用独占**模式下，如果你暂停并移除了某个应用，但该应用随后再次出现，可参考本指南进行排查。

## 适用情况

- 在**应用独占**模式下，暂停某个应用并将其移除后，该应用仍会再次出现在**选择独占应用**区域。
- 该应用在设置、应用市场和启动台中仍显示为**暂停**状态。

## 原因

在 Olares 1.12.5 中，当你将 GPU 模式切换为**应用独占**后，系统会自动选中一个应用作为独占应用。

如需移除该应用，必须先将其暂停。但如果该应用采用的是客户端/服务端（C/S）架构，暂停操作可能只暂停了客户端，服务端仍在后台运行。因此，该应用可能仍在占用 GPU 资源，从而再次出现在**选择独占应用**区域。

## 解决方案

1. 打开**设置** > **GPU**，确认当前**选择独占应用**区域显示的应用。

    ![独占模式下显示的应用](/images/zh/manual/help/ts-cs-app-reappears-stopped-app.png#bordered){width=90%}

2. 前往**设置** > **应用**，找到该应用，然后点击**恢复**。

    ![恢复应用](/images/zh/manual/help/ts-cs-app-reappears-resume.png#bordered){width=90%}

3. 等待片刻，待应用启动后，点击**暂停**。

    ![暂停应用](/images/zh/manual/help/ts-cs-app-reappears-stop.png#bordered){width=90%}

4. 返回**设置** > **GPU**，然后点击<i class="material-symbols-outlined">link_off</i>，再次移除该应用。

5. 刷新列表，确认该应用不再出现在**选择独占应用**区域。

    ![再次检查独占模式](/images/zh/manual/help/ts-cs-app-reappears-no-apps.png#bordered){width=90%}

6. 现在你可以恢复想要使用的应用，将其设置为独占应用。