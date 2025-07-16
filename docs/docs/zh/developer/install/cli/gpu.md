# `gpu`

## 命令说明

`gpu` 命令用于管理 GPU 功能，包括安装、启用、禁用和卸载 GPU 驱动及相关组件，以及查看 GPU 安装状态。

```bash
olares-cli gpu <子命令> [选项]
```

:::info
- 通常情况下，Olares 安装脚本会自动检测显卡硬件和 CUDA 驱动，并根据检测结果自动安装、启用显卡相关组件和服务。
- 当前仅支持 NVIDIA 显卡。
:::

## 子命令

| 子命令         | 描述                                                               |
|-------------|------------------------------------------------------------------|
| `install`   | 安装 GPU 驱动和相关依赖。|
| `enable`    | 启用 GPU 功能，支持 GPU 应用运行。                                           |
| `disable`   | 禁用 GPU 功能，停止支持 GPU 应用运行。                                         |
| `uninstall` | 卸载 GPU 驱动和相关程序。                                                  |
| `status`    | 查看 GPU 驱动和 CUDA 版本信息，以及 GPU 相关服务的运行状态。                           |

## 选项

| 选项         | 简写   | 用途                                                                                                                                                         | 是否必需 | 默认值         |
|--------------|------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|----------------|
| `--base-dir` | `-b` | 指定安装的基础目录。                                                                                                                                       | 否       | `$HOME/.olares`  |
| `--version`  | `-v` | 安装指定 Olares 版本的 GPU 驱动及组件。版本号格式为 `x.y.z`（如 `1.10.0`）或包含构建日期（如 `1.10.0-20241109`）。<br>可用版本请参考 [GitHub Releases](https://github.com/beclab/Olares/releases)。 | 否       | 当前已安装版本 |
| `--help`     | `-h` | 显示命令帮助信息。                                                                                                                                         | 否       | 无             |

## 使用示例
```bash
# 安装指定 Olares 版本的 GPU 驱动及依赖至指定目录
olares-cli gpu install --base-dir /home/olares/.olares --version 1.11.1-rc.4

# 启用 GPU 功能
olares-cli gpu enable

# 查看 GPU 驱动和服务状态
olares-cli gpu status

# 禁用 GPU 功能
olares-cli gpu disable

# 卸载 GPU 驱动及组件
olares-cli gpu uninstall

```


