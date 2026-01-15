# `upgrade`

## 命令说明

`upgrade`命令提供一组用于升级 Olares，检查升级准备情况与兼容性的工具。

```bash
olares-cli upgrade <子命令> [选项]
```

## 子命令

| 子命令 | 描述 | 
|--|--|
| `precheck` | 对 Olares 执行升级前检查。 |
| `spec` | 获取当前 CLI 版本对应的升级规格。 |
| `viable` | 判断是否可以基于指定基础版本直接执行升级。 |

## 全局选项

以下选项适用于`upgrade`主命令，需要时可被子命令继承。

| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--base-dir` | `-b` | 设置 Olares 安装包的基本目录。 | 否 | `$HOME/.olares` |
| `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--version` | `-v` | 设置要升级到的目标 Olares 版本，例如`1.10.0`、`1.10.0-20241109`。 | 否 | 无 |

## `viable` 专属选项

| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--base` | `-b` | 要检查的基础版本。 | 否 | 当前 Olares 系统版本 |

:::warning 选项冲突
主命令使用`-b`作为`--base-dir`的简写。但在执行`upgrade viable`时，`-b`指代的是`--base`。
:::

## 使用示例

```bash
# 检查当前系统是否可以直接升级
sudo olares-cli upgrade viable

# 检查从指定基础版本开始是否可直接升级
sudo olares-cli upgrade viable --base 1.9.0

# 执行升级前检查
sudo olares-cli upgrade precheck

# 查看当前 CLI 对应的升级规格
sudo olares-cli upgrade spec
```