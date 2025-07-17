---
description: ComfyUI 在 Olares 上的使用指南，通过节点式界面精确控制 AI 图像生成过程，创建可重用的工作流程。
---
# ComfyUI

ComfyUI 是一款基于节点的 Stable Diffusion 图形界面工具。它把 AI 绘图过程变成了可视化编程，让使用者能像搭建积木一样，通过连接各种功能节点来实现完整的绘图流程。从提示词编写、模型选择到后期处理，每个环节都能精确把控。

与 Stable Diffusion WebUI 简单直观的界面不同，ComfyUI 让你通过组合代表不同功能的节点，可以构建出自己的工作流程。这不仅让你能更好地掌控绘图过程，还可以把常用的复杂操作保存下来重复使用，也方便与他人分享。

## ComfyUI 能做什么？
通过 ComfyUI，你可以实现以下功能：
  
* 用可视化方式搭建和复用工作流
* 对绘图流程进行精细调优
* 自由组合不同模型和技术
* 导出导入工作流，方便分享
* 使用相同配置批量处理图片
* 添加高级图像后期效果

## 安装 ComfyUI 共享版
Olares 应用商店提供 ComfyUI 共享版，可允许同一 Olares 集群上的多个用户共享 ComfyUI 的模型、插件和工作流资源。它还提供了一个启动器（ComfyUI Launcher），帮助管理员用户管理 ComfyUI 资源和运行环境。

::: tip 注意
自 1.11.6 版本起，Olares 会使用 ComfyUI 共享版取代之前的集群范围应用。如果安装过 ComfyUI For Cluster 和对应的 ComfyUI 客户端，请卸载后再安装共享版。
::: 

要安装 ComfyUI 共享版：

1. 打开 Olares 应用市场，找到 ComfyUI 共享版，并点击**获取**。

   - 管理员将会在 Olares 桌面看到两个图标：一个是 ComfyUI 客户端界面，另一个是 ComfyUI 启动器。

    ![安装 ComfyUI 共享版](/images/manual/use-cases/install-comfyui.png){width=40%}

   - 成员用户只会在桌面看到 ComfyUI 客户端界面。

    :::tip 启动 ComfyUI 服务
    管理员必须从启动器启动 ComfyUI 服务之后，集群中的所有用户才能从客户端入口访问服务。
    :::

2. 点击 ComfyUI 图标打开界面。管理员也可以从启动器进入 ComfyUI 界面。

    ![ComfyUI](/images/manual/use-cases/comfyui.png#bordered)

## 了解更多

- [使用 ComfyUI 启动器管理 ComfyUI](comfyui-launcher.md)
- [Krita + ComfyUI 实时绘画](comfyui-for-krita.md)：了解如何利用 ComfyUI 助力 Krita 中的创意工作流。