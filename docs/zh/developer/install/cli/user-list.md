# `list`

## 命令说明

`list`子命令用于显示 Olares 系统中所有注册用户列表。你可以根据需要筛选用户并排序，以表格或 JSON 格式显示结果。

**别名**：`list`、`ls`、`l`

```bash
olares-cli user list [选项]
```

## 选项
| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--kubeconfig` | | 指定 kubeconfig 文件路径。 | 否 | 无 |
| `--no-headers` | | 输出结果不显示表头。 | 否 | 无 |
| `--output` | `-o`| 指定输出格式。<br>可选值：`table`、`json`。| 否 | `table` |
| `--reverse` | `-r` | 倒序排列输出结果。 | 否 | 无 |
| `--sort` | | 按指定字段对结果进行排序。<br>可选值：`name`、`role`、`create-time`、`memory`、`cpu`。 | 否 | 无 |

## 使用示例

```bash
# 以表格格式列出所有用户
sudo olares-cli user list

# 以 JSON 格式列出所有用户
sudo olares-cli user list -o json

# 按角色排序列出用户
sudo olares-cli user list --sort role
```