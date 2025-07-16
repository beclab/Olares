---
description: 在 Olares 上部署 Stable Diffusion 的指南，包括模型安装、多用户配置和系统优化，轻松实现 AI 图像生成。
---
# Stable Diffusion

Stable Diffusion 是基于潜在扩散模型（LDM）的新一代 AI 图像生成技术。它通过在低维空间处理图像生成过程，显著降低了计算需求，同时保持了高保真度的输出质量。

通过 Olares 平台，你可以轻松部署和管理 Stable Diffusion。无需关注文件系统和数据库等底层配置，直接进入创作环节。

借助 Olares 的多用户支持特性，团队成员可以共享同一个 Stable Diffusion 部署实例，并确保各自数据的私密性。这种方式避免了重复安装系统而过度消耗硬件资源的问题。

## Stable Diffusion 能做什么？

无论你是想扩展创作工具库的艺术家、将 AI 图像集成到工作流程中的开发者，还是对 AI 艺术创作充满好奇的探索者，Stable Diffusion 能支持：

* 将文字描述转换为图像
* 基于参考图片生成新图像
* 对已有图片进行局部重绘或扩展
* 图像风格迁移和艺术化处理
* 生成高清图像及超分辨率处理

## 安装 SD Web UI
:::info
从 Olares 1.11.6 开始，如果已安装 "SD Web UI For Cluster" 或 "SD Web UI" 客户端入口，需先卸载这些版本。
:::

1. 从应用市场里安装 SD Web UI 共享版。
2. 在桌面启动 SD Web UI 图标开始创作。请确保管理员已安装 SD Web UI 共享版。

## 多人使用避免冲突

模型检查点（checkpoint）默认采用全局共享机制——当有用户切换检查点后，其他用户的后续操作都会使用新选择的检查点。为避免互相影响，建议在创建任务时指定专用检查点。

![检查点配置说明](/images/manual/use-cases/sd-checkpoint.png)

1. 全局检查点设置
2. 单任务检查点设置

## 修改参数配置
在 Olares 中，启用 SD Web UI 默认使用 `--xformers` 参数，用于：
- 减少显存占用
- 加速图像生成
- 支持更高分辨率图像

但是， `--xformers` 会带来一定影响：
- 生成图像的风格多样性可能降低
- 对提示词的解释准确度略有下降

如需移除该参数，请按以下步骤操作：

:::info
只有 Olares 管理员可以通过控制面板应用调整系统参数。
:::

1. 在控制面板中，选择**浏览**。
2. 在管理员命名空间下找到 **sdwebui**。
3. 在**部署**中点击 **sdwebui**。
4. 点击右上角菜单 <i class="material-symbols-outlined">more_vert</i>，选择**编辑 YAML**。
5. 在 YAML 编辑器中找到并删除 `--xformers` 参数。默认的 YAML 文件内容类似如下：

    ```yaml {5}
    env:
      - name: CLI_ARGS
        value: >-
          --allow-code --enable-insecure-extension-access --api
          --no-hashing --gradio-queue --xformers
    ```

6. 点击**确定**使变更生效。

## 画廊

<table>
  <tr>
    <td><img src="/images/manual/use-cases/sd-example1.png" alt="示例图片 1" width="200" /></td>
    <td><img src="/images/manual/use-cases/sd-example2.png" alt="示例图片 2" width="200" /></td>
  </tr>
  <tr>
    <td><img src="/images/manual/use-cases/sd-example3.png" alt="示例图片 3" width="200" /></td>
    <td><img src="/images/manual/use-cases/sd-example4.png" alt="示例图片 4" width="200" /></td>
  </tr>
</table>