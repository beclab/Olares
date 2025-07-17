# `download`

## 命令说明
`download` 子命令用于下载 Restic 工具。Restic 是执行备份和恢复操作的必需工具，该命令可确保系统上已安装此工具。

```bash
olares-cli backups download [选项]
```
## 选项

| 选项                 | 简写   | 用途                        | 是否必需 | 默认值     |
|----------------------|------|-----------------------------|----------|------------|
| `--download-cdn-url` |      | 下载 Restic 工具的 CDN 地址。 | 否       | 系统默认 URL |
| `--help`             | `-h` | 显示命令帮助信息。                | 否       | 无         |

## 使用示例
```bash
# 通过自定义 CDN 地址下载 Restic
olares-cli backups download --download-cdn-url https://custom-cdn.example.com/restic
```