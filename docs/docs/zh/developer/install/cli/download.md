# `download`

## 命令说明

`download` 命令用于下载在本地机器上安装 Olares 所需的软件包和组件。它支持下载组件、检查安装包状态以及获取配置清单文件。
```bash
olares-cli download <子命令> [选项]
```

## 子命令

| 名称          | 简写  | 用途               | 示例                                     |
|-------------|-----|------------------|----------------------------------------|
| `check`     |     | 检查 Olares 安装包的状态 | `olares-cli download check`     |
| `component` | `c` | 下载 Olares 组件     | `olares-cli download component` |
| `wizard`    | `w` | 下载配置清单文件         | `olares-cli download wizard`    |

## 选项
| 选项                 | 简写   | 用途                                                                                                                                                 | 是否必需 | 默认值       |
|----------------------|------|------------------------------------------------------------------------------------------------------------------------------------------------------|----------|--------------|
| `--base-dir`         | `-b` | 设置 Olares 的基础目录。                                                                                                                               | 否       | `$HOME/.olares`|
| `--download-cdn-url` |      | 设置 CDN 加速下载地址，格式为 `https://example.cdn.com`。                                                                                                 | 否       | 默认地址     |
| `--help`             | `-h` | 显示帮助信息。                                                                                                                                         | 否       | 无           |
| `--kube`             |      | 指定 Kubernetes 类型。<br>支持 `k3s` 和 `k8s`。                                                                                                         | 否       | `k3s`        |
| `--version`          | `-v` | 指定 Olares 版本。<br>版本号格式为 `x.y.z`（如 `1.10.0`）或包含构建日期（如 `1.10.0-20241109`）。<br> 可用版本请参考 [GitHub Releases](https://github.com/beclab/Olares/releases)。 | 否       | 当前安装版本 |

## 使用示例
```bash
# 指定存储所有下载组件的基础目录
olares-cli download component -b /custom/path

# 指定用于下载 Olares 组件的 CDN 加速地址
olares-cli download component --download-cdn-url https://my.cdn.com

# 指定安装的 Kubernetes 类型
olares-cli download component --kube k8s

# 设置配置清单文件的路径
olares-cli download component --manifest /custom/path/manifest.json

# 指定要下载的 Olares 软件包和组件版本
olares-cli download component -v 1.11.0
```


