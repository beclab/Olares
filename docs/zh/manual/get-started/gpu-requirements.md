---
search: false
---

安装 Olares 无需 GPU，但大多数 AI 应用需要 GPU 才能正常运行。目前仅支持 NVIDIA 显卡。

- **架构**：Turing 或更新架构（GTX 16xx、RTX 20xx、30xx、40xx、50xx 系列及之后产品）。
:::info
旧架构 GPU 无法被 Olares 识别，依赖 GPU 的 AI 应用也无法运行。
:::
- **显存**：建议至少 8 GB。即使是受支持的显卡，如果显存过小，也会导致许多 AI 应用无法运行。

:::details 不确定显卡是否受支持？
在终端中运行以下命令，并查看输出中的代号前缀：

```bash
lspci | grep -i nvidia
```

示例输出：

```
3b:00.0 VGA compatible controller: NVIDIA Corporation AD102 [GeForce RTX 4090] (rev a1)
```

常见代号前缀与显卡架构及是否受支持的对应关系如下：

| 代号前缀 | 架构 | 是否支持 |
|:---|:---|:---:|
| GB | Blackwell | ✓ |
| AD | Ada Lovelace | ✓ |
| GA | Ampere | ✓ |
| TU | Turing | ✓ |
| GP | Pascal | ✗ |
| GM | Maxwell | ✗ |

你也可以参考 NVIDIA 开源驱动仓库中提供的[完整兼容 GPU 列表](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus)进行确认。
:::
