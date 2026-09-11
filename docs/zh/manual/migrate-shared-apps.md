---
outline: [2, 3]
description: 为 Olares 1.12.6 之前安装的旧版 v2 共享应用选择并完成正确的迁移方式。
head:
  - - meta
    - name: keywords
      content: Olares, 迁移共享应用, v2 共享应用, ComfyUI, Dify, OnlyOffice, SearXNG, Xinference, Ollama
---

# 迁移旧版共享应用

如果共享应用安装于 Olares 1.12.6 之前，并需要迁移到当前共享应用架构，请使用本指南。更新 Olares 后，旧版 v2 应用仍可继续运行，但无法直接升级到新架构。

当前架构的说明，参见[关于共享应用](olares/market/shared-apps.md)。

:::warning 卸载前检查应用数据
卸载旧版共享应用可能删除无法直接恢复到新版应用的数据。请先确认应用类型，并完成对应的导出或备份，再卸载旧应用。
:::

## 识别旧版 v2 应用

在应用市场中打开应用详情页，查看**信息**面板中的**兼容性**：

- 旧版 v2 共享应用通常显示 `Olares >=1.12.3-0, <1.12.6`。
- 当前共享应用显示 `Olares >=1.12.6-0`。

如果应用不符合上述版本范围，请不要直接使用本指南。先记录应用版本和 Chart 版本，再选择迁移方式。

## 选择迁移方式

| 应用或数据情况 | 迁移方式 |
|---|---|
| ComfyUI | 使用应用提供的自动迁移流程 |
| Falco、MTranServer 或其他没有用户数据的应用 | 确认无需保留数据后重新安装 |
| Dify、OnlyOffice、SearXNG 或 Xinference | 导出或备份应用数据，再恢复到新版应用 |
| 独立 Ollama 共享应用 | 在引擎基座中重新部署所需模型，再重新连接客户端 |

## 迁移 ComfyUI

按照 [ComfyUI 迁移说明](/zh/use-cases/comfyui-common-issues.md)执行受支持的迁移。确认新版数据位置后，再移除旧应用。

## 重新安装没有数据的应用

只有在确认应用中没有需要保留的用户数据、设置或工作流后，才能使用此方式。

1. 记录旧应用的名称和版本。
2. 卸载旧版 v2 应用。
3. 在应用市场中找到当前共享应用，确认兼容性范围以 `Olares >=1.12.6-0` 开头。
4. 安装当前共享应用，并从启动台打开。

## 手动迁移应用数据

不同应用能够保留的数据不同。卸载旧应用前，请完成对应的迁移指南：

- [迁移 Dify](/zh/use-cases/dify-upgrade.md)
- [迁移 OnlyOffice](/zh/use-cases/onlyoffice-migration.md)
- [迁移 SearXNG](/zh/use-cases/searxng.md)
- [迁移 Xinference](/zh/use-cases/xinference.md)

除非对应迁移指南明确要求，否则不要把旧应用的数据目录直接恢复到新版本。

## 从 Ollama 迁移到引擎基座

1. 记录独立 Ollama 共享应用中的模型，以及连接这些模型的客户端应用。
2. 使用[引擎基座应用](/zh/use-cases/llm-base-apps.md)部署每个所需模型。
3. 打开模型控制台，复制新的 **Base URL**。
4. 在每个客户端应用中，将 Ollama 端点替换为新的 Base URL。
5. 从每个客户端发送测试请求。确认客户端能正常使用新模型后，再移除旧 Ollama 应用。
