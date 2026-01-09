# `delete`

## 命令说明

`delete`子命令用于从 Olares 系统中永久删除一个已有的用户账号。

**别名**：`delete`、`d`、`del`、`rm`、`remove`

```bash
olares-cli user delete <用户名> [选项]
```

## 参数

| 参数 | 说明 | 是否必需|
|--|--|--|
| `<用户名>` | 指定要删除的用户名。通常为 Olares ID 中`@`符号之前的部分。<br>例如 `alice123@olares.com`中的`alice123`。| 是 |

## 选项
| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--kubeconfig` | | 指定 kubeconfig 文件路径。 | 否 | 无 |

## 使用示例

```bash
# 删除名为 alice123 的用户
sudo olares-cli user delete alice123
```