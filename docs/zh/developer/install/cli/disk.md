# `disk`

## 命令说明

`disk`命令提供了一组用于管理 Olares 系统存储资源的工具，主要用于基于 LVM 的存储配置管理。

```bash
olares-cli disk <subcommand>
```

## 子命令

| 子命令 | 描述 | 
|--|--|
| `extend` | 在基于 LVM 的安装环境中扩展 Olares 的存储容量。 |
| `list-unmounted` | 列出未挂载的磁盘。 |

## 参数标记

| 名称 | 简写 | 说明 | 
|--|--|--|
| `--help` | `-h` | 显示帮助信息。 |

## 使用示例

```bash
# 列出未挂载的磁盘
olares-cli disk list-unmounted

# 添加新检测到的未挂载磁盘来扩展 Olares 存储
olares-cli disk extend
```