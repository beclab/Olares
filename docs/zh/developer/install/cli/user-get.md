# `get`

## 命令说明

:::info 执行权限
该命令通常需要使用管理员权限（`sudo`）执行。
:::

`get`子命令用于获取 Olares 指定用户的详细信息。输出结果以表格或 JSON 格式显示。

```bash
olares-cli user get <用户名> [选项]
```

## 参数

| 参数 | 说明 | 是否必需|
|--|--|--|
| `<用户名>` | 指定要查询的用户名。通常为 Olares ID 中`@`符号之前的部分。<br>例如 `alice123@olares.com`中的`alice123`。| **是** |

## 选项
| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--kubeconfig` | | 指定 kubeconfig 文件路径。 | 否 | 无 |
| `--no-headers` | | 输出结果不显示表头。 | 否 | 无 |
| `--output` | `-o` | 指定输出格式。<br>可选值：`table`、`json`。 | 否 | `table` |

## 使用示例

```bash
# 以默认表格格式查看用户 alice 的信息
sudo olares-cli user get alice

# 以 JSON 格式查看用户 bob 的信息
sudo olares-cli user get bob -o json
```